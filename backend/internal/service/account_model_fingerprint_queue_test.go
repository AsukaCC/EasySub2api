//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestModelFingerprintQueueCapacityAndCancellation(t *testing.T) {
	svc, repo, upstream := newFingerprintTestService(t)
	upstream.entered, upstream.release = make(chan string, 8), make(chan struct{})
	jobs := make(map[string]*ModelFingerprintSnapshot)
	for _, id := range []string{"one", "two", "three", "four"} {
		job, err := svc.StartModelFingerprint(context.Background(), id, "gpt-6-astra", "admin")
		require.NoError(t, err)
		jobs[id] = job
		select {
		case entered := <-upstream.entered:
			require.Equal(t, id, entered)
		case <-time.After(3 * time.Second):
			t.Fatal("worker did not start")
		}
	}
	for _, id := range []string{"five", "six"} {
		job, err := svc.StartModelFingerprint(context.Background(), id, "gpt-6-astra", "admin")
		require.NoError(t, err)
		require.Equal(t, "queued", job.Status)
		jobs[id] = job
	}
	cancelled, err := svc.CancelModelFingerprint(context.Background(), "six", jobs["six"].ID)
	require.NoError(t, err)
	require.Equal(t, "cancelled", cancelled.Status)
	// A second instance must be able to interrupt the owning instance's transport.
	other, _, _ := newFingerprintTestService(t)
	other.accountRepo = repo
	cancelled, err = other.CancelModelFingerprint(context.Background(), "one", jobs["one"].ID)
	require.NoError(t, err)
	require.Equal(t, "cancelled", cancelled.Status)
	select {
	case entered := <-upstream.entered:
		require.Equal(t, "five", entered)
	case <-time.After(5 * time.Second):
		t.Fatal("next queued account did not start after cancellation")
	}
	close(upstream.release)
	for range 4 {
		require.Equal(t, "completed", waitFingerprint(t, repo.done).Status)
	}
	latest, err := svc.GetModelFingerprint(context.Background(), "one")
	require.NoError(t, err)
	require.Equal(t, "cancelled", latest.Status, "late worker save must not overwrite cancellation")
	upstream.mu.Lock()
	require.Empty(t, upstream.requests["six"])
	upstream.mu.Unlock()
	require.Eventually(t, func() bool {
		svc.modelFingerprintMu.Lock()
		defer svc.modelFingerprintMu.Unlock()
		return svc.modelFingerprintActive == 0
	}, time.Second, time.Millisecond)
}

func TestModelFingerprintQueueResumeAndAtomicClaim(t *testing.T) {
	first, repo, upstream := newFingerprintTestService(t)
	second, _, _ := newFingerprintTestService(t)
	second.accountRepo, second.httpUpstream = repo, upstream
	now := time.Now().UTC()
	job := &ModelFingerprintSnapshot{ID: "persisted-job", UserID: "admin", Status: "queued", Model: "gpt-6-astra", SamplingMode: "single_conversation", Total: 3, StartedAt: now, ExpiresAt: now.Add(time.Hour)}
	claimed, err := repo.ClaimModelFingerprint(context.Background(), "one", job)
	require.NoError(t, err)
	require.True(t, claimed)
	first.wakeModelFingerprintQueue()
	second.wakeModelFingerprintQueue()
	result := waitFingerprint(t, repo.done)
	require.Equal(t, job.ID, result.ID)
	require.Equal(t, "completed", result.Status)
	first.StopModelFingerprintQueue()
	second.StopModelFingerprintQueue()
	upstream.mu.Lock()
	defer upstream.mu.Unlock()
	require.Len(t, upstream.requests["one"], 3)
}

func TestModelFingerprintCancelDoesNotCancelNewJob(t *testing.T) {
	svc, repo, upstream := newFingerprintTestService(t)
	svc.modelFingerprintActive = 4
	old, err := svc.StartModelFingerprint(context.Background(), "one", "gpt-6-astra", "admin")
	require.NoError(t, err)
	_, err = svc.CancelModelFingerprint(context.Background(), "one", old.ID)
	require.NoError(t, err)
	newJob, err := svc.StartModelFingerprint(context.Background(), "one", "gpt-6-astra", "admin")
	require.NoError(t, err)
	_, err = svc.CancelModelFingerprint(context.Background(), "one", old.ID)
	require.Error(t, err)
	current, err := svc.GetModelFingerprint(context.Background(), "one")
	require.NoError(t, err)
	require.Equal(t, newJob.ID, current.ID)
	require.Equal(t, "queued", current.Status)
	require.Empty(t, upstream.requests)
	_, err = svc.CancelModelFingerprint(context.Background(), "one", newJob.ID)
	require.NoError(t, err)
	_, err = repo.ClaimModelFingerprint(context.Background(), "one", &ModelFingerprintSnapshot{ID: "third", Status: "queued", ExpiresAt: time.Now().Add(time.Hour)})
	require.NoError(t, err)
}

func TestModelFingerprintBusyAccountDoesNotBlockOtherQueuedAccounts(t *testing.T) {
	svc, repo, upstream := newFingerprintTestService(t)
	svc.modelFingerprintLoad = fingerprintLoadFunc(func(_ context.Context, accounts []AccountWithConcurrency) (map[string]*AccountLoadInfo, error) {
		load := &AccountLoadInfo{}
		if accounts[0].ID == "busy" {
			load.CurrentConcurrency = 1
		}
		return map[string]*AccountLoadInfo{accounts[0].ID: load}, nil
	})
	_, err := svc.StartModelFingerprint(context.Background(), "busy", "gpt-6-astra", "admin")
	require.NoError(t, err)
	_, err = svc.StartModelFingerprint(context.Background(), "idle", "gpt-6-astra", "admin")
	require.NoError(t, err)
	require.Equal(t, "completed", waitFingerprint(t, repo.done).Status)
	busy, err := svc.GetModelFingerprint(context.Background(), "busy")
	require.NoError(t, err)
	require.Equal(t, "queued", busy.Status)
	upstream.mu.Lock()
	defer upstream.mu.Unlock()
	require.Empty(t, upstream.requests["busy"])
	require.Len(t, upstream.requests["idle"], 3)
}
