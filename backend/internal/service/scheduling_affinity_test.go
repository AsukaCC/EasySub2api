//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/config"
	"github.com/stretchr/testify/require"
)

const (
	affinityStickyID = "11111111-1111-4111-8111-111111111111"
	affinityOtherID  = "22222222-2222-4222-8222-222222222222"
	affinityGroupID  = "33333333-3333-4333-8333-333333333333"
)

func affinityAccounts(platform string, stickyPriority int) []Account {
	now := time.Now()
	return []Account{
		{ID: affinityStickyID, Platform: platform, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true, Concurrency: 10, Priority: stickyPriority, LastUsedAt: &now, GroupIDs: []string{affinityGroupID}},
		{ID: affinityOtherID, Platform: platform, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true, Concurrency: 10, Priority: 0, GroupIDs: []string{affinityGroupID}},
	}
}

func TestGatewaySchedulingBeforeAffinity(t *testing.T) {
	for _, path := range []string{"legacy", "load", "batch_error", "wait", "routed_legacy", "routed_load"} {
		for _, scenario := range []string{"priority", "equal", "load", "reset", "excluded"} {
			if scenario == "load" && path != "load" && path != "routed_load" || scenario == "reset" && path != "load" {
				continue
			}
			t.Run(path+"/"+scenario, func(t *testing.T) {
				accounts := affinityAccounts(PlatformAnthropic, 10)
				want := affinityOtherID
				if scenario != "priority" {
					accounts[0].Priority = 0
				}
				if scenario == "equal" {
					want = affinityStickyID
				}
				repo := &mockAccountRepoForPlatform{accounts: accounts, accountsByID: map[string]*Account{}}
				for i := range repo.accounts {
					repo.accountsByID[repo.accounts[i].ID] = &repo.accounts[i]
				}
				group := &Group{ID: affinityGroupID, Name: "affinity", Platform: PlatformAnthropic, Status: StatusActive, Hydrated: true}
				model := "claude-sonnet-4-5"
				if path == "routed_load" || path == "routed_legacy" {
					group.ModelRoutingEnabled = true
					group.ModelRouting = map[string][]string{model: {affinityStickyID, affinityOtherID}}
				}
				cfg := testConfig()
				cfg.Gateway.Scheduling.LoadBatchEnabled = path != "legacy" && path != "routed_legacy"
				loads := map[string]*AccountLoadInfo{
					affinityStickyID: {AccountID: affinityStickyID, LoadRate: 10},
					affinityOtherID:  {AccountID: affinityOtherID, LoadRate: 10},
				}
				if scenario == "load" {
					loads[affinityStickyID].LoadRate = 80
				}
				if scenario == "reset" {
					soon, later := time.Now().Add(time.Hour), time.Now().Add(5*time.Hour)
					repo.accounts[0].SessionWindowEnd, repo.accounts[1].SessionWindowEnd = &later, &soon
					cfg.Gateway.Scheduling.PreferSoonestReset = true
				}
				var probes []string
				concurrency := schedulerTestConcurrencyCache{loadMap: loads, acquiredIDs: &probes}
				if path == "batch_error" {
					concurrency.loadBatchErr = errors.New("load unavailable")
				}
				if path == "wait" {
					concurrency.acquireResults = map[string]bool{affinityStickyID: false, affinityOtherID: false}
				}
				cache := &mockGatewayCacheForPlatform{sessionBindings: map[string]string{"affinity": affinityStickyID}}
				svc := &GatewayService{accountRepo: repo, groupRepo: &mockGroupRepoForGateway{groups: map[string]*Group{group.ID: group}}, cache: cache, cfg: cfg, concurrencyService: NewConcurrencyService(concurrency)}
				excluded := map[string]struct{}{}
				if scenario == "excluded" {
					excluded[affinityStickyID] = struct{}{}
				}
				result, err := svc.SelectAccountWithLoadAwareness(context.Background(), &group.ID, "affinity", model, excluded, "", "")
				require.NoError(t, err)
				require.Equal(t, want, result.Account.ID)
				require.Equal(t, path == "wait", result.WaitPlan != nil)
				if path != "wait" {
					require.Equal(t, []string{want}, probes, "affinity must not probe a worse candidate first")
				}
				if result.ReleaseFunc != nil {
					result.ReleaseFunc()
				}
			})
		}
	}
}

func TestOpenAILegacySchedulingBeforeAffinity(t *testing.T) {
	for _, path := range []string{"simple", "load", "batch_error", "wait"} {
		for _, scenario := range []string{"priority", "equal", "load", "busy"} {
			if scenario == "load" && path != "load" || scenario == "busy" && path != "load" && path != "batch_error" {
				continue
			}
			t.Run(path+"/"+scenario, func(t *testing.T) {
				accounts := affinityAccounts(PlatformOpenAI, 10)
				want := affinityOtherID
				if scenario != "priority" {
					accounts[0].Priority = 0
				}
				if scenario == "equal" {
					want = affinityStickyID
				}
				cfg := &config.Config{}
				cfg.Gateway.Scheduling.LoadBatchEnabled = true
				loads := map[string]*AccountLoadInfo{
					affinityStickyID: {AccountID: affinityStickyID, LoadRate: 10},
					affinityOtherID:  {AccountID: affinityOtherID, LoadRate: 10},
				}
				if scenario == "load" {
					loads[affinityStickyID].LoadRate = 80
				}
				concurrency := schedulerTestConcurrencyCache{loadMap: loads}
				if path == "batch_error" {
					concurrency.loadBatchErr = errors.New("load unavailable")
				}
				if scenario == "busy" {
					concurrency.acquireResults = map[string]bool{affinityStickyID: false}
				}
				if path == "wait" {
					concurrency.acquireResults = map[string]bool{affinityStickyID: false, affinityOtherID: false}
				}
				cache := &schedulerTestGatewayCache{sessionBindings: map[string]string{"openai:affinity": affinityStickyID}}
				svc := &OpenAIGatewayService{accountRepo: schedulerTestOpenAIAccountRepo{accounts: accounts}, cache: cache, cfg: cfg, concurrencyService: NewConcurrencyService(concurrency)}
				if path == "simple" {
					account, err := svc.SelectAccountForModelWithExclusions(context.Background(), &accounts[0].GroupIDs[0], "affinity", "gpt-5.6-sol", nil)
					require.NoError(t, err)
					require.Equal(t, want, account.ID)
					return
				}
				result, err := svc.SelectAccountWithLoadAwareness(context.Background(), &accounts[0].GroupIDs[0], "affinity", "gpt-5.6-sol", nil)
				require.NoError(t, err)
				require.Equal(t, want, result.Account.ID)
				require.Equal(t, path == "wait", result.WaitPlan != nil)
				if scenario == "busy" {
					require.Equal(t, affinityStickyID, cache.sessionBindings["openai:affinity"], "capacity spillover must keep durable affinity")
				}
				if result.ReleaseFunc != nil {
					result.ReleaseFunc()
				}
			})
		}
	}
}

func TestOpenAIAdvancedSchedulingBeforeAffinity(t *testing.T) {
	for _, weighted := range []bool{false, true} {
		for _, previous := range []bool{false, true} {
			for _, wait := range []bool{false, true} {
				t.Run(testBoolName("weighted", weighted)+"/"+testBoolName("previous", previous)+"/"+testBoolName("wait", wait), func(t *testing.T) {
					accounts := affinityAccounts(PlatformOpenAI, 100)
					cfg := &config.Config{}
					cfg.Gateway.OpenAIWS.LBTopK = 1
					cfg.Gateway.OpenAIWS.SchedulerScoreWeights.Priority = 1
					cfg.Gateway.OpenAIWS.SchedulerScoreWeights.PreviousResponse = 10000
					cfg.Gateway.OpenAIWS.SchedulerScoreWeights.SessionSticky = 10000
					cache := &schedulerTestGatewayCache{sessionBindings: map[string]string{"openai:affinity": affinityStickyID}}
					concurrency := schedulerTestConcurrencyCache{}
					if wait {
						concurrency.acquireResults = map[string]bool{affinityOtherID: false}
					}
					svc := &OpenAIGatewayService{accountRepo: schedulerTestOpenAIAccountRepo{accounts: accounts}, cache: cache, cfg: cfg, concurrencyService: NewConcurrencyService(concurrency)}
					req := OpenAIAccountScheduleRequest{GroupID: &accounts[0].GroupIDs[0], Platform: PlatformOpenAI, RequestedModel: "gpt-5.6-sol", SessionHash: "affinity", StickyAccountID: affinityStickyID, StickyWeighted: weighted}
					if previous {
						req.StickyPreviousAccountID, req.PreviousResponseCanMove = affinityStickyID, true
					}
					result, decision, err := newDefaultOpenAIAccountScheduler(svc, nil).Select(context.Background(), req)
					require.NoError(t, err)
					require.Equal(t, affinityOtherID, result.Account.ID)
					require.Equal(t, wait, result.WaitPlan != nil, "fallback must stay within the scheduled pool")
					require.False(t, decision.StickySessionHit)
					require.False(t, decision.StickyPreviousHit)
					if result.ReleaseFunc != nil {
						result.ReleaseFunc()
					}
				})
			}
		}
	}
}

func testBoolName(name string, value bool) string {
	if value {
		return name + "_on"
	}
	return name + "_off"
}

func TestGeminiCompatSchedulingBeforeAffinity(t *testing.T) {
	for _, samePriority := range []bool{false, true} {
		t.Run(testBoolName("equal", samePriority), func(t *testing.T) {
			accounts := affinityAccounts(PlatformGemini, 10)
			want := affinityOtherID
			if samePriority {
				accounts[0].Priority, want = 0, affinityStickyID
			}
			cache := &mockGatewayCacheForPlatform{sessionBindings: map[string]string{"gemini:affinity": affinityStickyID}}
			svc := &GeminiMessagesCompatService{accountRepo: &mockAccountRepoForPlatform{accounts: accounts}, cache: cache}
			result, err := svc.SelectAccountForModel(context.Background(), nil, "affinity", "gemini-2.5-pro")
			require.NoError(t, err)
			require.Equal(t, want, result.ID)
		})
	}
}

func TestOpenAIBaseSchedulingUnaffectedByAffinity(t *testing.T) {
	for _, factor := range []string{"priority", "load", "queue", "error", "ttft", "reset", "cost"} {
		t.Run(factor, func(t *testing.T) {
			accounts := affinityAccounts(PlatformOpenAI, 0)
			cfg := &config.Config{}
			cfg.Gateway.OpenAIWS.LBTopK = 2
			cfg.Gateway.OpenAIWS.SchedulerScoreWeights = config.GatewayOpenAIWSSchedulerScoreWeights{
				Priority: 1, Load: 1, Queue: 1, ErrorRate: 1, TTFT: 1, PreviousResponse: 10000, SessionSticky: 10000,
			}
			loads := map[string]*AccountLoadInfo{
				affinityStickyID: {AccountID: affinityStickyID},
				affinityOtherID:  {AccountID: affinityOtherID},
			}
			stats := newOpenAIAccountRuntimeStats()
			switch factor {
			case "priority":
				accounts[0].Priority = 10
			case "load":
				loads[affinityStickyID].LoadRate = 80
			case "queue":
				loads[affinityStickyID].WaitingCount = 5
			case "error":
				stats.report(affinityStickyID, false, nil)
			case "ttft":
				slow, fast := 10000, 1000
				stats.report(affinityStickyID, true, &slow)
				stats.report(affinityOtherID, true, &fast)
			case "reset":
				later, soon := time.Now().Add(5*time.Hour), time.Now().Add(time.Hour)
				accounts[0].SessionWindowEnd, accounts[1].SessionWindowEnd = &later, &soon
				cfg.Gateway.OpenAIWS.SchedulerScoreWeights.Reset = 1
			case "cost":
				for i, rate := range []float64{0.8, 0.03} {
					accounts[i].Extra = upstreamCostTestAccount("probe", UpstreamBillingProbeStatusOK, rate, time.Now().Add(-time.Minute), 30*time.Minute).Extra
				}
				cfg.Gateway.OpenAIWS.SchedulerScoreWeights.UpstreamCost = 1
			}
			scheduler := &defaultOpenAIAccountScheduler{service: &OpenAIGatewayService{cfg: cfg}, stats: stats}
			req := OpenAIAccountScheduleRequest{SessionHash: "base-before-affinity", StickyWeighted: true, PreviousResponseCanMove: true, UseUpstreamTokenCost: true}
			pool := []*Account{&accounts[0], &accounts[1]}
			base := scheduler.buildOpenAIAccountLoadPlan(context.Background(), req, pool, loads)
			req.StickyAccountID, req.StickyPreviousAccountID = affinityStickyID, affinityStickyID
			sticky := scheduler.buildOpenAIAccountLoadPlan(context.Background(), req, pool, loads)
			baseScores, stickyScores := openAIPlanScores(base), openAIPlanScores(sticky)
			require.Greater(t, stickyScores[affinityOtherID], stickyScores[affinityStickyID])
			for id, score := range baseScores {
				require.InDelta(t, score, stickyScores[id], 0.000001)
			}
			for i := range base.selectionOrder {
				require.Equal(t, base.selectionOrder[i].account.ID, sticky.selectionOrder[i].account.ID, "affinity must preserve weighted scheduling across unequal tiers")
			}
			cfg.Gateway.OpenAIWS.LBTopK = 1
			sticky = scheduler.buildOpenAIAccountLoadPlan(context.Background(), req, pool, loads)
			require.Equal(t, affinityOtherID, sticky.selectionOrder[0].account.ID, "affinity cannot force admission into Top K")
		})
	}
}

func TestOpenAISubscriptionPoolBeforeAffinity(t *testing.T) {
	for _, weighted := range []bool{false, true} {
		t.Run(testBoolName("weighted", weighted), func(t *testing.T) {
			accounts := affinityAccounts(PlatformOpenAI, 0)
			accounts[1].Type = AccountTypeOAuth
			accounts[1].Priority = 100
			accounts[1].Credentials = map[string]any{"plan_type": "plus"}
			svc := &OpenAIGatewayService{
				accountRepo:        schedulerTestOpenAIAccountRepo{accounts: accounts},
				cache:              &schedulerTestGatewayCache{sessionBindings: map[string]string{"openai:affinity": affinityStickyID}},
				concurrencyService: NewConcurrencyService(schedulerTestConcurrencyCache{}),
			}
			result, _, err := newDefaultOpenAIAccountScheduler(svc, nil).Select(context.Background(), OpenAIAccountScheduleRequest{
				GroupID: &accounts[0].GroupIDs[0], Platform: PlatformOpenAI, RequestedModel: "gpt-5.6-sol", SessionHash: "affinity",
				StickyAccountID: affinityStickyID, StickyWeighted: weighted, SubscriptionPriority: true,
			})
			require.NoError(t, err)
			require.Equal(t, affinityOtherID, result.Account.ID)
			result.ReleaseFunc()
		})
	}
}

func TestOpenAISchedulerScoreReportsAffinityWithinTier(t *testing.T) {
	for _, weighted := range []bool{false, true} {
		scores := buildOpenAIAccountSchedulerScoreSnapshot([]*Account{{ID: affinityStickyID}}, nil,
			GatewayOpenAIWSSchedulerScoreWeightsView{Priority: 10, Previous: 5, SessionSticky: 3}, weighted, nil)
		require.False(t, scores[affinityStickyID].StickyScoreInfinity)
		wantAffinity := 3.0
		if weighted {
			wantAffinity = 8
		}
		require.Equal(t, wantAffinity, scores[affinityStickyID].StickyScore)
		require.Equal(t, 10.0, scores[affinityStickyID].BaseScore)
	}
}
