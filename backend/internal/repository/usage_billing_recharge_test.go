package repository

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func rechargeBillingCommand() *service.UsageBillingCommand {
	return &service.UsageBillingCommand{
		RequestID: "request", APIKeyID: "key", UserID: "user", BalanceCost: 1,
		DynamicRatePlan: &service.UsageDynamicRatePlan{
			GroupID: "group", StandardCost: 10, AccountCost: 20, FallbackMultiplier: .2,
			Rules: []service.UsageDynamicRateRule{{RuleID: "offer", QuotaKey: "window", DiscountCoefficient: .5}},
		},
	}
}

func expectRechargeBillingClaim(mock sqlmock.Sqlmock) {
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO usage_billing_dedup").
		WithArgs("request", "key", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("dedup"))
	mock.ExpectQuery("SELECT request_fingerprint FROM usage_billing_dedup_archive").
		WithArgs("request", "key").WillReturnError(sql.ErrNoRows)
}

func expectRechargeBillingQuota(mock sqlmock.Sqlmock, quota float64) {
	mock.ExpectExec("INSERT INTO user_dynamic_rate_usage").
		WithArgs("user", "group", "offer", "window").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("(?s)SELECT used_amount.*FROM user_dynamic_rate_usage.*FOR UPDATE").
		WithArgs("user", "group", "offer", "window").
		WillReturnRows(sqlmock.NewRows([]string{"used_amount"}).AddRow(0))
	mock.ExpectExec("UPDATE user_dynamic_rate_usage").
		WithArgs(0.0, quota, "user", "group", "offer", "window").WillReturnResult(sqlmock.NewResult(0, 1))
}

func expectRechargeBillingDebit(mock sqlmock.Sqlmock, balance, bonus, cost, bonusUsed float64, failLedger bool) {
	mock.ExpectQuery("SELECT 1 FROM wallet_transactions").
		WithArgs("wallet-usage:key:request").WillReturnRows(sqlmock.NewRows([]string{"exists"}))
	expectDynamicRechargeWallet(mock, "user", balance, bonus)
	mock.ExpectQuery("(?s)SELECT id, remaining_amount.*expires_at > NOW.*FOR UPDATE").
		WithArgs("user").WillReturnRows(sqlmock.NewRows([]string{"id", "remaining_amount"}).AddRow("grant", bonus))
	if bonusUsed > 0 {
		mock.ExpectExec("UPDATE wallet_bonus_grants").WithArgs(bonusUsed, "grant").WillReturnResult(sqlmock.NewResult(0, 1))
	}
	rechargeUsed := walletMoney(cost - bonusUsed)
	afterBalance, afterBonus := walletMoney(balance-rechargeUsed), walletMoney(bonus-bonusUsed)
	mock.ExpectExec("UPDATE users SET recharge_balance = recharge_balance -").
		WithArgs(rechargeUsed, bonusUsed, "user").WillReturnResult(sqlmock.NewResult(0, 1))
	ledger := mock.ExpectExec("INSERT INTO wallet_transactions").
		WithArgs("user", "debit", -cost, -bonusUsed, -walletMoney(cost-bonusUsed), 0.0,
			balance, afterBalance, bonus, afterBonus, "api_usage", "request", "wallet-usage:key:request", "")
	if failLedger {
		ledger.WillReturnError(errors.New("ledger unavailable"))
		return
	}
	ledger.WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("(?s)SELECT recharge_balance, bonus_balance, frozen_recharge_balance, frozen_bonus_balance.*FROM users WHERE").
		WithArgs("user").WillReturnRows(sqlmock.NewRows([]string{"recharge_balance", "bonus_balance", "frozen_recharge_balance", "frozen_bonus_balance"}).AddRow(afterBalance, afterBonus, 0, 0))
	mock.ExpectQuery("SELECT expires_at, SUM").WithArgs("user").
		WillReturnRows(sqlmock.NewRows([]string{"expires_at", "remaining_amount"}))
}

func TestUsageBillingRechargeSettlementAndDedup(t *testing.T) {
	for _, tc := range []struct {
		name                                                        string
		recharge, bonus, cost, bonusUsed, discountedRecharge, quota float64
	}{
		{"discounted amount uses recharge bucket", 10, 10, 1, 0, 1, 20},
		{"mixed payment", .4, 10, 1.6, .4, 1.2, 8},
		{"bonus only pays regular rate", 0, 10, 2, 2, 0, 0},
		{"recharge covers remainder without overdraft", 2, .2, 1.6, .2, .4, 8},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()
			repo := &usageBillingRepository{db: db}
			balance := walletMoney(tc.recharge)
			expectRechargeBillingClaim(mock)
			expectDynamicRechargeWallet(mock, "user", balance, tc.bonus)
			if tc.quota > 0 {
				expectRechargeBillingQuota(mock, tc.quota)
			}
			expectRechargeBillingDebit(mock, balance, tc.bonus, tc.cost, tc.bonusUsed, false)
			mock.ExpectCommit()
			result, err := repo.Apply(context.Background(), rechargeBillingCommand())
			require.NoError(t, err)
			require.True(t, result.Applied)
			require.Equal(t, tc.cost, *result.FinalActualCost)
			require.Equal(t, tc.discountedRecharge, result.RechargeOnlyCost)
			require.Equal(t, walletMoney(math.Max(balance-(tc.cost-tc.bonusUsed), 0)+tc.bonus-tc.bonusUsed), *result.NewBalance)
			require.False(t, result.BalanceOverdrafted)

			// A duplicate claims no quota and performs no wallet mutation.
			cmd := rechargeBillingCommand()
			cmd.Normalize()
			mock.ExpectBegin()
			mock.ExpectQuery("INSERT INTO usage_billing_dedup").WillReturnError(sql.ErrNoRows)
			mock.ExpectQuery("SELECT request_fingerprint FROM usage_billing_dedup").
				WithArgs("request", "key").WillReturnRows(sqlmock.NewRows([]string{"request_fingerprint"}).AddRow(cmd.RequestFingerprint))
			mock.ExpectRollback()
			result, err = repo.Apply(context.Background(), cmd)
			require.NoError(t, err)
			require.False(t, result.Applied)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUsageBillingRechargeRollbackAndRetry(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := &usageBillingRepository{db: db}
	for _, fail := range []bool{true, false} {
		expectRechargeBillingClaim(mock)
		expectDynamicRechargeWallet(mock, "user", 10.4, 10)
		expectRechargeBillingQuota(mock, 8)
		expectRechargeBillingDebit(mock, 10.4, 10, 1.6, 1.2, fail)
		if fail {
			mock.ExpectRollback()
		} else {
			mock.ExpectCommit()
		}
		result, err := repo.Apply(context.Background(), rechargeBillingCommand())
		if fail {
			require.ErrorContains(t, err, "ledger unavailable")
			require.Nil(t, result)
		} else {
			require.NoError(t, err)
			require.True(t, result.Applied)
		}
	}
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestWalletRechargeReservationCannotUseBonusOrDebt(t *testing.T) {
	for _, balance := range []float64{0, -1} {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		mock.ExpectQuery("SELECT 1 FROM wallet_transactions").WillReturnRows(sqlmock.NewRows([]string{"exists"}))
		expectDynamicRechargeWallet(mock, "user", balance, 10)
		_, err = debitWalletTx(context.Background(), db, service.WalletDebitInput{
			UserID: "user", Amount: 1, RechargeOnlyAmount: 1, AllowOverdraft: true,
		}, "reservation")
		require.ErrorIs(t, err, service.ErrInsufficientBalance)
		mock.ExpectClose()
		require.NoError(t, db.Close())
		require.NoError(t, mock.ExpectationsWereMet())
	}
}

func TestDynamicRateBillingExcludesFrozenRecharge(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	mock.ExpectBegin()
	tx, err := db.Begin()
	require.NoError(t, err)
	mock.ExpectQuery("(?s)SELECT recharge_balance, bonus_balance, frozen_recharge_balance, frozen_bonus_balance.*FOR UPDATE").
		WithArgs("user").WillReturnRows(sqlmock.NewRows([]string{"recharge_balance", "bonus_balance", "frozen_recharge_balance", "frozen_bonus_balance"}).AddRow(0, 10, 100, 0))
	mock.ExpectQuery("(?s)SELECT id, remaining_amount.*expires_at <= NOW.*FOR UPDATE").
		WithArgs("user").WillReturnRows(sqlmock.NewRows([]string{"id", "remaining_amount"}))
	result := &service.UsageBillingApplyResult{}
	require.NoError(t, applyDynamicRateBilling(context.Background(), tx, rechargeBillingCommand(), result))
	require.Equal(t, 2.0, *result.FinalActualCost)
	require.Zero(t, result.RechargeOnlyCost)
	mock.ExpectCommit()
	require.NoError(t, tx.Commit())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDynamicRateBillingExpiresBonusBeforePricing(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	mock.ExpectBegin()
	tx, err := db.Begin()
	require.NoError(t, err)
	mock.ExpectQuery("(?s)SELECT recharge_balance, bonus_balance, frozen_recharge_balance, frozen_bonus_balance.*FOR UPDATE").
		WithArgs("user").WillReturnRows(sqlmock.NewRows([]string{"recharge_balance", "bonus_balance", "frozen_recharge_balance", "frozen_bonus_balance"}).AddRow(10.4, 10, 0, 0))
	mock.ExpectQuery("(?s)SELECT id, remaining_amount.*expires_at <= NOW.*FOR UPDATE").
		WithArgs("user").WillReturnRows(sqlmock.NewRows([]string{"id", "remaining_amount"}).AddRow("expired", 3))
	mock.ExpectExec("UPDATE wallet_bonus_grants").WithArgs("expired", 3.0).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE users").WithArgs(3.0, "user").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO wallet_transactions").
		WithArgs("user", "expire", -3.0, -3.0, 0.0, 0.0, 10.4, 7.4, 10.0, 7.0,
			"bonus_expiry", "", sqlmock.AnyArg(), "expired bonus balance").WillReturnResult(sqlmock.NewResult(0, 1))
	expectRechargeBillingQuota(mock, 8)
	result := &service.UsageBillingApplyResult{}
	require.NoError(t, applyDynamicRateBilling(context.Background(), tx, rechargeBillingCommand(), result))
	require.Equal(t, 1.6, *result.FinalActualCost)
	require.Equal(t, .4, result.RechargeOnlyCost)
	mock.ExpectCommit()
	require.NoError(t, tx.Commit())
	require.NoError(t, mock.ExpectationsWereMet())
}
