package service

import "time"

// March 1, 2027 midnight Pacific is 08:00 UTC (before daylight saving time).
var openAIGPT62UltrafastPromotionEnd = time.Date(2027, time.March, 1, 8, 0, 0, 0, time.UTC)

// Official USD/token prices, verified 2026-11-18:
// https://developers.openai.com/api/docs/pricing
func openAIGPT62FallbackPricing(model string) *LiteLLMModelPricing {
	input, output, read := 2e-6, 10e-6, 0.1e-6
	if model == "gpt-6.2-astra" {
		input, output, read = 8e-6, 40e-6, 0.4e-6
	}
	return &LiteLLMModelPricing{
		InputCostPerToken: input, InputCostPerTokenPriority: input * 2,
		OutputCostPerToken: output, OutputCostPerTokenPriority: output * 2,
		CacheReadInputTokenCost: read, CacheReadInputTokenCostPriority: read * 2,
		CacheCreationInputTokenCost: input * 1.25, CacheCreationInputTokenCostPriority: input * 2.5,
		LongContextInputTokenThreshold: 272_000, LongContextInputCostMultiplier: 2,
		LongContextOutputCostMultiplier: 1.5, SupportsServiceTier: true,
		LiteLLMProvider: "openai", Mode: "chat", SupportsPromptCaching: true,
	}
}
