//go:build unit

package service

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/config"
	"github.com/stretchr/testify/require"
)

// 日志头快照必须遮盖凭据与会话绑定值。
func TestSafeHeaderValueForLogRedactsCredentials(t *testing.T) {
	require.Equal(t, "Bearer [redacted]", safeHeaderValueForLog("Authorization", "Bearer sk-live-secret"))
	require.Equal(t, "[redacted]", safeHeaderValueForLog("x-api-key", "sk-ant-secret"))
	require.Equal(t, "[redacted]", safeHeaderValueForLog("Proxy-Authorization", "Basic dXNlcjpwYXNz"))
	require.Equal(t, "[redacted len=22]", safeHeaderValueForLog("Cookie", "__Secure-next-auth=abc"))
	require.Equal(t, "[redacted len=5]", safeHeaderValueForLog("Set-Cookie", "a=b;c"))
	require.Equal(t, "[redacted len=6]", safeHeaderValueForLog("chatgpt-account-id", "acc-id"))
	require.Equal(t, "[redacted len=4]", safeHeaderValueForLog("x-codex-turn-state", "blob"))
	require.Empty(t, safeHeaderValueForLog("cookie", "   "))
	require.Equal(t, "codex_cli_rs", safeHeaderValueForLog("originator", " codex_cli_rs "))
}

// 错误文案清洗：代理凭据、内嵌 Bearer 与敏感查询参数不得进入客户端响应或 ops 记录。
func TestSanitizeUpstreamErrorMessageStripsCredentials(t *testing.T) {
	msg := `Post "https://api.openai.com/v1/responses": proxyconnect tcp: dial http://alice:s3cr3t@10.0.0.9:3128: connection refused`
	got := sanitizeUpstreamErrorMessage(msg)
	require.NotContains(t, got, "alice")
	require.NotContains(t, got, "s3cr3t")
	require.Contains(t, got, "http://***@10.0.0.9:3128")
	require.Contains(t, got, "connection refused", "错误语义必须保留")

	got = sanitizeUpstreamErrorMessage(`socks5h://u:p%40ss@proxy.internal:1080: handshake failed`)
	require.Equal(t, `socks5h://***@proxy.internal:1080: handshake failed`, got)

	got = sanitizeUpstreamErrorMessage(`upstream rejected header Authorization: Bearer eyJhbGciOiJIUzI1NiJ9.payload.sig`)
	require.Equal(t, `upstream rejected header Authorization: Bearer ***`, got)

	got = sanitizeUpstreamErrorMessage(`GET https://example.com/v1?access_token=abc123&key=xyz&foo=bar`)
	require.Equal(t, `GET https://example.com/v1?access_token=***&key=***&foo=bar`, got)

	// 无凭据的文案保持原样
	plain := `dial tcp 1.2.3.4:443: i/o timeout`
	require.Equal(t, plain, sanitizeUpstreamErrorMessage(plain))
	require.Empty(t, sanitizeUpstreamErrorMessage(""))
}

// Proxy.RedactedURL 只保留 scheme://host:port。
func TestProxyRedactedURL(t *testing.T) {
	p := &Proxy{Protocol: "http", Host: "proxy.example.com", Port: 8080, Username: "user", Password: "p@ss"}
	require.Contains(t, p.URL(), "user", "URL() 保留凭据供拨号使用")
	require.Equal(t, "http://proxy.example.com:8080", p.RedactedURL())
	var nilProxy *Proxy
	require.Empty(t, nilProxy.RedactedURL())
}

// 运行时诊断：版本链来源判定与热路径一致，且不含任何凭据字段。
func TestGetCodexOutboundDiagnostics(t *testing.T) {
	SetCodexPrereleaseVersionAllowed(false)
	t.Cleanup(func() { SetCodexPrereleaseVersionAllowed(false) })

	cfg := &config.Config{}
	cfg.Gateway.OpenAIHTTP2.Enabled = true
	cfg.Gateway.TLSFingerprint.Enabled = false

	// 与 wire.go 一致：把 SettingService 的规范 UA 解析器注入进程级收口点。
	installResolver := func(t *testing.T, svc *SettingService) {
		t.Helper()
		SetCodexCanonicalUserAgentResolver(func() string {
			return svc.GetOpenAICodexCanonicalUserAgent(context.Background())
		})
		t.Cleanup(func() { SetCodexCanonicalUserAgentResolver(nil) })
	}

	t.Run("manual_override", func(t *testing.T) {
		svc := NewSettingService(&codexVersionSettingRepoStub{values: map[string]string{
			SettingKeyOpenAICodexClientVersion:          "0.150.0",
			SettingKeyOpenAICodexClientVersionSynced:    "0.153.4",
			SettingKeyOpenAICodexVersionAutoSyncEnabled: "false",
		}}, cfg)
		installResolver(t, svc)

		diag := svc.GetCodexOutboundDiagnostics(context.Background())
		require.Equal(t, CodexVersionSourceManualOverride, diag.VersionSource)
		require.True(t, diag.ManualOverrideEnabled)
		require.Equal(t, "0.150.0", diag.EffectiveVersion)
		require.Equal(t, "0.150.0", diag.ManualOverrideVersion)
		require.Equal(t, "0.153.4", diag.SyncedVersion)
		require.False(t, diag.AutoSyncEnabled)
		require.Equal(t, codexCLIVersion, diag.BuiltinDefaultVersion)
		require.Equal(t, codexUpstreamMinVersion, diag.MinimumSupportedVersion)
		require.Contains(t, diag.UserAgent, "/0.150.0")
		require.NotEmpty(t, diag.Originator)
		require.Equal(t, "http2", diag.ProtocolMode)
		require.True(t, diag.OpenAIHTTP2Enabled)
		require.False(t, diag.ProxyDirectFallbackAllowed)
	})

	t.Run("prerelease_override_rejected_falls_back_to_sync", func(t *testing.T) {
		svc := NewSettingService(&codexVersionSettingRepoStub{values: map[string]string{
			SettingKeyOpenAICodexClientVersion:       "0.154.0-alpha.3",
			SettingKeyOpenAICodexClientVersionSynced: "0.153.4",
		}}, cfg)
		installResolver(t, svc)

		diag := svc.GetCodexOutboundDiagnostics(context.Background())
		require.Equal(t, CodexVersionSourceAutoSync, diag.VersionSource)
		require.False(t, diag.ManualOverrideEnabled)
		require.Equal(t, "0.154.0-alpha.3", diag.RejectedManualOverrideVersion)
		require.Equal(t, "0.153.4", diag.EffectiveVersion)
		require.False(t, diag.PrereleaseAllowed)
	})

	t.Run("builtin_default", func(t *testing.T) {
		svc := NewSettingService(&codexVersionSettingRepoStub{values: map[string]string{}}, cfg)
		installResolver(t, svc)

		diag := svc.GetCodexOutboundDiagnostics(context.Background())
		require.Equal(t, CodexVersionSourceBuiltinDefault, diag.VersionSource)
		require.Equal(t, codexCLIVersion, diag.EffectiveVersion)
		require.True(t, diag.AutoSyncEnabled, "未配置时默认视为开启")
	})

	t.Run("tls_fingerprint_forces_http1_profile", func(t *testing.T) {
		tlsCfg := &config.Config{}
		tlsCfg.Gateway.TLSFingerprint.Enabled = true
		tlsCfg.Gateway.OpenAIHTTP2.Enabled = true
		svc := NewSettingService(&codexVersionSettingRepoStub{values: map[string]string{}}, tlsCfg)
		diag := svc.GetCodexOutboundDiagnostics(context.Background())
		require.Equal(t, "tls_fingerprint_http1", diag.ProtocolMode)
	})

	// 诊断结构不得含任何凭据字段名。
	t.Run("no_sensitive_fields", func(t *testing.T) {
		typ := reflect.TypeOf(CodexOutboundDiagnostics{})
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			name := strings.ToLower(field.Name + " " + field.Tag.Get("json"))
			for _, forbidden := range []string{"token", "cookie", "authorization", "password", "secret", "header_value"} {
				require.False(t, strings.Contains(name, forbidden),
					"diagnostics field %q must not carry credentials", field.Name)
			}
		}
	})
}
