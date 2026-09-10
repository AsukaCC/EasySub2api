//go:build unit

package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsCNProvider_MiniMax(t *testing.T) {
	assert.True(t, IsCNProvider(PlatformMiniMax))
	assert.True(t, (&Account{Platform: PlatformMiniMax}).IsMiniMax())
	assert.True(t, (&Account{Platform: PlatformMiniMax}).IsOpenAICompatible())
	assert.Contains(t, AllowedQuotaPlatforms, PlatformMiniMax)
	assert.Contains(t, AllowedSchedulingThresholdPlatforms, PlatformMiniMax)
	assert.Equal(t, PlatformMiniMax, NormalizeOpenAICompatiblePlatform(PlatformMiniMax))
	assert.True(t, isConcreteRequestPlatform(PlatformMiniMax))
	assert.True(t, IsUpstreamBillingProbeIdentity(PlatformMiniMax, AccountTypeAPIKey))
}

func TestMiniMaxQuotaURL(t *testing.T) {
	assert.Equal(t,
		"https://api.minimax.io/v1/api/openplatform/coding_plan/remains",
		minimaxQuotaURL("https://api.minimax.io/v1"))
	assert.Equal(t,
		"https://api.minimax.io/v1/api/openplatform/coding_plan/remains",
		minimaxQuotaURL("https://api.minimax.io/anthropic"))
	assert.Equal(t,
		"https://api.minimaxi.com/v1/api/openplatform/coding_plan/remains",
		minimaxQuotaURL("https://api.minimaxi.com/v1"))
	// 未知域名回落国内站。
	assert.Equal(t,
		"https://api.minimaxi.com/v1/api/openplatform/coding_plan/remains",
		minimaxQuotaURL("https://relay.example.com/v1"))
}

func TestParseMiniMaxUsageTiers(t *testing.T) {
	body := []byte(`{
		"model_remains": [
			{"model_name": "video", "current_interval_remaining_percent": 10, "end_time": 1780000000000},
			{
				"model_name": "general",
				"current_interval_remaining_percent": 37.5,
				"end_time": 1780000000000,
				"current_weekly_status": 1,
				"current_weekly_remaining_percent": 80,
				"weekly_end_time": 1780500000
			}
		]
	}`)
	tiers := parseMiniMaxUsageTiers(body)
	require.Len(t, tiers, 2)

	assert.Equal(t, "5h", tiers[0].Window)
	assert.InDelta(t, 62.5, tiers[0].UsedPercent, 0.0001)
	assert.Equal(t, time.UnixMilli(1780000000000).UTC().Format(time.RFC3339), tiers[0].ResetAt)

	assert.Equal(t, "weekly", tiers[1].Window)
	assert.InDelta(t, 20, tiers[1].UsedPercent, 0.0001)
	assert.Equal(t, time.Unix(1780500000, 0).UTC().Format(time.RFC3339), tiers[1].ResetAt)
}

func TestParseMiniMaxUsageTiers_WeeklyDisabledAndMissingGeneral(t *testing.T) {
	// current_weekly_status != 1 → 只返回 5h 窗口。
	body := []byte(`{"model_remains":[{"model_name":"general","current_interval_remaining_percent":100,"end_time":0,"current_weekly_status":0,"current_weekly_remaining_percent":50}]}`)
	tiers := parseMiniMaxUsageTiers(body)
	require.Len(t, tiers, 1)
	assert.Equal(t, "5h", tiers[0].Window)
	assert.InDelta(t, 0, tiers[0].UsedPercent, 0.0001)
	assert.Equal(t, "", tiers[0].ResetAt)

	// 没有 general 条目 → nil。
	assert.Nil(t, parseMiniMaxUsageTiers([]byte(`{"model_remains":[{"model_name":"video"}]}`)))
	assert.Nil(t, parseMiniMaxUsageTiers([]byte(`{"base_resp":{"status_code":1004}}`)))
}

func TestAccount_MiniMaxBaseURLs(t *testing.T) {
	chat := &Account{
		Platform:    PlatformMiniMax,
		Type:        AccountTypeAPIKey,
		Credentials: map[string]any{"api_key": "k", "account_mode": AccountModeCoding},
	}
	assert.Equal(t, DefaultMiniMaxBaseURL, chat.GetOpenAIBaseURL())
	assert.Equal(t, DefaultMiniMaxBaseURL, chat.GetOpenAIFormatBaseURL())
	assert.Equal(t, "", chat.GetAnthropicProtocolBaseURL())
	assert.True(t, chat.SupportsNativeCNResponses())
	assert.False(t, chat.UsesNativeCNResponses())

	responses := &Account{
		Platform:    PlatformMiniMax,
		Type:        AccountTypeAPIKey,
		Credentials: map[string]any{"api_key": "k", "api_protocol": APIProtocolResponses},
	}
	assert.Equal(t, APIProtocolResponses, responses.GetAPIProtocol())
	assert.True(t, responses.UsesNativeCNResponses())

	anthropic := &Account{
		Platform:    PlatformMiniMax,
		Type:        AccountTypeAPIKey,
		Credentials: map[string]any{"api_key": "k", "api_protocol": APIProtocolAnthropic},
	}
	assert.Equal(t, DefaultMiniMaxAnthropicBaseURL, anthropic.GetAnthropicProtocolBaseURL())
	// Anthropic 协议下 OpenAI 格式路径应回落到 Chat Completions 默认 base。
	assert.Equal(t, DefaultMiniMaxBaseURL, anthropic.GetOpenAIFormatBaseURL())

	custom := &Account{
		Platform:    PlatformMiniMax,
		Type:        AccountTypeAPIKey,
		Credentials: map[string]any{"api_key": "k", "api_protocol": APIProtocolAnthropic, "base_url": "https://api.minimax.io/anthropic"},
	}
	assert.Equal(t, "https://api.minimax.io/anthropic", custom.GetAnthropicProtocolBaseURL())
}

func TestAccount_GetCodingPlanProvider_MiniMax(t *testing.T) {
	newAcc := func(baseURL string, mode string) *Account {
		creds := map[string]any{"api_key": "k"}
		if mode != "" {
			creds["account_mode"] = mode
		}
		if baseURL != "" {
			creds["base_url"] = baseURL
		}
		return &Account{Platform: PlatformMiniMax, Type: AccountTypeAPIKey, Credentials: creds}
	}

	assert.Equal(t, PlatformMiniMax, newAcc("", AccountModeCoding).GetCodingPlanProvider(), "默认国内站")
	assert.Equal(t, PlatformMiniMax, newAcc("https://api.minimax.io/v1", AccountModeCoding).GetCodingPlanProvider())
	assert.Equal(t, PlatformMiniMax, newAcc("https://api.minimaxi.com/anthropic", AccountModeCoding).GetCodingPlanProvider())
	assert.Equal(t, PlatformMiniMax, newAcc("https://api.minimax.com/v1", AccountModeCoding).GetCodingPlanProvider())

	// payg 账号无 Coding Plan 额度端点。
	assert.Equal(t, "", newAcc("", AccountModePayG).GetCodingPlanProvider())
	// 自定义中转不得把第三方 Key 发往官方额度端点。
	assert.Equal(t, "", newAcc("https://relay.example.com/v1", AccountModeCoding).GetCodingPlanProvider())
}

func TestEvaluateAccountSchedulingThreshold_MiniMaxCodingPlan(t *testing.T) {
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	resetAt := now.Add(2 * time.Hour).Format(time.RFC3339)
	account := &Account{
		Platform:    PlatformMiniMax,
		Type:        AccountTypeAPIKey,
		Credentials: map[string]any{"api_key": "k", "account_mode": AccountModeCoding},
		Extra: map[string]any{
			cnExtraKey(PlatformMiniMax, cnExtraSuffix5hUsed):  92.0,
			cnExtraKey(PlatformMiniMax, cnExtraSuffix5hReset): resetAt,
		},
	}

	decision := EvaluateAccountSchedulingThreshold(account, map[string]int{PlatformMiniMax: 90}, now)
	require.True(t, decision.ShouldPause)
	assert.Equal(t, PlatformMiniMax, decision.Platform)
	assert.Equal(t, "5h", decision.Window)
	assert.Equal(t, PlatformMiniMax, decision.Scope)
	assert.InDelta(t, 92, decision.UsedPercent, 0.0001)
	require.NotNil(t, decision.Until)
	assert.Equal(t, resetAt, decision.Until.UTC().Format(time.RFC3339))

	// 未超阈值不停调。
	relaxed := EvaluateAccountSchedulingThreshold(account, map[string]int{PlatformMiniMax: 95}, now)
	assert.False(t, relaxed.ShouldPause)

	// 无快照（payg）不停调。
	payg := &Account{Platform: PlatformMiniMax, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "k"}}
	assert.False(t, EvaluateAccountSchedulingThreshold(payg, map[string]int{PlatformMiniMax: 50}, now).ShouldPause)
}
