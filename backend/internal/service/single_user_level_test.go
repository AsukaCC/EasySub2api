//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/pkg/ctxkey"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type singleLevelAssignments struct {
	stubLevelRulesRepo
	assigned []UserLevelRule
	err      error
}

func (r *singleLevelAssignments) GetAssignedLevelRulesBatch(_ context.Context, users []string) (map[string][]UserLevelRule, error) {
	out := make(map[string][]UserLevelRule)
	for _, user := range users {
		out[user] = r.assigned
	}
	return out, r.err
}

type retiredRateMustNotBeRead struct{ UserGroupRateRepository }

func (*retiredRateMustNotBeRead) GetByUserAndGroup(context.Context, string, string) (*float64, error) {
	panic("retired override was read")
}

func TestSingleUserLevelThreeFactorPricing(t *testing.T) {
	now := time.Now()
	rule := UserLevelRule{ID: "level", Enabled: true, WindowDays: 7, Tiers: []UserLevelTier{
		{ID: "base", SortOrder: 0, MinSpend: 0},
		{ID: "gold", SortOrder: 1, MinSpend: 10, DefaultMultiplier: userLevelFloatPtr(.75)},
	}}
	assignments := &singleLevelAssignments{assigned: []UserLevelRule{rule}}
	spend := &offerLevelRepo{spend: 15}
	svc := NewUserLevelService(spend, nil, nil, &retiredRateMustNotBeRead{}, nil)
	svc.rulesRepo = assignments
	discount := offerWindowRule(now.Add(-time.Hour), now.Add(time.Hour))
	discount.DiscountCoefficient = .5
	group := &Group{ID: "group", RateMultiplier: .2, DynamicRateRules: []GroupDynamicRateRule{discount}}
	plan, err := svc.ResolvePlan(context.Background(), "user", group, now)
	require.NoError(t, err)
	require.InDelta(t, .15, plan.NonDynamicMultiplier, 1e-12)
	require.InDelta(t, .075, plan.EffectiveMultiplier, 1e-12)
	require.Equal(t, .75, *plan.UserLevelMultiplier)
	require.Nil(t, plan.UserRateMultiplier)
	// A second request at the same instant must see a rule edit on another instance.
	assignments.assigned[0].Tiers[1].DefaultMultiplier = userLevelFloatPtr(.6)
	updated, err := svc.ResolvePlan(context.Background(), "user", group, now)
	require.NoError(t, err)
	require.InDelta(t, .06, updated.EffectiveMultiplier, 1e-12)
	require.InDelta(t, .075, plan.EffectiveMultiplier, 1e-12, "previous snapshot is immutable")
	spend.spend = 0
	updated, err = svc.ResolvePlan(context.Background(), "user", group, now)
	require.NoError(t, err)
	require.Equal(t, 1.0, *updated.UserLevelMultiplier, "unset base tier means one")
	require.InDelta(t, .1, updated.EffectiveMultiplier, 1e-12)
	group.DynamicRateRules = nil
	updated, err = svc.ResolvePlan(context.Background(), "user", group, now)
	require.NoError(t, err)
	require.InDelta(t, .2, updated.EffectiveMultiplier, 1e-12)
	assignments.err = errors.New("database unavailable")
	_, err = svc.ResolvePlan(context.Background(), "user", group, now)
	require.Error(t, err)
}

func TestSingleUserLevelRejectsMissingMultipleAndDisabledAssignments(t *testing.T) {
	for _, rules := range [][]UserLevelRule{nil, {{Enabled: false}}, {{Enabled: true}, {Enabled: true}}} {
		svc := NewUserLevelService(nil, nil, nil, nil, nil)
		svc.rulesRepo = &singleLevelAssignments{assigned: rules}
		_, err := svc.ResolveProfile(context.Background(), "user", time.Now())
		require.ErrorIs(t, err, ErrUserLevelRulesUnavailable)
	}
}

func TestSingleUserLevelLifecycleProtection(t *testing.T) {
	id := uuid.NewString()
	for _, isDefault := range []bool{true, false} {
		rule := &UserLevelRule{ID: id, Name: "Level", WindowDays: 7, Enabled: true, IsDefault: isDefault, AssignedUserCount: 1}
		repo := &stubLevelRulesRepo{rule: rule}
		svc := newLevelRuleService(repo)
		expected := ErrUserLevelRuleHasMembers
		if isDefault {
			expected = ErrUserLevelDefaultProtected
		}
		require.ErrorIs(t, svc.DeleteLevelRule(context.Background(), id), expected)
		disabled := false
		_, err := svc.UpdateLevelRule(context.Background(), id, UpdateUserLevelRuleInput{Enabled: &disabled})
		require.ErrorIs(t, err, expected)
		require.Nil(t, repo.updated)
		require.Empty(t, repo.deleted)
	}
	svc := newLevelRuleService(&stubLevelRulesRepo{})
	require.Error(t, svc.ReplaceUserLevelRules(context.Background(), uuid.NewString(), []string{id, uuid.NewString()}))
	_, err := svc.NormalizeDefaultLevelRuleIDs(context.Background(), nil)
	require.Error(t, err)
	created, err := svc.CreateLevelRule(context.Background(), CreateUserLevelRuleInput{Name: "New"})
	require.NoError(t, err)
	require.Equal(t, 1.0, *created.Tiers[0].DefaultMultiplier)
}

func TestSingleUserLevelPreservesProductPrecision(t *testing.T) {
	cmd := UsageBillingCommand{DynamicRatePlan: &UsageDynamicRatePlan{FallbackMultiplier: .1234 * .5678}}
	cmd.Normalize()
	require.InDelta(t, .07006652, cmd.DynamicRatePlan.FallbackMultiplier, 1e-12)
}

func TestSingleUserLevelProfitGateUsesSelectedSnapshot(t *testing.T) {
	group := &Group{ID: "group", Hydrated: true, Status: StatusActive, Platform: PlatformOpenAI,
		ProfitControlEnabled: true, RateMultiplier: .2, PeakRateEnabled: true, PeakRateMultiplier: 3}
	ctx := context.WithValue(context.Background(), ctxkey.UserID, "user")
	ctx, _ = WithGatewayTokenRequestPricing(ctx)
	ctx = context.WithValue(ctx, openAIProfitControlGateCtxKey{}, &openAIProfitControlGate{groupID: group.ID, threshold: 1})
	ctx = contextWithUserRatePlan(ctx, group, &UserRatePlan{GroupID: group.ID, EffectiveMultiplier: .075})
	resolved := (&GatewayService{}).withGatewayProfitControlGate(ctx, &group.ID)
	gate, ok := resolved.Value(openAIProfitControlGateCtxKey{}).(*openAIProfitControlGate)
	require.True(t, ok)
	require.NotNil(t, gate)
	require.InDelta(t, .075, gate.threshold, 1e-12)
}

func TestSingleUserLevelMediaUsesCreationSnapshotAndFinalCharge(t *testing.T) {
	usage := &openAIRecordUsageLogRepoStub{inserted: true}
	finalCost, finalRate := .1125, .1125
	billing := &openAIRecordUsageBillingRepoStub{result: &UsageBillingApplyResult{Applied: true, FinalActualCost: &finalCost, FinalRateMultiplier: &finalRate}}
	svc := newOpenAIRecordUsageServiceWithBillingRepoForTest(usage, billing, &openAIRecordUsageUserRepoStub{}, &openAIRecordUsageSubRepoStub{}, &retiredRateMustNotBeRead{})
	// A failing current rule lookup must not affect an already captured plan.
	svc.userLevelService = NewUserLevelService(nil, nil, nil, nil, nil)
	svc.userLevelService.rulesRepo = &singleLevelAssignments{err: errors.New("must not reload")}
	group := &Group{ID: "group", Hydrated: true, Platform: PlatformOpenAI, Status: StatusActive, RateMultiplier: .2,
		AudioRealtimePricePerMin: userLevelFloatPtr(1)}
	plan := &UserRatePlan{GroupID: group.ID, NonDynamicMultiplier: .15, EffectiveBaseMultiplier: .15, PeakMultiplier: 1,
		SelectedDynamicRuleID: "discount", DynamicCandidates: []DynamicRateCandidate{{RuleID: "discount", QuotaKey: "window", DiscountCoefficient: .5, PersonalQuotaAmount: 1}}}
	err := svc.RecordUsage(context.Background(), &OpenAIRecordUsageInput{
		Result: &OpenAIForwardResult{RequestID: "live-snapshot", Model: "gpt-live", AudioUsage: &AudioUsage{Mode: "realtime", DurationOrUnits: 1}},
		APIKey: &APIKey{ID: "key", GroupID: &group.ID, Group: group}, User: &User{ID: "user"}, Account: &Account{ID: "account"},
		RateMultiplierOverride: userLevelFloatPtr(.15), Selection: &AccountSelectionResult{BillingGroup: group, RatePlan: plan},
	})
	require.NoError(t, err)
	require.NotNil(t, billing.lastCmd.DynamicRatePlan)
	require.InDelta(t, .15, billing.lastCmd.DynamicRatePlan.FallbackMultiplier, 1e-12)
	require.Equal(t, finalCost, usage.lastLog.ActualCost)
	require.Equal(t, finalRate, usage.lastLog.RateMultiplier)
}
