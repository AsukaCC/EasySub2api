//go:build integration

package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	dbent "github.com/AsukaCC/EasySub2api/ent"
	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/stretchr/testify/require"
)

type walletBonusGrantLifecycle struct {
	original  float64
	remaining float64
	frozen    float64
	spent     float64
	expired   float64
	reversed  float64
	status    string
}

func loadWalletBonusGrantLifecycle(t *testing.T, ctx context.Context, client *dbent.Client, grantID string) walletBonusGrantLifecycle {
	t.Helper()

	rows, err := client.QueryContext(ctx, `
		SELECT original_amount::double precision,
		       remaining_amount::double precision,
		       frozen_amount::double precision,
		       spent_amount::double precision,
		       expired_amount::double precision,
		       reversed_amount::double precision,
		       status
		FROM wallet_bonus_grants
		WHERE id = $1
	`, grantID)
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()
	require.True(t, rows.Next(), "expected wallet bonus grant %s", grantID)

	var grant walletBonusGrantLifecycle
	require.NoError(t, rows.Scan(
		&grant.original,
		&grant.remaining,
		&grant.frozen,
		&grant.spent,
		&grant.expired,
		&grant.reversed,
		&grant.status,
	))
	require.NoError(t, rows.Err())
	require.InDelta(t, grant.original, grant.remaining+grant.frozen+grant.spent+grant.expired+grant.reversed, 1e-8,
		"grant lifecycle buckets must conserve the original amount")
	return grant
}

func requireGrantBuckets(t *testing.T, grant walletBonusGrantLifecycle, remaining, frozen, spent, expired, reversed float64) {
	t.Helper()
	require.InDelta(t, remaining, grant.remaining, 1e-8)
	require.InDelta(t, frozen, grant.frozen, 1e-8)
	require.InDelta(t, spent, grant.spent, 1e-8)
	require.InDelta(t, expired, grant.expired, 1e-8)
	require.InDelta(t, reversed, grant.reversed, 1e-8)
}

func TestUserWalletBonusLifecycle_CreditAgainstDebtAndAdminSetAreClassified(t *testing.T) {
	tx := testEntTx(t)
	ctx := dbent.NewTxContext(context.Background(), tx)
	client := tx.Client()
	repo := newUserRepositoryWithSQL(client, integrationDB)
	user := mustCreateUser(t, client, &service.User{
		Email:   fmt.Sprintf("wallet-debt-%d@example.com", time.Now().UnixNano()),
		Balance: -4,
	})
	expiresAt := time.Now().UTC().Add(24 * time.Hour)

	credit, err := repo.CreditWallet(ctx, service.WalletCreditInput{
		UserID: user.ID, Amount: 10, Kind: service.WalletKindBonus, ExpiresAt: &expiresAt,
		SourceType: "wallet_lifecycle_test", SourceID: newWalletUUID(), IdempotencyKey: newWalletUUID(),
	})
	require.NoError(t, err)
	require.NotNil(t, credit.BonusGrantID)
	require.InDelta(t, 6, credit.Summary.Balance, 1e-8)
	require.InDelta(t, 6, credit.Summary.BonusBalance, 1e-8)
	grant := loadWalletBonusGrantLifecycle(t, ctx, client, *credit.BonusGrantID)
	requireGrantBuckets(t, grant, 6, 0, 4, 0, 0)

	_, err = repo.SetWalletBalance(ctx, service.WalletSetInput{
		UserID: user.ID, RechargeAmount: 2, BonusAmount: 3, BonusExpiresAt: &expiresAt,
		SourceType: "wallet_lifecycle_test", SourceID: newWalletUUID(), IdempotencyKey: newWalletUUID(),
	})
	require.NoError(t, err)
	grant = loadWalletBonusGrantLifecycle(t, ctx, client, *credit.BonusGrantID)
	requireGrantBuckets(t, grant, 0, 0, 4, 0, 6)
	require.Equal(t, "adjusted", grant.status)
}

func TestUserWalletBonusLifecycle_CaptureAndRefundMoveSpentBackToRemaining(t *testing.T) {
	tx := testEntTx(t)
	ctx := dbent.NewTxContext(context.Background(), tx)
	client := tx.Client()
	repo := newUserRepositoryWithSQL(client, integrationDB)
	user := mustCreateUser(t, client, &service.User{
		Email: fmt.Sprintf("wallet-refund-%d@example.com", time.Now().UnixNano()),
	})
	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	credit, err := repo.CreditWallet(ctx, service.WalletCreditInput{
		UserID: user.ID, Amount: 10, Kind: service.WalletKindBonus, ExpiresAt: &expiresAt,
		SourceType: "wallet_lifecycle_test", SourceID: newWalletUUID(), IdempotencyKey: newWalletUUID(),
	})
	require.NoError(t, err)
	require.NotNil(t, credit.BonusGrantID)

	hold, err := repo.HoldWallet(ctx, service.WalletHoldInput{
		UserID: user.ID, Amount: 6, Purpose: "subscription", ReferenceID: newWalletUUID(),
	})
	require.NoError(t, err)
	requireGrantBuckets(t, loadWalletBonusGrantLifecycle(t, ctx, client, *credit.BonusGrantID), 4, 6, 0, 0, 0)

	_, err = repo.CaptureWalletHold(ctx, hold.HoldID, newWalletUUID())
	require.NoError(t, err)
	requireGrantBuckets(t, loadWalletBonusGrantLifecycle(t, ctx, client, *credit.BonusGrantID), 4, 0, 6, 0, 0)

	_, err = repo.RefundWalletHold(ctx, hold.HoldID, 2, newWalletUUID())
	require.NoError(t, err)
	requireGrantBuckets(t, loadWalletBonusGrantLifecycle(t, ctx, client, *credit.BonusGrantID), 6, 0, 4, 0, 0)
}

func TestUserWalletBonusLifecycle_ExpiredHoldReleaseAndRefundStayExpired(t *testing.T) {
	tests := []struct {
		name         string
		captureFirst bool
		refundAfter  float64
		wantSpent    float64
		wantExpired  float64
	}{
		{name: "release", wantExpired: 10},
		{name: "capture_then_refund", captureFirst: true, refundAfter: 2, wantSpent: 4, wantExpired: 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := testEntTx(t)
			ctx := dbent.NewTxContext(context.Background(), tx)
			client := tx.Client()
			repo := newUserRepositoryWithSQL(client, integrationDB)
			user := mustCreateUser(t, client, &service.User{
				Email: fmt.Sprintf("wallet-expired-%s-%d@example.com", tt.name, time.Now().UnixNano()),
			})
			expiresAt := time.Now().UTC().Add(24 * time.Hour)
			credit, err := repo.CreditWallet(ctx, service.WalletCreditInput{
				UserID: user.ID, Amount: 10, Kind: service.WalletKindBonus, ExpiresAt: &expiresAt,
				SourceType: "wallet_lifecycle_test", SourceID: newWalletUUID(), IdempotencyKey: newWalletUUID(),
			})
			require.NoError(t, err)
			require.NotNil(t, credit.BonusGrantID)
			hold, err := repo.HoldWallet(ctx, service.WalletHoldInput{
				UserID: user.ID, Amount: 6, Purpose: "subscription", ReferenceID: newWalletUUID(),
			})
			require.NoError(t, err)
			_, err = client.ExecContext(ctx, `UPDATE wallet_bonus_grants SET expires_at = NOW() - INTERVAL '1 minute' WHERE id = $1`, *credit.BonusGrantID)
			require.NoError(t, err)

			if tt.captureFirst {
				_, err = repo.CaptureWalletHold(ctx, hold.HoldID, newWalletUUID())
				require.NoError(t, err)
				_, err = repo.RefundWalletHold(ctx, hold.HoldID, tt.refundAfter, newWalletUUID())
				require.NoError(t, err)
			} else {
				_, err = repo.ReleaseWalletHold(ctx, hold.HoldID, newWalletUUID())
				require.NoError(t, err)
			}

			grant := loadWalletBonusGrantLifecycle(t, ctx, client, *credit.BonusGrantID)
			requireGrantBuckets(t, grant, 0, 0, tt.wantSpent, tt.wantExpired, 0)
			require.Equal(t, "expired", grant.status)
		})
	}
}

func TestUserWalletBonusLifecycle_RefundRecoveryCaptureIsReversedNotSpent(t *testing.T) {
	tx := testEntTx(t)
	ctx := dbent.NewTxContext(context.Background(), tx)
	client := tx.Client()
	repo := newUserRepositoryWithSQL(client, integrationDB)
	user := mustCreateUser(t, client, &service.User{
		Email: fmt.Sprintf("wallet-reversal-%d@example.com", time.Now().UnixNano()),
	})
	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	credit, err := repo.CreditWallet(ctx, service.WalletCreditInput{
		UserID: user.ID, Amount: 10, Kind: service.WalletKindBonus, ExpiresAt: &expiresAt,
		SourceType: "wallet_lifecycle_test", SourceID: newWalletUUID(), IdempotencyKey: newWalletUUID(),
	})
	require.NoError(t, err)
	require.NotNil(t, credit.BonusGrantID)
	hold, err := repo.HoldRefundPoints(ctx, service.RefundPointHoldInput{
		UserID: user.ID, RefundID: newWalletUUID(), BonusGrantID: *credit.BonusGrantID,
		BonusPoints: 6, RequestFingerprint: newWalletUUID(),
	})
	require.NoError(t, err)

	_, err = repo.CaptureWalletHold(ctx, hold.HoldID, newWalletUUID())
	require.NoError(t, err)
	grant := loadWalletBonusGrantLifecycle(t, ctx, client, *credit.BonusGrantID)
	requireGrantBuckets(t, grant, 4, 0, 0, 0, 6)
}
