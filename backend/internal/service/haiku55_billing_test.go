//go:build unit

package service

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHaiku55FallbackBillingBoundary(t *testing.T) {
	svc := newTestBillingService()
	for _, model := range []string{"claude-haiku-5-5", "anthropic/claude-haiku-5.5"} {
		for _, n := range []int{99999, 100000, 100001} {
			tokens := UsageTokens{InputTokens: n - 3000, OutputTokens: 500, CacheReadTokens: 2000, CacheCreationTokens: 1000}
			cost, err := svc.CalculateCost(model, tokens, 1)
			require.NoError(t, err)
			multiplier := 1.0
			if n > 100000 {
				multiplier = 5
			}
			require.InDelta(t, float64(tokens.InputTokens)*1e-7*multiplier, cost.InputCost, 1e-10)
			require.InDelta(t, 1000*1.25e-7*multiplier, cost.CacheCreationCost, 1e-10)
			require.InDelta(t, 2000*1e-8*multiplier, cost.CacheReadCost, 1e-10)
			require.InDelta(t, 500*5e-7*multiplier, cost.OutputCost, 1e-10)
			require.Equal(t, n > 100000, cost.LongContextBillingApplied)
		}
	}
}
