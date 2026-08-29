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

func queryAffiliateReservedFloat(t *testing.T, ctx context.Context, client *dbent.Client, query string, args ...any) float64 {
	t.Helper()
	rows, err := client.QueryContext(ctx, query, args...)
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()
	require.True(t, rows.Next())
	var value float64
	require.NoError(t, rows.Scan(&value))
	require.NoError(t, rows.Err())
	return value
}

func TestAffiliateRepository_AvailableQuotaExcludesReservedReversal(t *testing.T) {
	ctx := context.Background()
	tx := testEntTx(t)
	txCtx := dbent.NewTxContext(ctx, tx)
	client := tx.Client()

	user, err := client.User.Create().
		SetEmail(fmt.Sprintf("affiliate-reserved-%d@example.com", time.Now().UnixNano())).
		SetUsername("affiliate-reserved").
		SetPasswordHash("hash").
		SetRole(service.RoleUser).
		SetStatus(service.StatusActive).
		Save(txCtx)
	require.NoError(t, err)
	affCode := fmt.Sprintf("HLD%09d", time.Now().UnixNano()%1_000_000_000)
	_, err = client.ExecContext(txCtx, `
INSERT INTO user_affiliates
    (user_id, aff_code, aff_quota, aff_frozen_quota, aff_history_quota, created_at, updated_at)
VALUES ($1, $2, 12, 8, 20, NOW(), NOW())`, user.ID, affCode)
	require.NoError(t, err)
	_, err = client.ExecContext(txCtx, `
INSERT INTO user_affiliate_ledger
    (user_id, action, amount, reserved_reversal_amount, frozen_until, created_at, updated_at)
VALUES ($1, 'accrue', 8, 3, NOW() - INTERVAL '1 minute', NOW(), NOW())`, user.ID)
	require.NoError(t, err)

	repo := NewAffiliateRepository(client, integrationDB)
	overview, err := repo.GetAffiliateUserOverview(txCtx, user.ID)
	require.NoError(t, err)
	require.InDelta(t, 17.0, overview.AvailableQuota, 1e-9)

	thawed, err := repo.ThawFrozenQuota(txCtx, user.ID)
	require.NoError(t, err)
	require.InDelta(t, 8.0, thawed, 1e-9)
	summary, err := repo.EnsureUserAffiliate(txCtx, user.ID)
	require.NoError(t, err)
	require.InDelta(t, 17.0, summary.AffQuota, 1e-9)

	transferred, _, err := repo.TransferQuotaToBalance(txCtx, user.ID)
	require.NoError(t, err)
	require.InDelta(t, 17.0, transferred, 1e-9)
	require.InDelta(t, 3.0, queryAffiliateReservedFloat(t, txCtx, client,
		"SELECT aff_quota::double precision FROM user_affiliates WHERE user_id = $1", user.ID), 1e-9)
}
