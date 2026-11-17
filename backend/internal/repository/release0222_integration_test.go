//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRelease0222GroupUltrafastPersists(t *testing.T) {
	ctx := context.Background()
	repo := NewGroupRepository(integrationEntClient, integrationDB)
	group := &service.Group{Name: "ultrafast integration", Platform: service.PlatformOpenAI, Status: service.StatusActive, SubscriptionType: service.SubscriptionTypeStandard, RateMultiplier: 1, ForceOpenAIUltrafast: true}
	require.NoError(t, repo.Create(ctx, group))
	defer integrationDB.ExecContext(ctx, "DELETE FROM groups WHERE id=$1", group.ID)
	stored, err := repo.GetByID(ctx, group.ID)
	require.NoError(t, err)
	require.True(t, stored.ForceOpenAIUltrafast)
	stored.ForceOpenAIUltrafast = false
	require.NoError(t, repo.Update(ctx, stored))
	stored, err = repo.GetByID(ctx, group.ID)
	require.NoError(t, err)
	require.False(t, stored.ForceOpenAIUltrafast)
}

func TestRelease0222CooldownRecoveryChecksCurrentReason(t *testing.T) {
	ctx := context.Background()
	account := mustCreateAccount(t, integrationEntClient, &service.Account{Name: "recovery integration"})
	defer integrationDB.ExecContext(ctx, "DELETE FROM accounts WHERE id=$1", account.ID)
	repo := NewAccountRepository(integrationEntClient, integrationDB, nil).(*accountRepository)
	reset := time.Now().Add(time.Hour)
	require.NoError(t, repo.SetModelRateLimit(ctx, account.ID, "fable", reset, "window_exhausted"))
	require.NoError(t, repo.SetModelRateLimit(ctx, account.ID, "other", reset, "manual"))
	cleared, err := repo.ClearModelRateLimitIfReason(ctx, account.ID, "fable", "outdated_reason")
	require.NoError(t, err)
	require.False(t, cleared)
	cleared, err = repo.ClearModelRateLimitIfReason(ctx, account.ID, "fable", "window_exhausted")
	require.NoError(t, err)
	require.True(t, cleared)
	stored, err := repo.GetByID(ctx, account.ID)
	require.NoError(t, err)
	limits := stored.Extra["model_rate_limits"].(map[string]any)
	require.NotContains(t, limits, "fable")
	require.Contains(t, limits, "other")
}

func TestRelease0222UUIDRateSyncRollback(t *testing.T) {
	ctx := context.Background()
	user := mustCreateUser(t, integrationEntClient, &service.User{})
	group := mustCreateGroup(t, integrationEntClient, &service.Group{Name: "rate sync integration"})
	defer integrationDB.ExecContext(ctx, "DELETE FROM users WHERE id=$1", user.ID)
	defer integrationDB.ExecContext(ctx, "DELETE FROM groups WHERE id=$1", group.ID)
	repo := NewUserGroupRateRepository(integrationDB)
	require.NoError(t, repo.SyncGroupRateMultipliers(ctx, group.ID, []service.GroupRateMultiplierInput{{UserID: user.ID, RateMultiplier: 0}, {UserID: user.ID, RateMultiplier: 1.5}}))
	rates, err := repo.GetByUserID(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, 1.5, rates[group.ID])
	err = repo.SyncGroupRateMultipliers(ctx, group.ID, []service.GroupRateMultiplierInput{{UserID: uuid.NewString(), RateMultiplier: 2}})
	require.Error(t, err, "missing user must fail after the existing rows are cleared")
	rates, err = repo.GetByUserID(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, 1.5, rates[group.ID], "transaction must restore the existing rate")
	require.NoError(t, repo.SyncGroupRateMultipliers(ctx, group.ID, []service.GroupRateMultiplierInput{{UserID: user.ID, RateMultiplier: 0}}))
	rates, err = repo.GetByUserID(ctx, user.ID)
	require.NoError(t, err)
	require.Contains(t, rates, group.ID)
	require.Zero(t, rates[group.ID])
}
