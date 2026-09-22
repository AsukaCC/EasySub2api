package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/config"
	"github.com/AsukaCC/EasySub2api/internal/pkg/openai"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestGPT6SolLunaCatalogAndAliases(t *testing.T) {
	for _, model := range []string{"gpt-6-sol", "gpt-6-luna"} {
		t.Run(model, func(t *testing.T) {
			require.Contains(t, openai.DefaultModelIDs(), model)
			for _, alias := range []string{model, "openai/" + model, strings.ToUpper(model), model + "-max", model + "-2026-09-22", model + "-openai-compact"} {
				require.Equal(t, model, normalizeCodexModel(alias))
				require.Equal(t, model, normalizeModelNameForPricing(alias))
				require.False(t, isOpenAIGPT6AstraModel(alias))
				require.Equal(t, "max", normalizeOpenAIReasoningEffortForModel("max", alias))
			}
			d := newConfiguredCodexModelDescriptor(model)
			require.Equal(t, "medium", *d.DefaultReasoningLevel)
			require.Equal(t, reasoningLevels("none", "low", "medium", "high", "xhigh", "max"), d.SupportedReasoningLevels)
			require.EqualValues(t, 1050000, d.ContextWindow)
			require.EqualValues(t, 1050000, d.MaxContextWindow)
			require.Equal(t, []string{"text", "image"}, d.InputModalities)
			require.Len(t, d.ServiceTiers, 1)
			require.Equal(t, "priority", d.ServiceTiers[0].ID)
			require.True(t, isOpenAICodexImageInputModel(model))

			body := []byte(fmt.Sprintf("{\"models\":[{\"slug\":%q,\"use_responses_lite\":true}]}", model))
			adjusted, err := adjustAPIKeyCodexModelsManifest(body)
			require.NoError(t, err)
			require.False(t, gjson.GetBytes(adjusted, "models.0.use_responses_lite").Bool())

			account := &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
				Credentials: map[string]any{"model_mapping": map[string]any{"allowed": model}}}
			require.True(t, account.IsModelSupported("allowed"))
			require.False(t, account.IsModelSupported(model), "explicit whitelist must still apply")
		})
	}
	require.Equal(t, "gpt-6-astra", normalizeCodexModel("gpt-6"))
	require.False(t, isOpenAIGPT6Model("gpt-6-unknown"))
}

func TestGPT6SolLunaForwardPreservesModelAndEffort(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, model := range []string{"gpt-6-sol", "gpt-6-luna"} {
		for _, effort := range []string{"none", "max"} {
			for _, accountType := range []string{AccountTypeAPIKey, AccountTypeOAuth} {
				t.Run(model+"/"+effort+"/"+accountType, func(t *testing.T) {
					upstream := &httpUpstreamRecorder{resp: &http.Response{
						StatusCode: http.StatusOK,
						Header:     http.Header{"Content-Type": []string{"application/json"}},
						Body:       io.NopCloser(strings.NewReader(`{"usage":{"input_tokens":1,"output_tokens":2}}`)),
					}}
					svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
					account := &Account{ID: "gpt6-test", Platform: PlatformOpenAI, Type: accountType,
						Status: StatusActive, Schedulable: true, Concurrency: 1,
						Credentials: map[string]any{
							"api_key": "fixture", "access_token": "fixture", "chatgpt_account_id": "fixture",
							"model_mapping": map[string]any{"alias": model},
						},
						Extra: map[string]any{"use_responses_api": true},
					}
					c, _ := gin.CreateTestContext(httptest.NewRecorder())
					c.Request = httptest.NewRequest(http.MethodPost, "/openai/v1/responses", nil)
					SetOpenAIClientTransport(c, OpenAIClientTransportHTTP)
					body := []byte(fmt.Sprintf("{\"model\":\"alias\",\"stream\":false,\"input\":\"hello\",\"reasoning\":{\"effort\":%q},\"tools\":[{\"type\":\"function\",\"name\":\"lookup\",\"parameters\":{\"type\":\"object\",\"properties\":{}}}]}", effort))
					result, err := svc.Forward(context.Background(), c, account, body)
					require.NoError(t, err)
					require.NotNil(t, result)
					require.Equal(t, model, gjson.GetBytes(upstream.lastBody, "model").String())
					require.Equal(t, effort, gjson.GetBytes(upstream.lastBody, "reasoning.effort").String())
					require.Equal(t, "lookup", gjson.GetBytes(upstream.lastBody, "tools.0.name").String())
					if effort == "max" {
						require.NotNil(t, result.ReasoningEffort)
						require.Equal(t, effort, *result.ReasoningEffort)
					}
				})
			}
		}
	}
}

func TestGPT6SolLunaPricing(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "resources", "model-pricing", "model_prices_and_context_window.json"))
	require.NoError(t, err)
	catalog := &PricingService{}
	catalog.pricingData, err = catalog.parsePricingData(data)
	require.NoError(t, err)
	astraOnly := &PricingService{pricingData: map[string]*LiteLLMModelPricing{
		"gpt-6":       openAIGPT6AstraFallbackPricing,
		"gpt-6-astra": openAIGPT6AstraFallbackPricing,
	}}
	for _, tc := range []struct {
		model         string
		input, output float64
	}{{"gpt-6-sol", 2e-6, 10e-6}, {"gpt-6-luna", 0.1e-6, 0.5e-6}} {
		for name, pricingService := range map[string]*PricingService{"static": nil, "catalog": catalog, "astra-only": astraOnly} {
			t.Run(tc.model+"/"+name, func(t *testing.T) {
				svc := NewBillingService(&config.Config{}, pricingService)
				require.True(t, svc.HasIdentifiedTokenPricing(tc.model))
				for _, alias := range []string{tc.model, "openai/" + tc.model, tc.model + "-2026-09-22"} {
					p, err := svc.GetModelPricing(alias)
					require.NoError(t, err)
					require.InDelta(t, tc.input, p.InputPricePerToken, 1e-15)
					require.InDelta(t, tc.output, p.OutputPricePerToken, 1e-15)
					require.InDelta(t, tc.input*1.25, p.CacheCreationPricePerToken, 1e-15)
					require.InDelta(t, tc.input/10, p.CacheReadPricePerToken, 1e-15)
				}
				for _, tier := range []struct {
					name  string
					scale float64
				}{{"", 1}, {"priority", 2}, {"fast", 2}, {"flex", 0.5}} {
					for _, cached := range []int{72000, 72001} {
						tokens := UsageTokens{InputTokens: 100000, CacheCreationTokens: 100000, CacheReadTokens: cached, OutputTokens: 10}
						// Gateway normalizes the public "fast" alias before billing.
						cost, err := svc.CalculateCostWithServiceTier(tc.model, tokens, 1, normalizedOpenAIServiceTierValue(tier.name))
						require.NoError(t, err)
						inputScale, outputScale := tier.scale, tier.scale
						if cached > 72000 {
							inputScale *= 2
							outputScale *= 1.5
						}
						require.InDelta(t, 100000*tc.input*inputScale, cost.InputCost, 1e-12)
						require.InDelta(t, 100000*tc.input*1.25*inputScale, cost.CacheCreationCost, 1e-12)
						require.InDelta(t, float64(cached)*tc.input/10*inputScale, cost.CacheReadCost, 1e-12)
						require.InDelta(t, 10*tc.output*outputScale, cost.OutputCost, 1e-12)
					}
				}
			})
		}
		t.Run(tc.model+"/dynamic-and-explicit-prices", func(t *testing.T) {
			svc := NewBillingService(&config.Config{}, &PricingService{pricingData: map[string]*LiteLLMModelPricing{
				tc.model: {InputCostPerToken: tc.input * 3, OutputCostPerToken: tc.output * 3},
			}})
			p, err := svc.GetModelPricing(tc.model)
			require.NoError(t, err)
			require.InDelta(t, tc.input*3, p.InputPricePerToken, 1e-15)
			require.InDelta(t, tc.input*3*1.25, p.CacheCreationPricePerToken, 1e-15)
			require.InDelta(t, tc.input*3*2*1.25, p.CacheCreationPricePerTokenPriority, 1e-15)
			free := 0.0
			p, err = svc.GetModelPricingWithChannel(tc.model, &ChannelModelPricing{CacheWritePrice: &free})
			require.NoError(t, err)
			require.Zero(t, p.CacheCreationPricePerToken)
			original, err := svc.GetModelPricing(tc.model)
			require.NoError(t, err)
			require.Positive(t, original.CacheCreationPricePerToken, "channel overrides must not mutate shared pricing")
		})
	}
	require.True(t, json.Valid(data))
}
