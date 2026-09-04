package service

import (
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestBoundAccountModelRoutes(t *testing.T) {
	ruleID := "01990f3d-8198-7000-8000-000000000001"

	t.Run("empty rule bypasses legacy account mappings", func(t *testing.T) {
		account := &Account{
			Platform:    PlatformOpenAI,
			Type:        AccountTypeOAuth,
			ModelRuleID: &ruleID,
			Credentials: map[string]any{
				"model_mapping": map[string]any{"gpt-5.4": "gpt-5.4"},
			},
			Extra: map[string]any{"openai_passthrough": true},
		}

		require.True(t, account.IsModelSupported("custom/provider-model"))
		require.Equal(t, "custom/provider-model", account.GetMappedModel("custom/provider-model"))
		require.False(t, account.IsOpenAIPassthroughEnabled())
	})

	t.Run("exact route wins and longest wildcard is used", func(t *testing.T) {
		account := &Account{
			Platform:    PlatformOpenAI,
			ModelRuleID: &ruleID,
			ModelRoutes: []domain.AccountModelRoute{
				{RequestModel: "gpt-*", UpstreamModel: "generic", ReasoningEffort: "medium"},
				{RequestModel: "gpt-5.*", UpstreamModel: "gpt-5-family", ReasoningEffort: "high"},
				{RequestModel: "gpt-5.6", UpstreamModel: "gpt-5.6-sol", ReasoningEffort: "xhigh"},
			},
		}

		require.True(t, account.IsModelSupported("gpt-5.6"))
		require.Equal(t, "gpt-5.6-sol", account.GetMappedModel("gpt-5.6"))
		require.Equal(t, "xhigh", account.GetModelReasoningEffort("gpt-5.6"))
		require.Equal(t, "gpt-5-family", account.GetMappedModel("gpt-5.5"))
		require.Equal(t, "high", account.GetModelReasoningEffort("gpt-5.5"))
		require.False(t, account.IsModelSupported("claude-sonnet-4-6"))
		require.Equal(t, "claude-sonnet-4-6", account.GetMappedModel("claude-sonnet-4-6"))
	})
}

func TestAccountModelRuleTierMismatch(t *testing.T) {
	ruleID := "01990f3d-8198-7000-8000-000000000002"
	ruleTier := "pro"
	account := &Account{
		SubscriptionTier:          "team",
		ModelRuleID:               &ruleID,
		ModelRuleSubscriptionTier: &ruleTier,
	}

	require.True(t, account.HasModelRuleTierMismatch())
	account.SubscriptionTier = "pro"
	require.False(t, account.HasModelRuleTierMismatch())
	account.ModelRuleSubscriptionTier = nil
	require.False(t, account.HasModelRuleTierMismatch())
}

func TestResolveAccountSubscriptionTier(t *testing.T) {
	require.Equal(t, "pro_team", ResolveAccountSubscriptionTier(&Account{
		Platform:    PlatformOpenAI,
		Credentials: map[string]any{"plan_type": " Pro / Team "},
	}))
	require.Equal(t, "standard", ResolveAccountSubscriptionTier(&Account{
		Platform:    PlatformGemini,
		Credentials: map[string]any{"tier_id": "STANDARD", "plan_type": "fallback"},
	}))
	require.Empty(t, ResolveAccountSubscriptionTier(&Account{Platform: PlatformAnthropic}))
}
