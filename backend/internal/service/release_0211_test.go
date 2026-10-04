//go:build unit

package service

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGPT61ValidationAndAliases(t *testing.T) {
	require.Equal(t, "ultra", normalizeOpenAIReasoningEffortForModel("ULTRA", "gpt-6.1-sol"))
	_, _, err := ApplyOpenAIReasoningEffortPolicy([]byte(`{"model":"gpt-6.1-sol","reasoning":{"effort":"ultra"}}`), "max", nil, ReasoningEffortOverLimitDeny)
	require.Error(t, err, "ultra must not bypass a group's reasoning ceiling")
	_, err = applyOpenAIWSReasoningEffortPolicyForModel([]byte(`{"model":"gpt-6.1-sol","reasoning":{"effort":"none"}}`), nil, "")
	require.Error(t, err)
	for _, model := range []string{"gpt-6.1-sol", "openai/gpt-6.1-sol", "gpt61-sol", "gpt6.1-sol", "gpt-6.1-sol-2026-09-25"} {
		for _, field := range []string{`"reasoning":{"effort":"none"}`, `"reasoning_effort":"minimal"`, `"output_config":{"effort":"none"}`, `"thinking":{"type":"disabled"}`} {
			require.Error(t, validateGPT61SolReasoningEffort([]byte(`{"model":"`+model+`",`+field+`}`)), model)
		}
		require.Equal(t, "gpt-6.1-sol", normalizeCodexModel(model))
	}
	for _, effort := range []string{"low", "medium", "high", "xhigh", "max", "ultra"} {
		require.NoError(t, validateGPT61SolReasoningEffort([]byte(`{"model":"gpt-6.1-sol","reasoning":{"effort":"`+effort+`"}}`)))
	}
	require.Error(t, validateGPT61SolReasoningEffort([]byte(`{"model":"alias","reasoning_effort":"none"}`), "gpt-6.1-sol"))
	require.Error(t, validateGPT61SolReasoningEffort([]byte(`{"model":"gpt-6.1-sol-minimal"}`)))
	require.NoError(t, validateGPT61SolReasoningEffort([]byte(`{"model":"gpt-5.6-sol","reasoning_effort":"none"}`)))
}

func TestGPT61CompatibilityReturnsBadRequest(t *testing.T) {
	svc := &OpenAIGatewayService{}
	for _, protocol := range []string{"chat", "messages", "raw"} {
		t.Run(protocol, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			body := []byte(`{"model":"gpt-6.1-sol","reasoning_effort":"none"}`)
			var err error
			switch protocol {
			case "chat":
				_, err = svc.ForwardAsChatCompletions(context.Background(), c, nil, body, "", "")
			case "messages":
				_, err = svc.ForwardAsAnthropic(context.Background(), c, nil, body, "", "")
			case "raw":
				_, err = svc.forwardAsRawChatCompletions(context.Background(), c, nil, body, "")
			}
			require.Error(t, err)
			require.Equal(t, 400, w.Code)
			require.Contains(t, w.Body.String(), "invalid_request_error")
		})
	}
}

func TestRelease0211Pricing(t *testing.T) {
	svc := NewBillingService(&config.Config{}, nil)
	tokens := UsageTokens{InputTokens: 1000, OutputTokens: 100, CacheReadTokens: 100, CacheCreationTokens: 100}
	for _, model := range []string{"gpt-6-astra", "gpt-6.1-sol"} {
		base, err := svc.CalculateCostWithServiceTier(model, tokens, 1, "")
		require.NoError(t, err)
		fast, err := svc.CalculateCostWithServiceTier(model, tokens, 1, "priority")
		require.NoError(t, err)
		require.InDelta(t, base.ActualCost*2, fast.ActualCost, 1e-10)
		if model == "gpt-6-astra" {
			ultra, err := svc.CalculateCostWithServiceTier(model, tokens, 1, "ultrafast")
			require.NoError(t, err)
			require.InDelta(t, base.ActualCost*6, ultra.ActualCost, 1e-10)
			// Group token prices must also retain Astra's Ultrafast tier.
			in, out := 1e-6, 2e-6
			group := &Group{ModelPricing: []ChannelModelPricing{{Models: []string{model}, BillingMode: BillingModeToken, InputPrice: &in, OutputPrice: &out}}}
			cost, err := svc.CalculateCostUnified(CostInput{Ctx: context.Background(), Model: model, Group: group, Tokens: UsageTokens{InputTokens: 1000, OutputTokens: 100}, RateMultiplier: 2, ServiceTier: "ultrafast", Resolver: NewModelPricingResolver(nil, svc)})
			require.NoError(t, err)
			require.InDelta(t, (1000*in+100*out)*6*2, cost.ActualCost, 1e-10)
		}
	}
	p, err := svc.GetModelPricing("gpt-6.1-sol")
	require.NoError(t, err)
	require.Equal(t, 272000, p.LongContextInputThreshold)
	d := newConfiguredCodexModelDescriptor("gpt-6.1-sol")
	require.Equal(t, int64(272000), d.ContextWindow)
	require.Equal(t, int64(872000), d.MaxContextWindow)
	require.Len(t, d.SupportedReasoningLevels, 6)
	disabled := false
	applyUpstreamModelMetadataToCodexDescriptor(&d, codexModelMetadataOverride{ID: "gpt-6.1-sol", Reasoning: &disabled})
	require.Len(t, d.SupportedReasoningLevels, 6)
	require.Equal(t, "low", *d.DefaultReasoningLevel)
	caps := accountCodexToolCapabilities(&Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey}, "gpt-6.1-sol")
	require.JSONEq(t, "false", string(caps["use_responses_lite"]))
}

func TestInflightEstimateAstraTierMatchesBilling(t *testing.T) {
	svc := newInflightEstimateGateway(t, nil)
	key := &APIKey{User: &User{ID: "user"}}
	base, ok := svc.EstimateInflightReservation(context.Background(), key, InflightEstimateRequest{Model: "gpt-6-astra", BodyBytes: 4000, MaxTokens: 1000})
	require.True(t, ok)
	ultra, ok := svc.EstimateInflightReservation(context.Background(), key, InflightEstimateRequest{Model: "gpt-6-astra", BodyBytes: 4000, MaxTokens: 1000, ServiceTier: "ultrafast"})
	require.True(t, ok)
	require.InDelta(t, base*6, ultra, 1e-10)
}

type releaseKeyCountRepo struct {
	APIKeyRepository
	count int64
	err   error
}

func (r releaseKeyCountRepo) CountByUserID(context.Context, string) (int64, error) {
	return r.count, r.err
}

type releaseKeyCounter struct {
	APIKeyCache
	count int64
	err   error
}

func (r releaseKeyCounter) IncrementCreateCount(context.Context, string, time.Duration) (int64, error) {
	return r.count, r.err
}

func TestAPIKeyCreateLimits(t *testing.T) {
	cfg := &config.Config{}
	cfg.APIKeyCreate.MaxActivePerUser = 200
	cfg.APIKeyCreate.MaxPerUserPerHour = 60
	svc := &APIKeyService{cfg: cfg, apiKeyRepo: releaseKeyCountRepo{count: 200}}
	require.ErrorIs(t, svc.checkAPIKeyCreateLimits(context.Background(), "user"), ErrAPIKeyCountExceeded)
	svc.apiKeyRepo = releaseKeyCountRepo{count: 199}
	svc.cache = releaseKeyCounter{count: 61}
	require.ErrorIs(t, svc.checkAPIKeyCreateLimits(context.Background(), "user"), ErrAPIKeyCreateLimited)
	svc.cache = releaseKeyCounter{count: 60}
	require.NoError(t, svc.checkAPIKeyCreateLimits(context.Background(), "user"))
	svc.cache = releaseKeyCounter{err: errors.New("redis unavailable")}
	require.NoError(t, svc.checkAPIKeyCreateLimits(context.Background(), "user"))
	svc.apiKeyRepo = releaseKeyCountRepo{err: errors.New("db unavailable")}
	require.Error(t, svc.checkAPIKeyCreateLimits(context.Background(), "user"))
	cfg.APIKeyCreate.MaxActivePerUser = 0
	cfg.APIKeyCreate.MaxPerUserPerHour = 0
	require.NoError(t, svc.checkAPIKeyCreateLimits(context.Background(), "user"))
}
