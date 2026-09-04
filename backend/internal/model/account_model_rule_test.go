package model

import (
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestAccountModelRuleValidateReasoningEfforts(t *testing.T) {
	t.Run("normalizes OpenAI mapping efforts", func(t *testing.T) {
		rule := &AccountModelRule{
			Name:     "OpenAI",
			Platform: domain.PlatformOpenAI,
			Mapping: map[string]string{
				"gpt-*": "gpt-5.6",
			},
			ReasoningEfforts: map[string]string{
				"gpt-*": " X-HIGH ",
			},
		}

		require.NoError(t, rule.Validate())
		require.Equal(t, map[string]string{"gpt-*": "xhigh"}, rule.ReasoningEfforts)
	})

	t.Run("rejects efforts for non OpenAI rules", func(t *testing.T) {
		rule := &AccountModelRule{
			Name:     "Anthropic",
			Platform: domain.PlatformAnthropic,
			Mapping:  map[string]string{"claude-*": "claude-sonnet-4-6"},
			ReasoningEfforts: map[string]string{
				"claude-*": "high",
			},
		}

		require.ErrorContains(t, rule.Validate(), "only supported for OpenAI")
	})

	t.Run("rejects efforts without a mapping source", func(t *testing.T) {
		rule := &AccountModelRule{
			Name:             "OpenAI",
			Platform:         domain.PlatformOpenAI,
			Mapping:          map[string]string{"gpt-5.6": "gpt-5.6"},
			ReasoningEfforts: map[string]string{"gpt-latest": "high"},
		}

		require.ErrorContains(t, rule.Validate(), "must exist in mapping")
	})
}
