//go:build unit

package service

import (
	"context"
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

func TestListActiveDynamicRateOffersEmptyWhenBelowSpendThreshold(t *testing.T) {
	now := time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)
	rule := offerWindowRule(now.Add(-time.Hour), now.Add(time.Hour))
	rule.ActivationSpend = 100
	svc := newOfferService(offerTestGroup(rule), &offerLevelRepo{}, true)

	offers, err := svc.ListActiveDynamicRateOffers(context.Background(), "user-1", []string{"group-1"}, now)
	require.NoError(t, err)
	require.Empty(t, offers)
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

func TestListActiveDynamicRateOffersEmptyWhenQuotaExhausted(t *testing.T) {
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
	require.Empty(t, offers)
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
