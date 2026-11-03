//go:build unit

package service

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fingerprintLoadFunc func(context.Context, []AccountWithConcurrency) (map[string]*AccountLoadInfo, error)

func (f fingerprintLoadFunc) GetAccountsLoadBatchFresh(ctx context.Context, accounts []AccountWithConcurrency) (map[string]*AccountLoadInfo, error) {
	return f(ctx, accounts)
}

func TestModelFingerprintEligibility(t *testing.T) {
	future, past := time.Now().Add(time.Hour), time.Now().Add(-time.Hour)
	for _, tc := range []struct {
		name, reason string
		edit         func(*Account)
		load         *AccountLoadInfo
		err          error
	}{
		{name: "idle", load: &AccountLoadInfo{}},
		{name: "active request", load: &AccountLoadInfo{CurrentConcurrency: 1}, reason: "account_busy"},
		{name: "queued request", load: &AccountLoadInfo{WaitingCount: 1}, reason: "account_busy"},
		{name: "load error", err: errors.New("redis unavailable"), reason: "account_activity_unknown"},
		{name: "missing load", reason: "account_activity_unknown"},
		{name: "disabled", edit: func(a *Account) { a.Status = StatusDisabled }, reason: "account_unavailable"},
		{name: "error", edit: func(a *Account) { a.Status = StatusError }, reason: "account_unavailable"},
		{name: "paused", edit: func(a *Account) { a.Schedulable = false }, reason: "account_unavailable"},
		{name: "expired", edit: func(a *Account) { a.ExpiresAt, a.AutoPauseOnExpired = &past, true }, reason: "account_unavailable"},
		{name: "rate limited", edit: func(a *Account) { a.RateLimitResetAt = &future }, reason: "account_unavailable"},
		{name: "overloaded", edit: func(a *Account) { a.OverloadUntil = &future }, reason: "account_unavailable"},
		{name: "cooldown", edit: func(a *Account) { a.TempUnschedulableUntil = &future }, reason: "account_unavailable"},
		{name: "codex 5h full", edit: func(a *Account) { a.Extra["codex_5h_used_percent"] = 100.0 }, reason: "quota_exhausted"},
		{name: "codex 7d overfull", edit: func(a *Account) { a.Extra["codex_7d_used_percent"] = "101" }, reason: "quota_exhausted"},
		{name: "codex not full", edit: func(a *Account) { a.Extra["codex_5h_used_percent"] = 99.9 }, load: &AccountLoadInfo{}},
		{name: "codex reset", edit: func(a *Account) {
			a.Extra["codex_5h_used_percent"] = 100
			a.Extra["codex_5h_reset_at"] = past.Format(time.RFC3339)
		}, load: &AccountLoadInfo{}},
		{name: "codex relative reset", edit: func(a *Account) {
			a.Extra["codex_7d_used_percent"] = 100
			a.Extra["codex_7d_reset_after_seconds"] = 60
			a.Extra["codex_usage_updated_at"] = past.Format(time.RFC3339)
		}, load: &AccountLoadInfo{}},
		{name: "claude 5h full", edit: func(a *Account) { a.Extra["session_window_utilization"] = 1.0; a.SessionWindowEnd = &future }, reason: "quota_exhausted"},
		{name: "claude 5h reset", edit: func(a *Account) { a.Extra["session_window_utilization"] = 1.0; a.SessionWindowEnd = &past }, load: &AccountLoadInfo{}},
		{name: "claude rejected window", edit: func(a *Account) { a.SessionWindowStatus = "rejected"; a.SessionWindowEnd = &future }, reason: "quota_exhausted"},
		{name: "claude rejected window reset", edit: func(a *Account) { a.SessionWindowStatus = "rejected"; a.SessionWindowEnd = &past }, load: &AccountLoadInfo{}},
		{name: "claude 7d full", edit: func(a *Account) {
			a.Extra["passive_usage_7d_utilization"] = 1.0
			a.Extra["passive_usage_7d_reset"] = future.Unix()
		}, reason: "quota_exhausted"},
		{name: "claude 7d reset", edit: func(a *Account) {
			a.Extra["passive_usage_7d_utilization"] = 1.0
			a.Extra["passive_usage_7d_reset"] = past.Unix()
		}, load: &AccountLoadInfo{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := &Account{ID: "one", Status: StatusActive, Schedulable: true, Extra: map[string]any{}}
			if tc.edit != nil {
				tc.edit(a)
			}
			svc := &AccountTestService{modelFingerprintLoad: fingerprintLoadFunc(func(context.Context, []AccountWithConcurrency) (map[string]*AccountLoadInfo, error) {
				return map[string]*AccountLoadInfo{"one": tc.load}, tc.err
			})}
			require.Equal(t, tc.reason, svc.fingerprintSkipReason(context.Background(), a))
		})
	}
	svc := &AccountTestService{}
	require.Equal(t, "account_activity_unknown", svc.fingerprintSkipReason(context.Background(), &Account{Status: StatusActive, Schedulable: true}))
}

func TestModelFingerprintSkippedWithoutUpstreamOrUsage(t *testing.T) {
	for _, source := range []string{"manual", "scheduled"} {
		t.Run(source, func(t *testing.T) {
			svc, repo, upstream := newFingerprintTestService(t)
			require.NoError(t, repo.UpdateExtra(context.Background(), "one", map[string]any{"codex_7d_used_percent": 100}))
			result, err := svc.StartModelFingerprint(context.Background(), "one", "gpt-6-astra", "admin", ModelFingerprintOptions{APIKeyID: "local-key", Protocol: "chat", Source: source})
			require.NoError(t, err)
			require.Equal(t, "skipped", result.Status)
			require.Equal(t, "quota_exhausted", result.Error)
			require.Equal(t, source, result.Source)
			require.Zero(t, result.Completed)
			require.NotNil(t, result.FinishedAt)
			require.Equal(t, result.ID, waitFingerprint(t, repo.done).ID)
			require.Empty(t, upstream.requests)
			require.Empty(t, svc.modelFingerprintUsage.(*fingerprintUsageStub).logs)
			require.Zero(t, svc.modelFingerprintActive)
		})
	}
}

func TestModelFingerprintScheduledBusyQueueAndIdleRecovery(t *testing.T) {
	svc, base, upstream := newFingerprintTestService(t)
	var busy atomic.Bool
	busy.Store(true)
	svc.modelFingerprintLoad = fingerprintLoadFunc(func(context.Context, []AccountWithConcurrency) (map[string]*AccountLoadInfo, error) {
		load := &AccountLoadInfo{}
		if busy.Load() {
			load.CurrentConcurrency = 1
		}
		return map[string]*AccountLoadInfo{"one": load}, nil
	})
	svc.accountRepo = &fingerprintScheduleTestRepo{fingerprintTestRepo: base, plans: []ModelFingerprintSchedule{{
		AccountID: "one", UserID: "admin", Enabled: true,
		Options: ModelFingerprintOptions{APIKeyID: "local-key", Model: "gpt-6-astra", Protocol: "chat"},
	}}}
	svc.RunDueModelFingerprints(context.Background())
	result, err := svc.GetModelFingerprint(context.Background(), "one")
	require.NoError(t, err)
	require.Equal(t, "queued", result.Status)
	require.Equal(t, "account_busy", result.QueueReason)
	require.Equal(t, "scheduled", result.Source)
	require.Empty(t, upstream.requests)
	svc.httpUpstream = &fingerprintAPITransport{}
	busy.Store(false)
	svc.wakeModelFingerprintQueue()
	require.Equal(t, "completed", waitFingerprint(t, base.done).Status)
}

func TestModelFingerprintYieldsBetweenSamples(t *testing.T) {
	for _, reason := range []string{"account_busy", "quota_exhausted", "account_unavailable"} {
		t.Run(reason, func(t *testing.T) {
			svc, repo, upstream := newFingerprintTestService(t)
			var busy atomic.Bool
			svc.modelFingerprintLoad = fingerprintLoadFunc(func(context.Context, []AccountWithConcurrency) (map[string]*AccountLoadInfo, error) {
				load := &AccountLoadInfo{}
				if busy.Load() {
					load.CurrentConcurrency = 1
				}
				return map[string]*AccountLoadInfo{"one": load}, nil
			})
			upstream.entered, upstream.release = make(chan string, 1), make(chan struct{})
			_, err := svc.StartModelFingerprint(context.Background(), "one", "gpt-6-astra", "admin")
			require.NoError(t, err)
			select {
			case <-upstream.entered:
			case <-time.After(3 * time.Second):
				t.Fatal("first sample did not start")
			}
			switch reason {
			case "account_busy":
				busy.Store(true)
			case "quota_exhausted":
				require.NoError(t, repo.UpdateExtra(context.Background(), "one", map[string]any{"codex_5h_used_percent": 100}))
			case "account_unavailable":
				require.NoError(t, repo.UpdateExtra(context.Background(), "one", map[string]any{"quota_limit": 1, "quota_used": 1}))
			}
			close(upstream.release)
			if reason == "account_busy" {
				require.Eventually(t, func() bool {
					result, _ := svc.GetModelFingerprint(context.Background(), "one")
					return result != nil && result.Status == "waiting"
				}, time.Second, time.Millisecond)
				svc.modelFingerprintMu.Lock()
				require.Zero(t, svc.modelFingerprintActive)
				svc.modelFingerprintMu.Unlock()
				busy.Store(false)
				require.Equal(t, "completed", waitFingerprint(t, repo.done).Status)
				return
			}
			result := waitFingerprint(t, repo.done)
			require.Equal(t, "skipped", result.Status)
			require.Equal(t, reason, result.Error)
			require.Equal(t, 1, result.Completed)
			require.Len(t, upstream.requests["one"], 1)
			require.Len(t, svc.modelFingerprintUsage.(*fingerprintUsageStub).logs, 1)
		})
	}
}
