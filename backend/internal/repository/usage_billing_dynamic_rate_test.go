package repository

import (
	"context"
	"math"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func expectDynamicRechargeWallet(mock sqlmock.Sqlmock, userID string, balance, bonus float64) {
	mock.ExpectQuery("(?s)SELECT recharge_balance, bonus_balance, frozen_recharge_balance, frozen_bonus_balance.*FOR UPDATE").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"recharge_balance", "bonus_balance", "frozen_recharge_balance", "frozen_bonus_balance"}).AddRow(balance, bonus, 0, 0))
	mock.ExpectQuery("(?s)SELECT id, remaining_amount.*expires_at <= NOW.*FOR UPDATE").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "remaining_amount"}))
}

func TestDynamicRateBillingRechargeOnly(t *testing.T) {
	for _, tc := range []struct {
		name                                      string
		recharge, bonus, limit, legacyLimit, used float64
		subscription                              bool
		wantCost, wantRecharge, wantQuota         float64
	}{
		{name: "discounted amount uses recharge bucket", recharge: 10, bonus: 10, wantCost: 1, wantRecharge: 1, wantQuota: 20},
		{name: "recharge exhausted then normal bonus", recharge: .4, bonus: 10, wantCost: 1.6, wantRecharge: .4, wantQuota: 8},
		{name: "bonus only", bonus: 10, wantCost: 2},
		{name: "no funded balance", wantCost: 2},
		{name: "debt gets no discount", recharge: -.5, wantCost: 2},
		{name: "partial personal quota", recharge: 10, bonus: 10, limit: 10, used: 5, wantCost: 1.75, wantRecharge: .25, wantQuota: 5},
		{name: "exhausted personal quota", recharge: 10, bonus: 10, limit: 10, used: 10, wantCost: 2},
		{name: "legacy quota is per user", recharge: 10, legacyLimit: 8, used: 4, wantCost: 1.8, wantRecharge: .2, wantQuota: 4},
		{name: "subscription is not recharge points", subscription: true, wantCost: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()
			mock.ExpectBegin()
			tx, err := db.BeginTx(context.Background(), nil)
			require.NoError(t, err)
			if !tc.subscription {
				expectDynamicRechargeWallet(mock, "user", tc.recharge, tc.bonus)
			}
			if tc.recharge > 0 && !tc.subscription {
				mock.ExpectExec("INSERT INTO user_dynamic_rate_usage").WithArgs("user", "group", "discount", "window").WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectQuery("(?s)SELECT used_amount.*FROM user_dynamic_rate_usage.*FOR UPDATE").
					WithArgs("user", "group", "discount", "window").WillReturnRows(sqlmock.NewRows([]string{"used_amount"}).AddRow(tc.used))
				if tc.wantQuota > 0 {
					limit := tc.limit
					if limit == 0 {
						limit = tc.legacyLimit
					}
					mock.ExpectExec("UPDATE user_dynamic_rate_usage").WithArgs(limit, tc.wantQuota, "user", "group", "discount", "window").
						WillReturnResult(sqlmock.NewResult(0, 1))
				}
			}
			mock.ExpectCommit()
			cmd := &service.UsageBillingCommand{
				UserID: "user", BalanceCost: 1, APIKeyQuotaCost: 1, APIKeyRateLimitCost: 1, AccountQuotaCost: 20,
				DynamicRatePlan: &service.UsageDynamicRatePlan{GroupID: "group", StandardCost: 10, AccountCost: 20, FallbackMultiplier: .2,
					Rules: []service.UsageDynamicRateRule{
						{RuleID: "higher-price", QuotaKey: "window", DiscountCoefficient: .8},
						{RuleID: "discount", QuotaKey: "window", DiscountCoefficient: .5, PersonalQuotaAmount: tc.limit, SharedQuotaAmount: tc.legacyLimit},
					},
				},
			}
			if tc.subscription {
				cmd.BalanceCost = 0
				subscriptionID := "subscription"
				cmd.SubscriptionID = &subscriptionID
				cmd.SubscriptionCost = 1
			}
			result := &service.UsageBillingApplyResult{}
			require.NoError(t, applyDynamicRateBilling(context.Background(), tx, cmd, result))
			require.NotNil(t, result.FinalActualCost)
			require.InDelta(t, tc.wantCost, *result.FinalActualCost, 1e-8)
			require.InDelta(t, tc.wantRecharge, result.RechargeOnlyCost, 1e-8)
			require.Equal(t, service.QuantizeRateMultiplier(tc.wantCost/10), *result.FinalRateMultiplier)
			require.Equal(t, *result.FinalActualCost, cmd.APIKeyQuotaCost)
			require.Equal(t, *result.FinalActualCost, cmd.APIKeyRateLimitCost)
			require.Equal(t, float64(20), cmd.AccountQuotaCost)
			if tc.subscription {
				require.Equal(t, *result.FinalActualCost, cmd.SubscriptionCost)
			} else {
				require.Equal(t, *result.FinalActualCost, cmd.BalanceCost)
			}
			require.NoError(t, tx.Commit())
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDynamicRateBillingRejectsNonFinitePlan(t *testing.T) {
	for _, value := range []float64{math.NaN(), math.Inf(1), -1} {
		cmd := &service.UsageBillingCommand{UserID: "user", BalanceCost: 1,
			DynamicRatePlan: &service.UsageDynamicRatePlan{GroupID: "group", StandardCost: 1, FallbackMultiplier: value}}
		require.Error(t, applyDynamicRateBilling(context.Background(), nil, cmd, &service.UsageBillingApplyResult{}))
	}
}

func TestDynamicRateBillingPreservesSmallRechargeBoundary(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	mock.ExpectBegin()
	tx, err := db.BeginTx(context.Background(), nil)
	require.NoError(t, err)
	expectDynamicRechargeWallet(mock, "user", 1.00000001, 1)
	mock.ExpectExec("INSERT INTO user_dynamic_rate_usage").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("(?s)SELECT used_amount.*FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"used_amount"}).AddRow(0))
	mock.ExpectExec("UPDATE user_dynamic_rate_usage").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	cmd := &service.UsageBillingCommand{UserID: "user", BalanceCost: .1,
		DynamicRatePlan: &service.UsageDynamicRatePlan{GroupID: "group", StandardCost: 1, AccountCost: 1, FallbackMultiplier: .2,
			Rules: []service.UsageDynamicRateRule{{RuleID: "discount", QuotaKey: "window", DiscountCoefficient: .5}}}}
	result := &service.UsageBillingApplyResult{}
	require.NoError(t, applyDynamicRateBilling(context.Background(), tx, cmd, result))
	require.Equal(t, 0.00000001, result.RechargeOnlyCost)
	require.Equal(t, .19999999, *result.FinalActualCost)
	require.NoError(t, tx.Commit())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDynamicRateBillingRegularCostSurvivesRoundedDiscountEstimate(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	mock.ExpectBegin()
	tx, err := db.Begin()
	require.NoError(t, err)
	expectDynamicRechargeWallet(mock, "user", 1, 1)
	cmd := rechargeBillingCommand()
	cmd.BalanceCost, cmd.APIKeyQuotaCost, cmd.APIKeyRateLimitCost = 1e-9, 1e-9, 1e-9
	cmd.DynamicRatePlan.StandardCost = 1e-8
	cmd.DynamicRatePlan.FallbackMultiplier = 1
	cmd.Normalize()
	require.Zero(t, cmd.BalanceCost)
	result := &service.UsageBillingApplyResult{}
	require.NoError(t, applyDynamicRateBilling(context.Background(), tx, cmd, result))
	require.Equal(t, 1e-8, *result.FinalActualCost)
	require.Equal(t, 1e-8, cmd.BalanceCost)
	require.Equal(t, cmd.BalanceCost, cmd.APIKeyQuotaCost)
	require.Equal(t, cmd.BalanceCost, cmd.APIKeyRateLimitCost)
	mock.ExpectCommit()
	require.NoError(t, tx.Commit())
	require.NoError(t, mock.ExpectationsWereMet())
}
