//go:build unit

package service

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/pkg/tlsfingerprint"
	"github.com/stretchr/testify/require"
)

type fingerprintTimeoutTransport struct {
	HTTPUpstream
	waitForCancellation bool
	calls               int
	canceled            bool
}

func (u *fingerprintTimeoutTransport) DoWithTLS(req *http.Request, _ string, _ string, _ int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	u.calls++
	if u.waitForCancellation {
		<-req.Context().Done()
		u.canceled = true
	}
	return nil, context.DeadlineExceeded
}

func TestModelFingerprintTimeoutStopsWithoutRetry(t *testing.T) {
	for _, wait := range []bool{false, true} {
		t.Run(map[bool]string{false: "upstream timeout", true: "job deadline cancels transport"}[wait], func(t *testing.T) {
			svc, repo, _ := newFingerprintTestService()
			transport := &fingerprintTimeoutTransport{waitForCancellation: wait}
			svc.httpUpstream = transport
			now := time.Now()
			snapshot := &ModelFingerprintSnapshot{ID: "timeout-job", Model: "gpt-6-astra", Protocol: "chat", SamplingMode: "independent", Status: "running", StartedAt: now, ExpiresAt: now.Add(15*time.Second + 50*time.Millisecond)}
			svc.runModelFingerprint("one", "admin", snapshot, repo)
			result := waitFingerprint(t, repo.done)
			require.Equal(t, "failed", result.Status)
			require.Equal(t, "timeout", result.Error)
			require.Equal(t, 1, transport.calls)
			require.Equal(t, wait, transport.canceled)
			require.NotNil(t, result.FinishedAt)
			require.Len(t, svc.modelFingerprintUsage.(*fingerprintUsageStub).logs, 1)
		})
	}
}

func TestModelFingerprintTimeoutReleasesWorker(t *testing.T) {
	svc, repo, _ := newFingerprintTestService()
	svc.httpUpstream = &fingerprintAPITransport{}
	_, err := svc.GetFingerprintModels(context.Background(), "one")
	require.NoError(t, err)
	svc.httpUpstream = &fingerprintTimeoutTransport{}
	_, err = svc.StartModelFingerprint(context.Background(), "one", "gpt-6-astra", "admin", ModelFingerprintOptions{APIKeyID: "local-key", Protocol: "chat"})
	require.NoError(t, err)
	require.Equal(t, "timeout", waitFingerprint(t, repo.done).Error)
	require.Eventually(t, func() bool {
		svc.modelFingerprintMu.Lock()
		defer svc.modelFingerprintMu.Unlock()
		return svc.modelFingerprintActive == 0
	}, time.Second, time.Millisecond)
}
