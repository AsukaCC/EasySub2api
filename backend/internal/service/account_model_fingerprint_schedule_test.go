//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fingerprintScheduleTestRepo struct {
	*fingerprintTestRepo
	modelFingerprintHistoryStore
	plans   []ModelFingerprintSchedule
	saved   *ModelFingerprintSchedule
	cleaned bool
}

func (r *fingerprintScheduleTestRepo) ListDueModelFingerprints(context.Context, time.Time) ([]ModelFingerprintSchedule, error) {
	return r.plans, nil
}
func (r *fingerprintScheduleTestRepo) SetModelFingerprintSchedule(_ context.Context, p *ModelFingerprintSchedule) error {
	copy := *p
	r.saved = &copy
	return nil
}
func (r *fingerprintScheduleTestRepo) GetModelFingerprintSchedule(context.Context, string) (*ModelFingerprintSchedule, error) {
	copy := *r.saved
	return &copy, nil
}
func (r *fingerprintScheduleTestRepo) CleanupModelFingerprintHistory(context.Context, time.Time) error {
	r.cleaned = true
	return nil
}

func TestModelFingerprintScheduleConfigAndRun(t *testing.T) {
	svc, base, _ := newFingerprintTestService()
	repo := &fingerprintScheduleTestRepo{fingerprintTestRepo: base}
	svc.accountRepo = repo
	svc.httpUpstream = &fingerprintAPITransport{}
	plan := &ModelFingerprintSchedule{AccountID: "one", UserID: "admin", Enabled: true, Options: ModelFingerprintOptions{APIKeyID: "local-key", Model: "gpt-6-astra", Protocol: "chat", ReasoningEffort: "low"}}
	require.NoError(t, svc.SetFingerprintSchedule(context.Background(), plan))
	require.True(t, repo.saved.NextRunAt.After(time.Now()))
	require.Equal(t, 0, repo.saved.NextRunAt.Second())
	require.Equal(t, 0, repo.saved.NextRunAt.Minute())
	require.Equal(t, &ModelFingerprintTimeWindow{EndHour: 24, Timezone: "Asia/Shanghai"}, repo.saved.Window)
	repo.plans = []ModelFingerprintSchedule{*repo.saved}
	svc.RunDueModelFingerprints(context.Background())
	result := waitFingerprint(t, base.done)
	require.True(t, repo.cleaned)
	require.Equal(t, "scheduled", result.Source)
	require.Equal(t, "completed", result.Status)
	require.Equal(t, "admin", result.UserID)
	require.Equal(t, "local-key", result.APIKeyID)
}

func TestModelFingerprintTimeWindow(t *testing.T) {
	for _, tc := range []struct {
		name, now, next, timezone string
		start, end                int
		inside                    bool
	}{
		{"before opening", "2026-10-09T00:30:00Z", "2026-10-09T01:00:00Z", "Asia/Shanghai", 9, 18, false},
		{"start inclusive", "2026-10-09T01:00:00Z", "2026-10-09T02:00:00Z", "Asia/Shanghai", 9, 18, true},
		{"hourly", "2026-10-09T01:30:00Z", "2026-10-09T02:00:00Z", "Asia/Shanghai", 9, 18, true},
		{"last hour", "2026-10-09T09:00:00Z", "2026-10-10T01:00:00Z", "Asia/Shanghai", 9, 18, true},
		{"end exclusive", "2026-10-09T10:00:00Z", "2026-10-10T01:00:00Z", "Asia/Shanghai", 9, 18, false},
		{"all day rollover", "2026-10-09T15:59:59Z", "2026-10-09T16:00:00Z", "Asia/Shanghai", 0, 24, true},
		{"one hour daily", "2026-10-09T01:00:00Z", "2026-10-10T01:00:00Z", "Asia/Shanghai", 9, 10, true},
		{"half hour offset", "2026-10-09T03:00:00Z", "2026-10-09T03:30:00Z", "Asia/Kolkata", 9, 18, false},
		{"spring DST skipped hour", "2026-03-08T06:30:00Z", "2026-03-09T06:00:00Z", "America/New_York", 2, 3, false},
		{"fall DST repeated hour", "2026-11-01T05:00:00Z", "2026-11-01T06:00:00Z", "America/New_York", 1, 2, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			now, err := time.Parse(time.RFC3339, tc.now)
			require.NoError(t, err)
			window := ModelFingerprintTimeWindow{StartHour: tc.start, EndHour: tc.end, Timezone: tc.timezone}
			next, err := window.NextRun(now)
			require.NoError(t, err)
			require.Equal(t, tc.next, next.Format(time.RFC3339))
			require.Equal(t, tc.inside, window.Contains(now))
		})
	}
}

func TestModelFingerprintScheduleRejectsInvalidWindow(t *testing.T) {
	for _, window := range []ModelFingerprintTimeWindow{
		{StartHour: -1, EndHour: 24, Timezone: "UTC"},
		{StartHour: 24, EndHour: 24, Timezone: "UTC"},
		{StartHour: 0, EndHour: 25, Timezone: "UTC"},
		{StartHour: 9, EndHour: 9, Timezone: "UTC"},
		{StartHour: 18, EndHour: 9, Timezone: "UTC"},
		{StartHour: 0, EndHour: 24, Timezone: "invalid-zone"},
		{StartHour: 0, EndHour: 24, Timezone: "Local"},
	} {
		svc := &AccountTestService{}
		err := svc.SetFingerprintSchedule(context.Background(), &ModelFingerprintSchedule{Enabled: true, Window: &window})
		require.Error(t, err)
	}
}

func TestModelFingerprintScheduleReadsLegacyAndSavedWindows(t *testing.T) {
	svc, base, _ := newFingerprintTestService()
	due := time.Date(2099, 10, 9, 1, 30, 0, 0, time.UTC)
	repo := &fingerprintScheduleTestRepo{fingerprintTestRepo: base, saved: &ModelFingerprintSchedule{
		AccountID: "one", Enabled: true, NextRunAt: due,
	}}
	svc.accountRepo = repo
	plan, err := svc.GetFingerprintSchedule(context.Background(), "one")
	require.NoError(t, err)
	require.Equal(t, &ModelFingerprintTimeWindow{EndHour: 24, Timezone: "Asia/Shanghai"}, plan.Window)
	require.Equal(t, due.Add(30*time.Minute), plan.NextRunAt)
	repo.saved.Window = &ModelFingerprintTimeWindow{StartHour: 18, EndHour: 20, Timezone: "Asia/Shanghai"}
	plan, err = svc.GetFingerprintSchedule(context.Background(), "one")
	require.NoError(t, err)
	require.Equal(t, repo.saved.Window, plan.Window)
	require.Equal(t, time.Date(2099, 10, 9, 10, 0, 0, 0, time.UTC), plan.NextRunAt)
}

func TestModelFingerprintSchedulePreparationFailureIsRecorded(t *testing.T) {
	svc, repo, _ := newFingerprintTestService()
	svc.httpUpstream = &fingerprintAPITransport{}
	result, err := svc.StartModelFingerprint(context.Background(), "one", "removed-model", "admin", ModelFingerprintOptions{APIKeyID: "local-key", Protocol: "chat", Source: "scheduled"})
	require.NoError(t, err)
	require.Equal(t, "failed", result.Status)
	require.Equal(t, "preparation_failed", result.Error)
	require.Equal(t, 0, result.Completed)
	require.Equal(t, result.ID, waitFingerprint(t, repo.done).ID)
	require.Equal(t, 0, svc.modelFingerprintActive)
}

func TestModelFingerprintScheduleBusyWorkersStillCleanHistory(t *testing.T) {
	svc, base, _ := newFingerprintTestService()
	repo := &fingerprintScheduleTestRepo{fingerprintTestRepo: base, plans: []ModelFingerprintSchedule{{AccountID: "one"}}}
	svc.accountRepo = repo
	svc.modelFingerprintActive = 4
	svc.RunDueModelFingerprints(context.Background())
	require.True(t, repo.cleaned)
	require.Empty(t, base.snapshots)
}
