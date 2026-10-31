//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/config"
	"github.com/AsukaCC/EasySub2api/internal/pkg/ctxkey"
	"github.com/stretchr/testify/require"
)

type keySchedulingTestCache struct {
	GatewayCache
	activity map[string]AccountKeyActivity
}

func (c *keySchedulingTestCache) GetAccountKeyActivity(_ context.Context, ids []string, _ string) (map[string]AccountKeyActivity, error) {
	result := make(map[string]AccountKeyActivity, len(ids))
	for _, id := range ids {
		result[id] = c.activity[id]
	}
	return result, nil
}

func TestKeySchedulingOrder(t *testing.T) {
	now := time.Now()
	recent, idle := now.Add(-time.Minute), now.Add(-10*time.Minute)
	for _, tc := range []struct {
		name         string
		a, b         AccountKeyActivity
		lastA, lastB *time.Time
		priorityB    int
		want         string
	}{
		{"empty_before_shared", AccountKeyActivity{OtherKeys: 1}, AccountKeyActivity{}, &recent, &recent, 0, "b"},
		{"second_round", AccountKeyActivity{OtherKeys: 2}, AccountKeyActivity{OtherKeys: 1}, &recent, &recent, 0, "b"},
		{"own_key_before_empty", AccountKeyActivity{}, AccountKeyActivity{OwnRequests: 3}, &recent, &recent, 0, "b"},
		{"empty_before_shared_own", AccountKeyActivity{OtherKeys: 1, OwnRequests: 1}, AccountKeyActivity{}, &recent, &recent, 0, "b"},
		{"idle_compensation", AccountKeyActivity{}, AccountKeyActivity{}, &recent, &idle, 0, "b"},
		{"recent_selection_clears_idle", AccountKeyActivity{}, AccountKeyActivity{LastScheduledAt: now}, &recent, &idle, 0, "a"},
		{"same_tier_keeps_factors_and_affinity", AccountKeyActivity{}, AccountKeyActivity{}, &recent, &recent, 0, "a"},
		{"priority_is_preserved", AccountKeyActivity{OtherKeys: 2}, AccountKeyActivity{}, &recent, &idle, 1, "a"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			accounts := []*Account{{ID: "a", LastUsedAt: tc.lastA}, {ID: "b", Priority: tc.priorityB, LastUsedAt: tc.lastB}}
			ctx := context.WithValue(context.Background(), keySchedulingContextKey{}, &keySchedulingSnapshot{now: now, activity: map[string]AccountKeyActivity{"a": tc.a, "b": tc.b}})
			applyKeySchedulingOrder(ctx, accounts, schedulingAccount, sameAccountPriority)
			require.Equal(t, tc.want, accounts[0].ID)
		})
	}
}

func TestKeySchedulingGatewayPaths(t *testing.T) {
	for _, path := range []string{"legacy", "legacy_busy", "legacy_sticky_busy", "legacy_wait", "load", "batch_error", "wait", "routed_legacy", "routed_load"} {
		t.Run(path, func(t *testing.T) {
			accounts := affinityAccounts(PlatformAnthropic, 0)
			repo := &mockAccountRepoForPlatform{accounts: accounts, accountsByID: map[string]*Account{}}
			for i := range accounts {
				repo.accountsByID[accounts[i].ID] = &repo.accounts[i]
			}
			group := &Group{ID: affinityGroupID, Platform: PlatformAnthropic, Status: StatusActive, Hydrated: true}
			model := "claude-sonnet-4-5"
			if path == "routed_legacy" || path == "routed_load" {
				group.ModelRoutingEnabled = true
				group.ModelRouting = map[string][]string{model: {affinityStickyID, affinityOtherID}}
			}
			cfg := testConfig()
			cfg.Gateway.Scheduling.LoadBatchEnabled = path != "legacy" && path != "routed_legacy" && path != "legacy_busy" && path != "legacy_wait" && path != "legacy_sticky_busy"
			concurrency := schedulerTestConcurrencyCache{loadMap: map[string]*AccountLoadInfo{affinityStickyID: {LoadRate: 1}, affinityOtherID: {LoadRate: 40}}}
			if path == "batch_error" {
				concurrency.loadBatchErr = errors.New("unavailable")
			}
			if path == "wait" || path == "legacy_wait" {
				concurrency.acquireResults = map[string]bool{affinityStickyID: false, affinityOtherID: false}
			}
			if path == "legacy_busy" {
				concurrency.acquireResults = map[string]bool{affinityOtherID: false}
			}
			cache := &keySchedulingTestCache{GatewayCache: &mockGatewayCacheForPlatform{sessionBindings: map[string]string{"affinity": affinityStickyID}}, activity: map[string]AccountKeyActivity{affinityStickyID: {OtherKeys: 1}}}
			if path == "legacy_sticky_busy" {
				cache.activity[affinityStickyID] = AccountKeyActivity{OwnRequests: 1}
				concurrency.acquireResults = map[string]bool{affinityStickyID: false}
			}
			svc := &GatewayService{accountRepo: repo, groupRepo: &mockGroupRepoForGateway{groups: map[string]*Group{group.ID: group}}, cache: cache, cfg: cfg, concurrencyService: NewConcurrencyService(concurrency)}
			ctx := context.WithValue(context.Background(), ctxkey.APIKeyID, "key")
			result, err := svc.SelectAccountWithLoadAwareness(ctx, &group.ID, "affinity", model, nil, "", "")
			require.NoError(t, err)
			want := affinityOtherID
			if path == "legacy_busy" {
				want = affinityStickyID
			}
			require.Equal(t, want, result.Account.ID)
			if path == "legacy_sticky_busy" {
				binding, bindErr := cache.GetSessionAccountID(ctx, group.ID, "affinity")
				require.NoError(t, bindErr)
				require.Equal(t, affinityStickyID, binding)
			}
			require.Equal(t, path == "wait" || path == "legacy_wait", result.WaitPlan != nil)
			if result.ReleaseFunc != nil {
				result.ReleaseFunc()
			}
		})
	}
}

func TestKeySchedulingAdvancedTopKAndAffinity(t *testing.T) {
	for _, topK := range []int{1, 2} {
		accounts := affinityAccounts(PlatformOpenAI, 0)
		cfg := &config.Config{}
		cfg.Gateway.OpenAIWS.LBTopK = topK
		cfg.Gateway.OpenAIWS.SchedulerScoreWeights = config.GatewayOpenAIWSSchedulerScoreWeights{Load: 100, SessionSticky: 10000, PreviousResponse: 10000}
		scheduler := &defaultOpenAIAccountScheduler{service: &OpenAIGatewayService{cfg: cfg}}
		ctx := context.WithValue(context.Background(), keySchedulingContextKey{}, &keySchedulingSnapshot{now: time.Now(), activity: map[string]AccountKeyActivity{affinityStickyID: {OtherKeys: 1}, affinityOtherID: {}}})
		loads := map[string]*AccountLoadInfo{affinityStickyID: {LoadRate: 1}, affinityOtherID: {LoadRate: 80}}
		plan := scheduler.buildOpenAIAccountLoadPlan(ctx, OpenAIAccountScheduleRequest{StickyAccountID: affinityStickyID, StickyPreviousAccountID: affinityStickyID, PreviousResponseCanMove: true, StickyWeighted: true}, []*Account{&accounts[0], &accounts[1]}, loads)
		require.Equal(t, affinityOtherID, plan.selectionOrder[0].account.ID)
		require.Len(t, plan.selectionOrder, 2, "retain alternatives when a concurrent key wins the first account")
	}
}

func TestKeySchedulingOpenAIAndGeminiPaths(t *testing.T) {
	for _, path := range []string{"simple", "legacy", "legacy_busy", "legacy_sticky_busy", "legacy_wait", "load", "batch_error", "wait", "advanced", "gemini"} {
		t.Run(path, func(t *testing.T) {
			accounts := affinityAccounts(PlatformOpenAI, 0)
			baseCache := &schedulerTestGatewayCache{sessionBindings: map[string]string{"openai:affinity": affinityStickyID}}
			cache := &keySchedulingTestCache{GatewayCache: baseCache, activity: map[string]AccountKeyActivity{affinityStickyID: {OtherKeys: 1}}}
			ctx := context.WithValue(context.Background(), ctxkey.APIKeyID, "key")
			if path == "gemini" {
				accounts = affinityAccounts(PlatformGemini, 0)
				cache.GatewayCache = &mockGatewayCacheForPlatform{sessionBindings: map[string]string{"gemini:affinity": affinityStickyID}}
				svc := &GeminiMessagesCompatService{accountRepo: &mockAccountRepoForPlatform{accounts: accounts}, cache: cache}
				result, err := svc.SelectAccountForModel(ctx, nil, "affinity", "gemini-2.5-pro")
				require.NoError(t, err)
				require.Equal(t, affinityOtherID, result.ID)
				return
			}
			cfg := &config.Config{}
			cfg.Gateway.Scheduling.LoadBatchEnabled = path != "legacy" && path != "legacy_busy" && path != "legacy_wait" && path != "legacy_sticky_busy"
			cfg.Gateway.OpenAIWS.LBTopK = 1
			concurrency := schedulerTestConcurrencyCache{loadMap: map[string]*AccountLoadInfo{affinityStickyID: {LoadRate: 1}, affinityOtherID: {LoadRate: 40}}}
			if path == "batch_error" {
				concurrency.loadBatchErr = errors.New("unavailable")
			}
			if path == "wait" || path == "legacy_wait" {
				concurrency.acquireResults = map[string]bool{affinityStickyID: false, affinityOtherID: false}
			}
			if path == "legacy_busy" {
				concurrency.acquireResults = map[string]bool{affinityOtherID: false}
			}
			if path == "legacy_sticky_busy" {
				cache.activity[affinityStickyID] = AccountKeyActivity{OwnRequests: 1}
				concurrency.acquireResults = map[string]bool{affinityStickyID: false}
			}
			svc := &OpenAIGatewayService{accountRepo: schedulerTestOpenAIAccountRepo{accounts: accounts}, cache: cache, cfg: cfg, concurrencyService: NewConcurrencyService(concurrency)}
			if path == "simple" {
				account, err := svc.SelectAccountForModelWithExclusions(ctx, &accounts[0].GroupIDs[0], "affinity", "gpt-5.6-sol", nil)
				require.NoError(t, err)
				require.Equal(t, affinityOtherID, account.ID)
				return
			}
			var result *AccountSelectionResult
			var err error
			if path == "advanced" {
				result, _, err = newDefaultOpenAIAccountScheduler(svc, nil).Select(ctx, OpenAIAccountScheduleRequest{GroupID: &accounts[0].GroupIDs[0], Platform: PlatformOpenAI, RequestedModel: "gpt-5.6-sol", SessionHash: "affinity", StickyAccountID: affinityStickyID})
			} else {
				result, err = svc.SelectAccountWithLoadAwareness(ctx, &accounts[0].GroupIDs[0], "affinity", "gpt-5.6-sol", nil)
			}
			require.NoError(t, err)
			want := affinityOtherID
			if path == "legacy_busy" {
				want = affinityStickyID
			}
			require.Equal(t, want, result.Account.ID)
			if path == "legacy_sticky_busy" {
				require.Equal(t, affinityStickyID, baseCache.sessionBindings["openai:affinity"])
			}
			require.Equal(t, path == "wait" || path == "legacy_wait", result.WaitPlan != nil)
			if result.ReleaseFunc != nil {
				result.ReleaseFunc()
			}
		})
	}
}

type keySchedulingRejectLeaseCache struct {
	schedulerTestConcurrencyCache
	err error
}

func (c keySchedulingRejectLeaseCache) AcquireAccountKeyLease(context.Context, string, string, string, int, []AccountKeyPeer) (bool, error) {
	return false, c.err
}
func (c keySchedulingRejectLeaseCache) RefreshAccountKeyLease(context.Context, string, string, string) error {
	return nil
}
func (c keySchedulingRejectLeaseCache) ReleaseAccountKeyLease(context.Context, string, string, string) error {
	return nil
}

func TestKeySchedulingRejectedClaimReleasesSlot(t *testing.T) {
	for _, leaseErr := range []error{nil, errors.New("redis unavailable")} {
		var released []string
		cache := keySchedulingRejectLeaseCache{schedulerTestConcurrencyCache: schedulerTestConcurrencyCache{releasedIDs: &released}, err: leaseErr}
		ctx := context.WithValue(context.Background(), ctxkey.APIKeyID, "key")
		result, err := NewConcurrencyService(cache).AcquireAccountSlot(ctx, "account", 10)
		require.ErrorIs(t, err, leaseErr)
		require.False(t, result.Acquired)
		require.Equal(t, []string{"account"}, released)
	}
}
