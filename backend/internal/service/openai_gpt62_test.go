//go:build unit

package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/pkg/apicompat"
	"github.com/AsukaCC/EasySub2api/internal/pkg/openai"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestGPT62IdentityAndCapabilities(t *testing.T) {
	for _, base := range []string{"gpt-6.2-astra", "gpt-6.2-sol"} {
		require.Contains(t, openai.DefaultModelIDs(), base)
		for _, suffix := range []string{"", "-max", "-openai-compact", "-2026-11-17"} {
			model := "openai/" + base + suffix
			require.Equal(t, base, normalizeCodexModel(model))
			require.Equal(t, base, NormalizeOpenAICompatRequestedModel(model))
			require.Equal(t, base, normalizeModelNameForPricing(model))
			require.True(t, openai.RequiresResponsesForToolCalls(model))
			require.True(t, openAIModelSupportsUltrafast(model))
			require.ErrorContains(t, validateGPT61SolChatOnlyRequest([]byte(`{}`), model, true), "requires Responses")
		}
		for _, body := range []string{`{"reasoning":{"effort":"none"}}`, `{"reasoning_effort":"minimal"}`, `{"thinking":{"type":"disabled"}}`, `{"output_config":{"effort":"none"}}`} {
			require.Error(t, validateGPT61SolReasoningEffort([]byte(body), base))
			require.Error(t, validateGPT61CompatRequest([]byte(body), base))
		}
		require.Error(t, validateGPT61SolReasoningEffort([]byte(`{}`), base+"-none"))
		_, err := apicompat.ChatCompletionsToResponses(&apicompat.ChatCompletionsRequest{Model: base, ReasoningEffort: "none"})
		require.Error(t, err)
		request := &apicompat.AnthropicRequest{Model: base + "-max"}
		applyOpenAICompatModelNormalization(request)
		converted, err := apicompat.AnthropicToResponses(request)
		require.NoError(t, err)
		require.Equal(t, "max", converted.Reasoning.Effort)
		descriptor := newConfiguredCodexModelDescriptor(base)
		require.Equal(t, "medium", *descriptor.DefaultReasoningLevel)
		require.Equal(t, int64(1_050_000), descriptor.ContextWindow)
		require.Equal(t, []string{"text", "image"}, descriptor.InputModalities)
		require.Equal(t, reasoningLevels("low", "medium", "high", "xhigh", "max"), descriptor.SupportedReasoningLevels)
		require.Len(t, descriptor.ServiceTiers, 2)
	}
	require.Equal(t, "gpt-6.2-sol", normalizeCodexModel("openai/GPT62_SOL-max"))
	for _, model := range []string{"gpt-6.2-sol-pro", "gpt-6.2-astra-unknown", "gpt-6.2-terra", "gpt-6.1-sol", "gpt-6-astra"} {
		require.False(t, openai.IsGPT62ModelSpelling(model))
	}
	prices := &PricingService{pricingData: map[string]*LiteLLMModelPricing{}}
	require.Nil(t, prices.GetModelPricing("gpt-6.2-sol-pro"))
}

func TestGPT62OfficialBilling(t *testing.T) {
	data, err := os.ReadFile("../../resources/model-pricing/model_prices_and_context_window.json")
	require.NoError(t, err)
	var catalog map[string]*LiteLLMModelPricing
	require.NoError(t, json.Unmarshal(data, &catalog))
	for _, dynamic := range []bool{false, true} {
		var pricingService *PricingService
		if dynamic {
			pricingService = &PricingService{pricingData: catalog}
		}
		billing := NewBillingService(nil, pricingService)
		for _, model := range []string{"gpt-6.2-sol", "gpt-6.2-astra"} {
			unit := 1.0
			if model == "gpt-6.2-astra" {
				unit = 4
			}
			for _, resolver := range []*ModelPricingResolver{nil, NewModelPricingResolver(nil, billing)} {
				for _, count := range []int{272000, 272001} {
					for _, tier := range []struct {
						name       string
						at         time.Time
						multiplier float64
					}{
						{"default", openAIGPT62UltrafastPromotionEnd, 1},
						{"priority", openAIGPT62UltrafastPromotionEnd, 2},
						{"ultrafast", openAIGPT62UltrafastPromotionEnd.Add(-time.Nanosecond), 3},
						{"ultrafast", openAIGPT62UltrafastPromotionEnd, 6},
					} {
						tokens := UsageTokens{InputTokens: count - 4000, CacheReadTokens: 1000, CacheCreationTokens: 3000, CacheCreation30mTokens: 1000, CacheCreation12hTokens: 2000, OutputTokens: 1000}
						cost, err := billing.CalculateCostUnified(CostInput{
							Ctx: context.Background(), Model: "openai/" + model + "-max", Tokens: tokens,
							RateMultiplier: 0.4, PricingAt: tier.at, ServiceTier: tier.name, Resolver: resolver,
						})
						require.NoError(t, err)
						inputFactor, outputFactor := 1.0, 1.0
						if count > 272000 {
							inputFactor, outputFactor = 2, 1.5
						}
						require.InDelta(t, float64(count-4000)*2e-6*unit*inputFactor*tier.multiplier, cost.InputCost, 1e-10)
						require.InDelta(t, 0.0001*unit*inputFactor*tier.multiplier, cost.CacheReadCost, 1e-10)
						require.InDelta(t, 0.0125*unit*inputFactor*tier.multiplier, cost.CacheCreationCost, 1e-10)
						require.InDelta(t, 0.01*unit*outputFactor*tier.multiplier, cost.OutputCost, 1e-10)
						require.InDelta(t, cost.TotalCost*0.4, cost.ActualCost, 1e-8)
					}
				}
			}
		}
	}
}

func TestGPT62CustomPricesAndLegacyRates(t *testing.T) {
	billing := NewBillingService(nil, nil)
	zero := 0.0
	for _, model := range []string{"gpt-6.2-sol", "gpt-6.2-astra"} {
		p, err := billing.GetModelPricingWithChannel(model, &ChannelModelPricing{InputPrice: &zero, OutputPrice: &zero, CacheReadPrice: &zero, CacheWritePrice: &zero, CacheWrite12hPrice: &zero})
		require.NoError(t, err)
		for _, tier := range []string{"default", "priority", "ultrafast"} {
			cost := billing.computeTokenBreakdown(p, UsageTokens{InputTokens: 300000, OutputTokens: 100, CacheReadTokens: 100, CacheCreationTokens: 100, CacheCreation12hTokens: 100}, 1, tier, true)
			require.Zero(t, cost.TotalCost)
		}
	}
	for _, tc := range []struct {
		model string
		input float64
	}{{"gpt-6-astra", 10e-6}, {"gpt-6.1-sol", 2e-6}} {
		p, err := billing.GetModelPricing(tc.model)
		require.NoError(t, err)
		require.Equal(t, tc.input, p.InputPricePerToken)
		require.Equal(t, float64(6), configuredServiceTierMultiplier("ultrafast", p, openAIGPT62UltrafastPromotionEnd.Add(-time.Hour)))
	}
}

func TestGPT62ResponsesImageToolForwarding(t *testing.T) {
	for _, model := range []string{"gpt-6.2-sol", "gpt-6.2-astra"} {
		for _, explicitImageTool := range []bool{false, true} {
			upstream := &httpUpstreamRecorder{resp: &http.Response{
				StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"application/json"}},
				Body: io.NopCloser(strings.NewReader(`{"id":"resp_62","model":"` + model + `","output":[],"usage":{"input_tokens":1,"output_tokens":1}}`)),
			}}
			svc := newOpenAIImageGenerationControlTestService(upstream)
			c, _ := newOpenAIImageGenerationControlTestContext(true, "codex_cli_rs/0.169.0")
			account := newOpenAIImageGenerationControlTestAccount()
			account.Type = AccountTypeAPIKey
			payload := map[string]any{"model": model, "input": "draw a square", "stream": false, "reasoning": map[string]any{"effort": "high"}}
			if explicitImageTool {
				payload["tools"] = []any{map[string]any{"type": "image_generation", "model": "gpt-image-2", "quality": "low"}}
				c.Request.Header.Set(responsesLiteHeader, "true")
			}
			body, err := json.Marshal(payload)
			require.NoError(t, err)
			_, err = svc.Forward(context.Background(), c, account, body)
			require.NoError(t, err)
			require.NotNil(t, upstream.lastReq)
			require.Equal(t, model, gjson.GetBytes(upstream.lastBody, "model").String())
			require.Equal(t, explicitImageTool, hasNativeImageGenerationTool(upstream.lastBody))
			if explicitImageTool {
				require.Empty(t, upstream.lastReq.Header.Get(responsesLiteHeader))
				require.Equal(t, "gpt-image-2", gjson.GetBytes(upstream.lastBody, `tools.#(type=="image_generation").model`).String())
			}
		}
	}
}
