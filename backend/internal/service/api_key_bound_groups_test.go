package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAPIKeyBoundGroupIDsPrefersExplicitGroupIDs(t *testing.T) {
	primary := "g-openai"
	key := &APIKey{
		GroupID:  &primary,
		GroupIDs: []string{"g-openai", "g-grok"},
		Group:    &Group{ID: primary, Platform: PlatformOpenAI},
	}
	require.Equal(t, []string{"g-openai", "g-grok"}, key.BoundGroupIDs())
	require.True(t, key.HasMultipleBoundGroups())
}

func TestAPIKeyBoundGroupIDsFallsBackToPrimaryGroup(t *testing.T) {
	primary := "g-openai"
	key := &APIKey{
		GroupID: &primary,
		Group:   &Group{ID: primary, Platform: PlatformOpenAI},
	}
	require.Equal(t, []string{"g-openai"}, key.BoundGroupIDs())
	require.False(t, key.HasMultipleBoundGroups())
}
