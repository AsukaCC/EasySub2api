//go:build unit

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/pkg/ctxkey"
	"github.com/stretchr/testify/require"
)

func TestRefreshWSTurnBillingUsesSelectedGroupAndFreezesPlan(t *testing.T) {
	group := &Group{ID: "selected", Platform: PlatformOpenAI, Status: StatusActive, Hydrated: true, RateMultiplier: .8, ProfitControlEnabled: true}
	repo := &dynamicRateSummaryGroupRepository{group: group}
	assignments := &singleLevelAssignments{assigned: []UserLevelRule{{ID: "level", Enabled: true, WindowDays: 7, Tiers: []UserLevelTier{{ID: "base", DefaultMultiplier: userLevelFloatPtr(.5)}}}}}
	levels := NewUserLevelService(&offerLevelRepo{}, nil, repo, nil, nil)
	levels.rulesRepo = assignments
	svc := &OpenAIGatewayService{userLevelService: levels}
	entry := &Group{ID: "entry", Platform: PlatformOpenAI, RateMultiplier: 10}
	key := &APIKey{UserID: "user", Group: entry}
	selected := &AccountSelectionResult{BillingGroup: group}
	ctx := context.WithValue(context.Background(), ctxkey.UserID, "user")
	turnCtx, first, at, err := svc.RefreshWSTurnBilling(ctx, key, selected)
	require.NoError(t, err)
	require.Equal(t, "selected", first.BillingGroup.ID)
	require.Equal(t, at, OpenAIPricingAtFromContext(turnCtx))
	require.NotNil(t, first.profitGate)
	require.Equal(t, at, first.profitGate.pricingAt)
	require.InDelta(t, .4, first.RatePlan.EffectiveMultiplier, 1e-12)
	require.InDelta(t, .4, first.profitGate.threshold, 1e-12)

	updated := *group
	updated.RateMultiplier = .2
	repo.group = &updated
	_, second, _, err := svc.RefreshWSTurnBilling(ctx, key, selected)
	require.NoError(t, err)
	require.InDelta(t, .1, second.RatePlan.EffectiveMultiplier, 1e-12)
	require.InDelta(t, .4, first.RatePlan.EffectiveMultiplier, 1e-12)
	require.Same(t, entry, key.Group)
	require.Same(t, group, selected.BillingGroup)

	assignments.err = errors.New("unavailable")
	_, snapshot, _, err := svc.RefreshWSTurnBilling(ctx, key, selected)
	require.Error(t, err)
	require.Nil(t, snapshot)
	assignments.err = nil
	updated.Platform = PlatformAnthropic
	_, _, _, err = svc.RefreshWSTurnBilling(ctx, key, selected)
	require.Error(t, err)
}
