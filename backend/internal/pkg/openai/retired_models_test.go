package openai

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRetiredGPTFamilies(t *testing.T) {
	for _, model := range []string{
		"gpt-5.4", "gpt-5.5", "gpt-5.4-mini", "gpt-5.4-nano", "gpt-5.5-pro", "gpt54-mini", "gpt55pro",
		"gpt-5.4-2026-03-05", "gpt-5.5-pro-2026-09-01", "openai/gpt5.4mini", "vendor/team/GPT5.5",
		" GPT_5_4_NANO ", "gpt--5.4-high", "gpt 5.5 pro", "gpt-5.4[1m]", "gpt-5.5:latest",
	} {
		require.True(t, IsRetiredModel(model), model)
	}
	for _, model := range []string{"gpt-5.6-sol", "gpt-5.6-luna", "gpt-6-astra", "gpt-5.3-codex-spark", "gpt-5.2", "gpt-image-2.5-flare", "gpt-5.40", "gpt-5.50", "claude-opus-4-8", "my-model", ""} {
		require.False(t, IsRetiredModel(model), model)
	}
	for _, model := range DefaultModelIDs() {
		require.False(t, IsRetiredModel(model), model)
	}
	require.False(t, IsRetiredModel(DefaultTestModel))
}

func TestRetiredManifestFilteringPreservesMetadata(t *testing.T) {
	input := []byte(`{"models":[{"slug":"gpt-5.5-pro","unknown":1},{"slug":"gpt-5.6-sol","unknown":{"x":[1,2]}}],"data":[{"id":"openai/gpt5.4nano"},{"id":"gpt-image-2"}],"cursor":"retained"}`)
	body, changed, err := FilterRetiredManifest(input)
	require.NoError(t, err)
	require.True(t, changed)
	var parsed struct {
		Models []struct {
			Slug    string
			Unknown map[string]any
		}
		Data   []struct{ ID string }
		Cursor string
	}
	require.NoError(t, json.Unmarshal(body, &parsed))
	require.Len(t, parsed.Models, 1)
	require.Equal(t, "gpt-5.6-sol", parsed.Models[0].Slug)
	require.Contains(t, parsed.Models[0].Unknown, "x")
	require.Len(t, parsed.Data, 1)
	require.Equal(t, "retained", parsed.Cursor)
	next, changed, err := FilterRetiredManifest(body)
	require.NoError(t, err)
	require.False(t, changed)
	require.Equal(t, body, next)
}
