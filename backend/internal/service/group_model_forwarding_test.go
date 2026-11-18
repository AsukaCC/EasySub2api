//go:build unit

package service

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestGroupModelForwardingSingleHopAndValidation(t *testing.T) {
	rules, err := normalizeGroupModelForwarding(map[string]string{
		"gpt-5.4": "target-a", "old-*": "target-b", "old-specific-*": "target-c",
		"old-exact": "target-d", "target-d": "target-e",
	})
	require.NoError(t, err)
	group := &Group{Platform: PlatformOpenAI, ModelForwarding: rules}
	for source, want := range map[string]string{
		"gpt-5.4": "target-a", "old-other": "target-b", "old-specific-name": "target-c", "old-exact": "target-d",
	} {
		got, matched := group.ResolveForwardedModel(source)
		require.True(t, matched)
		require.Equal(t, want, got)
	}
	for _, invalid := range []map[string]string{
		{"a": "gpt-5.4"}, {"a": "target-*"}, {"a*b": "target"}, {"a": "x", " a ": "y"},
	} {
		_, err := normalizeGroupModelForwarding(invalid)
		require.Error(t, err)
	}
}

func TestGroupModelForwardingFiltersCodexManifest(t *testing.T) {
	body := []byte(`{"models":[{"slug":"target","extra":1},{"slug":"old-one"}],"other":"keep"}`)
	group := &Group{Platform: PlatformOpenAI, ModelForwarding: map[string]string{"old-*": "target"}}
	for _, custom := range []bool{false, true} {
		group.ModelsListConfig = GroupModelsListConfig{Enabled: custom, Models: []string{"old-one", "old-two", "target"}}
		got, changed, err := mergeConfiguredCodexModelsManifest(body, []string{"old-two"}, group)
		require.NoError(t, err)
		require.True(t, changed)
		require.JSONEq(t, `{"models":[{"slug":"target","extra":1}],"other":"keep"}`, string(got))
		require.Len(t, gjson.GetBytes(body, "models").Array(), 2)
	}
}
