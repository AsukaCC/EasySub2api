package model

import (
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestAccountModelRuleValidateRoutes(t *testing.T) {
	t.Run("normalizes OpenAI route efforts and tier", func(t *testing.T) {
		tier := " Pro / Team "
		rule := &AccountModelRule{
			Name:             " OpenAI routing ",
			Platform:         domain.PlatformOpenAI,
			SubscriptionTier: &tier,
			ModelRoutes: []domain.AccountModelRoute{{
				RequestModel:    " gpt-* ",
				UpstreamModel:   " gpt-5.6 ",
				ReasoningEffort: " X-HIGH ",
			}},
		}

		require.NoError(t, rule.Validate())
		require.Equal(t, "OpenAI routing", rule.Name)
		require.Equal(t, "pro_team", *rule.SubscriptionTier)
		require.Equal(t, []domain.AccountModelRoute{{
			RequestModel:    "gpt-*",
			UpstreamModel:   "gpt-5.6",
			ReasoningEffort: "xhigh",
		}}, rule.ModelRoutes)
	})

	t.Run("allows an empty passthrough route set", func(t *testing.T) {
		rule := &AccountModelRule{Name: "Passthrough", Platform: domain.PlatformAnthropic}

		require.NoError(t, rule.Validate())
		require.Empty(t, rule.ModelRoutes)
	})

	t.Run("rejects efforts for non OpenAI rules", func(t *testing.T) {
		rule := &AccountModelRule{
			Name:     "Anthropic",
			Platform: domain.PlatformAnthropic,
			ModelRoutes: []domain.AccountModelRoute{{
				RequestModel:    "claude-*",
				UpstreamModel:   "claude-sonnet-4-6",
				ReasoningEffort: "high",
			}},
		}

		require.ErrorContains(t, rule.Validate(), "only supported for OpenAI")
	})

	t.Run("rejects duplicate and invalid wildcard routes", func(t *testing.T) {
		rule := &AccountModelRule{
			Name:     "Invalid",
			Platform: domain.PlatformOpenAI,
			ModelRoutes: []domain.AccountModelRoute{{
				RequestModel:  "gpt-*-preview",
				UpstreamModel: "gpt-5.6",
			}},
		}

		require.ErrorContains(t, rule.Validate(), "wildcard")

		rule.ModelRoutes = []domain.AccountModelRoute{
			{RequestModel: "gpt-5.6", UpstreamModel: "gpt-5.6"},
			{RequestModel: "gpt-5.6", UpstreamModel: "gpt-5.6-sol"},
		}
		require.ErrorContains(t, rule.Validate(), "duplicate")
	})
}
