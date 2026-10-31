//go:build unit

package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/pkg/ctxkey"
	"github.com/stretchr/testify/require"
)

type sessionPriorityTestCache struct {
	GatewayCache
	now              time.Time
	values           map[string]string
	expires          map[string]time.Time
	failPriorityRead bool
}

func newSessionPriorityTestCache() *sessionPriorityTestCache {
	return &sessionPriorityTestCache{now: time.Now(), values: map[string]string{}, expires: map[string]time.Time{}}
}

func (c *sessionPriorityTestCache) GetSessionAccountID(_ context.Context, group, key string) (string, error) {
	if c.failPriorityRead && strings.HasPrefix(key, "scheduling-priority:") {
		return "", errors.New("redis unavailable")
	}
	k := group + "/" + key
	if !c.now.Before(c.expires[k]) {
		return "", ErrStickySessionNotFound
	}
	return c.values[k], nil
}

func (c *sessionPriorityTestCache) SetSessionAccountID(_ context.Context, group, key, id string, ttl time.Duration) error {
	k := group + "/" + key
	c.values[k], c.expires[k] = id, c.now.Add(ttl)
	return nil
}

func (c *sessionPriorityTestCache) DeleteSessionAccountID(_ context.Context, group, key string) error {
	delete(c.values, group+"/"+key)
	delete(c.expires, group+"/"+key)
	return nil
}

func (c *sessionPriorityTestCache) GetAccountKeyActivity(_ context.Context, ids []string, _ string) (map[string]AccountKeyActivity, error) {
	activity := map[string]AccountKeyActivity{}
	for _, id := range ids {
		activity[id] = AccountKeyActivity{}
	}
	activity[affinityStickyID] = AccountKeyActivity{OtherKeys: 2, OwnRequests: 1}
	return activity, nil
}

func TestSessionPrioritySchedulingPaths(t *testing.T) {
	paths := []string{"gateway_legacy", "gateway_load", "gateway_batch_error", "gateway_routed", "openai_legacy", "openai_load", "openai_batch_error", "openai_advanced", "openai_weighted", "gemini"}
	for _, path := range paths {
		for _, scenario := range []string{"active", "same_priority", "expired", "excluded", "disabled", "cache_error", "busy", "wait"} {
			if path == "gemini" && (scenario == "busy" || scenario == "wait") {
				continue
			}
			t.Run(path+"/"+scenario, func(t *testing.T) {
				platform, cacheKey := PlatformOpenAI, "openai:session"
				if strings.HasPrefix(path, "gateway") {
					platform, cacheKey = PlatformAnthropic, "session"
				}
				if path == "gemini" {
					platform, cacheKey = PlatformGemini, "gemini:session"
				}
				accounts := affinityAccounts(platform, 10)
				if scenario == "same_priority" {
					accounts[0].Priority = 0
				}
				if scenario == "disabled" {
					accounts[0].Schedulable = false
				}
				cache := newSessionPriorityTestCache()
				ctx := context.WithValue(context.Background(), ctxkey.APIKeyID, "key")
				require.NoError(t, cache.SetSessionAccountID(ctx, affinityGroupID, cacheKey, affinityStickyID, time.Hour))
				markSessionSchedulingSuccess(ctx, cache, affinityGroupID, cacheKey, affinityStickyID)
				deadlineKey := affinityGroupID + "/" + sessionSchedulingPriorityKey(cacheKey)
				deadline := cache.expires[deadlineKey]
				cache.now = cache.now.Add(29*time.Minute + 59*time.Second)
				if scenario == "expired" {
					cache.now = deadline
				}
				cache.failPriorityRead = scenario == "cache_error"
				excluded := map[string]struct{}{}
				if scenario == "excluded" {
					excluded[affinityStickyID] = struct{}{}
				}
				cfg := testConfig()
				cfg.Gateway.Scheduling.LoadBatchEnabled = !strings.HasSuffix(path, "legacy")
				cfg.Gateway.OpenAIWS.LBTopK = 1
				cfg.Gateway.OpenAIWS.SchedulerScoreWeights.Load = 100
				concurrency := schedulerTestConcurrencyCache{loadMap: map[string]*AccountLoadInfo{affinityStickyID: {LoadRate: 90}, affinityOtherID: {LoadRate: 0}}}
				if strings.HasSuffix(path, "batch_error") {
					concurrency.loadBatchErr = errors.New("unavailable")
				}
				if scenario == "busy" || scenario == "wait" {
					concurrency.acquireResults = map[string]bool{affinityStickyID: false}
				}
				if scenario == "wait" {
					concurrency.acquireResults[affinityOtherID] = false
				}
				group := &Group{ID: affinityGroupID, Platform: platform, Status: StatusActive, Hydrated: true}
				groupRepo := &mockGroupRepoForGateway{groups: map[string]*Group{group.ID: group}}
				var result *AccountSelectionResult
				var err error
				if platform == PlatformAnthropic {
					repo := &mockAccountRepoForPlatform{accounts: accounts, accountsByID: map[string]*Account{}}
					for i := range repo.accounts {
						repo.accountsByID[repo.accounts[i].ID] = &repo.accounts[i]
					}
					if path == "gateway_routed" {
						group.ModelRoutingEnabled = true
						group.ModelRouting = map[string][]string{"claude-sonnet-4-5": {affinityStickyID, affinityOtherID}}
					}
					svc := &GatewayService{accountRepo: repo, groupRepo: groupRepo, cache: cache, cfg: cfg, concurrencyService: NewConcurrencyService(concurrency)}
					result, err = svc.SelectAccountWithLoadAwareness(ctx, &group.ID, "session", "claude-sonnet-4-5", excluded, "", "")
				} else if path == "gemini" {
					svc := &GeminiMessagesCompatService{accountRepo: &mockAccountRepoForPlatform{accounts: accounts}, groupRepo: groupRepo, cache: cache}
					var account *Account
					account, err = svc.SelectAccountForModelWithExclusions(ctx, &group.ID, "session", "gemini-2.5-pro", excluded)
					result = &AccountSelectionResult{Account: account, Acquired: true}
				} else {
					svc := &OpenAIGatewayService{accountRepo: schedulerTestOpenAIAccountRepo{accounts: accounts}, cache: cache, cfg: cfg, concurrencyService: NewConcurrencyService(concurrency)}
					if path == "openai_advanced" || path == "openai_weighted" {
						result, _, err = newDefaultOpenAIAccountScheduler(svc, nil).Select(ctx, OpenAIAccountScheduleRequest{GroupID: &group.ID, Platform: platform, SessionHash: "session", RequestedModel: "gpt-5.6-sol", ExcludedIDs: excluded, StickyWeighted: path == "openai_weighted"})
					} else {
						result, err = svc.SelectAccountWithLoadAwareness(ctx, &group.ID, "session", "gpt-5.6-sol", excluded)
					}
				}
				require.NoError(t, err)
				require.NotNil(t, result.Account)
				if result.ReleaseFunc != nil {
					defer result.ReleaseFunc()
				}
				if scenario == "wait" {
					require.False(t, result.Acquired)
					require.NotNil(t, result.WaitPlan)
					require.Equal(t, deadline, cache.expires[deadlineKey], "waiting must not renew the window")
					return
				}
				want := affinityOtherID
				if scenario == "active" || scenario == "same_priority" {
					want = affinityStickyID
				}
				require.Equal(t, want, result.Account.ID)
				if want == affinityStickyID {
					require.Equal(t, cache.now.Add(30*time.Minute), cache.expires[deadlineKey], "successful scheduling renews the window")
				}
			})
		}
	}
}

func TestSessionPrioritySlidingWindowAndIsolation(t *testing.T) {
	cache := newSessionPriorityTestCache()
	ctx := context.Background()
	group := affinityGroupID
	key := "openai:session"
	require.NoError(t, cache.SetSessionAccountID(ctx, group, key, affinityStickyID, time.Hour))
	markSessionSchedulingSuccess(ctx, cache, group, key, affinityStickyID)
	cache.now = cache.now.Add(20 * time.Minute)
	active := withSessionSchedulingAffinity(ctx, cache, &group, key, affinityStickyID)
	renewSessionSchedulingAffinity(active, affinityStickyID)
	cache.now = cache.now.Add(29*time.Minute + 59*time.Second)
	require.Equal(t, affinityStickyID, sessionSchedulingPriorityID(withSessionSchedulingAffinity(ctx, cache, &group, key, affinityStickyID)))
	require.Empty(t, sessionSchedulingPriorityID(withSessionSchedulingAffinity(ctx, cache, nil, key, affinityStickyID)))
	require.Empty(t, sessionSchedulingPriorityID(withSessionSchedulingAffinity(ctx, cache, &group, "openai:other", affinityStickyID)))
	require.Empty(t, sessionSchedulingPriorityID(withSessionSchedulingAffinity(ctx, cache, &group, key, affinityOtherID)))
	cache.now = cache.now.Add(time.Second)
	require.Empty(t, sessionSchedulingPriorityID(withSessionSchedulingAffinity(ctx, cache, &group, key, affinityStickyID)))
	bound, err := cache.GetSessionAccountID(ctx, group, key)
	require.NoError(t, err)
	require.Equal(t, affinityStickyID, bound, "protocol binding remains after scheduling priority expires")
}

func TestSessionPriorityTerminalAdmission(t *testing.T) {
	for _, platform := range []string{PlatformAnthropic, PlatformOpenAI} {
		for _, gated := range []bool{false, true} {
			t.Run(platform+"/"+testBoolName("profit_gate", gated), func(t *testing.T) {
				cache := newSessionPriorityTestCache()
				ctx := context.Background()
				group, key := affinityGroupID, "session"
				if platform == PlatformOpenAI {
					key = "openai:" + key
				}
				if gated {
					ctx = context.WithValue(ctx, openAIProfitControlGateCtxKey{}, &openAIProfitControlGate{groupID: group, platform: platform, threshold: 0.5})
				}
				var bind, eager func(context.Context, *string, string, string) error
				if platform == PlatformOpenAI {
					cfg := testConfig()
					cfg.Gateway.OpenAIWS.StickySessionTTLSeconds = 60
					svc := &OpenAIGatewayService{cache: cache, cfg: cfg}
					bind, eager = svc.BindStickySessionAfterProfitAdmission, svc.bindOpenAIStickySessionDuringSelection
				} else {
					svc := &GatewayService{cache: cache}
					bind, eager = svc.BindStickySessionAfterProfitAdmission, svc.bindGatewayStickySessionDuringSelection
				}
				require.NoError(t, eager(ctx, &group, "session", affinityStickyID))
				_, err := cache.GetSessionAccountID(ctx, group, sessionSchedulingPriorityKey(key))
				require.ErrorIs(t, err, ErrStickySessionNotFound, "speculative binding must not start the priority window")
				require.NoError(t, bind(ctx, &group, "session", affinityStickyID))
				deadlineKey := group + "/" + sessionSchedulingPriorityKey(key)
				require.Equal(t, cache.now.Add(30*time.Minute), cache.expires[deadlineKey])
				cache.now = cache.now.Add(20 * time.Minute)
				bound, bindErr := cache.GetSessionAccountID(ctx, group, key)
				require.NoError(t, bindErr)
				require.Equal(t, affinityStickyID, bound)
				active := withSessionSchedulingAffinity(ctx, cache, &group, key, affinityStickyID)
				require.Equal(t, affinityStickyID, sessionSchedulingPriorityID(active), "short configured binding TTL must not truncate the priority window")
				if gated {
					deadline := cache.expires[deadlineKey]
					renewSessionSchedulingAffinity(active, affinityStickyID)
					require.Equal(t, deadline, cache.expires[deadlineKey], "profit admission must happen before renewal")
					require.NoError(t, bind(ctx, &group, "session", affinityOtherID))
					require.Equal(t, deadline, cache.expires[deadlineKey], "fallback must not renew the preserved original binding")
				}
				require.NoError(t, bind(ctx, &group, "session", affinityStickyID))
				require.Equal(t, cache.now.Add(30*time.Minute), cache.expires[deadlineKey])
			})
		}
	}
}

func TestSessionPriorityPreservesCompactCapabilityTiers(t *testing.T) {
	for _, path := range []string{"simple", "load", "advanced"} {
		for _, unsupported := range []bool{false, true} {
			t.Run(path+"/"+testBoolName("unsupported", unsupported), func(t *testing.T) {
				accounts := affinityAccounts(PlatformOpenAI, 0)
				accounts[1].Extra = map[string]any{"openai_compact_supported": true}
				if unsupported {
					accounts[0].Extra = map[string]any{"openai_compact_supported": false}
				}
				ctx := context.Background()
				group := affinityGroupID
				cache := newSessionPriorityTestCache()
				require.NoError(t, cache.SetSessionAccountID(ctx, group, "openai:session", affinityStickyID, time.Hour))
				markSessionSchedulingSuccess(ctx, cache, group, "openai:session", affinityStickyID)
				cfg := testConfig()
				cfg.Gateway.Scheduling.LoadBatchEnabled = true
				cfg.Gateway.OpenAIWS.LBTopK = 1
				svc := &OpenAIGatewayService{accountRepo: schedulerTestOpenAIAccountRepo{accounts: accounts}, cache: cache, cfg: cfg, concurrencyService: NewConcurrencyService(schedulerTestConcurrencyCache{})}
				var result *AccountSelectionResult
				var err error
				if path == "simple" {
					var account *Account
					account, err = svc.selectAccountForModelWithExclusions(ctx, &group, PlatformOpenAI, "session", "gpt-5.6-sol", nil, true, "", "", false)
					result = &AccountSelectionResult{Account: account}
				} else if path == "load" {
					result, err = svc.selectAccountWithLoadAwareness(ctx, &group, PlatformOpenAI, "session", "gpt-5.6-sol", nil, true, "", false)
				} else {
					result, _, err = newDefaultOpenAIAccountScheduler(svc, nil).Select(ctx, OpenAIAccountScheduleRequest{GroupID: &group, Platform: PlatformOpenAI, SessionHash: "session", RequestedModel: "gpt-5.6-sol", RequireCompact: true})
				}
				require.NoError(t, err)
				if result.ReleaseFunc != nil {
					defer result.ReleaseFunc()
				}
				require.Equal(t, affinityOtherID, result.Account.ID)
			})
		}
	}
}

type sessionPriorityLeaseCache struct {
	schedulerTestConcurrencyCache
	expected []int
	peers    [][]AccountKeyPeer
}

func (c *sessionPriorityLeaseCache) AcquireAccountKeyLease(_ context.Context, _, _, _ string, expected int, peers []AccountKeyPeer) (bool, error) {
	c.expected = append(c.expected, expected)
	c.peers = append(c.peers, peers)
	return true, nil
}
func (*sessionPriorityLeaseCache) RefreshAccountKeyLease(context.Context, string, string, string) error {
	return nil
}
func (*sessionPriorityLeaseCache) ReleaseAccountKeyLease(context.Context, string, string, string) error {
	return nil
}

func TestSessionPriorityLeaseKeepsIndependentConcurrency(t *testing.T) {
	cache := &sessionPriorityLeaseCache{}
	ctx := context.WithValue(context.Background(), ctxkey.APIKeyID, "key")
	ctx = context.WithValue(ctx, keySchedulingContextKey{}, &keySchedulingSnapshot{activity: map[string]AccountKeyActivity{"sticky": {OtherKeys: 2}}, peers: map[string][]AccountKeyPeer{"sticky": {{ID: "empty", MaxConcurrency: 10}}}})
	ctx = context.WithValue(ctx, sessionSchedulingAffinityKey{}, &sessionSchedulingAffinity{priorityID: "sticky"})
	for i := 0; i < 2; i++ {
		result, err := NewConcurrencyService(cache).AcquireAccountSlot(ctx, "sticky", 10)
		require.NoError(t, err)
		require.True(t, result.Acquired)
		defer result.ReleaseFunc()
	}
	require.Equal(t, []int{-1, -1}, cache.expected)
	for _, peers := range cache.peers {
		require.Empty(t, peers)
	}
	cache.acquireResults = map[string]bool{"sticky": false}
	result, err := NewConcurrencyService(cache).AcquireAccountSlot(ctx, "sticky", 10)
	require.NoError(t, err)
	require.False(t, result.Acquired, "active affinity cannot bypass the account concurrency limit")
	require.Len(t, cache.expected, 2, "every admitted request has a separate lease")
}
