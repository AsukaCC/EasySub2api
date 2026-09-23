package repository

import (
	"context"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestDebitWalletConsumesBonusThenRecharge(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT 1 FROM wallet_transactions").
		WithArgs("wallet-debit:test").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}))
	mock.ExpectQuery("(?s)SELECT recharge_balance, bonus_balance, frozen_recharge_balance, frozen_bonus_balance.*FOR UPDATE").
		WithArgs("user").
		WillReturnRows(sqlmock.NewRows([]string{"recharge_balance", "bonus_balance", "frozen_recharge_balance", "frozen_bonus_balance"}).AddRow(2.0, 3.0, 0.0, 0.0))
	mock.ExpectQuery("(?s)SELECT id, remaining_amount.*expires_at <= NOW.*FOR UPDATE").
		WithArgs("user").
		WillReturnRows(sqlmock.NewRows([]string{"id", "remaining_amount"}))
	mock.ExpectQuery("(?s)SELECT id, remaining_amount.*expires_at > NOW.*FOR UPDATE").
		WithArgs("user").
		WillReturnRows(sqlmock.NewRows([]string{"id", "remaining_amount"}).AddRow("grant-1", 3.0))
	mock.ExpectExec("UPDATE wallet_bonus_grants").
		WithArgs(3.0, "grant-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE users SET recharge_balance = recharge_balance -").
		WithArgs(2.0, 3.0, "user").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO wallet_transactions").
		WithArgs("user", "debit", -5.0, -3.0, -2.0, 0.0, 2.0, 0.0, 3.0, 0.0, "", "", "wallet-debit:test", "").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("(?s)SELECT recharge_balance, bonus_balance, frozen_recharge_balance, frozen_bonus_balance.*FROM users WHERE").
		WithArgs("user").
		WillReturnRows(sqlmock.NewRows([]string{"recharge_balance", "bonus_balance", "frozen_recharge_balance", "frozen_bonus_balance"}).AddRow(0.0, 0.0, 0.0, 0.0))
	mock.ExpectQuery("SELECT expires_at, SUM").
		WithArgs("user").
		WillReturnRows(sqlmock.NewRows([]string{"expires_at", "remaining_amount"}))

	result, err := debitWalletTx(context.Background(), db, service.WalletDebitInput{
		UserID: "user", Amount: 5, AllowOverdraft: false,
	}, "wallet-debit:test")
	require.NoError(t, err)
	require.Equal(t, 3.0, result.BonusAmount)
	require.Equal(t, 2.0, result.RechargeAmount)
	require.Equal(t, 0.0, result.Summary.RechargeBalance)
	require.Equal(t, 0.0, result.Summary.BonusBalance)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDebitWalletRejectsWhenRechargeBucketCannotCoverRemainder(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT 1 FROM wallet_transactions").
		WithArgs("wallet-debit:insufficient").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}))
	mock.ExpectQuery("(?s)SELECT recharge_balance, bonus_balance, frozen_recharge_balance, frozen_bonus_balance.*FOR UPDATE").
		WithArgs("user").
		WillReturnRows(sqlmock.NewRows([]string{"recharge_balance", "bonus_balance", "frozen_recharge_balance", "frozen_bonus_balance"}).AddRow(1.0, 2.0, 0.0, 0.0))
	mock.ExpectQuery("(?s)SELECT id, remaining_amount.*expires_at <= NOW.*FOR UPDATE").
		WithArgs("user").
		WillReturnRows(sqlmock.NewRows([]string{"id", "remaining_amount"}))

	_, err = debitWalletTx(context.Background(), db, service.WalletDebitInput{
		UserID: "user", Amount: 5, AllowOverdraft: false,
	}, "wallet-debit:insufficient")
	require.ErrorIs(t, err, service.ErrInsufficientBalance)
	require.NoError(t, mock.ExpectationsWereMet())
}
