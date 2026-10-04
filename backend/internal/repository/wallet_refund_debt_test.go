package repository

import (
	"context"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/AsukaCC/EasySub2api/ent"
	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func refundDebtRepo(t *testing.T) (*userRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	t.Cleanup(func() { _ = client.Close() })
	return newUserRepositoryWithSQL(client, db), mock
}

func debtWalletRow(recharge, bonus, frozenRecharge, frozenBonus float64) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"recharge_balance", "bonus_balance", "frozen_recharge_balance", "frozen_bonus_balance"}).AddRow(recharge, bonus, frozenRecharge, frozenBonus)
}

func TestRefundDebtReserveAndResolve(t *testing.T) {
	for _, outcome := range []string{"release", "capture"} {
		for _, source := range []string{"missing", "consumed", "partial"} {
			t.Run(source+"/"+outcome, func(t *testing.T) {
				repo, mock := refundDebtRepo(t)
				ctx := context.Background()
				available := 0.0
				if source == "partial" {
					available = 5
				}
				mock.ExpectBegin()
				mock.ExpectQuery("(?s)SELECT id, status, amount.*FROM wallet_holds").WithArgs("refund").WillReturnRows(sqlmock.NewRows([]string{"id", "status", "amount", "bonus_amount", "recharge_amount", "request_fingerprint"}))
				mock.ExpectQuery("(?s)SELECT recharge_balance.*FOR UPDATE").WithArgs("user").WillReturnRows(debtWalletRow(100, available, 0, 0))
				grants := sqlmock.NewRows([]string{"remaining_amount"})
				if source != "missing" {
					grants.AddRow(available)
				}
				mock.ExpectQuery("(?s)SELECT remaining_amount.*FROM wallet_bonus_grants").WithArgs("grant", "user").WillReturnRows(grants)
				mock.ExpectExec("INSERT INTO wallet_holds").WithArgs(sqlmock.AnyArg(), "user", "refund", 120.0, 20.0, 100.0, "fingerprint").WillReturnResult(sqlmock.NewResult(0, 1))
				if available > 0 {
					mock.ExpectExec("UPDATE wallet_bonus_grants").WithArgs(available, "grant").WillReturnResult(sqlmock.NewResult(0, 1))
					mock.ExpectExec("INSERT INTO wallet_hold_allocations").WithArgs(sqlmock.AnyArg(), "grant", available).WillReturnResult(sqlmock.NewResult(0, 1))
				}
				mock.ExpectExec("UPDATE users SET").WithArgs(100.0, 20.0, "user").WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec("INSERT INTO wallet_transactions").WithArgs("user", "hold", -120.0, -20.0, -100.0, 120.0, 100.0, 0.0, available, available-20, "payment_refund", "refund", "wallet-hold:payment_refund:refund", "").WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectQuery("(?s)SELECT recharge_balance.*FROM users WHERE").WithArgs("user").WillReturnRows(debtWalletRow(0, available-20, 100, 20))
				mock.ExpectCommit()
				hold, err := repo.HoldRefundPoints(ctx, service.RefundPointHoldInput{UserID: "user", RefundID: "refund", BonusGrantID: "grant", BasePoints: 100, BonusPoints: 20, RequestFingerprint: "fingerprint"})
				require.NoError(t, err)
				require.Equal(t, available-20, hold.Summary.BonusBalance)
				require.Zero(t, hold.Summary.AvailableBalance)
				mock.ExpectBegin()
				mock.ExpectQuery("(?s)SELECT id, user_id, purpose.*FROM wallet_holds").WithArgs(hold.HoldID).WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "purpose", "reference_id", "status", "amount", "bonus_amount", "recharge_amount", "refunded_amount", "refunded_bonus_amount", "refunded_recharge_amount"}).AddRow(hold.HoldID, "user", "payment_refund", "refund", "held", 120, 20, 100, 0, 0, 0))
				mock.ExpectQuery("(?s)SELECT recharge_balance.*FOR UPDATE").WithArgs("user").WillReturnRows(debtWalletRow(0, available-20, 100, 20))
				if outcome == "release" {
					allocations := sqlmock.NewRows([]string{"bonus_grant_id", "amount", "restore"})
					if available > 0 {
						allocations.AddRow("grant", available, true)
					}
					mock.ExpectQuery("(?s)SELECT a.bonus_grant_id.*FROM wallet_hold_allocations").WithArgs(hold.HoldID).WillReturnRows(allocations)
					if available > 0 {
						mock.ExpectExec("UPDATE wallet_bonus_grants").WithArgs(available, available, "grant").WillReturnResult(sqlmock.NewResult(0, 1))
					}
					mock.ExpectExec("UPDATE users SET").WithArgs(100.0, 20.0, 100.0, 20.0, "user").WillReturnResult(sqlmock.NewResult(0, 1))
					mock.ExpectExec("UPDATE wallet_holds").WithArgs(hold.HoldID).WillReturnResult(sqlmock.NewResult(0, 1))
					mock.ExpectExec("INSERT INTO wallet_transactions").WithArgs("user", "release", 120.0, 20.0, 100.0, -120.0, 0.0, 100.0, available-20, available, "payment_refund", "refund", "resolve", "release wallet hold").WillReturnResult(sqlmock.NewResult(0, 1))
					mock.ExpectQuery("(?s)SELECT recharge_balance.*FROM users WHERE").WithArgs("user").WillReturnRows(debtWalletRow(100, available, 0, 0))
				} else {
					mock.ExpectExec("UPDATE wallet_bonus_grants g").WithArgs(hold.HoldID, "payment_refund").WillReturnResult(sqlmock.NewResult(0, 1))
					mock.ExpectExec("UPDATE users SET").WithArgs(100.0, 20.0, "user").WillReturnResult(sqlmock.NewResult(0, 1))
					mock.ExpectExec("UPDATE wallet_holds").WithArgs(hold.HoldID).WillReturnResult(sqlmock.NewResult(0, 1))
					mock.ExpectExec("INSERT INTO wallet_transactions").WithArgs("user", "capture", -120.0, -20.0, -100.0, -120.0, 0.0, 0.0, available-20, available-20, "payment_refund", "refund", "resolve", "capture wallet hold").WillReturnResult(sqlmock.NewResult(0, 1))
					mock.ExpectQuery("(?s)SELECT recharge_balance.*FROM users WHERE").WithArgs("user").WillReturnRows(debtWalletRow(0, available-20, 0, 0))
				}
				mock.ExpectCommit()
				var resolved service.WalletHoldResult
				if outcome == "release" {
					resolved, err = repo.ReleaseWalletHold(ctx, hold.HoldID, "resolve")
				} else {
					resolved, err = repo.CaptureWalletHold(ctx, hold.HoldID, "resolve")
				}
				require.NoError(t, err)
				require.True(t, resolved.Applied)
				require.Zero(t, resolved.Summary.FrozenBonus)
				mock.ExpectBegin()
				mock.ExpectQuery("(?s)SELECT id, user_id, purpose.*FROM wallet_holds").WithArgs(hold.HoldID).WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "purpose", "reference_id", "status", "amount", "bonus_amount", "recharge_amount", "refunded_amount", "refunded_bonus_amount", "refunded_recharge_amount"}).AddRow(hold.HoldID, "user", "payment_refund", "refund", resolved.Status, 120, 20, 100, 0, 0, 0))
				mock.ExpectQuery("(?s)SELECT recharge_balance.*FROM users WHERE").WithArgs("user").WillReturnRows(debtWalletRow(resolved.Summary.RechargeBalance, resolved.Summary.BonusBalance, 0, 0))
				mock.ExpectCommit()
				if outcome == "release" {
					resolved, err = repo.ReleaseWalletHold(ctx, hold.HoldID, "resolve")
				} else {
					resolved, err = repo.CaptureWalletHold(ctx, hold.HoldID, "resolve")
				}
				require.NoError(t, err)
				require.False(t, resolved.Applied)
				require.NoError(t, mock.ExpectationsWereMet())
			})
		}
	}
}

func TestPermanentBonusCreditPreservesRechargeAndRepaysBonusDebt(t *testing.T) {
	repo, mock := refundDebtRepo(t)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT 1 FROM wallet_transactions").WithArgs("credit").WillReturnRows(sqlmock.NewRows([]string{"exists"}))
	mock.ExpectQuery("(?s)SELECT recharge_balance.*FOR UPDATE").WithArgs("user").WillReturnRows(debtWalletRow(100, -20, 0, 0))
	mock.ExpectExec("UPDATE users").WithArgs(0.0, 10.0, 0.0, "user").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO wallet_bonus_grants").WithArgs(sqlmock.AnyArg(), "user", 10.0, 10.0, 0.0, nil, "", "", "active").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO wallet_transactions").WithArgs("user", "credit", 10.0, 10.0, 0.0, 0.0, 100.0, 100.0, -20.0, -10.0, "", "", "credit", "").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("(?s)SELECT recharge_balance.*FROM users WHERE").WithArgs("user").WillReturnRows(debtWalletRow(100, -10, 0, 0))
	mock.ExpectCommit()
	result, err := repo.CreditWallet(context.Background(), service.WalletCreditInput{UserID: "user", Kind: service.WalletKindBonus, Amount: 10, IdempotencyKey: "credit"})
	require.NoError(t, err)
	require.Equal(t, 90.0, result.Summary.AvailableBalance)
	require.Equal(t, -10.0, result.Summary.BonusBalance)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUnifiedRefundCanUseBonusSourceAndCreateNetDebt(t *testing.T) {
	repo, mock := refundDebtRepo(t)
	mock.ExpectBegin()
	mock.ExpectQuery("(?s)SELECT id, status, amount.*FROM wallet_holds").WithArgs("refund").WillReturnRows(sqlmock.NewRows([]string{"id", "status", "amount", "bonus_amount", "recharge_amount", "request_fingerprint"}))
	mock.ExpectQuery("(?s)SELECT recharge_balance.*FOR UPDATE").WithArgs("user").WillReturnRows(debtWalletRow(0, 100, 0, 0))
	mock.ExpectExec("INSERT INTO wallet_holds").WithArgs(sqlmock.AnyArg(), "user", "refund", 110.0, 10.0, 100.0, "").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE users SET").WithArgs(100.0, 10.0, "user").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO wallet_transactions").WithArgs("user", "hold", -110.0, -10.0, -100.0, 110.0, 0.0, -100.0, 100.0, 90.0, "payment_refund", "refund", "wallet-hold:payment_refund:refund", "").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("(?s)SELECT recharge_balance.*FROM users WHERE").WithArgs("user").WillReturnRows(debtWalletRow(-100, 90, 100, 10))
	mock.ExpectCommit()
	hold, err := repo.HoldRefundPoints(context.Background(), service.RefundPointHoldInput{UserID: "user", RefundID: "refund", BasePoints: 100, BonusPoints: 10})
	require.NoError(t, err)
	require.Equal(t, -10.0, hold.Summary.Balance)
	require.Equal(t, 10.0, hold.Summary.OverdraftAmount)
	require.Zero(t, hold.Summary.AvailableBalance)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUnifiedRefundPrincipalCannotSpendExistingDebt(t *testing.T) {
	repo, mock := refundDebtRepo(t)
	mock.ExpectBegin()
	mock.ExpectQuery("(?s)SELECT id, status, amount.*FROM wallet_holds").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery("(?s)SELECT recharge_balance.*FOR UPDATE").WithArgs("user").WillReturnRows(debtWalletRow(100, -20, 0, 0))
	mock.ExpectRollback()
	_, err := repo.HoldRefundPoints(context.Background(), service.RefundPointHoldInput{UserID: "user", RefundID: "refund", BasePoints: 100, BonusPoints: 10})
	require.ErrorIs(t, err, service.ErrInsufficientBalance)
	require.NoError(t, mock.ExpectationsWereMet())
}
