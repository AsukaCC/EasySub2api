package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/config"
	"github.com/AsukaCC/EasySub2api/internal/pkg/tlsfingerprint"
	"github.com/stretchr/testify/require"
)

func TestCodexContextLimitsRoundTrip(t *testing.T) {
	for _, tc := range []struct {
		name, fields    string
		window, maximum int64
	}{
		{"distinct", `"context_window":272000,"max_context_window":872000`, 272000, 872000},
		{"legacy", `"context_window":272000`, 272000, 272000},
		{"maximum only", `"max_context_window":872000`, 872000, 872000},
		{"invalid maximum", `"context_window":272000,"max_context_window":-1`, 272000, 272000},
		{"clamped default", `"context_window":272000,"max_context_window":128000`, 128000, 128000},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var entry upstreamModelCapabilityEntry
			require.NoError(t, json.Unmarshal([]byte("{"+tc.fields+"}"), &entry))
			metadata := upstreamMetadataFromCapabilityEntry("custom-model", entry)
			body, err := json.Marshal(metadata)
			require.NoError(t, err)
			var restored UpstreamModelMetadata
			require.NoError(t, json.Unmarshal(body, &restored))
			descriptor := newConfiguredCodexModelDescriptor("custom-model")
			applyUpstreamModelMetadataToCodexDescriptor(&descriptor, intersectUpstreamModelMetadata("custom-model", []UpstreamModelMetadata{restored}))
			require.Equal(t, tc.window, descriptor.ContextWindow)
			require.Equal(t, tc.maximum, descriptor.MaxContextWindow)
		})
	}
}

func TestCodexContextLimitsIntersection(t *testing.T) {
	for _, tc := range []struct {
		name    string
		second  UpstreamModelMetadata
		maximum int64
	}{
		{"explicit", UpstreamModelMetadata{ContextWindow: 300000, MaxContextWindow: 512000}, 512000},
		{"legacy", UpstreamModelMetadata{ContextWindow: 300000}, 300000},
		{"unknown", UpstreamModelMetadata{}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			candidates := []UpstreamModelMetadata{{ContextWindow: 272000, MaxContextWindow: 872000}, tc.second}
			for range 2 {
				result := intersectUpstreamModelMetadata("custom-model", candidates)
				require.Equal(t, tc.maximum, result.MaxContextWindow)
				candidates[0], candidates[1] = candidates[1], candidates[0]
			}
		})
	}
}

func TestCodexContextLimitsRegistryAndCustomOverride(t *testing.T) {
	registry := UpstreamModelMetadata{ContextWindow: 1000000, MaxContextWindow: 1000000}
	direct := UpstreamModelMetadata{ContextWindow: 272000, MaxContextWindow: 872000}
	merged, _ := mergeUpstreamModelMetadata(direct, registry)
	require.Equal(t, direct.ContextWindow, merged.ContextWindow)
	require.Equal(t, direct.MaxContextWindow, merged.MaxContextWindow)
	merged, _ = mergeUpstreamModelMetadata(UpstreamModelMetadata{}, registry)
	require.Equal(t, registry.MaxContextWindow, merged.MaxContextWindow)
	descriptor := newConfiguredCodexModelDescriptor("gpt-6-astra")
	applyUpstreamModelMetadataToCodexDescriptor(&descriptor, codexModelMetadataOverride{UpstreamModelMetadata: direct})
	require.EqualValues(t, configuredCodexGPT6AstraContext, descriptor.MaxContextWindow)
}

type contextLimitsRepo struct {
	AccountRepository
	writes int
}

func (r *contextLimitsRepo) UpdateExtra(context.Context, string, map[string]any) error {
	r.writes++
	return nil
}

type contextLimitsTransport struct {
	HTTPUpstream
	status int
	body   string
}

func (u *contextLimitsTransport) DoWithTLS(*http.Request, string, string, int, *tlsfingerprint.Profile) (*http.Response, error) {
	return &http.Response{StatusCode: u.status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(u.body))}, nil
}

func TestCodexContextLimitsSyncPreservesPartialAndFailedSnapshots(t *testing.T) {
	account := &Account{ID: "account-1", Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
		Credentials: map[string]any{"api_key": "fixture", "base_url": "https://provider.example"}}
	reasoning := false
	account.SetUpstreamModelMetadataSnapshot(UpstreamModelMetadataSnapshot{Models: map[string]UpstreamModelMetadata{
		"custom-model": {ID: "custom-model", ContextWindow: 272000, MaxContextWindow: 872000, Reasoning: &reasoning, InputModalities: []string{"text"}},
	}})
	repo := &contextLimitsRepo{}
	transport := &contextLimitsTransport{status: 200, body: `{"data":[{"id":"custom-model","context_window":272000,"reasoning":false,"input_modalities":["text"]}]}`}
	svc := &AccountTestService{accountRepo: repo, httpUpstream: transport, cfg: &config.Config{}}
	for range 2 {
		_, err := svc.SyncUpstreamModelCatalog(context.Background(), account)
		require.NoError(t, err)
		metadata, ok := account.GetUpstreamModelMetadata("custom-model")
		require.True(t, ok)
		require.EqualValues(t, 872000, metadata.MaxContextWindow)
	}
	require.Equal(t, 2, repo.writes)
	transport.status = 503
	_, err := svc.SyncUpstreamModelCatalog(context.Background(), account)
	require.Error(t, err)
	require.Equal(t, 2, repo.writes)
	metadata, _ := account.GetUpstreamModelMetadata("custom-model")
	require.EqualValues(t, 872000, metadata.MaxContextWindow)
}
