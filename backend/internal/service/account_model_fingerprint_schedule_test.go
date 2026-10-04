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
func (r *fingerprintScheduleTestRepo) CleanupModelFingerprintHistory(context.Context, time.Time) error {
	r.cleaned = true
	return nil
}

func TestModelFingerprintScheduleConfigAndRun(t *testing.T) {
	svc, base, _ := newFingerprintTestService()
	repo := &fingerprintScheduleTestRepo{fingerprintTestRepo: base}
	svc.accountRepo = repo
	svc.httpUpstream = &fingerprintAPITransport{}
	plan := &ModelFingerprintSchedule{AccountID: "one", UserID: "admin", Enabled: true, Options: ModelFingerprintOptions{Model: "gpt-6-astra", Protocol: "chat", ReasoningEffort: "low"}}
	require.NoError(t, svc.SetFingerprintSchedule(context.Background(), plan))
	require.True(t, repo.saved.NextRunAt.After(time.Now()))
	require.Equal(t, 0, repo.saved.NextRunAt.Second())
	require.Equal(t, 0, repo.saved.NextRunAt.Minute()%30)
	repo.plans = []ModelFingerprintSchedule{*repo.saved}
	svc.RunDueModelFingerprints(context.Background())
	result := waitFingerprint(t, base.done)
	require.True(t, repo.cleaned)
	require.Equal(t, "scheduled", result.Source)
	require.Equal(t, "completed", result.Status)
	require.Equal(t, "admin", result.UserID)
}

func TestModelFingerprintSchedulePreparationFailureIsRecorded(t *testing.T) {
	svc, repo, _ := newFingerprintTestService()
	svc.httpUpstream = &fingerprintAPITransport{}
	result, err := svc.StartModelFingerprint(context.Background(), "one", "removed-model", "admin", ModelFingerprintOptions{Protocol: "chat", Source: "scheduled"})
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
