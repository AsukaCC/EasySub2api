//go:build unit

package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// WS 会话 turn-state 绑定带账号溯源：failover 换号后不得回放旧账号铸造的 blob。
func TestOpenAIWSStateStore_SessionTurnStateAccountScoped(t *testing.T) {
	store := NewOpenAIWSStateStore(nil)
	store.BindSessionTurnStateForAccount("g1", "sess-hash", "acc-A", "blob-A", time.Minute)

	state, ok := store.GetSessionTurnStateForAccount("g1", "sess-hash", "acc-A")
	require.True(t, ok)
	require.Equal(t, "blob-A", state)

	// 换号到 acc-B：旧账号的 blob 视为不存在
	_, ok = store.GetSessionTurnStateForAccount("g1", "sess-hash", "acc-B")
	require.False(t, ok, "failover 后旧账号 turn-state 不得进入新账号")

	// 旧式无溯源读取仍能拿到值（兼容既有调用方）
	state, ok = store.GetSessionTurnState("g1", "sess-hash")
	require.True(t, ok)
	require.Equal(t, "blob-A", state)

	// 旧式无溯源绑定：任何账号都可读取（历史数据兼容）
	store.BindSessionTurnState("g1", "legacy-hash", "blob-legacy", time.Minute)
	state, ok = store.GetSessionTurnStateForAccount("g1", "legacy-hash", "acc-B")
	require.True(t, ok)
	require.Equal(t, "blob-legacy", state)

	// 新账号铸造后覆盖绑定，acc-A 不再能回放
	store.BindSessionTurnStateForAccount("g1", "sess-hash", "acc-B", "blob-B", time.Minute)
	_, ok = store.GetSessionTurnStateForAccount("g1", "sess-hash", "acc-A")
	require.False(t, ok)
	state, ok = store.GetSessionTurnStateForAccount("g1", "sess-hash", "acc-B")
	require.True(t, ok)
	require.Equal(t, "blob-B", state)
}

// WS 握手与 HTTP 出站共用 turn-state 守卫：客户端回带的异账号 blob 在握手前被剥离。
func TestBuildOpenAIWSHeaders_StripsForeignTurnStateOnFailover(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &OpenAIGatewayService{}
	decision := OpenAIWSProtocolDecision{Transport: OpenAIUpstreamTransportResponsesWebsocketV2}

	newCtx := func() *gin.Context {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodGet, "/v1/responses", nil)
		c.Request.Header.Set("session-id", "sess-ws-failover")
		c.Set("api_key", &APIKey{ID: "key-7"})
		return c
	}
	oauth := func(id string) *Account {
		return &Account{
			ID:          id,
			Platform:    PlatformOpenAI,
			Type:        AccountTypeOAuth,
			Credentials: map[string]any{"chatgpt_account_id": "chatgpt-" + id},
		}
	}

	// 账号 A 在上一轮铸造了 blob-A 并已交付客户端
	c := newCtx()
	upstream := http.Header{}
	upstream.Set("x-codex-turn-state", "blob-A")
	svc.relayOpenAICodexTurnState(c, oauth("acc-A"), upstream)

	// 同账号握手：保留
	same, _, err := svc.buildOpenAIWSHeaders(
		context.Background(), newCtx(), oauth("acc-A"), "token", decision,
		true, "blob-A", "", "", "gpt-5.6-codex", "",
	)
	require.NoError(t, err)
	require.Equal(t, "blob-A", same.Get("x-codex-turn-state"))

	// failover 换到账号 B：客户端仍回带 blob-A，必须剥离
	foreign, _, err := svc.buildOpenAIWSHeaders(
		context.Background(), newCtx(), oauth("acc-B"), "token", decision,
		true, "blob-A", "", "", "gpt-5.6-codex", "",
	)
	require.NoError(t, err)
	require.Empty(t, foreign.Get("x-codex-turn-state"), "WS 握手不得携带旧账号铸造的 turn-state")
	// 身份三元组仍按新账号规范收口
	require.NotEmpty(t, foreign.Get("originator"))
	require.NotEmpty(t, foreign.Get("version"))
	require.NotEmpty(t, foreign.Get("user-agent"))
}

// API Key OpenAI 路径不注入 OAuth Codex 专属身份头，也不携带平台品牌头。
func TestBuildOpenAIWSHeaders_APIKeyAccountHasNoCodexIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &OpenAIGatewayService{}
	decision := OpenAIWSProtocolDecision{Transport: OpenAIUpstreamTransportResponsesWebsocketV2}

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/responses", nil)
	c.Request.Header.Set("X-EasySub2api", "1")
	c.Request.Header.Set("originator", "codex_cli_rs")
	c.Request.Header.Set("version", "0.153.4")

	account := &Account{
		ID:       "acc-key",
		Platform: PlatformOpenAI,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			credKeyHeaderOverrideEnabled: true,
			credKeyHeaderOverrides: map[string]any{
				"X-Sub2api": "1",
				"X-Gateway": "edge",
				"X-Tenant":  "acme",
			},
		},
	}
	headers, _, err := svc.buildOpenAIWSHeaders(
		context.Background(), c, account, "sk-test", decision,
		false, "", "", "", "gpt-5", "",
	)
	require.NoError(t, err)
	require.Empty(t, headers.Get("version"), "API Key 路径不注入 Codex version 头")
	require.Empty(t, headers.Get("chatgpt-account-id"))
	for existing := range headers {
		lower := strings.ToLower(existing)
		require.NotContains(t, lower, "sub2api", "WS 握手不得携带平台品牌头: %s", existing)
		require.NotEqual(t, "x-gateway", lower)
	}
	require.Equal(t, "acme", getHeaderRaw(headers, "x-tenant"), "兼容性覆写保留")
	require.Empty(t, headers.Get("originator"), "客户端自报 Codex originator 不应透传到 API Key 上游")
}
