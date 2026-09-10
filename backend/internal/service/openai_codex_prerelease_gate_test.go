//go:build unit

package service

import (
	"context"
	"net/http"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/pkg/openai"
	"github.com/stretchr/testify/require"
)

// 生产模式（默认）：预发布版本在所有来源上都被拒绝并回退；仅显式开启时放行。
func TestAcceptCodexClientVersionPrereleaseGate(t *testing.T) {
	SetCodexPrereleaseVersionAllowed(false)
	t.Cleanup(func() { SetCodexPrereleaseVersionAllowed(false) })

	require.Equal(t, "0.153.4", AcceptCodexClientVersion("0.153.4"))
	require.Empty(t, AcceptCodexClientVersion("0.154.0-alpha.3"), "生产模式必须拒绝预发布版本")
	require.Empty(t, AcceptCodexClientVersion("0.154.0-beta.1"))
	require.Empty(t, AcceptCodexClientVersion("latest"))
	require.False(t, IsCodexPrereleaseVersion("0.153.4"))
	require.True(t, IsCodexPrereleaseVersion("0.154.0-alpha.3"))

	SetCodexPrereleaseVersionAllowed(true)
	require.True(t, CodexPrereleaseVersionAllowed())
	require.Equal(t, "0.154.0-alpha.3", AcceptCodexClientVersion("0.154.0-alpha.3"), "显式开启后放行")
}

// 面板覆写 / 同步值为预发布形态时，热路径版本解析必须回退到下一来源。
func TestGetOpenAICodexClientVersionRejectsPrereleaseInProduction(t *testing.T) {
	SetCodexPrereleaseVersionAllowed(false)
	t.Cleanup(func() { SetCodexPrereleaseVersionAllowed(false) })

	t.Run("覆写为预发布回退同步值", func(t *testing.T) {
		svc := NewSettingService(&codexVersionSettingRepoStub{values: map[string]string{
			SettingKeyOpenAICodexClientVersion:       "0.154.0-alpha.3",
			SettingKeyOpenAICodexClientVersionSynced: "0.153.4",
		}}, nil)
		require.Equal(t, "0.153.4", svc.GetOpenAICodexClientVersion(context.Background()))
	})

	t.Run("同步值为预发布回退内置常量", func(t *testing.T) {
		svc := NewSettingService(&codexVersionSettingRepoStub{values: map[string]string{
			SettingKeyOpenAICodexClientVersionSynced: "0.154.0-alpha.3",
		}}, nil)
		require.Equal(t, codexCLIVersion, svc.GetOpenAICodexClientVersion(context.Background()))
	})

	t.Run("开启放行后预发布覆写生效", func(t *testing.T) {
		SetCodexPrereleaseVersionAllowed(true)
		t.Cleanup(func() { SetCodexPrereleaseVersionAllowed(false) })
		svc := NewSettingService(&codexVersionSettingRepoStub{values: map[string]string{
			SettingKeyOpenAICodexClientVersion: "0.154.0-alpha.3",
		}}, nil)
		require.Equal(t, "0.154.0-alpha.3", svc.GetOpenAICodexClientVersion(context.Background()))
	})
}

// 规范 UA 解析器给出预发布版本段时，出站身份三元组必须整体回退到稳定版并保持自洽。
func TestResolveCodexOutboundIdentityRejectsPrereleaseUAVersion(t *testing.T) {
	SetCodexPrereleaseVersionAllowed(false)
	t.Cleanup(func() { SetCodexPrereleaseVersionAllowed(false) })
	SetCodexCanonicalUserAgentResolver(func() string {
		return "codex_cli_rs/0.154.0-alpha.3" + codexCLIUserAgentSuffix
	})
	t.Cleanup(func() { SetCodexCanonicalUserAgentResolver(nil) })

	h := http.Header{}
	h.Set("originator", "codex_cli_rs")
	enforceCodexIdentityHeaders(h)

	require.Equal(t, codexCLIVersion, h.Get("version"))
	require.Equal(t, "codex_cli_rs", h.Get("originator"))
	require.Equal(t, "codex_cli_rs/"+codexCLIVersion+codexCLIUserAgentSuffix, h.Get("user-agent"))
	require.Equal(t, codexCLIVersion, openai.CodexUserAgentVersion(h.Get("user-agent")),
		"UA 版本段与 version 头必须同源")
}

// 编译期兜底版本本身必须是稳定版，且不低于上游门槛。
func TestBuiltinCodexVersionIsStableRelease(t *testing.T) {
	require.Equal(t, "0.154.0", codexCLIVersion)
	require.False(t, IsCodexPrereleaseVersion(codexCLIVersion))
	require.GreaterOrEqual(t, CompareVersions(codexCLIVersion, codexUpstreamMinVersion), 0)
}
