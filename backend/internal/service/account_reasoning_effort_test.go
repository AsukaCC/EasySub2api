package service

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestAccountGetModelReasoningEffort(t *testing.T) {
	account := &Account{
		Platform: PlatformOpenAI,
		Credentials: map[string]any{
			"model_reasoning_efforts": map[string]any{
				"gpt-*":     "medium",
				"gpt-5.*":   "high",
				"gpt-5.6":   "x-high",
				"bad-model": "unsupported",
			},
		},
	}

	require.Equal(t, "xhigh", account.GetModelReasoningEffort("gpt-5.6"))
	require.Equal(t, "high", account.GetModelReasoningEffort("gpt-5.5"))
	require.Equal(t, "medium", account.GetModelReasoningEffort("gpt-4.1"))
	require.Empty(t, account.GetModelReasoningEffort("bad-model"))
	require.Empty(t, account.GetModelReasoningEffort("claude-sonnet-4-6"))

	account.Platform = PlatformAnthropic
	require.Empty(t, account.GetModelReasoningEffort("gpt-5.6"))
}

func TestApplyAccountModelReasoningEffort(t *testing.T) {
	account := &Account{
		Platform: PlatformOpenAI,
		Credentials: map[string]any{
			"model_reasoning_efforts": map[string]string{"gpt-5.6": "high"},
		},
	}

	t.Run("responses", func(t *testing.T) {
		got, changed := ApplyAccountModelReasoningEffort([]byte(`{"model":"gpt-5.6","reasoning":{"effort":"low"}}`), account, "gpt-5.6")
		require.True(t, changed)
		require.Equal(t, "high", gjson.GetBytes(got, "reasoning.effort").String())
	})

	t.Run("chat completions", func(t *testing.T) {
		got, changed := ApplyAccountModelReasoningEffort([]byte(`{"model":"gpt-5.6","messages":[]}`), account, "gpt-5.6")
		require.True(t, changed)
		require.Equal(t, "high", gjson.GetBytes(got, "reasoning_effort").String())
	})

	t.Run("messages bridge", func(t *testing.T) {
		got, changed := ApplyAccountModelReasoningEffortForMessages([]byte(`{"model":"gpt-5.6","messages":[]}`), account, "gpt-5.6")
		require.True(t, changed)
		require.Equal(t, "high", gjson.GetBytes(got, "output_config.effort").String())
	})
}

func TestAccountModelReasoningEffortRespectsGroupCeiling(t *testing.T) {
	account := &Account{
		Platform: PlatformOpenAI,
		Credentials: map[string]any{
			"model_reasoning_efforts": map[string]string{"gpt-5.6": "xhigh"},
		},
	}

	forced, changed := ApplyAccountModelReasoningEffort([]byte(`{"model":"gpt-5.6"}`), account, "gpt-5.6")
	require.True(t, changed)
	capped, changed, err := ApplyOpenAIReasoningEffortPolicyForModel(
		forced,
		"medium",
		nil,
		ReasoningEffortOverLimitDowngrade,
		"gpt-5.6",
	)
	require.NoError(t, err)
	require.True(t, changed)
	require.Equal(t, "medium", gjson.GetBytes(capped, "reasoning.effort").String())

	wsBody, err := applyOpenAIWSReasoningEffortPolicyForModel(
		[]byte(`{"model":"gpt-5.6","reasoning":{"effort":"low"}}`),
		&OpenAIWSIngressHooks{
			ModelReasoningEffort: account.GetModelReasoningEffort,
			MaxReasoningEffort:   "medium",
		},
		"gpt-5.6",
	)
	require.NoError(t, err)
	require.Equal(t, "medium", gjson.GetBytes(wsBody, "reasoning.effort").String())
}
