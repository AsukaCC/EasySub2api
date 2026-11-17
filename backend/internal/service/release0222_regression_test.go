//go:build unit

package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/config"
	"github.com/AsukaCC/EasySub2api/internal/pkg/ctxkey"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestRelease0222ForcedTier(t *testing.T) {
	group := &Group{ID: "group", Hydrated: true, Platform: PlatformComposite, Status: "active", ForceOpenAIFast: true, ForceOpenAIUltrafast: true}
	ctx := context.WithValue(context.Background(), ctxkey.Group, group)
	account := &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
	svc := &OpenAIGatewayService{}
	for _, tc := range []struct{ model, want string }{{"gpt-6-astra", "ultrafast"}, {"gpt-6.1-sol", "ultrafast"}, {"gpt-6-luna", "priority"}} {
		body, err := svc.applyOpenAIFastPolicyToBody(ctx, account, tc.model, []byte(`{"service_tier":"flex"}`))
		require.NoError(t, err)
		require.Equal(t, tc.want, gjson.GetBytes(body, "service_tier").String())
	}
	group.ForceOpenAIFast = false
	require.Empty(t, openAIGroupForcedServiceTier(ctx, account, "gpt-6-luna"))
	require.Empty(t, openAIGroupForcedServiceTier(ctx, &Account{Platform: PlatformAnthropic}, "gpt-6-astra"))
	group.Hydrated = false
	require.Empty(t, openAIGroupForcedServiceTier(ctx, account, "gpt-6-astra"))
}

func TestRelease0222AdditionalToolsIDs(t *testing.T) {
	id := "at_12345678-1234-1234-1234-123456789abc"
	input := []any{map[string]any{"type": "additional_tools", "id": id, "tools": []any{}}, map[string]any{"type": "item_reference", "id": id}}
	filtered := filterCodexInputWithOptions(input, codexInputFilterOptions{PreserveReferences: true})
	for _, item := range filtered {
		require.Equal(t, "at_12345678123412341234123456789abc", item.(map[string]any)["id"])
	}
	require.Equal(t, id, input[0].(map[string]any)["id"], "input must remain unchanged")
	filtered = filterCodexInputWithOptions(input, codexInputFilterOptions{})
	require.Len(t, filtered, 1)
	require.NotEmpty(t, filtered[0].(map[string]any)["id"])
	body, _ := json.Marshal(map[string]any{"input": input})
	out, changed, err := normalizeOpenAIResponsesLiteToolsPayload(body)
	require.NoError(t, err)
	require.True(t, changed)
	require.Equal(t, gjson.GetBytes(out, "input.0.id").String(), gjson.GetBytes(out, "input.1.id").String())
	require.Equal(t, "at_opaque-provider-id", normalizeCodexAdditionalToolsID("at_opaque-provider-id"))
}

func TestRelease0222ExplicitDeepSeekRates(t *testing.T) {
	bs := NewBillingService(&config.Config{}, nil)
	flags := tokenPriceOverridesFromPatch(json.RawMessage(`{"input_cost_per_token":0,"output_cost_per_token":0.000123}`))
	price := &ModelPricing{InputPricePerToken: 0, OutputPricePerToken: 0.000123, priceOverrides: flags}
	resolved := bs.applyModelSpecificPricingPolicy("deepseek-v4-pro", price)
	peak := applyDeepSeekPeakPricing("deepseek-v4-pro", resolved, time.Date(2026, 11, 16, 2, 0, 0, 0, time.UTC), true)
	require.Zero(t, peak.InputPricePerToken)
	require.Equal(t, price.OutputPricePerToken, peak.OutputPricePerToken)
	require.Positive(t, peak.CacheReadPricePerToken)
	for _, model := range []string{"gemini-4-carbon-medium", "gemini-4-hydrogen", "Qwen/Qwen3-Embedding-8B", "Qwen/Qwen3-Reranker-4B"} {
		require.NotNil(t, bs.getFallbackPricing(model), model)
	}
	require.Nil(t, bs.getFallbackPricing("Qwen/Qwen3-Reranker-2B"))
	for _, model := range []string{"gpt-6-astra", "gpt-6.1-sol"} {
		price, err := bs.GetModelPricing(model)
		require.NoError(t, err)
		require.Equal(t, float64(6), price.UltrafastMultiplier)
	}
}

func TestRelease0222WeeklyQuotaSurvivesRollover(t *testing.T) {
	now := time.Now()
	extra := map[string]any{"passive_usage_7d_reset": float64(now.Add(time.Hour).Unix()), "passive_usage_7d_oi_reset": float64(now.Add(-time.Hour).Unix())}
	clears := rolloverPassiveUsageClears(extra, now)
	require.Contains(t, clears, "session_window_utilization")
	require.NotContains(t, clears, "passive_usage_7d_utilization")
	require.Contains(t, clears, "passive_usage_7d_oi_utilization")
	require.NotContains(t, clears, "passive_usage_sampled_at")
}

func TestRelease0222QuotaObservationTime(t *testing.T) {
	now := time.Now()
	observed := now.Add(-time.Hour).UTC().Truncate(time.Second)
	usage := &UsageInfo{UpdatedAt: &now}
	applyExtraToUsage(usage, map[string]any{"codex_usage_updated_at": observed.Format(time.RFC3339), "codex_5h_used_percent": 75.0}, now)
	require.Equal(t, observed, *usage.UpdatedAt)
	applyExtraToUsage(usage, nil, now)
	require.Nil(t, usage.UpdatedAt)
	require.False(t, hasOpenAIQuotaObservation(map[string]any{"codex_usage_updated_at": now}))
	require.Empty(t, usageQuotaTiers(&UsageInfo{FiveHour: &UsageProgress{UtilizationUnknown: true}}))
}

type release0222FableRepo struct {
	AccountRepository
	called bool
}

func (r *release0222FableRepo) ClearModelRateLimitIfReason(_ context.Context, _, scope, reason string) (bool, error) {
	r.called = scope == anthropicFableRateLimitKey && reason == anthropicFableWindowReason
	return r.called, nil
}

func TestRelease0222FableRecoveryIsScoped(t *testing.T) {
	limits := map[string]any{anthropicFableRateLimitKey: map[string]any{"reason": anthropicFableWindowReason, "rate_limit_reset_at": time.Now().Add(time.Hour).Format(time.RFC3339)}, "other": map[string]any{"reason": "manual"}}
	account := &Account{ID: "account", Extra: map[string]any{modelRateLimitsKey: limits}}
	repo := &release0222FableRepo{}
	require.True(t, recoverAnthropicFableWindowLimit(context.Background(), repo, account, 0))
	require.True(t, repo.called)
	require.Contains(t, account.Extra[modelRateLimitsKey], "other")
	require.Contains(t, limits, anthropicFableRateLimitKey, "shared snapshot must not be mutated")
}

func TestRelease0222StreamingErrorAfterKeepalive(t *testing.T) {
	for _, started := range []bool{false, true} {
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		if started {
			c.Header("Content-Type", "text/event-stream")
			c.Writer.Write([]byte(": ping\n\n"))
			c.Writer.Flush()
		}
		writeGatewayErrorResponse(c, 502, "upstream_error", "try later")
		if started {
			require.Equal(t, 200, recorder.Code)
			require.Contains(t, recorder.Body.String(), "event:error\n")
			require.Contains(t, recorder.Body.String(), `"message":"try later"`)
		} else {
			require.Equal(t, 502, recorder.Code)
			require.True(t, json.Valid(recorder.Body.Bytes()))
		}
	}
}

func TestRelease0222SpoolFailureKeepsBoundedMemory(t *testing.T) {
	stage := newOpenAIFirstOutputStage(70 * 1024)
	defer stage.Close()
	stage.createTemp = func() (*os.File, error) { return nil, os.ErrPermission }
	payload := bytes.Repeat([]byte("x"), 68*1024)
	_, err := stage.Write(payload)
	require.NoError(t, err)
	require.Nil(t, stage.tempFile)
	_, err = stage.Write(bytes.Repeat([]byte("y"), 3*1024))
	require.ErrorIs(t, err, errOpenAIFirstOutputStageLimit)
	var output bytes.Buffer
	require.NoError(t, stage.CommitTo(&output))
	require.Equal(t, payload, output.Bytes())
}

func TestRelease0222BedrockSchedulingModelAndErrors(t *testing.T) {
	account := &Account{Credentials: map[string]any{"model_mapping": map[string]any{"client-model": "scheduled-model"}}}
	ctx := WithGatewaySchedulingModel(context.Background(), "client-model")
	require.Equal(t, "scheduled-model", bedrockSchedulingModelKey(ctx, account, "channel-model"))
	require.True(t, isBedrockAccountBlockedMessage("Operation not allowed."))
	require.False(t, isBedrockAccountBlockedMessage("Unexpected field: operation not allowed"))
	require.True(t, isBedrockContentFilterError(503, []byte(`{"message":"Blocked by content filtering policy"}`)))
	require.False(t, isBedrockContentFilterError(400, []byte(`{"message":"Blocked by content filtering policy"}`)))
}

func TestRelease0222QuotaClearsOnlyAbsentWindows(t *testing.T) {
	tiers := []CNQuotaTier{{Window: "monthly", UsedPercent: 55}}
	updates := cnQuotaExtraUpdates(PlatformKimi, tiers, time.Now())
	clearAbsentCNQuotaWindows(PlatformKimi, tiers, updates)
	require.Equal(t, float64(55), updates["kimi_monthly_used_percent"])
	for _, key := range []string{"kimi_5h_used_percent", "kimi_weekly_used_percent", "kimi_monthly_reset_at"} {
		require.Contains(t, updates, key)
		require.Nil(t, updates[key])
	}
}

func TestRelease0222SSEEscapedUsageIsParsed(t *testing.T) {
	require.True(t, sseEventNeedsJSONParse("content_block_delta", []byte(`{"u\u0073age":{"output_tokens":5}}`), false, false))
	require.False(t, sseEventNeedsJSONParse("content_block_delta", []byte(`{"delta":{"text":"hello"}}`), false, false))
}
