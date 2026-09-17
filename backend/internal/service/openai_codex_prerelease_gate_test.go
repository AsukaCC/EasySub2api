//go:build unit

package service

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAcceptCodexClientVersionPrereleaseGate(t *testing.T) {
	SetCodexPrereleaseVersionAllowed(false)
	for _, version := range []string{"0.154.0", "0.154.0-alpha.3", "0.154.0-beta.1"} {
		require.Equal(t, version, AcceptCodexClientVersion(version))
	}
	require.Empty(t, AcceptCodexClientVersion("latest"))
	require.Empty(t, AcceptCodexClientVersion("0.100.0-alpha.1"))
	require.True(t, CodexPrereleaseVersionAllowed())
}
func TestGetOpenAICodexClientVersionAcceptsPrerelease(t *testing.T) {
	svc := NewSettingService(&codexVersionSettingRepoStub{values: map[string]string{
		SettingKeyOpenAICodexClientVersion:       "0.154.0-alpha.3",
		SettingKeyOpenAICodexClientVersionSynced: "0.153.4",
	}}, nil)
	require.Equal(t, "0.154.0-alpha.3", svc.GetOpenAICodexClientVersion(context.Background()))
}
func TestBuiltinCodexVersionIsStableRelease(t *testing.T) {
	require.Equal(t, "0.154.0", codexCLIVersion)
	require.False(t, IsCodexPrereleaseVersion(codexCLIVersion))
}
