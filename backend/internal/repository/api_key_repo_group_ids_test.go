package repository

import (
	"context"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestHydrateGroupIDsMutatesTheCallersAPIKey(t *testing.T) {
	primary := "g-openai"
	key := &service.APIKey{ID: "k-1", GroupID: &primary}
	repo := &apiKeyRepository{}

	require.NoError(t, repo.hydrateGroupIDs(context.Background(), []*service.APIKey{key}))
	require.Equal(t, []string{"g-openai"}, key.GroupIDs)
}

func TestHydrateGroupIDsDoesNotLeaveADiscardedCopyUnchanged(t *testing.T) {
	primary := "g-openai"
	key := &service.APIKey{ID: "k-1", GroupID: &primary}
	repo := &apiKeyRepository{}

	// The previous []APIKey{*key} call copied the struct; hydrate must be given
	// a pointer so GetByKey / GetByKeyForAuth actually observe bound groups.
	require.NoError(t, repo.hydrateGroupIDs(context.Background(), []*service.APIKey{key}))
	require.Equal(t, []string{primary}, key.BoundGroupIDs())
}
