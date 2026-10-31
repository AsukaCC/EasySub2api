//go:build unit

package service

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/config"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestAccountConnectionCustomPromptReachesUpstream(t *testing.T) {
	for _, tc := range []struct {
		name, platform, model, mode, promptPath, response string
	}{
		{"claude", PlatformAnthropic, "claude-sonnet-4-6", "", "messages.0.content", "data: {\"type\":\"message_stop\"}\n\n"},
		{"responses", PlatformOpenAI, "gpt-6-astra", "", "input.0.content.0.text", "data: {\"type\":\"response.completed\"}\n\n"},
		{"compact", PlatformOpenAI, "gpt-6-astra", "compact", "input.0.content", `{"output":[{"type":"compaction","encrypted_content":"opaque"}]}`},
		{"gemini", PlatformGemini, "gemini-3.1-pro-preview", "", "contents.0.parts.0.text", "data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"READY\"}]},\"finishReason\":\"STOP\"}]}\n\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			account := &Account{ID: "prompt-account", Platform: tc.platform, Type: AccountTypeAPIKey, Status: StatusActive,
				Credentials: map[string]any{"api_key": "test-key", "base_url": "https://example.com"}}
			repo := &openAIAccountTestRepo{mockAccountRepoForGemini: mockAccountRepoForGemini{accountsByID: map[string]*Account{account.ID: account}}}
			upstream := &queuedHTTPUpstream{responses: []*http.Response{newJSONResponse(http.StatusOK, tc.response)}}
			svc := &AccountTestService{accountRepo: repo, httpUpstream: upstream, cfg: &config.Config{}}
			c, _ := newTestContext()
			require.NoError(t, svc.TestAccountConnection(c, account.ID, tc.model, "Reply READY", tc.mode))
			require.Len(t, upstream.requests, 1)
			body, err := io.ReadAll(upstream.requests[0].Body)
			require.NoError(t, err)
			require.Equal(t, "Reply READY", gjson.GetBytes(body, tc.promptPath).String())
			if tc.mode == "compact" {
				require.Equal(t, "compaction_trigger", gjson.GetBytes(body, "input.1.type").String())
			}
			deadline, ok := upstream.requests[0].Context().Deadline()
			require.True(t, ok)
			require.WithinDuration(t, time.Now().Add(3*time.Minute), deadline, time.Second)
		})
	}
}

func TestAntigravityConnectionCustomPrompt(t *testing.T) {
	svc := &AntigravityGatewayService{}
	gemini, err := svc.buildGeminiTestRequest("project", "gemini-3.1-pro-preview", "Reply READY")
	require.NoError(t, err)
	require.Equal(t, "Reply READY", gjson.GetBytes(gemini, "request.contents.0.parts.0.text").String())
	claude, err := svc.buildClaudeTestRequest("project", "claude-sonnet-4-6", "Reply READY")
	require.NoError(t, err)
	require.Equal(t, "Reply READY", gjson.GetBytes(claude, "request.contents.0.parts.0.text").String())
}

func TestAccountConnectionGrokCustomPrompt(t *testing.T) {
	account := &Account{ID: "grok-prompt", Platform: PlatformGrok, Type: AccountTypeOAuth, Status: StatusActive,
		Credentials: map[string]any{"access_token": "test-only", "refresh_token": "test-only", "expires_at": time.Now().Add(time.Hour).Format(time.RFC3339)}}
	repo := &mockAccountRepoForGemini{accountsByID: map[string]*Account{account.ID: account}}
	upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: http.StatusOK,
		Header: http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:   io.NopCloser(strings.NewReader("data: {\"type\":\"response.completed\"}\n\n"))}}
	svc := &AccountTestService{accountRepo: repo, grokTokenProvider: NewGrokTokenProvider(repo, nil), httpUpstream: upstream}
	c, _ := newTestContext()
	require.NoError(t, svc.TestAccountConnection(c, account.ID, "grok-4.5", "Reply READY", "text"))
	require.Equal(t, "Reply READY", gjson.GetBytes(upstream.lastBody, "input.0.content.0.text").String())
}

func TestAccountConnectionDeadlineReturnsTimeout(t *testing.T) {
	svc, repo, _ := newFingerprintTestService()
	upstream := &fingerprintTimeoutTransport{waitForCancellation: true}
	svc.httpUpstream = upstream
	c, recorder := newTestContext()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 50*time.Millisecond)
	defer cancel()
	c.Request = c.Request.WithContext(ctx)
	require.ErrorIs(t, svc.TestAccountConnection(c, "one", "gpt-6-astra", "READY", ""), context.DeadlineExceeded)
	require.True(t, upstream.canceled)
	require.Contains(t, recorder.Body.String(), "ACCOUNT_TEST_TIMEOUT")
	require.Empty(t, repo.snapshots)
}
