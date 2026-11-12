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

func TestAccountOpenAIImageSizeTiers(t *testing.T) {
	cases := []struct {
		name  string
		extra map[string]any
		want  []string
	}{
		{name: "missing", extra: nil, want: nil},
		{name: "empty", extra: map[string]any{AccountExtraOpenAIImageSizeTiers: []any{}}, want: nil},
		{name: "all tiers", extra: map[string]any{AccountExtraOpenAIImageSizeTiers: []any{"4K", "1K", "2K"}}, want: nil},
		{name: "unknown only", extra: map[string]any{AccountExtraOpenAIImageSizeTiers: []any{"8K", 4}}, want: nil},
		{name: "not a list", extra: map[string]any{AccountExtraOpenAIImageSizeTiers: "4K"}, want: nil},
		{name: "ordered and case-insensitive", extra: map[string]any{AccountExtraOpenAIImageSizeTiers: []any{" 2k", "1K", "2K"}}, want: []string{"1K", "2K"}},
		{name: "string slice", extra: map[string]any{AccountExtraOpenAIImageSizeTiers: []string{"4K"}}, want: []string{"4K"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			account := &Account{Platform: PlatformOpenAI, Extra: tc.extra}
			require.Equal(t, tc.want, account.OpenAIImageSizeTiers())
		})
	}

	lowOnly := &Account{Extra: map[string]any{AccountExtraOpenAIImageSizeTiers: []any{"1K", "2K"}}}
	require.True(t, lowOnly.SupportsOpenAIImageSizeTier("2K"))
	require.False(t, lowOnly.SupportsOpenAIImageSizeTier("4K"))
	require.True(t, (&Account{}).SupportsOpenAIImageSizeTier("4K"))
}

func TestAccountOpenAIImageTestSize(t *testing.T) {
	require.Empty(t, (&Account{}).OpenAIImageTestSize())
	require.Equal(t, "1024x1024", (&Account{Extra: map[string]any{AccountExtraOpenAIImageSizeTiers: []any{"2K", "1K"}}}).OpenAIImageTestSize())
	require.Equal(t, "1536x1024", (&Account{Extra: map[string]any{AccountExtraOpenAIImageSizeTiers: []any{"2K", "4K"}}}).OpenAIImageTestSize())
	require.Equal(t, "3840x2160", (&Account{Extra: map[string]any{AccountExtraOpenAIImageSizeTiers: []any{"4K"}}}).OpenAIImageTestSize())
}

func TestOpenAIResponsesImageRequestSize(t *testing.T) {
	require.Equal(t, "3840x2160", OpenAIResponsesImageRequestSize([]byte(`{"model":"gpt-5.6-sol","tools":[{"type":"web_search"},{"type":"image_generation","size":"3840x2160"}]}`)))
	require.Equal(t, "1024x1024", OpenAIResponsesImageRequestSize([]byte(`{"model":"gpt-5.6-sol","size":"1024x1024","tools":[{"type":"image_generation"}]}`)))
	require.Empty(t, OpenAIResponsesImageRequestSize([]byte(`{"model":"gpt-5.6-sol","tools":[{"type":"image_generation"}]}`)))
}

func newOpenAIImageSizeTestAccount(id string, priority int, tiers ...string) Account {
	extra := map[string]any{}
	if len(tiers) > 0 {
		values := make([]any, 0, len(tiers))
		for _, tier := range tiers {
			values = append(values, tier)
		}
		extra[AccountExtraOpenAIImageSizeTiers] = values
	}
	return Account{
		ID:          id,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 1,
		Priority:    priority,
		Extra:       extra,
	}
}

type openAIImageSizeSchedulerMode struct {
	name      string
	advanced  bool
	loadBatch bool
}

var openAIImageSizeSchedulerModes = []openAIImageSizeSchedulerMode{
	{name: "advanced", advanced: true, loadBatch: true},
	{name: "legacy_load_batch", loadBatch: true},
	{name: "legacy_sequential"},
}

func newOpenAIImageSizeSchedulerTestService(mode openAIImageSizeSchedulerMode, accounts []Account, cache *schedulerTestGatewayCache, concurrency schedulerTestConcurrencyCache) *OpenAIGatewayService {
	cfg := &config.Config{}
	cfg.Gateway.Scheduling.LoadBatchEnabled = mode.loadBatch
	cfg.Gateway.Scheduling.StickySessionMaxWaiting = 3
	cfg.Gateway.Scheduling.FallbackMaxWaiting = 3
	if cache == nil {
		cache = &schedulerTestGatewayCache{}
	}
	for key, id := range cache.sessionBindings {
		cache.sessionBindings[sessionSchedulingPriorityKey(key)] = id
	}
	svc := &OpenAIGatewayService{
		accountRepo:        schedulerTestOpenAIAccountRepo{accounts: accounts},
		cache:              cache,
		cfg:                cfg,
		concurrencyService: NewConcurrencyService(concurrency),
	}
	if mode.advanced {
		svc.rateLimitService = newOpenAIAdvancedSchedulerRateLimitService("true")
	} else {
		svc.rateLimitService = newOpenAIAdvancedSchedulerRateLimitService("false")
	}
	return svc
}

func selectOpenAIImageSizeTestAccount(t *testing.T, svc *OpenAIGatewayService, groupID string, sessionHash string, size string) (*AccountSelectionResult, error) {
	t.Helper()
	ctx := WithOpenAIImageSizeRouting(context.Background(), size)
	selection, _, err := svc.SelectAccountWithSchedulerForImages(ctx, &groupID, sessionHash, "gpt-image-2", nil, OpenAIImagesCapabilityNative)
	return selection, err
}

func TestSelectAccountWithSchedulerForImages_ExplicitSizeFiltersByTier(t *testing.T) {
	for _, mode := range openAIImageSizeSchedulerModes {
		t.Run(mode.name, func(t *testing.T) {
			groupID := "00000000-0000-0000-0000-000000093001"
			accounts := []Account{
				newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000081001", 0, "1K", "2K"),
				newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000081002", 5, "4K"),
			}
			svc := newOpenAIImageSizeSchedulerTestService(mode, accounts, nil, schedulerTestConcurrencyCache{})

			selection, err := selectOpenAIImageSizeTestAccount(t, svc, groupID, "", "3840x2160")
			require.NoError(t, err)
			require.Equal(t, "00000000-0000-0000-0000-000000081002", selection.Account.ID)

			selection, err = selectOpenAIImageSizeTestAccount(t, svc, groupID, "", "1536x1024")
			require.NoError(t, err)
			require.Equal(t, "00000000-0000-0000-0000-000000081001", selection.Account.ID)

			selection, err = selectOpenAIImageSizeTestAccount(t, svc, groupID, "", "4K")
			require.NoError(t, err)
			require.Equal(t, "00000000-0000-0000-0000-000000081002", selection.Account.ID)
		})
	}
}

func TestSelectAccountWithSchedulerForImages_ExplicitSizeWithoutTierAccountFails(t *testing.T) {
	for _, mode := range openAIImageSizeSchedulerModes {
		t.Run(mode.name, func(t *testing.T) {
			groupID := "00000000-0000-0000-0000-000000093002"
			accounts := []Account{
				newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000081011", 0, "1K", "2K"),
				newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000081012", 0, "2K"),
			}
			svc := newOpenAIImageSizeSchedulerTestService(mode, accounts, nil, schedulerTestConcurrencyCache{})

			selection, err := selectOpenAIImageSizeTestAccount(t, svc, groupID, "", "2880x2880")
			require.Error(t, err)
			require.True(t, errors.Is(err, ErrNoAvailableAccounts))
			require.Nil(t, selection)
		})
	}
}

func TestSelectAccountWithSchedulerForImages_UnspecifiedSizePreferenceOrder(t *testing.T) {
	cases := []struct {
		name     string
		accounts []Account
		want     string
	}{
		{
			name: "unrestricted first",
			accounts: []Account{
				newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000081021", 0, "4K"),
				newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000081022", 0, "1K", "2K"),
				newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000081023", 9),
			},
			want: "00000000-0000-0000-0000-000000081023",
		},
		{
			name: "then 4K",
			accounts: []Account{
				newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000081031", 0, "1K", "2K"),
				newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000081032", 9, "2K", "4K"),
			},
			want: "00000000-0000-0000-0000-000000081032",
		},
		{
			name: "then 2K",
			accounts: []Account{
				newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000081041", 0, "1K"),
				newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000081042", 9, "2K"),
			},
			want: "00000000-0000-0000-0000-000000081042",
		},
		{
			name: "then 1K",
			accounts: []Account{
				newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000081051", 0, "1K"),
			},
			want: "00000000-0000-0000-0000-000000081051",
		},
	}
	for _, mode := range openAIImageSizeSchedulerModes {
		for _, tc := range cases {
			t.Run(mode.name+"/"+tc.name, func(t *testing.T) {
				svc := newOpenAIImageSizeSchedulerTestService(mode, tc.accounts, nil, schedulerTestConcurrencyCache{})
				for _, size := range []string{"", "auto", "3:2"} {
					selection, err := selectOpenAIImageSizeTestAccount(t, svc, "00000000-0000-0000-0000-000000093003", "", size)
					require.NoError(t, err, size)
					require.Equal(t, tc.want, selection.Account.ID, size)
				}
			})
		}
	}
}

func TestSelectAccountWithSchedulerForImages_UnspecifiedSizeSkipsBusyPreferredAccount(t *testing.T) {
	for _, mode := range openAIImageSizeSchedulerModes {
		t.Run(mode.name, func(t *testing.T) {
			accounts := []Account{
				newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000081061", 0),
				newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000081062", 9, "4K"),
			}
			concurrency := schedulerTestConcurrencyCache{acquireResults: map[string]bool{"00000000-0000-0000-0000-000000081061": false}}
			svc := newOpenAIImageSizeSchedulerTestService(mode, accounts, nil, concurrency)

			selection, err := selectOpenAIImageSizeTestAccount(t, svc, "00000000-0000-0000-0000-000000093004", "", "")
			require.NoError(t, err)
			require.Equal(t, "00000000-0000-0000-0000-000000081062", selection.Account.ID)
			require.True(t, selection.Acquired)
		})
	}
}

func TestSelectAccountWithSchedulerForImages_UnspecifiedSizeSkipsBusyPreferredAccountWithSession(t *testing.T) {
	for _, mode := range openAIImageSizeSchedulerModes {
		t.Run(mode.name, func(t *testing.T) {
			groupID := "00000000-0000-0000-0000-000000093009"
			accounts := []Account{
				newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000081111", 0),
				newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000081112", 9, "4K"),
			}
			for i := range accounts {
				accounts[i].GroupIDs = []string{groupID}
			}
			cache := &schedulerTestGatewayCache{}
			concurrency := schedulerTestConcurrencyCache{acquireResults: map[string]bool{"00000000-0000-0000-0000-000000081111": false}}
			svc := newOpenAIImageSizeSchedulerTestService(mode, accounts, cache, concurrency)

			selection, err := selectOpenAIImageSizeTestAccount(t, svc, groupID, "image-size-busy-session", "")
			require.NoError(t, err)
			require.Equal(t, "00000000-0000-0000-0000-000000081112", selection.Account.ID)
			require.True(t, selection.Acquired)
		})
	}
}

func TestSelectAccountWithSchedulerForImages_SpilloverToRestrictedAccountKeepsStickyBinding(t *testing.T) {
	for _, mode := range openAIImageSizeSchedulerModes {
		t.Run(mode.name, func(t *testing.T) {
			groupID := "00000000-0000-0000-0000-000000093011"
			sessionHash := "image-size-spillover"
			accounts := []Account{
				newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000081131", 0),
				newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000081132", 9, "4K"),
			}
			for i := range accounts {
				accounts[i].GroupIDs = []string{groupID}
			}
			cache := &schedulerTestGatewayCache{sessionBindings: map[string]string{"openai:" + sessionHash: "00000000-0000-0000-0000-000000081131"}}
			concurrency := schedulerTestConcurrencyCache{
				acquireResults: map[string]bool{"00000000-0000-0000-0000-000000081131": false},
				waitCounts:     map[string]int{"00000000-0000-0000-0000-000000081131": 3},
				loadMap:        map[string]*AccountLoadInfo{"00000000-0000-0000-0000-000000081131": {AccountID: "00000000-0000-0000-0000-000000081131", LoadRate: 100, WaitingCount: 3}},
			}
			svc := newOpenAIImageSizeSchedulerTestService(mode, accounts, cache, concurrency)

			selection, err := selectOpenAIImageSizeTestAccount(t, svc, groupID, sessionHash, "")
			require.NoError(t, err)
			require.NotNil(t, selection.Account)
			require.Equal(t, "00000000-0000-0000-0000-000000081131", cache.sessionBindings["openai:"+sessionHash], "a spillover pick must not rebind the session")
			if mode.loadBatch {
				require.Equal(t, "00000000-0000-0000-0000-000000081132", selection.Account.ID)
				require.True(t, selection.Acquired)
			}
		})
	}
}

func TestSelectAccountWithSchedulerForImages_SpilloverWithoutBindingBindsServedAccount(t *testing.T) {
	for _, mode := range openAIImageSizeSchedulerModes {
		t.Run(mode.name, func(t *testing.T) {
			groupID := "00000000-0000-0000-0000-000000093012"
			sessionHash := "image-size-spillover-unbound"
			accounts := []Account{
				newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000081141", 0),
				newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000081142", 9, "4K"),
			}
			for i := range accounts {
				accounts[i].GroupIDs = []string{groupID}
			}
			cache := &schedulerTestGatewayCache{}
			concurrency := schedulerTestConcurrencyCache{
				acquireResults: map[string]bool{"00000000-0000-0000-0000-000000081141": false},
				waitCounts:     map[string]int{"00000000-0000-0000-0000-000000081141": 3},
				loadMap:        map[string]*AccountLoadInfo{"00000000-0000-0000-0000-000000081141": {AccountID: "00000000-0000-0000-0000-000000081141", LoadRate: 100, WaitingCount: 3}},
			}
			svc := newOpenAIImageSizeSchedulerTestService(mode, accounts, cache, concurrency)

			selection, err := selectOpenAIImageSizeTestAccount(t, svc, groupID, sessionHash, "")
			require.NoError(t, err)
			require.Equal(t, "00000000-0000-0000-0000-000000081142", selection.Account.ID)
			require.Equal(t, "00000000-0000-0000-0000-000000081142", cache.sessionBindings["openai:"+sessionHash])
		})
	}
}

func TestSelectAccountWithSchedulerForImages_SpilloverReplacesClearedStaleBinding(t *testing.T) {
	for _, mode := range openAIImageSizeSchedulerModes {
		t.Run(mode.name, func(t *testing.T) {
			groupID := "00000000-0000-0000-0000-000000093014"
			sessionHash := "image-size-spillover-stale"
			accounts := []Account{
				newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000081161", 0),
				newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000081162", 9, "4K"),
				newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000081163", 0),
			}
			accounts[2].Schedulable = false
			for i := range accounts {
				accounts[i].GroupIDs = []string{groupID}
			}
			cache := &schedulerTestGatewayCache{sessionBindings: map[string]string{"openai:" + sessionHash: "00000000-0000-0000-0000-000000081163"}}
			concurrency := schedulerTestConcurrencyCache{
				acquireResults: map[string]bool{"00000000-0000-0000-0000-000000081161": false},
				waitCounts:     map[string]int{"00000000-0000-0000-0000-000000081161": 3},
				loadMap:        map[string]*AccountLoadInfo{"00000000-0000-0000-0000-000000081161": {AccountID: "00000000-0000-0000-0000-000000081161", LoadRate: 100, WaitingCount: 3}},
			}
			svc := newOpenAIImageSizeSchedulerTestService(mode, accounts, cache, concurrency)

			selection, err := selectOpenAIImageSizeTestAccount(t, svc, groupID, sessionHash, "")
			require.NoError(t, err)
			require.Equal(t, "00000000-0000-0000-0000-000000081162", selection.Account.ID)
			require.True(t, selection.Acquired)
			require.Equal(t, "00000000-0000-0000-0000-000000081162", cache.sessionBindings["openai:"+sessionHash], "a binding that is no longer usable is replaced by the served account")
		})
	}
}

func TestSelectAccountWithSchedulerForImages_AllPassesQueuedKeepBindingOnReturnedAccount(t *testing.T) {
	for _, mode := range openAIImageSizeSchedulerModes {
		t.Run(mode.name, func(t *testing.T) {
			groupID := "00000000-0000-0000-0000-000000093015"
			sessionHash := "image-size-all-queued"
			accounts := []Account{
				newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000081171", 0),
				newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000081172", 9, "4K"),
			}
			for i := range accounts {
				accounts[i].GroupIDs = []string{groupID}
			}
			cache := &schedulerTestGatewayCache{}
			concurrency := schedulerTestConcurrencyCache{
				acquireResults: map[string]bool{"00000000-0000-0000-0000-000000081171": false, "00000000-0000-0000-0000-000000081172": false},
				waitCounts:     map[string]int{"00000000-0000-0000-0000-000000081171": 3, "00000000-0000-0000-0000-000000081172": 3},
				loadMap: map[string]*AccountLoadInfo{
					"00000000-0000-0000-0000-000000081171": {AccountID: "00000000-0000-0000-0000-000000081171", LoadRate: 100, WaitingCount: 3},
					"00000000-0000-0000-0000-000000081172": {AccountID: "00000000-0000-0000-0000-000000081172", LoadRate: 100, WaitingCount: 3},
				},
			}
			svc := newOpenAIImageSizeSchedulerTestService(mode, accounts, cache, concurrency)

			selection, err := selectOpenAIImageSizeTestAccount(t, svc, groupID, sessionHash, "")
			require.NoError(t, err)
			require.Equal(t, "00000000-0000-0000-0000-000000081171", selection.Account.ID)
			require.NotNil(t, selection.WaitPlan)
			bound, ok := cache.sessionBindings["openai:"+sessionHash]
			if mode.loadBatch {
				require.False(t, ok, "a queued selection does not bind the session")
			} else {
				require.Equal(t, "00000000-0000-0000-0000-000000081171", bound, "eager sequential binding stays on the returned queued account")
			}
		})
	}
}

func TestSelectAccountWithSchedulerForImages_UnspecifiedSizePrefersHealthyProxyOverQuarantinedUnrestricted(t *testing.T) {
	for _, mode := range openAIImageSizeSchedulerModes {
		t.Run(mode.name, func(t *testing.T) {
			quarantinedProxy := "00000000-0000-0000-0000-000000000071"
			healthyProxy := "00000000-0000-0000-0000-000000000072"
			unrestricted := newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000081121", 0)
			unrestricted.ProxyID = &quarantinedProxy
			highOnly := newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000081122", 9, "4K")
			highOnly.ProxyID = &healthyProxy
			svc := newOpenAIImageSizeSchedulerTestService(mode, []Account{unrestricted, highOnly}, nil, schedulerTestConcurrencyCache{})
			svc.openaiProxyStreamCircuit = newOpenAIProxyStreamCircuit(openAIProxyStreamCircuitSettings{
				failureThreshold: 1,
				failureWindow:    time.Minute,
				quarantineTTL:    10 * time.Minute,
				maxEntries:       16,
			})
			svc.openaiProxyStreamCircuit.recordFailure(quarantinedProxy, time.Now())

			selection, err := selectOpenAIImageSizeTestAccount(t, svc, "00000000-0000-0000-0000-000000093010", "", "")
			require.NoError(t, err)
			require.Equal(t, "00000000-0000-0000-0000-000000081122", selection.Account.ID, "a restricted account on a healthy proxy beats the quarantine fail-open")

			svc.accountRepo = schedulerTestOpenAIAccountRepo{accounts: []Account{unrestricted}}
			selection, err = selectOpenAIImageSizeTestAccount(t, svc, "00000000-0000-0000-0000-000000093010", "", "")
			require.NoError(t, err)
			require.Equal(t, "00000000-0000-0000-0000-000000081121", selection.Account.ID, "fail-open still serves when nothing else can")
		})
	}
}

func TestSelectAccountWithSchedulerForImages_StickyAccountUnderSizeRouting(t *testing.T) {
	for _, mode := range openAIImageSizeSchedulerModes {
		t.Run(mode.name, func(t *testing.T) {
			groupID := "00000000-0000-0000-0000-000000093005"
			sessionHash := "image-size-sticky"
			accounts := []Account{
				newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000081071", 0),
				newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000081072", 9, "1K", "2K"),
			}
			for i := range accounts {
				accounts[i].GroupIDs = []string{groupID}
			}
			cache := &schedulerTestGatewayCache{sessionBindings: map[string]string{
				"openai:" + sessionHash: "00000000-0000-0000-0000-000000081072",
			}}
			svc := newOpenAIImageSizeSchedulerTestService(mode, accounts, cache, schedulerTestConcurrencyCache{})

			selection, err := selectOpenAIImageSizeTestAccount(t, svc, groupID, sessionHash, "")
			require.NoError(t, err)
			require.Equal(t, "00000000-0000-0000-0000-000000081072", selection.Account.ID, "unspecified size keeps the sticky account")

			cache.sessionBindings["openai:"+sessionHash] = "00000000-0000-0000-0000-000000081072"
			selection, err = selectOpenAIImageSizeTestAccount(t, svc, groupID, sessionHash, "3840x2160")
			require.NoError(t, err)
			require.Equal(t, "00000000-0000-0000-0000-000000081071", selection.Account.ID, "explicit 4K leaves a 1K/2K-only sticky account")
		})
	}
}

func TestSelectAccountWithSchedulerForImages_BusyStickyAccountMatchesUnroutedSelection(t *testing.T) {
	for _, mode := range openAIImageSizeSchedulerModes {
		t.Run(mode.name, func(t *testing.T) {
			groupID := "00000000-0000-0000-0000-000000093013"
			sessionHash := "image-size-busy-sticky"
			selectWith := func(ctx context.Context) (string, bool) {
				accounts := []Account{
					newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000081151", 0),
					newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000081152", 9, "4K"),
				}
				for i := range accounts {
					accounts[i].GroupIDs = []string{groupID}
				}
				cache := &schedulerTestGatewayCache{sessionBindings: map[string]string{"openai:" + sessionHash: "00000000-0000-0000-0000-000000081151"}}
				concurrency := schedulerTestConcurrencyCache{
					acquireResults: map[string]bool{"00000000-0000-0000-0000-000000081151": false},
					loadMap:        map[string]*AccountLoadInfo{"00000000-0000-0000-0000-000000081151": {AccountID: "00000000-0000-0000-0000-000000081151", LoadRate: 100}},
				}
				svc := newOpenAIImageSizeSchedulerTestService(mode, accounts, cache, concurrency)
				selection, _, err := svc.SelectAccountWithSchedulerForImages(ctx, &groupID, sessionHash, "gpt-image-2", nil, OpenAIImagesCapabilityNative)
				require.NoError(t, err)
				require.NotNil(t, selection.Account)
				return selection.Account.ID, selection.Acquired
			}
			routedID, routedAcquired := selectWith(WithOpenAIImageSizeRouting(context.Background(), ""))
			unroutedID, unroutedAcquired := selectWith(context.Background())
			require.Equal(t, unroutedID, routedID, "a busy sticky account is waited on or escaped exactly as without size routing")
			require.Equal(t, unroutedAcquired, routedAcquired)
		})
	}
}

func TestSelectAccountWithScheduler_TextRequestIgnoresImageSizeTiers(t *testing.T) {
	for _, mode := range openAIImageSizeSchedulerModes {
		t.Run(mode.name, func(t *testing.T) {
			groupID := "00000000-0000-0000-0000-000000093006"
			accounts := []Account{newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000081081", 0, "4K")}
			svc := newOpenAIImageSizeSchedulerTestService(mode, accounts, nil, schedulerTestConcurrencyCache{})

			selection, _, err := svc.SelectAccountWithScheduler(context.Background(), &groupID, "", "", "gpt-5.6-sol", nil, OpenAIUpstreamTransportAny, false)
			require.NoError(t, err)
			require.Equal(t, "00000000-0000-0000-0000-000000081081", selection.Account.ID)
		})
	}
}

func TestSelectOpenAIAccountByImageSizePasses_StopsWhenNoAccountVetoed(t *testing.T) {
	restricted := newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000000001", 0, "4K")
	unrestricted := newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000000002", 0)

	calls := 0
	_, _, err := selectOpenAIAccountByImageSizePasses(WithOpenAIImageSizeRouting(context.Background(), ""), func(ctx context.Context) (*AccountSelectionResult, OpenAIAccountScheduleDecision, error) {
		calls++
		vetoed, _ := openAIImageSizeVetoReason(ctx, &unrestricted)
		require.False(t, vetoed)
		return nil, OpenAIAccountScheduleDecision{}, ErrNoAvailableAccounts
	})
	require.ErrorIs(t, err, ErrNoAvailableAccounts)
	require.Equal(t, 1, calls)

	calls = 0
	_, _, err = selectOpenAIAccountByImageSizePasses(WithOpenAIImageSizeRouting(context.Background(), ""), func(ctx context.Context) (*AccountSelectionResult, OpenAIAccountScheduleDecision, error) {
		calls++
		_, _ = openAIImageSizeVetoReason(ctx, &restricted)
		return nil, OpenAIAccountScheduleDecision{}, ErrNoAvailableAccounts
	})
	require.ErrorIs(t, err, ErrNoAvailableAccounts)
	require.Equal(t, 2, calls, "the 4K pass admits the 4K-only account, so the 2K and 1K passes are skipped")
}

func TestSelectOpenAIAccountByImageSizePasses_ReturnsEarliestWaitPlan(t *testing.T) {
	restricted := newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000000001", 0, "1K")
	calls := 0
	selection, _, err := selectOpenAIAccountByImageSizePasses(WithOpenAIImageSizeRouting(context.Background(), "auto"), func(ctx context.Context) (*AccountSelectionResult, OpenAIAccountScheduleDecision, error) {
		calls++
		_, _ = openAIImageSizeVetoReason(ctx, &restricted)
		if calls == 1 {
			return &AccountSelectionResult{Account: &Account{ID: "00000000-0000-0000-0000-000000000010"}, WaitPlan: &AccountWaitPlan{AccountID: "00000000-0000-0000-0000-000000000010"}}, OpenAIAccountScheduleDecision{}, nil
		}
		if calls == 2 {
			return &AccountSelectionResult{Account: &Account{ID: "00000000-0000-0000-0000-000000000011"}, WaitPlan: &AccountWaitPlan{AccountID: "00000000-0000-0000-0000-000000000011"}}, OpenAIAccountScheduleDecision{}, nil
		}
		return nil, OpenAIAccountScheduleDecision{}, ErrNoAvailableAccounts
	})
	require.NoError(t, err)
	require.Equal(t, "00000000-0000-0000-0000-000000000010", selection.Account.ID)
	require.Equal(t, 2, calls, "the unrestricted pass jumps straight to the 1K-only rank")
}

func TestOpenAIImageSizeTierUnsupported(t *testing.T) {
	groupID := "00000000-0000-0000-0000-000000093007"
	lowOnly := newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000081091", 0, "1K", "2K")
	highOnly := newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000081092", 0, "4K")
	textOnly := newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000081093", 0)
	textOnly.Credentials = map[string]any{"model_mapping": map[string]any{"gpt-5.6-sol": "gpt-5.6-sol"}}

	svc := &OpenAIGatewayService{accountRepo: schedulerTestOpenAIAccountRepo{accounts: []Account{lowOnly, textOnly}}}
	tier, unsupported := svc.OpenAIImageSizeTierUnsupported(context.Background(), &groupID, "gpt-image-2", "3840x2160", "", OpenAIImagesCapabilityNative)
	require.True(t, unsupported)
	require.Equal(t, "4K", tier)

	_, unsupported = svc.OpenAIImageSizeTierUnsupported(context.Background(), &groupID, "gpt-image-2", "1024x1024", "", OpenAIImagesCapabilityNative)
	require.False(t, unsupported)
	_, unsupported = svc.OpenAIImageSizeTierUnsupported(context.Background(), &groupID, "gpt-image-2", "auto", "", OpenAIImagesCapabilityNative)
	require.False(t, unsupported)

	svc.accountRepo = schedulerTestOpenAIAccountRepo{accounts: []Account{lowOnly, highOnly}}
	_, unsupported = svc.OpenAIImageSizeTierUnsupported(context.Background(), &groupID, "gpt-image-2", "3840x2160", "", OpenAIImagesCapabilityNative)
	require.False(t, unsupported)

	svc.accountRepo = schedulerTestOpenAIAccountRepo{accounts: []Account{textOnly}}
	_, unsupported = svc.OpenAIImageSizeTierUnsupported(context.Background(), &groupID, "gpt-image-2", "3840x2160", "", OpenAIImagesCapabilityNative)
	require.False(t, unsupported, "no account serves the model: leave classification to model_not_found")

	chatOnly := newOpenAIImageSizeTestAccount("00000000-0000-0000-0000-000000081094", 0)
	chatOnly.Extra["openai_responses_supported"] = false
	svc.accountRepo = schedulerTestOpenAIAccountRepo{accounts: []Account{chatOnly, lowOnly}}
	_, unsupported = svc.OpenAIImageSizeTierUnsupported(context.Background(), &groupID, "gpt-5.6-sol", "3840x2160", OpenAIEndpointCapabilityResponses, "")
	require.True(t, unsupported, "an unrestricted account without the Responses capability cannot serve the request")
	_, unsupported = svc.OpenAIImageSizeTierUnsupported(context.Background(), &groupID, "gpt-5.6-sol", "3840x2160", OpenAIEndpointCapabilityChatCompletions, "")
	require.False(t, unsupported)
}

// Local availability lookup also serves model-plaza and endpoint diagnostics.
func (r schedulerTestOpenAIAccountRepo) ListModelAvailabilityCandidates(_ context.Context, _ *string, _ []string, _ bool) ([]Account, error) {
	return r.accounts, nil
}
