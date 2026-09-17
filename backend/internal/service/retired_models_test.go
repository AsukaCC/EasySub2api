package service

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/pkg/openai"
	"github.com/stretchr/testify/require"
)

func TestRetiredModelsRejectPassthroughAndMappings(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		a := &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Extra: map[string]any{"openai_passthrough": passthrough},
			Credentials: map[string]any{"model_mapping": map[string]any{"old-alias": "gpt-5.5-pro", "gpt-5.4": "gpt-5.6-sol", "active": "gpt-5.6-sol"}}}
		require.False(t, a.IsModelSupported("gpt-5.4"))
		if !passthrough {
			require.False(t, a.IsModelSupported("old-alias"))
			require.ErrorIs(t, CheckActiveAccountModel(a, "old-alias"), ErrModelRetired)
		} else {
			require.NoError(t, CheckActiveAccountModel(a, "old-alias"), "passthrough ignores stale normal mappings")
		}
		require.NoError(t, CheckActiveAccountModel(a, "active"))
		for _, body := range []string{`{"model":"gpt5.4nano"}`, `{"type":"session.update","session":{"model":"gpt-5.5"}}`, `{"type":"response.create","response":{"model":"gpt-5.5-pro"}}`} {
			require.ErrorIs(t, checkActivePayloadModels(a, []byte(body)), ErrModelRetired)
		}
	}
}

func TestRetiredModelsCannotBeNormalizedIntoActiveModels(t *testing.T) {
	for _, model := range []string{"gpt-5.4", "gpt-5.4-mini", "openai/gpt5.5-pro", "gpt-5.5-2026-09-01"} {
		require.Empty(t, normalizeKnownOpenAICodexModel(model))
		mapped, known := normalizeKnownCodexModel(model)
		require.False(t, known)
		require.Empty(t, mapped)
		require.Equal(t, model, normalizeCodexModel(model))
	}
	for name, target := range codexModelMap {
		require.False(t, openai.IsRetiredModel(name))
		require.False(t, openai.IsRetiredModel(target))
	}
	require.Equal(t, openai.DefaultTestModel, normalizeCodexModel(""))
}

func TestRetiredModelsRejectBeforeUpstreamConstruction(t *testing.T) {
	s := &OpenAIGatewayService{}
	a := &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
	body := []byte(`{"model":"gpt-5.4-mini","input":"test"}`)
	_, err := s.Forward(context.Background(), nil, a, body)
	require.ErrorIs(t, err, ErrModelRetired)
	_, err = s.buildUpstreamRequest(context.Background(), nil, a, body, "", false, "", false)
	require.ErrorIs(t, err, ErrModelRetired)
	_, err = s.buildUpstreamRequestOpenAIPassthrough(context.Background(), nil, a, body, "")
	require.ErrorIs(t, err, ErrModelRetired)
}

func TestRetiredModelsHiddenFromManifestAndFreshETag(t *testing.T) {
	body := `{"models":[{"slug":"gpt-5.4-mini"},{"slug":"gpt-5.6-sol","custom":true}]}`
	s := &OpenAIGatewayService{httpUpstream: &codexModelsHTTPUpstreamStub{do: func(*http.Request, string, string, int) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: http.Header{"Etag": []string{`"before-retirement"`}}, Body: io.NopCloser(strings.NewReader(body))}, nil
	}}}
	m, err := s.fetchCodexModelsManifestUpstream(context.Background(), codexModelsManifestRequest{
		url: "https://upstream.example/models", headers: http.Header{}, useAPIKeyUpstream: true,
		credentialAccount: &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey},
	}, "")
	require.NoError(t, err)
	require.NotContains(t, string(m.Body), "gpt-5.4")
	require.Contains(t, string(m.Body), "gpt-5.6-sol")
	require.NotEqual(t, `"before-retirement"`, m.ETag)
	require.False(t, codexModelsManifestForClient(m, `"before-retirement"`).NotModified)
	require.True(t, codexModelsManifestForClient(m, m.ETag).NotModified)
	generated, err := BuildCodexModelsManifest([]string{"gpt-5.4", "gpt-5.5-pro", "gpt-5.6-sol"})
	require.NoError(t, err)
	require.NotContains(t, string(generated), "gpt-5.4")
	require.NotContains(t, string(generated), "gpt-5.5")
}

func TestRetiredHistoricalPricingRemainsAvailable(t *testing.T) {
	s := NewBillingService(nil, nil)
	for _, model := range []string{"gpt-5.4", "openai/gpt5.4", "gpt5.5", "gpt-5.4-mini", "gpt-5.4-nano"} {
		pricing, err := s.GetModelPricing(model)
		require.NoError(t, err)
		require.Positive(t, pricing.InputPricePerToken)
		require.ErrorIs(t, CheckActiveModel(model), ErrModelRetired)
	}
	pricing, err := s.GetModelPricing("gpt5.5")
	require.NoError(t, err)
	require.InDelta(t, pricing.InputPricePerToken*2.5, pricing.InputPricePerTokenPriority, 1e-12)
	require.InDelta(t, pricing.OutputPricePerToken*2.5, pricing.OutputPricePerTokenPriority, 1e-12)
}

func TestRetiredModelsRejectBeforeSchedulerDependencies(t *testing.T) {
	ctx := context.Background()
	gateway := &GatewayService{}
	_, err := gateway.SelectAccountForModel(ctx, nil, "", "gpt-5.4")
	require.ErrorIs(t, err, ErrModelRetired)
	_, err = gateway.SelectAccountWithLoadAwareness(ctx, nil, "", "gpt55-pro", nil, "", "")
	require.ErrorIs(t, err, ErrModelRetired)
	openAI := &OpenAIGatewayService{}
	_, _, err = openAI.SelectAccountWithScheduler(ctx, nil, "", "", "openai/gpt5.4-mini", nil, OpenAIUpstreamTransportAny, false)
	require.ErrorIs(t, err, ErrModelRetired)
	require.ErrorIs(t, ValidateLiveCallRequest(&LiveCallRequest{SDP: "test", Session: []byte(`{"model":"gpt-5.5-pro"}`)}), ErrModelRetired)
}

func TestRetiredPublicModelCannotBeReintroducedByCompositeRewrite(t *testing.T) {
	ctx := WithCompositeRouteDecision(context.Background(), CompositeRouteDecision{
		Matched: true, PublicModel: "gpt-5.4-mini", UpstreamModel: "gpt-5.6-sol", TargetPlatform: PlatformOpenAI,
	})
	require.ErrorIs(t, CheckActiveRequestModel(ctx, "gpt-5.6-sol"), ErrModelRetired)
	ctx = WithCompositeRouteDecision(context.Background(), CompositeRouteDecision{
		Matched: true, PublicModel: "active-alias", UpstreamModel: "gpt-5.5-pro", TargetPlatform: PlatformOpenAI,
	})
	require.ErrorIs(t, CheckActiveRequestModel(ctx, "active-alias"), ErrModelRetired)
}
