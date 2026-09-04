package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func absoluteRuleForTest(start, end string) GroupDynamicRateRule {
	return GroupDynamicRateRule{
		ID:         uuid.NewString(),
		Name:       "absolute window",
		Enabled:    true,
		StartAt:    start,
		EndAt:      end,
		Multiplier: 0.8,
	}
}

func TestNormalizeDynamicRateRulesNormalizesAbsoluteWindowToUTC(t *testing.T) {
	rules, err := NormalizeDynamicRateRules([]GroupDynamicRateRule{{
		ID:                  uuid.NewString(),
		Name:                "discount",
		Enabled:             true,
		StartAt:             "2026-09-01T10:00:00.123456789+08:00",
		EndAt:               "2026-09-01T12:00:00+08:00",
		Multiplier:          0.87555,
		SharedQuotaAmount:   1.234567895,
		PersonalQuotaAmount: 2.345678895,
	}})

	require.NoError(t, err)
	require.Len(t, rules, 1)
	require.Equal(t, "2026-09-01T02:00:00.123456789Z", rules[0].StartAt)
	require.Equal(t, "2026-09-01T04:00:00Z", rules[0].EndAt)
	require.Equal(t, 0.8756, rules[0].Multiplier)
	require.Equal(t, 1.2345679, rules[0].SharedQuotaAmount)
	require.Equal(t, 2.3456789, rules[0].PersonalQuotaAmount)
	require.Empty(t, rules[0].StartTime)
	require.Empty(t, rules[0].EndTime)
	require.Empty(t, rules[0].QuotaAmount)
}

func TestNormalizeDynamicRateRulesRejectsIncompleteOrInvalidAbsoluteWindow(t *testing.T) {
	tests := []struct {
		name string
		rule GroupDynamicRateRule
		want string
	}{
		{
			name: "missing window",
			rule: GroupDynamicRateRule{ID: uuid.NewString(), Name: "missing", Multiplier: 1},
			want: "start_at and end_at are required",
		},
		{
			name: "partial window",
			rule: GroupDynamicRateRule{ID: uuid.NewString(), Name: "partial", StartAt: "2026-09-01T00:00:00Z", Multiplier: 1},
			want: "start_at and end_at are required",
		},
		{
			name: "reversed window",
			rule: absoluteRuleForTest("2026-09-01T02:00:00Z", "2026-09-01T01:00:00Z"),
			want: "start_at and end_at must be valid RFC3339",
		},
		{
			name: "missing timezone",
			rule: absoluteRuleForTest("2026-09-01T01:00:00", "2026-09-01T02:00:00"),
			want: "start_at and end_at must be valid RFC3339",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NormalizeDynamicRateRules([]GroupDynamicRateRule{test.rule})
			require.Error(t, err)
			require.Contains(t, err.Error(), test.want)
		})
	}
}

func TestDynamicRuleAppliesUsesAbsoluteHalfOpenInterval(t *testing.T) {
	start := time.Date(2026, 9, 1, 2, 0, 0, 123456789, time.UTC)
	end := start.Add(2 * time.Hour)
	rule := absoluteRuleForTest(start.Format(time.RFC3339Nano), end.Format(time.RFC3339Nano))
	profile := UserLevelProfile{Level: 1}

	key, applies := dynamicRuleApplies(rule, profile, start)
	require.True(t, applies)
	require.Equal(t, start.Format(time.RFC3339Nano), key)

	_, applies = dynamicRuleApplies(rule, profile, end.Add(-time.Nanosecond))
	require.True(t, applies)
	_, applies = dynamicRuleApplies(rule, profile, end)
	require.False(t, applies)
	_, applies = dynamicRuleApplies(rule, profile, start.Add(-time.Nanosecond))
	require.False(t, applies)
}

func TestLegacyDynamicRuleDoesNotApplyAndIsMarkedLegacy(t *testing.T) {
	rule := GroupDynamicRateRule{
		ID:         uuid.NewString(),
		Name:       "old daily rule",
		Enabled:    true,
		Timezone:   "Asia/Shanghai",
		StartTime:  "09:00",
		EndTime:    "10:00",
		Multiplier: 0.8,
	}
	profile := UserLevelProfile{Level: 1}

	_, applies := dynamicRuleApplies(rule, profile, time.Date(2026, 9, 1, 1, 30, 0, 0, time.UTC))
	require.False(t, applies)
	require.Equal(t, DynamicRateStatusLegacy, dynamicRateRuleStatus(rule, time.Now()))

	rules, err := NormalizeDynamicRateRules([]GroupDynamicRateRule{rule})
	require.NoError(t, err)
	require.Len(t, rules, 1)
	require.True(t, strings.HasPrefix(rules[0].StartTime, "09:"))
}

func TestDynamicRateRuleStatusTracksAbsoluteWindow(t *testing.T) {
	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	rule := absoluteRuleForTest(start.Format(time.RFC3339), end.Format(time.RFC3339))
	require.Equal(t, DynamicRateStatusNotStarted, dynamicRateRuleStatus(rule, start.Add(-time.Nanosecond)))
	require.Equal(t, DynamicRateStatusActive, dynamicRateRuleStatus(rule, start))
	require.Equal(t, DynamicRateStatusExpired, dynamicRateRuleStatus(rule, end))
}

type dynamicRateSummaryRepository struct {
	UserLevelRepository
	shared map[DynamicRateUsageKey]float64
}

func (r *dynamicRateSummaryRepository) GetSharedDynamicRateUsage(_ context.Context, _ string, _ []DynamicRateUsageKey) (map[DynamicRateUsageKey]float64, error) {
	return r.shared, nil
}

type dynamicRateSummaryGroupRepository struct {
	GroupRepository
	group *Group
}

func (r *dynamicRateSummaryGroupRepository) GetByIDLite(context.Context, string) (*Group, error) {
	return r.group, nil
}

func TestGetDynamicRateUsageSummaryReportsSharedRemainderAndStates(t *testing.T) {
	now := time.Date(2026, 9, 3, 1, 0, 0, 0, time.UTC)
	active := absoluteRuleForTest("2026-09-03T00:00:00Z", "2026-09-03T02:00:00Z")
	active.SharedQuotaAmount = 2
	future := absoluteRuleForTest("2026-09-04T00:00:00Z", "2026-09-04T02:00:00Z")
	legacy := GroupDynamicRateRule{
		ID:         uuid.NewString(),
		Name:       "legacy",
		Enabled:    true,
		Timezone:   "Asia/Shanghai",
		StartTime:  "09:00",
		EndTime:    "10:00",
		Multiplier: 0.9,
	}

	start, _, quotaKey, ok := parseDynamicRateWindow(active)
	require.True(t, ok)
	require.Equal(t, start.Format(time.RFC3339Nano), quotaKey)
	require.Equal(t, now.Add(-time.Hour), start)
	repo := &dynamicRateSummaryRepository{shared: map[DynamicRateUsageKey]float64{
		{RuleID: active.ID, QuotaKey: quotaKey}: 1.25,
	}}
	groupRepo := &dynamicRateSummaryGroupRepository{group: &Group{
		ID:               "group-1",
		DynamicRateRules: []GroupDynamicRateRule{active, future, legacy},
	}}
	service := NewUserLevelService(repo, nil, groupRepo, nil, nil)

	result, err := service.GetDynamicRateUsageSummary(context.Background(), "group-1", now)
	require.NoError(t, err)
	require.Len(t, result, 3)
	require.Equal(t, DynamicRateStatusActive, result[0].Status)
	require.Equal(t, 1.25, result[0].SharedUsedAmount)
	require.NotNil(t, result[0].SharedRemainingAmount)
	require.Equal(t, 0.75, *result[0].SharedRemainingAmount)
	require.Equal(t, DynamicRateStatusNotStarted, result[1].Status)
	require.Nil(t, result[1].SharedRemainingAmount)
	require.Equal(t, DynamicRateStatusLegacy, result[2].Status)
}
