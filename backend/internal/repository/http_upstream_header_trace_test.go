package repository

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/AsukaCC/EasySub2api/internal/pkg/ctxkey"
	"github.com/AsukaCC/EasySub2api/internal/service"
)

func TestUpstreamHeaderTraceCapturesWrittenHeadersAndRedactsCredentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(server.Close)

	ctx := context.WithValue(t.Context(), ctxkey.RequestID, "capture-me")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, server.URL+"/v1/responses?secret=query", strings.NewReader("{}"))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer actual-production-token")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Session_ID", "actual-session-id")
	req.Header.Set("User-Agent", "codex_cli_rs/1.2.3")

	gate := &upstreamHeaderTraceGate{requestID: "capture-me", maxAttempts: 1}
	tracedReq, capture := gate.attach(req, service.HTTPUpstreamProfileOpenAI, "account-42", upstreamProtocolModeOpenAIH2, directProxyKey)
	require.NotNil(t, capture)

	resp, err := http.DefaultClient.Do(tracedReq)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	capture.finish(resp, nil)

	snapshot := capture.snapshot()
	require.True(t, snapshot.headerWriteCompleted)
	require.True(t, snapshot.requestWriteCompleted)
	require.False(t, snapshot.requestWriteError)
	headers := strings.Join(snapshot.headers, "\n")
	require.Contains(t, headers, "Authorization: Bearer [redacted")
	require.Contains(t, headers, "Content-Type: application/json")
	require.Contains(t, headers, "User-Agent: codex_cli_rs/1.2.3")
	require.Contains(t, headers, "Accept-Encoding: gzip")
	require.NotContains(t, headers, "actual-production-token")
	require.NotContains(t, headers, "actual-session-id")
	require.Equal(t, "/v1/responses", capture.targetPath)
}

func TestUpstreamHeaderTraceRequiresMatchingRequestID(t *testing.T) {
	gate := &upstreamHeaderTraceGate{requestID: "capture-me", maxAttempts: 1}

	requestWithID := func(requestID string) *http.Request {
		req, err := http.NewRequestWithContext(
			context.WithValue(t.Context(), ctxkey.RequestID, requestID),
			http.MethodGet,
			"https://example.com/v1/responses",
			nil,
		)
		require.NoError(t, err)
		return req
	}

	req := requestWithID("other")
	tracedReq, capture := gate.attach(req, service.HTTPUpstreamProfileOpenAI, "account-1", upstreamProtocolModeOpenAIH2, directProxyKey)
	require.Same(t, req, tracedReq)
	require.Nil(t, capture)

	req = requestWithID("capture-me")
	tracedReq, capture = gate.attach(req, service.HTTPUpstreamProfileDefault, "account-1", upstreamProtocolModeDefault, directProxyKey)
	require.NotSame(t, req, tracedReq)
	require.NotNil(t, capture)

	tracedReq, capture = gate.attach(req, service.HTTPUpstreamProfileOpenAI, "account-1", upstreamProtocolModeOpenAIH2, directProxyKey)
	require.Same(t, req, tracedReq)
	require.Nil(t, capture)
}

func TestNewUpstreamHeaderTraceGateFromEnv(t *testing.T) {
	t.Setenv(upstreamHeaderTraceRequestIDEnv, " request-123 ")
	t.Setenv(upstreamHeaderTraceMaxAttemptsEnv, "999")

	gate := newUpstreamHeaderTraceGateFromEnv()
	require.NotNil(t, gate)
	require.Equal(t, "request-123", gate.requestID)
	require.Equal(t, uint32(maxUpstreamHeaderTraceAttempts), gate.maxAttempts)

	t.Setenv(upstreamHeaderTraceRequestIDEnv, "")
	require.Nil(t, newUpstreamHeaderTraceGateFromEnv())
}

func TestSanitizeUpstreamHeaderTraceValue(t *testing.T) {
	require.Equal(t, "application/json", sanitizeUpstreamHeaderTraceValue("Content-Type", "application/json"))
	require.Equal(t, "/v1/responses", sanitizeUpstreamHeaderTraceValue(":path", "/v1/responses?token=secret"))
	require.NotContains(t, sanitizeUpstreamHeaderTraceValue("Authorization", "Bearer secret"), "secret")
	require.NotContains(t, sanitizeUpstreamHeaderTraceValue("X-Custom-Token", "secret"), "secret")
	require.Empty(t, fingerprintUpstreamHeaderTraceValue(""))
}
