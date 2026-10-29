//go:build unit

package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRelease0219ModelsRefreshReusesFreshResult(t *testing.T) {
	// A caller observed a miss, but a previous flight finished before its
	// refresh started. A missing HTTP client ensures a second fetch would fail.
	svc := &OpenAIGatewayService{}
	manifest := &CodexModelsManifest{Body: []byte(`{"models":[{"slug":"gpt-6.1-terra"}]}`)}
	svc.codexModelsManifestCache.set("shared-key", manifest, time.Now())
	select {
	case result := <-svc.refreshCachedAPIKeyCodexModelsManifest("shared-key", codexModelsManifestRequest{}):
		require.NoError(t, result.Err)
		require.Equal(t, manifest, result.Val)
	case <-time.After(2 * time.Second):
		t.Fatal("fresh manifest was not reused")
	}
}
