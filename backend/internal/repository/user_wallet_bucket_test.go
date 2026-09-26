package repository

import (
	"context"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestWalletDebitCovered(t *testing.T) {
	tests := []struct {
		name           string
		recharge       float64
		bonus          float64
		amount         float64
		rechargeOnly   float64
		allowOverdraft bool
		want           bool
	}{
		{name: "bonus covers charge while recharge is already overdrawn", recharge: -5, bonus: 10, amount: 1, want: true},
		{name: "overdrawn recharge without bonus cannot pay", recharge: -5, amount: 1, want: false},
		{name: "usage overdraft may deepen recharge debt", recharge: 1, amount: 5, allowOverdraft: true, want: true},
		{name: "funded buckets cannot cover remainder", recharge: 1, bonus: 2, amount: 5, want: false},
		{name: "recharge-only discount cannot use bonus or debt", recharge: 0, bonus: 10, amount: 1, rechargeOnly: 1, allowOverdraft: true, want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := walletDebitCovered(walletUserRow{recharge: tc.recharge, bonus: tc.bonus}, tc.amount, tc.rechargeOnly, tc.allowOverdraft)
			require.Equal(t, tc.want, got)
		})
	}
}

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

func TestDebitWalletAllowsOverdraftOnRechargeRemainder(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT 1 FROM wallet_transactions").
		WithArgs("wallet-debit:overdraft").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}))
	mock.ExpectQuery("(?s)SELECT recharge_balance, bonus_balance, frozen_recharge_balance, frozen_bonus_balance.*FOR UPDATE").
		WithArgs("user").
		WillReturnRows(sqlmock.NewRows([]string{"recharge_balance", "bonus_balance", "frozen_recharge_balance", "frozen_bonus_balance"}).AddRow(1.0, 0.0, 0.0, 0.0))
	mock.ExpectQuery("(?s)SELECT id, remaining_amount.*expires_at <= NOW.*FOR UPDATE").
		WithArgs("user").
		WillReturnRows(sqlmock.NewRows([]string{"id", "remaining_amount"}))
	mock.ExpectQuery("(?s)SELECT id, remaining_amount.*expires_at > NOW.*FOR UPDATE").
		WithArgs("user").
		WillReturnRows(sqlmock.NewRows([]string{"id", "remaining_amount"}))
	mock.ExpectExec("UPDATE users SET recharge_balance = recharge_balance -").
		WithArgs(5.0, 0.0, "user").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO wallet_transactions").
		WithArgs("user", "debit", -5.0, 0.0, -5.0, 0.0, 1.0, -4.0, 0.0, 0.0, "", "", "wallet-debit:overdraft", "").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("(?s)SELECT recharge_balance, bonus_balance, frozen_recharge_balance, frozen_bonus_balance.*FROM users WHERE").
		WithArgs("user").
		WillReturnRows(sqlmock.NewRows([]string{"recharge_balance", "bonus_balance", "frozen_recharge_balance", "frozen_bonus_balance"}).AddRow(-4.0, 0.0, 0.0, 0.0))
	mock.ExpectQuery("SELECT expires_at, SUM").
		WithArgs("user").
		WillReturnRows(sqlmock.NewRows([]string{"expires_at", "remaining_amount"}))

	result, err := debitWalletTx(context.Background(), db, service.WalletDebitInput{
		UserID: "user", Amount: 5, AllowOverdraft: true,
	}, "wallet-debit:overdraft")
	require.NoError(t, err)
	require.Equal(t, 0.0, result.BonusAmount)
	require.Equal(t, 5.0, result.RechargeAmount)
	require.Equal(t, 0.0, result.Summary.AvailableBalance)
	require.Equal(t, 4.0, result.Summary.OverdraftAmount)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDebitWalletUsesBonusWhenRechargeBucketIsNegative(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT 1 FROM wallet_transactions").
		WithArgs("wallet-debit:bonus-covers-debt").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}))
	mock.ExpectQuery("(?s)SELECT recharge_balance, bonus_balance, frozen_recharge_balance, frozen_bonus_balance.*FOR UPDATE").
		WithArgs("user").
		WillReturnRows(sqlmock.NewRows([]string{"recharge_balance", "bonus_balance", "frozen_recharge_balance", "frozen_bonus_balance"}).AddRow(-5.0, 10.0, 0.0, 0.0))
	mock.ExpectQuery("(?s)SELECT id, remaining_amount.*expires_at <= NOW.*FOR UPDATE").
		WithArgs("user").
		WillReturnRows(sqlmock.NewRows([]string{"id", "remaining_amount"}))
	mock.ExpectQuery("(?s)SELECT id, remaining_amount.*expires_at > NOW.*FOR UPDATE").
		WithArgs("user").
		WillReturnRows(sqlmock.NewRows([]string{"id", "remaining_amount"}).AddRow("grant-1", 10.0))
	mock.ExpectExec("UPDATE wallet_bonus_grants").
		WithArgs(1.0, "grant-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE users SET recharge_balance = recharge_balance -").
		WithArgs(0.0, 1.0, "user").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO wallet_transactions").
		WithArgs("user", "debit", -1.0, -1.0, 0.0, 0.0, -5.0, -5.0, 10.0, 9.0, "", "", "wallet-debit:bonus-covers-debt", "").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("(?s)SELECT recharge_balance, bonus_balance, frozen_recharge_balance, frozen_bonus_balance.*FROM users WHERE").
		WithArgs("user").
		WillReturnRows(sqlmock.NewRows([]string{"recharge_balance", "bonus_balance", "frozen_recharge_balance", "frozen_bonus_balance"}).AddRow(-5.0, 9.0, 0.0, 0.0))
	mock.ExpectQuery("SELECT expires_at, SUM").
		WithArgs("user").
		WillReturnRows(sqlmock.NewRows([]string{"expires_at", "remaining_amount"}))

	result, err := debitWalletTx(context.Background(), db, service.WalletDebitInput{
		UserID: "user", Amount: 1, AllowOverdraft: false,
	}, "wallet-debit:bonus-covers-debt")
	require.NoError(t, err)
	require.Equal(t, 1.0, result.BonusAmount)
	require.Equal(t, 0.0, result.RechargeAmount)
	require.Equal(t, 9.0, result.Summary.AvailableBalance)
	require.Equal(t, 5.0, result.Summary.OverdraftAmount)
	require.NoError(t, mock.ExpectationsWereMet())
}
