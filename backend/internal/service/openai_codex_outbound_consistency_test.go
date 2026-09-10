package service

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// Codex 出站头自洽性回归：session 头形态、compat 桥接身份、OpenAI-Beta 策略、
// installation-id 头体同源。

var uuidV4Pattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestIsolateOpenAISessionHeader_UUIDShape(t *testing.T) {
	require.Equal(t, "", isolateOpenAISessionHeader("k1", ""))
	require.Equal(t, "", isolateOpenAISessionHeader("k1", "   "))

	a := isolateOpenAISessionHeader("k1", "sess-abc")
	require.Regexp(t, uuidV4Pattern, a, "上游 session 头必须是 UUIDv4 形态，而非裸哈希")
	require.Equal(t, a, isolateOpenAISessionHeader("k1", "sess-abc"), "同一 (apiKey, raw) 必须确定性")
	require.NotEqual(t, a, isolateOpenAISessionHeader("k2", "sess-abc"), "不同 API Key 不得碰撞")
	require.NotEqual(t, a, isolateOpenAISessionHeader("k1", "sess-xyz"))

	// 与 Messages / Chat Completions 桥接既有形态完全一致，避免同一会话在不同入口出现两个 ID。
	require.Equal(t, generateSessionUUID(isolateOpenAISessionID("k1", "sess-abc")), a)
}

func TestEnsureCodexIdentityHeaders_NoLegacyResponsesBeta(t *testing.T) {
	h := make(http.Header)
	ensureCodexIdentityHeaders(h)
	identity := resolveCodexOutboundIdentity("")
	require.Equal(t, identity.userAgent, h.Get("user-agent"))
	require.Equal(t, identity.originator, h.Get("originator"))
	require.Equal(t, identity.version, h.Get("version"))
	require.Empty(t, h.Values("OpenAI-Beta"), "HTTP 推理面不得发送 OpenAI-Beta: responses=experimental")

	// 客户端带来的 legacy token 被剥离，独立的 beta 协商保留。
	h = make(http.Header)
	h.Set("OpenAI-Beta", "responses=experimental, other=1")
	ensureCodexIdentityHeaders(h)
	require.Equal(t, []string{"other=1"}, h.Values("OpenAI-Beta"))
}

func TestBuildUpstreamRequest_CompatBridgeKeepsCanonicalIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	account := &Account{
		ID:       "acc-compat",
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"chatgpt_account_id": "chatgpt-acc-1",
		},
		Extra: map[string]any{"openai_device_id": "11111111-2222-4333-8444-555555555555"},
	}
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	// anthropic-cache-* prompt_cache_key 触发 compat 桥接分支。
	body := []byte(`{"model":"gpt-5.4","instructions":"x","input":[],"stream":true,"prompt_cache_key":"anthropic-cache-abc"}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(body))
	c.Request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh) claude-cli/1.0")
	c.Request.Header.Set("session_id", "client-session")

	svc := &OpenAIGatewayService{}
	req, err := svc.buildUpstreamRequest(context.Background(), c, account, body, "token", true, "anthropic-cache-abc", false)
	require.NoError(t, err)

	identity := resolveCodexOutboundIdentity("")
	require.Equal(t, identity.originator, req.Header.Get("originator"), "compat 桥接不得丢失 originator")
	require.Equal(t, identity.version, req.Header.Get("version"), "compat 桥接必须带 version")
	require.Equal(t, identity.userAgent, req.Header.Get("user-agent"), "客户端浏览器型 UA 不得到达 chatgpt.com")
	require.Empty(t, req.Header.Values("OpenAI-Beta"))
	require.Equal(t, "chatgpt-acc-1", req.Header.Get("chatgpt-account-id"))
	require.Regexp(t, uuidV4Pattern, req.Header.Get("session_id"))
	require.Equal(t, isolateOpenAISessionHeader("", "anthropic-cache-abc"), req.Header.Get("session_id"))
	// 头侧 installation-id 与 body 侧 client_metadata 同源（账号 device_id）。
	require.Equal(t, "11111111-2222-4333-8444-555555555555", req.Header.Get("x-codex-installation-id"))
}

func TestBuildUpstreamRequest_SessionHeadersAreUUIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	account := &Account{
		ID:          "acc-oauth",
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Credentials: map[string]any{"chatgpt_account_id": "chatgpt-acc-2"},
	}
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	body := []byte(`{"model":"gpt-5.4","instructions":"x","input":[],"stream":true,"prompt_cache_key":"pc-1"}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(body))
	c.Request.Header.Set("originator", "codex_cli_rs")
	c.Request.Header.Set("User-Agent", "codex_cli_rs/0.150.0 (Mac OS 15; arm64) iTerm.app")

	svc := &OpenAIGatewayService{}
	req, err := svc.buildUpstreamRequest(context.Background(), c, account, body, "token", true, "pc-1", true)
	require.NoError(t, err)
	require.Regexp(t, uuidV4Pattern, req.Header.Get("session_id"))
	require.Regexp(t, uuidV4Pattern, req.Header.Get("conversation_id"))
	require.Equal(t, req.Header.Get("session_id"), req.Header.Get("conversation_id"))
	require.Empty(t, req.Header.Values("OpenAI-Beta"))
	// 未配置 device_id 时不臆造 installation-id。
	require.Empty(t, req.Header.Get("x-codex-installation-id"))
}

func TestApplyCodexInstallationIDHeaderFallback(t *testing.T) {
	oauth := &Account{
		ID:       "acc-dev",
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Extra:    map[string]any{"openai_device_id": "dev-1"},
	}
	h := make(http.Header)
	applyCodexInstallationIDHeaderFallback(oauth, h)
	require.Equal(t, "dev-1", h.Get("x-codex-installation-id"))

	// 客户端已带真实安装 ID 时不覆盖。
	h = make(http.Header)
	h.Set("x-codex-installation-id", "client-inst")
	applyCodexInstallationIDHeaderFallback(oauth, h)
	require.Equal(t, "client-inst", h.Get("x-codex-installation-id"))

	// api_key 账号不适用（Platform API 不是 Codex 身份面）。
	apiKey := &Account{ID: "acc-key", Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Extra: map[string]any{"openai_device_id": "dev-2"}}
	h = make(http.Header)
	applyCodexInstallationIDHeaderFallback(apiKey, h)
	require.Empty(t, h.Get("x-codex-installation-id"))

	applyCodexInstallationIDHeaderFallback(nil, h)
	applyCodexInstallationIDHeaderFallback(oauth, nil)
}
