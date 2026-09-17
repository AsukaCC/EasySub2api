//go:build unit

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestStickyEscapeAlternativesAndLastResort(t *testing.T) {
	for _, tc := range []struct {
		name            string
		other           bool
		otherExcluded   bool
		stickyExcluded  bool
		otherBusy       bool
		wantID          string
		wantWait        bool
		wantUnavailable bool
	}{
		{name: "available alternative", other: true, wantID: "other"},
		{name: "single account remains usable", wantID: "sticky"},
		{name: "caller exclusion preserved", other: true, otherExcluded: true, wantID: "sticky"},
		{name: "queue alternative before original", other: true, otherBusy: true, wantID: "other", wantWait: true},
		{name: "excluded original never becomes last resort", other: true, otherExcluded: true, stickyExcluded: true, wantUnavailable: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			group := "escape-group"
			accounts := []Account{{ID: "sticky", Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true, Concurrency: 1, GroupIDs: []string{group}}}
			if tc.other {
				accounts = append(accounts, Account{ID: "other", Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true, Concurrency: 1, GroupIDs: []string{group}})
			}
			cache := &schedulerTestGatewayCache{sessionBindings: map[string]string{"openai:escape-session": "sticky"}}
			var acquired []string
			concurrency := schedulerTestConcurrencyCache{acquiredIDs: &acquired, acquireResults: map[string]bool{"sticky": true, "other": !tc.otherBusy}}
			cfg := &config.Config{}
			cfg.Gateway.OpenAIScheduler.StickyEscapeEnabled = true
			cfg.Gateway.OpenAIScheduler.StickyEscapeTTFTMs = 15000
			cfg.Gateway.OpenAIScheduler.StickyEscapeErrorRate = 0.5
			svc := &OpenAIGatewayService{accountRepo: schedulerTestOpenAIAccountRepo{accounts: accounts}, cache: cache, cfg: cfg,
				rateLimitService: newOpenAIAdvancedSchedulerRateLimitService("true"), concurrencyService: NewConcurrencyService(concurrency), openaiAccountStats: newOpenAIAccountRuntimeStats()}
			slow := 20000
			svc.openaiAccountStats.report("sticky", true, &slow)
			excluded := map[string]struct{}{"unrelated": {}}
			if tc.otherExcluded {
				excluded["other"] = struct{}{}
			}
			if tc.stickyExcluded {
				excluded["sticky"] = struct{}{}
			}
			before := cloneExcludedAccountIDs(excluded)
			result, decision, err := svc.SelectAccountWithScheduler(context.Background(), &group, "", "escape-session", "gpt-5.6-sol", excluded, OpenAIUpstreamTransportAny, false)
			if tc.wantUnavailable {
				require.ErrorIs(t, err, ErrNoAvailableAccounts)
				require.Nil(t, result)
				require.Empty(t, acquired)
				require.Equal(t, before, excluded)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, tc.wantID, result.Account.ID)
			require.Equal(t, tc.wantWait, result.WaitPlan != nil)
			require.Equal(t, before, excluded)
			require.Equal(t, "sticky", cache.sessionBindings["openai:escape-session"])
			require.Equal(t, openAIAccountScheduleLayerLoadBalance, decision.Layer)
			if tc.wantID == "other" {
				require.NotContains(t, acquired, "sticky")
			}
			if result.ReleaseFunc != nil {
				result.ReleaseFunc()
			}
		})
	}
}

type stickyEscapeAcquireErrorCache struct {
	schedulerTestConcurrencyCache
	err error
}

func (c stickyEscapeAcquireErrorCache) AcquireAccountSlot(context.Context, string, int, string) (bool, error) {
	return false, c.err
}

func TestStickyEscapeDoesNotMaskOperationalErrors(t *testing.T) {
	group := "escape-group"
	accounts := []Account{
		{ID: "sticky", Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true, Concurrency: 1, GroupIDs: []string{group}},
		{ID: "other", Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true, Concurrency: 1, GroupIDs: []string{group}},
	}
	cfg := &config.Config{}
	cfg.Gateway.OpenAIScheduler.StickyEscapeEnabled = true
	cfg.Gateway.OpenAIScheduler.StickyEscapeTTFTMs = 15000
	want := errors.New("slot cache unavailable")
	svc := &OpenAIGatewayService{accountRepo: schedulerTestOpenAIAccountRepo{accounts: accounts},
		cache: &schedulerTestGatewayCache{sessionBindings: map[string]string{"openai:escape-session": "sticky"}}, cfg: cfg,
		rateLimitService: newOpenAIAdvancedSchedulerRateLimitService("true"), openaiAccountStats: newOpenAIAccountRuntimeStats(),
		concurrencyService: NewConcurrencyService(stickyEscapeAcquireErrorCache{err: want})}
	slow := 20000
	svc.openaiAccountStats.report("sticky", true, &slow)
	result, _, err := svc.SelectAccountWithScheduler(context.Background(), &group, "", "escape-session", "gpt-5.6-sol", nil, OpenAIUpstreamTransportAny, false)
	require.ErrorIs(t, err, want)
	require.Nil(t, result)
}

func TestGatewayPlatformIsolationIgnoresLegacyMixedFlag(t *testing.T) {
	svc := &GatewayService{}
	account := &Account{Platform: PlatformAntigravity, Extra: map[string]any{"mixed_scheduling": true}}
	require.False(t, svc.isAccountAllowedForPlatform(account, PlatformGemini, true))
	require.True(t, svc.isAccountAllowedForPlatform(account, PlatformAntigravity, false))
}
