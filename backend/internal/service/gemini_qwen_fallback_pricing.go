package service

func (s *BillingService) initCurrentGeminiAndQwenPricing() {
	s.fallbackPrices["gemini-4-carbon"] = &ModelPricing{
		InputPricePerToken: 2e-6, OutputPricePerToken: 10e-6, CacheReadPricePerToken: 0.1e-6,
		InputPricePerTokenPriority: 3.6e-6, OutputPricePerTokenPriority: 18e-6, CacheReadPricePerTokenPriority: 0.18e-6,
	}
	s.fallbackPrices["gemini-4-hydrogen"] = &ModelPricing{
		InputPricePerToken: 1.5e-6, OutputPricePerToken: 7.5e-6, CacheReadPricePerToken: 0.075e-6,
		InputPricePerTokenPriority: 2.7e-6, OutputPricePerTokenPriority: 13.5e-6, CacheReadPricePerTokenPriority: 0.135e-6,
	}
	for _, family := range []string{"qwen3-embedding-", "qwen3-reranker-"} {
		for size, price := range map[string]float64{"8b": 0.04e-6, "4b": 0.02e-6, "0.6b": 0.01e-6, "0-6b": 0.01e-6} {
			s.fallbackPrices[family+size] = &ModelPricing{InputPricePerToken: price}
		}
	}
}
