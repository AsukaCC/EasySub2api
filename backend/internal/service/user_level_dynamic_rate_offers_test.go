//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type offerLevelRepo struct {
	spend float64
	used  map[DynamicRateUsageKey]float64
}

func (r *offerLevelRepo) GetRollingSpend(context.Context, string, time.Time, time.Time) (float64, error) {
	return r.spend, nil
}

func (r *offerLevelRepo) GetRollingSpendBatch(_ context.Context, userIDs []string, _, _ time.Time) (map[string]float64, error) {
	out := make(map[string]float64, len(userIDs))
	for _, id := range userIDs {
		out[id] = r.spend
	}
	return out, nil
}

func (r *offerLevelRepo) GetDynamicRateUsage(_ context.Context, _, _ string, keys []DynamicRateUsageKey) (map[DynamicRateUsageKey]float64, error) {
	out := make(map[DynamicRateUsageKey]float64, len(keys))
	for _, key := range keys {
		out[key] = r.used[key]
	}
	return out, nil
}

func (r *offerLevelRepo) GetSharedDynamicRateUsage(context.Context, string, []DynamicRateUsageKey) (map[DynamicRateUsageKey]float64, error) {
	return map[DynamicRateUsageKey]float64{}, nil
}

type offerGroupRepo struct {
	GroupRepository
	group *Group
}

func (r *offerGroupRepo) GetByIDLite(context.Context, string) (*Group, error) {
	return r.group, nil
}

func (r *offerGroupRepo) ListActive(context.Context) ([]Group, error) {
	if r.group == nil {
		return nil, nil
	}
	return []Group{*r.group}, nil
}

func offerWindowRule(start, end time.Time) GroupDynamicRateRule {
	return GroupDynamicRateRule{
		ID:                  "rule-1",
		Name:                "weekend",
		Enabled:             true,
		StartAt:             start.Format(time.RFC3339Nano),
		EndAt:               end.Format(time.RFC3339Nano),
		DiscountCoefficient: 0.8,
	}
}

func offerTestGroup(rule GroupDynamicRateRule) *Group {
	return &Group{
		ID:               "group-1",
		Name:             "VIP",
		RateMultiplier:   1,
		Status:           StatusActive,
		DynamicRateRules: []GroupDynamicRateRule{rule},
	}
}

func newOfferService(group *Group, repo *offerLevelRepo, withRules bool) *UserLevelService {
	svc := NewUserLevelService(repo, nil, &offerGroupRepo{group: group}, nil, nil)
	_ = withRules
	svc.rulesRepo = &singleLevelAssignments{assigned: []UserLevelRule{{ID: "default", WindowDays: 7, Enabled: true,
		Tiers: []UserLevelTier{{ID: "base", SortOrder: 0, MinSpend: 0}}}}}
	return svc
}

func TestListActiveDynamicRateOffersIncludesSpendThreshold(t *testing.T) {
	now := time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)
	rule := offerWindowRule(now.Add(-time.Hour), now.Add(time.Hour))
	rule.ActivationSpend = 100
	svc := newOfferService(offerTestGroup(rule), &offerLevelRepo{}, true)

	offers, err := svc.ListActiveDynamicRateOffers(context.Background(), "user-1", []string{"group-1"}, now)
	require.NoError(t, err)
	require.Len(t, offers, 1)
	require.Equal(t, "below_threshold", offers[0].Status)
	require.Equal(t, 100.0, offers[0].ActivationSpend)
	require.Zero(t, offers[0].Usage7d)
}

func TestListActiveDynamicRateOffersEmptyWhenOutsideWindow(t *testing.T) {
	now := time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name  string
		start time.Time
		end   time.Time
	}{
		{name: "not started", start: now.Add(time.Hour), end: now.Add(2 * time.Hour)},
		{name: "expired", start: now.Add(-2 * time.Hour), end: now.Add(-time.Hour)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			svc := newOfferService(offerTestGroup(offerWindowRule(test.start, test.end)), &offerLevelRepo{}, false)
			offers, err := svc.ListActiveDynamicRateOffers(context.Background(), "user-1", []string{"group-1"}, now)
			require.NoError(t, err)
			require.Empty(t, offers)
		})
	}
}

func TestListActiveDynamicRateOffersIncludesExhaustedQuota(t *testing.T) {
	now := time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)
	start := now.Add(-time.Hour)
	rule := offerWindowRule(start, now.Add(time.Hour))
	rule.PersonalQuotaAmount = 10
	svc := newOfferService(offerTestGroup(rule), &offerLevelRepo{
		used: map[DynamicRateUsageKey]float64{
			{RuleID: rule.ID, QuotaKey: start.Format(time.RFC3339Nano)}: 10,
		},
	}, false)

	offers, err := svc.ListActiveDynamicRateOffers(context.Background(), "user-1", []string{"group-1"}, now)
	require.NoError(t, err)
	require.Len(t, offers, 1)
	require.Equal(t, "quota_exhausted", offers[0].Status)
	require.Equal(t, 10.0, offers[0].PersonalQuotaAmount)
	require.Equal(t, 10.0, offers[0].PersonalUsedAmount)
}

func TestListActiveDynamicRateOffersReturnsSelectedWindow(t *testing.T) {
	now := time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)
	start := now.Add(-time.Hour)
	end := now.Add(2 * time.Hour)
	rule := offerWindowRule(start, end)
	rule.ActivationSpend = 50
	svc := newOfferService(offerTestGroup(rule), &offerLevelRepo{spend: 80}, true)

	offers, err := svc.ListActiveDynamicRateOffers(context.Background(), "user-1", []string{"group-1"}, now)
	require.NoError(t, err)
	require.Len(t, offers, 1)
	require.Equal(t, "participating", offers[0].Status)
	require.Equal(t, "group-1", offers[0].GroupID)
	require.Equal(t, "VIP", offers[0].GroupName)
	require.Equal(t, "rule-1", offers[0].RuleID)
	require.InDelta(t, 0.8, offers[0].DiscountCoefficient, 1e-12)
	require.Equal(t, "weekend", offers[0].RuleName)
	require.True(t, offers[0].StartAt.Equal(start.UTC()))
	require.True(t, offers[0].EndAt.Equal(end.UTC()))
}

type offerGroupsRepo struct {
	GroupRepository
	groups map[string]*Group
}

func (r *offerGroupsRepo) GetByIDLite(_ context.Context, id string) (*Group, error) {
	return r.groups[id], nil
}

func (r *offerGroupsRepo) ListActive(context.Context) ([]Group, error) {
	groups := make([]Group, 0, len(r.groups))
	for _, group := range r.groups {
		groups = append(groups, *group)
	}
	return groups, nil
}

func TestListActiveDynamicRateOffersIncludesDiscountForEveryGroup(t *testing.T) {
	now := time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)
	rule := offerWindowRule(now.Add(-time.Hour), now.Add(time.Hour))
	groups := &offerGroupsRepo{groups: make(map[string]*Group)}
	for id, coefficient := range map[string]float64{"group-1": .5, "group-2": .85, "no-discount": 1} {
		copy := rule
		copy.DiscountCoefficient = coefficient
		group := offerTestGroup(copy)
		group.ID = id
		groups.groups[id] = group
	}
	// Overlapping windows select the best discount; never emit duplicate groups.
	other := rule
	other.ID = "less-discount"
	other.DiscountCoefficient = .9
	groups.groups["group-1"].DynamicRateRules = append(groups.groups["group-1"].DynamicRateRules, other)
	svc := newOfferService(nil, &offerLevelRepo{}, true)
	svc.groupRepo = groups
	offers, err := svc.ListActiveDynamicRateOffers(context.Background(), "user", []string{"group-1", "group-2", "no-discount", "group-1"}, now)
	require.NoError(t, err)
	require.Len(t, offers, 2)
	byGroup := make(map[string]float64)
	for _, offer := range offers {
		byGroup[offer.GroupID] = offer.DiscountCoefficient
	}
	require.Equal(t, map[string]float64{"group-1": .5, "group-2": .85}, byGroup)
}

func TestListActiveDynamicRateOffersVisibleWithoutGroupAccess(t *testing.T) {
	now := time.Now()
	for _, subscription := range []bool{false, true} {
		group := offerTestGroup(offerWindowRule(now.Add(-time.Hour), now.Add(time.Hour)))
		wantStatus := "group_unavailable"
		if subscription {
			group.SubscriptionType = "subscription"
			wantStatus = "subscription_required"
		}
		svc := newOfferService(group, &offerLevelRepo{}, true)
		offers, err := svc.ListActiveDynamicRateOffers(context.Background(), "user", nil, now)
		require.NoError(t, err)
		require.Len(t, offers, 1)
		require.Equal(t, wantStatus, offers[0].Status)
	}
}

func TestListActiveDynamicRateOffersVisibleWithoutLevelConfiguration(t *testing.T) {
	now := time.Now()
	group := offerTestGroup(offerWindowRule(now.Add(-time.Hour), now.Add(time.Hour)))
	svc := newOfferService(group, &offerLevelRepo{spend: 25}, true)
	svc.rulesRepo = &singleLevelAssignments{}
	offers, err := svc.ListActiveDynamicRateOffers(context.Background(), "user", []string{group.ID}, now)
	require.NoError(t, err)
	require.Len(t, offers, 1)
	require.Equal(t, "level_required", offers[0].Status)
	require.Equal(t, 25.0, offers[0].Usage7d)
}

type offerSubscriptionRepo struct {
	UserSubscriptionRepository
	sub *UserSubscription
	err error
}

func (r *offerSubscriptionRepo) GetActiveByUserIDAndGroupID(context.Context, string, string) (*UserSubscription, error) {
	return r.sub, r.err
}

func TestListActiveDynamicRateOffersSubscriptionConditions(t *testing.T) {
	now := time.Now()
	for _, test := range []struct {
		name   string
		sub    *UserSubscription
		err    error
		status string
	}{
		{name: "missing", err: ErrSubscriptionNotFound, status: "subscription_required"},
		{name: "expired", sub: &UserSubscription{Status: SubscriptionStatusActive, ExpiresAt: now.Add(-time.Hour)}, status: "subscription_required"},
		{name: "active", sub: &UserSubscription{Status: SubscriptionStatusActive, ExpiresAt: now.Add(time.Hour)}, status: "participating"},
		{name: "quota exceeded", sub: &UserSubscription{Status: SubscriptionStatusActive, ExpiresAt: now.Add(time.Hour), DailyUsageUSD: 11}, status: "subscription_limited"},
		{name: "repository error", err: errors.New("unavailable")},
	} {
		t.Run(test.name, func(t *testing.T) {
			group := offerTestGroup(offerWindowRule(now.Add(-time.Hour), now.Add(time.Hour)))
			group.SubscriptionType = "subscription"
			group.DailyLimitUSD = userLevelFloatPtr(10)
			svc := newOfferService(group, &offerLevelRepo{}, true)
			svc.subRepo = &offerSubscriptionRepo{sub: test.sub, err: test.err}
			offers, err := svc.ListActiveDynamicRateOffers(context.Background(), "user", []string{group.ID}, now)
			if test.status == "" {
				require.ErrorIs(t, err, test.err)
				require.Empty(t, offers)
				return
			}
			require.NoError(t, err)
			require.Len(t, offers, 1)
			require.Equal(t, test.status, offers[0].Status)
		})
	}
}

func TestListActiveDynamicRateOffersSelectionMatchesBilling(t *testing.T) {
	now := time.Now()
	rule := offerWindowRule(now.Add(-time.Hour), now.Add(time.Hour))
	rule.ActivationSpend = 100
	group := offerTestGroup(rule)
	better := rule
	better.ID = "better"
	better.ActivationSpend = 200
	better.DiscountCoefficient = .5
	group.DynamicRateRules = append(group.DynamicRateRules, better)
	for _, spend := range []float64{99, 100, 199, 200} {
		svc := newOfferService(group, &offerLevelRepo{spend: spend}, true)
		offers, err := svc.ListActiveDynamicRateOffers(context.Background(), "user", []string{group.ID}, now)
		require.NoError(t, err)
		require.Len(t, offers, 1)
		plan, err := svc.ResolvePlan(context.Background(), "user", group, now)
		require.NoError(t, err)
		if spend < 100 {
			require.Equal(t, "below_threshold", offers[0].Status)
			require.Empty(t, plan.SelectedDynamicRuleID)
		} else {
			require.Equal(t, "participating", offers[0].Status)
			require.Equal(t, plan.SelectedDynamicRuleID, offers[0].RuleID)
		}
	}
}

func TestListActiveDynamicRateOffersExcludesDisabledAndInvalidRules(t *testing.T) {
	now := time.Now()
	for _, mutate := range []func(*Group){
		func(g *Group) { g.Status = "inactive" },
		func(g *Group) { g.DynamicRateRules[0].Enabled = false },
		func(g *Group) { g.DynamicRateRules[0].StartAt = "invalid" },
		func(g *Group) { g.DynamicRateRules[0].DiscountCoefficient = 1 },
		func(g *Group) { g.DynamicRateRules[0].DiscountCoefficient = .001 },
	} {
		group := offerTestGroup(offerWindowRule(now.Add(-time.Hour), now.Add(time.Hour)))
		mutate(group)
		svc := newOfferService(group, &offerLevelRepo{}, true)
		offers, err := svc.ListActiveDynamicRateOffers(context.Background(), "user", []string{group.ID}, now)
		require.NoError(t, err)
		require.Empty(t, offers)
	}
}
