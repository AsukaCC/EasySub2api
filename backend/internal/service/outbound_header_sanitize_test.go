//go:build unit

package service

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// 平台品牌 / 基础设施头在任何大小写、任何来源下都必须被出站终态清理剥离，
// 而协议头与凭据头不得被误删。
func TestSanitizeOpenAIOutboundHeaders(t *testing.T) {
	h := http.Header{}
	h.Set("Authorization", "Bearer token-with-sub2api-inside")
	h.Set("Chatgpt-Account-Id", "sub2api-looking-id")
	h.Set("User-Agent", "codex_cli_rs/0.153.4 (Ubuntu 22.4.0; x86_64) xterm-256color")
	h.Set("originator", "codex_cli_rs")
	h.Set("version", "0.153.4")
	h.Set("OpenAI-Beta", "responses=experimental")
	h.Set("X-Codex-Turn-State", "blob")
	// denylist：不同大小写
	h.Set("X-EasySub2api", "1")
	h.Set("x-sub2api", "1")
	h.Set("X-Gateway", "edge-1")
	h.Set("X-Relay", "r1")
	h.Set("X-Provider", "p1")
	h.Set("X-Powered-By", "Go")
	h.Set("X-Server", "node-3")
	h.Set("X-Request-Source", "web")
	h.Set("Via", "1.1 proxy")
	h.Set("Forwarded", "for=1.2.3.4")
	// 头名 / 头值含品牌字样
	h.Set("X-Custom-EasySub2API-Trace", "abc")
	h.Set("X-Client-Tag", "EasySub2api/1.0")
	h.Set("X-Meta", "powered by Sub2API")

	sanitizeOpenAIOutboundHeaders(h)

	for _, removed := range []string{
		"X-EasySub2api", "x-sub2api", "X-Gateway", "X-Relay", "X-Provider", "X-Powered-By",
		"X-Server", "X-Request-Source", "Via", "Forwarded",
		"X-Custom-EasySub2API-Trace", "X-Client-Tag", "X-Meta",
	} {
		require.Empty(t, h.Values(removed), "header %q must be stripped", removed)
	}
	// 协议头与凭据头保留；凭据值即便偶然含品牌字样也不得被删（否则上游 401）。
	require.Equal(t, "Bearer token-with-sub2api-inside", h.Get("Authorization"))
	require.Equal(t, "sub2api-looking-id", h.Get("Chatgpt-Account-Id"))
	require.Equal(t, "codex_cli_rs", h.Get("originator"))
	require.Equal(t, "0.153.4", h.Get("version"))
	require.Equal(t, "responses=experimental", h.Get("OpenAI-Beta"))
	require.Equal(t, "blob", h.Get("X-Codex-Turn-State"))
	require.NotEmpty(t, h.Get("User-Agent"))

	// nil 安全
	sanitizeOpenAIOutboundHeaders(nil)
}

func TestSanitizeOutboundHeaderMap(t *testing.T) {
	m := map[string]string{
		"authorization": "Bearer x",
		"originator":    "Codex Desktop",
		"X-Powered-By":  "EasySub2api",
		"via":           "1.1 gw",
		"x-trace":       "sub2api-edge",
	}
	sanitizeOutboundHeaderMap(m)
	require.Equal(t, map[string]string{
		"authorization": "Bearer x",
		"originator":    "Codex Desktop",
	}, m)
	sanitizeOutboundHeaderMap(nil)
}

// 账号级 Header Override 不能成为平台品牌头的出站通道：保存即拒绝、应用即跳过。
func TestHeaderOverrideRejectsPlatformBrandHeaders(t *testing.T) {
	for _, tc := range []struct {
		name, value string
	}{
		{"X-EasySub2api", "1"},
		{"x-sub2api", "1"},
		{"X-Gateway", "edge"},
		{"Via", "1.1 gw"},
		{"Forwarded", "for=1.1.1.1"},
		{"X-Powered-By", "gin"},
		{"X-Trace", "EasySub2api/1.0"},
		{"X-Anything-Sub2API", "x"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			creds := map[string]any{
				credKeyHeaderOverrideEnabled: true,
				credKeyHeaderOverrides:       map[string]any{tc.name: tc.value},
			}
			err := NormalizeHeaderOverrideCredentials(creds)
			require.Error(t, err, "保存路径必须拒绝 %q", tc.name)

			// 应用路径（兜底历史落库数据）：跳过该条目，不写入出站头。
			account := &Account{
				Platform: PlatformOpenAI,
				Type:     AccountTypeAPIKey,
				Credentials: map[string]any{
					credKeyHeaderOverrideEnabled: true,
					credKeyHeaderOverrides:       map[string]any{tc.name: tc.value, "x-allowed": "ok"},
				},
			}
			h := http.Header{}
			account.ApplyHeaderOverrides(h)
			for existing := range h {
				require.False(t, strings.EqualFold(existing, tc.name), "override %q must be skipped", tc.name)
			}
			// 覆写以小写 wire casing 写入，按原样键读取
			require.Equal(t, "ok", getHeaderRaw(h, "x-allowed"))
		})
	}
}

// 合法的兼容性覆写（如中间层准入头）仍然生效，避免过度拦截。
func TestHeaderOverrideKeepsBenignHeaders(t *testing.T) {
	creds := map[string]any{
		credKeyHeaderOverrideEnabled: true,
		credKeyHeaderOverrides:       map[string]any{"X-Tenant": "acme", "OpenAI-Organization": "org_1"},
	}
	require.NoError(t, NormalizeHeaderOverrideCredentials(creds))
	account := &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: creds}
	h := http.Header{}
	account.ApplyHeaderOverrides(h)
	sanitizeOpenAIOutboundHeaders(h)
	require.Equal(t, "acme", getHeaderRaw(h, "x-tenant"))
	require.Equal(t, "org_1", getHeaderRaw(h, "openai-organization"))
}

// Live / 额度探针等非主网关路径同样经过终态清理。
func TestLiveAndQuotaProbeHeadersCarryNoPlatformBrand(t *testing.T) {
	live := http.Header{}
	live.Set("X-EasySub2api", "1")
	live.Set("Via", "1.1 gw")
	applyLiveUpstreamIdentityHeaders(live)
	require.Empty(t, live.Get("X-EasySub2api"))
	require.Empty(t, live.Get("Via"))
	require.Equal(t, "quicksilver=v2", live.Get("OpenAI-Alpha"))
	require.Empty(t, live.Get("OpenAI-Beta"))
	require.NotEmpty(t, live.Get("originator"))
	require.NotEmpty(t, live.Get("user-agent"))

	quota := buildCodexCommonHeaders("tok", "acc", false)
	sanitizeOutboundHeaderMap(quota)
	for name, value := range quota {
		require.False(t, isInternalOutboundHeader(name, value), "quota probe header %q leaks platform identity", name)
	}
	require.Equal(t, openaiQuotaCodexOriginator, quota["originator"])
}
