//go:build integration

package repository

import (
	"context"
	"sync"
	"testing"
	"time"

	dbent "github.com/AsukaCC/EasySub2api/ent"
	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestPermanentPointsRefundAndLedger(t *testing.T) {
	tx := testEntTx(t)
	ctx := dbent.NewTxContext(context.Background(), tx)
	client := tx.Client()
	repo := newUserRepositoryWithSQL(client, integrationDB)
	user := mustCreateUser(t, client, &service.User{Email: newWalletUUID() + "@points.example.com", Username: "point-ledger-fixture"})
	credit := func(kind string, amount float64) service.WalletMutationResult {
		result, err := repo.CreditWallet(ctx, service.WalletCreditInput{UserID: user.ID, Kind: kind, Amount: amount, IdempotencyKey: newWalletUUID()})
		require.NoError(t, err)
		return result
	}
	credit(service.WalletKindRecharge, 100)
	bonus := credit(service.WalletKindBonus, 20)
	// Even a legacy timestamp must not trigger an expiry debit.
	_, err := client.ExecContext(ctx, "UPDATE wallet_bonus_grants SET expires_at = NOW() - INTERVAL '1 day' WHERE id = $1", *bonus.BonusGrantID)
	require.NoError(t, err)
	wallet, err := repo.GetWalletSummary(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, 20.0, wallet.BonusBalance)
	_, err = repo.DebitWallet(ctx, service.WalletDebitInput{UserID: user.ID, Amount: 20, IdempotencyKey: newWalletUUID()})
	require.NoError(t, err)
	input := service.RefundPointHoldInput{UserID: user.ID, RefundID: newWalletUUID(), BonusGrantID: *bonus.BonusGrantID, BasePoints: 100, BonusPoints: 20, RequestFingerprint: "same-request"}
	hold, err := repo.HoldRefundPoints(ctx, input)
	require.NoError(t, err)
	require.Equal(t, -20.0, hold.Summary.BonusBalance)
	duplicate, err := repo.HoldRefundPoints(ctx, input)
	require.NoError(t, err)
	require.False(t, duplicate.Applied)
	for i := 0; i < 2; i++ {
		result, err := repo.CaptureWalletHold(ctx, hold.HoldID, "capture:"+input.RefundID)
		require.NoError(t, err)
		require.Equal(t, i == 0, result.Applied)
		require.Equal(t, -20.0, result.Summary.BonusBalance)
	}
	recharged := credit(service.WalletKindRecharge, 100)
	require.Equal(t, 80.0, recharged.Summary.AvailableBalance)
	require.Equal(t, -20.0, recharged.Summary.BonusBalance)
	credited := credit(service.WalletKindBonus, 10)
	require.Equal(t, -10.0, credited.Summary.BonusBalance)
	require.Equal(t, 90.0, credited.Summary.AvailableBalance)
	from, to := time.Now().UTC().Add(-time.Hour), time.Now().UTC().Add(time.Hour)
	filter := service.PointChangeFilter{UserID: user.ID, Keyword: "point-ledger-fixture", PointType: "bonus", Direction: "decrease", StartTime: &from, EndTime: &to, Page: 1, PageSize: 100}
	page, err := repo.ListPointChanges(ctx, filter)
	require.NoError(t, err)
	require.EqualValues(t, 2, page.Total) // consumption and refund hold; capture is not another deduction
	for _, item := range page.Items {
		require.Equal(t, -20.0, item.Amount)
		require.NotNil(t, item.BalanceBefore)
		require.NotNil(t, item.BalanceAfter)
		if item.Action == "hold" {
			require.Equal(t, -120.0, *item.BalanceAfter-*item.BalanceBefore)
		} else {
			require.Equal(t, item.Amount, *item.BalanceAfter-*item.BalanceBefore)
		}
	}
	filter.Direction = ""
	filter.PointType = ""
	all, err := repo.ListPointChanges(ctx, filter)
	require.NoError(t, err)
	require.EqualValues(t, 9, all.Total) // four credits, one debit, two hold and two capture rows
	filter.Keyword = "missing-user"
	empty, err := repo.ListPointChanges(ctx, filter)
	require.NoError(t, err)
	require.Empty(t, empty.Items)
	require.Zero(t, empty.Total)
}

func TestPermanentPointsConcurrentRefundAndConsumption(t *testing.T) {
	ctx := context.Background()
	client := integrationEntClient
	repo := newUserRepositoryWithSQL(client, integrationDB)
	user := mustCreateUser(t, client, &service.User{Email: newWalletUUID() + "@points.example.com"})
	t.Cleanup(func() { _ = client.User.DeleteOneID(user.ID).Exec(ctx) })
	_, err := repo.CreditWallet(ctx, service.WalletCreditInput{UserID: user.ID, Kind: service.WalletKindRecharge, Amount: 100, IdempotencyKey: newWalletUUID()})
	require.NoError(t, err)
	var holdErr, debitErr error
	var wg sync.WaitGroup
	start := make(chan struct{})
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		_, holdErr = repo.HoldRefundPoints(ctx, service.RefundPointHoldInput{UserID: user.ID, RefundID: newWalletUUID(), BasePoints: 100, BonusPoints: 20})
	}()
	go func() {
		defer wg.Done()
		<-start
		_, debitErr = repo.DebitWallet(ctx, service.WalletDebitInput{UserID: user.ID, Amount: 60, IdempotencyKey: newWalletUUID()})
	}()
	close(start)
	wg.Wait()
	wallet, err := repo.GetWalletSummary(ctx, user.ID)
	require.NoError(t, err)
	if holdErr == nil {
		require.ErrorIs(t, debitErr, service.ErrInsufficientBalance)
		require.Equal(t, 0.0, wallet.RechargeBalance)
		require.Equal(t, -20.0, wallet.BonusBalance)
	} else {
		require.NoError(t, debitErr)
		require.ErrorIs(t, holdErr, service.ErrInsufficientBalance)
		require.Equal(t, 40.0, wallet.RechargeBalance)
		require.Zero(t, wallet.BonusBalance)
	}
}
