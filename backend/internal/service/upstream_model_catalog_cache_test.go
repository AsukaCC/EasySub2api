//go:build unit

package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/pkg/tlsfingerprint"
	"github.com/stretchr/testify/require"
)

type modelCatalogTransport struct {
	HTTPUpstream
	mu     sync.Mutex
	calls  int
	status int
	body   string
}

func (u *modelCatalogTransport) DoWithTLS(req *http.Request, _ string, _ string, _ int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	if req.URL.Host == "models.dev" {
		return newJSONResponse(200, `{}`), nil
	}
	u.calls++
	return newJSONResponse(u.status, u.body), nil
}

func TestUpstreamModelCatalogPersistsAcrossServicesAndKeyChanges(t *testing.T) {
	svc, repo, _ := newFingerprintTestService()
	transport := &modelCatalogTransport{status: 200, body: `{"data":[{"id":"gpt-6-astra","supported_reasoning_levels":["low","high"]},{"id":"gpt-image-2"}]}`}
	svc.httpUpstream = transport
	models, err := svc.GetFingerprintModelsForKey(context.Background(), "one", "admin", "local-key")
	require.NoError(t, err)
	require.Len(t, models, 1)
	require.Equal(t, []string{"low", "high"}, models[0].ReasoningLevels)
	account, err := repo.GetByID(context.Background(), "one")
	require.NoError(t, err)
	require.Len(t, cachedUpstreamModelCatalog(account).Models, 2)
	encoded, err := json.Marshal(account.Extra[UpstreamModelCatalogExtraKey])
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "test-only")

	// A new service instance represents a restart, without an in-memory catalog.
	restarted := &AccountTestService{accountRepo: repo, httpUpstream: transport, modelFingerprintKeys: svc.modelFingerprintKeys, modelFingerprintGroups: svc.modelFingerprintGroups}
	got, err := restarted.GetFingerprintModelsForKey(context.Background(), "one", "admin", "local-key")
	require.NoError(t, err)
	require.Equal(t, models, got)
	require.Equal(t, 1, transport.calls)
	svc.modelFingerprintKeys.(*fingerprintKeyRepo).keys[0].Status = StatusAPIKeyDisabled
	_, err = restarted.GetFingerprintModelsForKey(context.Background(), "one", "admin", "local-key")
	require.Error(t, err, "cached models must still require an active selected key")
	require.Equal(t, 1, transport.calls)
}

func TestUpstreamModelCatalogSyncRefreshesIncompleteModelsAndPreservesOnFailure(t *testing.T) {
	svc, repo, _ := newFingerprintTestService()
	transport := &modelCatalogTransport{status: 200, body: `{"data":[{"id":"custom-one"}]}`}
	svc.httpUpstream = transport
	account, _ := repo.GetByID(context.Background(), "one")
	catalog, err := svc.SyncUpstreamModelCatalog(context.Background(), account)
	require.NoError(t, err)
	require.NotEmpty(t, catalog.Warnings)
	account, _ = repo.GetByID(context.Background(), "one")
	require.Equal(t, []string{"custom-one"}, cachedUpstreamModelCatalog(account).Models)
	_, err = svc.FetchOpenAIAccountModels(context.Background(), account)
	require.NoError(t, err)
	require.Equal(t, 1, transport.calls)

	transport.body = `{"data":[{"id":"custom-two"}]}`
	_, err = svc.SyncUpstreamModelCatalog(context.Background(), account)
	require.NoError(t, err)
	account, _ = repo.GetByID(context.Background(), "one")
	require.Equal(t, []string{"custom-two"}, cachedUpstreamModelCatalog(account).Models)
	require.Equal(t, 2, transport.calls, "explicit synchronization must bypass the cache")
	transport.status = 503
	_, err = svc.SyncUpstreamModelCatalog(context.Background(), account)
	require.Error(t, err)
	account, _ = repo.GetByID(context.Background(), "one")
	require.Equal(t, []string{"custom-two"}, cachedUpstreamModelCatalog(account).Models)
}

func TestUpstreamModelCatalogInvalidatesConnectionChanges(t *testing.T) {
	svc, repo, _ := newFingerprintTestService()
	account, _ := repo.GetByID(context.Background(), "one")
	require.NoError(t, svc.saveUpstreamModelCatalog(context.Background(), account, &UpstreamModelCatalog{Models: []string{"gpt-6-astra"}}, false, nil))
	for _, field := range []string{"base_url", "api_key", "api_protocol", "header_overrides"} {
		t.Run(field, func(t *testing.T) {
			copy, _ := repo.GetByID(context.Background(), "one")
			copy.Credentials[field] = "changed"
			require.Nil(t, cachedUpstreamModelCatalog(copy))
		})
	}
	account, _ = repo.GetByID(context.Background(), "one")
	account.Platform = PlatformAnthropic
	require.Nil(t, cachedUpstreamModelCatalog(account))
	account, _ = repo.GetByID(context.Background(), "one")
	account.Type = AccountTypeOAuth
	require.Nil(t, cachedUpstreamModelCatalog(account))
	account, _ = repo.GetByID(context.Background(), "one")
	account.Credentials["model_mapping"] = map[string]any{"alias": "gpt-6-astra"}
	require.NotNil(t, cachedUpstreamModelCatalog(account), "allowlist edits must not invalidate a live upstream catalog")
}

func TestUpstreamModelCatalogConcurrentMissAndAccountIsolation(t *testing.T) {
	svc, repo, _ := newFingerprintTestService()
	transport := &modelCatalogTransport{status: 200, body: `{"data":[{"id":"gpt-6-astra"}]}`}
	svc.httpUpstream = transport
	var wg sync.WaitGroup
	results := make(chan error, 12)
	for range 12 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			account, err := repo.GetByID(context.Background(), "one")
			if err == nil {
				_, err = svc.accountModelCatalog(context.Background(), account)
			}
			results <- err
		}()
	}
	wg.Wait()
	close(results)
	for err := range results {
		require.NoError(t, err)
	}
	require.Equal(t, 1, transport.calls)
	account, _ := repo.GetByID(context.Background(), "two")
	_, err := svc.accountModelCatalog(context.Background(), account)
	require.NoError(t, err)
	require.Equal(t, 2, transport.calls)
}

func TestUpstreamModelCatalogConfiguredFallbackTracksMapping(t *testing.T) {
	svc, repo, _ := newFingerprintTestService()
	repo.mapping = map[string]any{"alias": "custom-one"}
	transport := &modelCatalogTransport{status: 404, body: `{}`}
	svc.httpUpstream = transport
	account, _ := repo.GetByID(context.Background(), "one")
	catalog, err := svc.accountModelCatalog(context.Background(), account)
	require.NoError(t, err)
	require.Equal(t, []string{"custom-one"}, catalog.Models)
	account, _ = repo.GetByID(context.Background(), "one")
	require.NotNil(t, cachedUpstreamModelCatalog(account))
	repo.mapping = map[string]any{"alias": "custom-two"}
	account, _ = repo.GetByID(context.Background(), "one")
	require.Nil(t, cachedUpstreamModelCatalog(account))
	catalog, err = svc.accountModelCatalog(context.Background(), account)
	require.NoError(t, err)
	require.Equal(t, []string{"custom-two"}, catalog.Models)
	require.Equal(t, 2, transport.calls)
}

type failingCatalogRepo struct{ *fingerprintTestRepo }

func (r *failingCatalogRepo) UpdateExtra(context.Context, string, map[string]any) error {
	return errors.New("storage offline")
}

func TestUpstreamModelCatalogStorageFailureIsNotReportedAsSuccess(t *testing.T) {
	svc, repo, _ := newFingerprintTestService()
	svc.accountRepo = &failingCatalogRepo{repo}
	svc.httpUpstream = &modelCatalogTransport{status: 200, body: `{"data":[{"id":"gpt-6-astra"}]}`}
	account, _ := repo.GetByID(context.Background(), "one")
	_, err := svc.accountModelCatalog(context.Background(), account)
	require.Error(t, err)
	var syncErr *UpstreamModelSyncError
	require.ErrorAs(t, err, &syncErr)
	require.Equal(t, UpstreamModelSyncErrorInternal, syncErr.Kind)
	account, _ = repo.GetByID(context.Background(), "one")
	require.Nil(t, cachedUpstreamModelCatalog(account))
}
