//go:build unit

package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/pkg/tlsfingerprint"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type fingerprintAPITransport struct {
	HTTPUpstream
	mu       sync.Mutex
	payloads []map[string]any
	paths    []string
	hosts    []string
	status   int
	fallback bool
	invalid  bool
}

func (u *fingerprintAPITransport) DoWithTLS(req *http.Request, _ string, _ string, _ int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	body := `{"data":[{"id":"gpt-6-astra","supported_reasoning_levels":["low","high"]},{"id":"gpt-image-2"}]}`
	status := http.StatusOK
	if req.Method == http.MethodPost {
		var payload map[string]any
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			return nil, err
		}
		u.payloads = append(u.payloads, payload)
		u.paths = append(u.paths, req.URL.Path)
		u.hosts = append(u.hosts, req.URL.Hostname())
		text := strings.Repeat("17,42,138,251,", 60)
		if u.invalid {
			text = "No numbers"
		}
		data := map[string]any{"model": "reported-model", "usage": map[string]int{"input_tokens": 40, "output_tokens": 80}}
		if req.URL.Path == "/v1/messages" {
			data["content"] = []map[string]string{{"type": "text", "text": text}}
			data["stop_reason"] = "end_turn"
		} else {
			data["choices"] = []any{map[string]any{"finish_reason": "stop", "message": map[string]string{"content": text}}}
		}
		encoded, _ := json.Marshal(data)
		body = string(encoded)
		if u.status != 0 {
			status = u.status
		}
		if u.fallback && req.URL.Path == "/v1/chat/completions" {
			status = 404
		}
	}
	return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
}

func TestModelFingerprintAPIProtocols(t *testing.T) {
	for _, protocol := range []string{"chat", "anthropic", "auto"} {
		t.Run(protocol, func(t *testing.T) {
			svc, repo, _ := newFingerprintTestService()
			transport := &fingerprintAPITransport{fallback: protocol == "auto"}
			svc.httpUpstream = transport
			models, err := svc.GetFingerprintModels(context.Background(), "one")
			require.NoError(t, err)
			require.Len(t, models, 1)
			require.Equal(t, []string{"low", "high"}, models[0].ReasoningLevels)
			_, err = svc.StartModelFingerprint(context.Background(), "one", "gpt-6-astra", "admin", ModelFingerprintOptions{APIKeyID: "local-key", Protocol: protocol, ReasoningEffort: "high"})
			require.NoError(t, err)
			result := waitFingerprint(t, repo.done)
			require.Equal(t, "completed", result.Status)
			require.Equal(t, "independent", result.SamplingMode)
			require.Equal(t, 3, result.Valid)
			expected := 3
			if protocol == "auto" {
				expected = 4
			}
			require.Equal(t, expected, result.Completed)
			transport.mu.Lock()
			defer transport.mu.Unlock()
			require.Len(t, transport.payloads, expected)
			for i, payload := range transport.payloads {
				require.Len(t, payload["messages"], 1)
				require.Equal(t, false, payload["stream"])
				require.NotContains(t, payload, "system")
				if transport.paths[i] == "/v1/chat/completions" {
					require.Equal(t, "high", payload["reasoning_effort"])
				} else {
					require.Equal(t, "high", payload["output_config"].(map[string]any)["effort"])
				}
			}
			usage := svc.modelFingerprintUsage.(*fingerprintUsageStub)
			usage.mu.Lock()
			defer usage.mu.Unlock()
			require.Len(t, usage.logs, expected)
			for _, log := range usage.logs {
				require.Equal(t, RequestTypeTest, log.RequestType)
				require.Equal(t, "admin", log.UserID)
				require.Equal(t, "local-key", log.APIKeyID)
				require.Equal(t, "fingerprint-group", *log.GroupID)
				require.NotNil(t, log.UpstreamEndpoint)
				require.Equal(t, "/admin/accounts/:id/model-fingerprint", *log.InboundEndpoint)
				require.Equal(t, 0.0, log.ActualCost)
			}
		})
	}
}

func TestModelFingerprintAcceptsUpstreamAPIKeyAccounts(t *testing.T) {
	for _, platform := range []string{PlatformOpenAI, PlatformAnthropic} {
		t.Run(platform, func(t *testing.T) {
			account := &Account{
				Platform:    platform,
				Type:        AccountTypeUpstream,
				Credentials: map[string]any{"api_key": "upstream-test-key", "base_url": "https://relay.example.com"},
			}
			require.True(t, fingerprintAPIAccount(account))
			if platform != PlatformAnthropic {
				require.Equal(t, "upstream-test-key", account.GetOpenAIProtocolAPIKey())
			}
		})
	}
}

func TestModelFingerprintRejectsNonOpenAIOrAnthropicAccounts(t *testing.T) {
	for _, platform := range []string{PlatformGemini, PlatformTypeSafe, PlatformGrok, PlatformKimi, PlatformAntigravity} {
		t.Run(platform, func(t *testing.T) {
			account := &Account{
				Platform:    platform,
				Type:        AccountTypeAPIKey,
				Credentials: map[string]any{"api_key": "native-key", "base_url": "https://relay.example.com"},
			}
			require.False(t, fingerprintAPIAccount(account))
		})
	}
}

func TestModelFingerprintAPIValidationAndRetryBound(t *testing.T) {
	svc, repo, _ := newFingerprintTestService()
	transport := &fingerprintAPITransport{invalid: true}
	svc.httpUpstream = transport
	for _, option := range []ModelFingerprintOptions{
		{APIKeyID: "local-key", Protocol: "responses"}, {APIKeyID: "local-key", Protocol: "chat", ReasoningEffort: "unsupported"},
	} {
		_, err := svc.StartModelFingerprint(context.Background(), "one", "gpt-6-astra", "admin", option)
		require.Error(t, err)
	}
	_, err := svc.StartModelFingerprint(context.Background(), "one", "unadvertised", "admin", ModelFingerprintOptions{APIKeyID: "local-key", Protocol: "chat"})
	require.Error(t, err)
	require.Empty(t, transport.payloads)
	_, err = svc.StartModelFingerprint(context.Background(), "one", "gpt-6-astra", "admin", ModelFingerprintOptions{APIKeyID: "local-key", Protocol: "chat"})
	require.NoError(t, err)
	result := waitFingerprint(t, repo.done)
	require.Equal(t, "insufficient_samples", result.Error)
	require.Equal(t, 6, result.Completed)
}

func TestModelFingerprintAPIAuthFailureDoesNotRetry(t *testing.T) {
	svc, repo, _ := newFingerprintTestService()
	svc.httpUpstream = &fingerprintAPITransport{status: 401}
	_, err := svc.StartModelFingerprint(context.Background(), "one", "gpt-6-astra", "admin", ModelFingerprintOptions{APIKeyID: "local-key", Protocol: "auto"})
	require.NoError(t, err)
	result := waitFingerprint(t, repo.done)
	require.Equal(t, "upstream_failed", result.Error)
	require.Equal(t, 1, result.Completed)
}

func TestModelFingerprintAPIDefaultHostsAndCustomRelay(t *testing.T) {
	for _, tc := range []struct{ platform, base, host string }{
		{PlatformOpenAI, "", "api.openai.com"},
		{PlatformAnthropic, "", "api.anthropic.com"},
	} {
		t.Run(tc.platform, func(t *testing.T) {
			svc, _, _ := newFingerprintTestService()
			transport := &fingerprintAPITransport{}
			svc.httpUpstream = transport
			account := &Account{ID: "account", Platform: tc.platform, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "test-only", "base_url": tc.base}}
			probe := &modelFingerprintProbe{protocol: "chat", model: "gpt-6-astra", prompt: "test", expected: 300, cancel: func() {}, conversation: &modelFingerprintConversation{id: "test"}, startedAt: time.Now()}
			ctx := context.WithValue(context.Background(), modelFingerprintContextKey{}, probe)
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, "/test", nil).WithContext(ctx)
			require.NoError(t, svc.testFingerprintAPI(c, account))
			require.Equal(t, []string{tc.host}, transport.hosts)
			require.Nil(t, probe.firstTokenMs)
		})
	}
}
