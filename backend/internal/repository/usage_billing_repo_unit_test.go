//go:build unit

package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"github.com/AsukaCC/EasySub2api/internal/service"
)

const (
	walletTransactionExistsSQL = `SELECT 1 FROM wallet_transactions WHERE idempotency_key = $1`
	lockWalletUserSQL          = `(?s)SELECT balance, bonus_balance, frozen_balance, frozen_bonus_balance.*FROM users.*WHERE id = $1 AND deleted_at IS NULL.*FOR UPDATE`
	expiredWalletBonusSQL      = `(?s)SELECT id, remaining_amount.*FROM wallet_bonus_grants.*expires_at <= NOW().*FOR UPDATE`
	availableWalletBonusSQL    = `(?s)SELECT id, remaining_amount.*FROM wallet_bonus_grants.*expires_at > NOW().*FOR UPDATE`
	updateWalletBalanceSQL     = `(?s)UPDATE users SET balance = balance - $1, bonus_balance = bonus_balance - $2, updated_at = NOW().*WHERE id = $3`
	insertWalletTransactionSQL = `(?s)INSERT INTO wallet_transactions.*VALUES`
	loadWalletSummarySQL       = `(?s)SELECT balance, bonus_balance, frozen_balance, frozen_bonus_balance.*FROM users WHERE id = $1 AND deleted_at IS NULL`
	nextExpiringWalletBonusSQL = `(?s)SELECT expires_at, SUM(remaining_amount).*FROM wallet_bonus_grants.*WHERE user_id = $1.*LIMIT 1`
)

func expectWalletDebitQueries(mock sqlmock.Sqlmock, userID, idempotencyKey string, balanceBefore, balanceAfter float64) {
	mock.ExpectQuery(walletTransactionExistsSQL).
		WithArgs(idempotencyKey).
		WillReturnRows(sqlmock.NewRows([]string{"?column?"}))
	mock.ExpectQuery(lockWalletUserSQL).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance", "bonus_balance", "frozen_balance", "frozen_bonus_balance"}).
			AddRow(balanceBefore, 0.0, 0.0, 0.0))
	mock.ExpectQuery(expiredWalletBonusSQL).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "remaining_amount"}))
	mock.ExpectQuery(availableWalletBonusSQL).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "remaining_amount"}))
	mock.ExpectExec(updateWalletBalanceSQL).
		WithArgs(balanceBefore-balanceAfter, 0.0, userID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(insertWalletTransactionSQL).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(loadWalletSummarySQL).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance", "bonus_balance", "frozen_balance", "frozen_bonus_balance"}).
			AddRow(balanceAfter, 0.0, 0.0, 0.0))
	mock.ExpectQuery(nextExpiringWalletBonusSQL).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"expires_at", "remaining_amount"}))
}

func TestApplyUsageBillingEffects_UsesStringWalletIDAndFlagsOverdraft(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	const (
		userID    = "20000000-0000-0000-0000-000000000042"
		apiKeyID  = "70000000-0000-0000-0000-000000000007"
		requestID = "req-wallet-overdraft"
	)
	mock.ExpectBegin()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	expectWalletDebitQueries(mock, userID, "wallet-usage:"+apiKeyID+":"+requestID, 5.0, -5.0)
	mock.ExpectCommit()

	result := &service.UsageBillingApplyResult{Applied: true}
	err = (&usageBillingRepository{}).applyUsageBillingEffects(ctx, tx, &service.UsageBillingCommand{
		RequestID:   requestID,
		APIKeyID:    apiKeyID,
		UserID:      userID,
		BalanceCost: 10,
	}, result)
	require.NoError(t, err)
	require.NotNil(t, result.NewBalance)
	require.InDelta(t, -5.0, *result.NewBalance, 0.000001)
	require.True(t, result.BalanceOverdrafted)
	require.NoError(t, tx.Commit())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestApplyUsageBillingEffects_ReturnsUserNotFoundForUnknownStringID(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	const (
		userID    = "20000000-0000-0000-0000-000000000404"
		apiKeyID  = "70000000-0000-0000-0000-000000000007"
		requestID = "req-wallet-missing-user"
	)
	mock.ExpectBegin()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	mock.ExpectQuery(walletTransactionExistsSQL).
		WithArgs("wallet-usage:" + apiKeyID + ":" + requestID).
		WillReturnRows(sqlmock.NewRows([]string{"?column?"}))
	mock.ExpectQuery(lockWalletUserSQL).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance", "bonus_balance", "frozen_balance", "frozen_bonus_balance"}))
	mock.ExpectRollback()

	result := &service.UsageBillingApplyResult{Applied: true}
	err = (&usageBillingRepository{}).applyUsageBillingEffects(ctx, tx, &service.UsageBillingCommand{
		RequestID:   requestID,
		APIKeyID:    apiKeyID,
		UserID:      userID,
		BalanceCost: 1,
	}, result)
	require.ErrorIs(t, err, service.ErrUserNotFound)
	require.Nil(t, result.NewBalance)
	require.NoError(t, tx.Rollback())
	require.NoError(t, mock.ExpectationsWereMet())
}
