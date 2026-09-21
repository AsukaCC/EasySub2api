package service

import "context"

// Retain only upstream-reported usage. Early sampling cancellation may omit
// final counters; do not estimate tokens or make another request to obtain them.
func collectModelFingerprintUsage(ctx context.Context, data map[string]any) {
	p := fingerprintProbe(ctx)
	if p == nil {
		return
	}
	for _, key := range []string{"response", "message"} {
		if nested, ok := data[key].(map[string]any); ok {
			data = nested
		}
	}
	if model, ok := data["model"].(string); ok && model != "" {
		p.usage.UpstreamResponseModel = &model
	}
	usage, _ := data["usage"].(map[string]any)
	count := func(m map[string]any, key string) int {
		v, _ := m[key].(float64)
		return max(0, int(v))
	}
	if usage != nil {
		if _, ok := usage["input_tokens"]; ok {
			p.usage.InputTokens = count(usage, "input_tokens")
		}
		if _, ok := usage["output_tokens"]; ok {
			p.usage.OutputTokens = count(usage, "output_tokens")
		}
		if _, ok := usage["prompt_tokens"]; ok {
			p.usage.InputTokens = count(usage, "prompt_tokens")
			p.usage.OutputTokens = count(usage, "completion_tokens")
		}
		for _, key := range []string{"input_tokens_details", "prompt_tokens_details"} {
			if details, ok := usage[key].(map[string]any); ok {
				p.usage.CacheReadTokens = count(details, "cached_tokens")
				p.usage.InputTokens = max(0, p.usage.InputTokens-p.usage.CacheReadTokens)
			}
		}
		if _, ok := usage["cache_read_input_tokens"]; ok {
			p.usage.CacheReadTokens = count(usage, "cache_read_input_tokens")
		}
		if _, ok := usage["cache_creation_input_tokens"]; ok {
			p.usage.CacheCreationTokens = count(usage, "cache_creation_input_tokens")
		}
	}
	if usage, ok := data["usageMetadata"].(map[string]any); ok {
		p.usage.CacheReadTokens = count(usage, "cachedContentTokenCount")
		p.usage.InputTokens = max(0, count(usage, "promptTokenCount")-p.usage.CacheReadTokens)
		p.usage.OutputTokens = count(usage, "candidatesTokenCount") + count(usage, "thoughtsTokenCount")
	}
}
