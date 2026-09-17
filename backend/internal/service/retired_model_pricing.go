package service

import (
	"github.com/AsukaCC/EasySub2api/internal/pkg/openai"
	"strings"
)

// Billing normalization deliberately retains historical identities. Never use
// this helper for routing, model discovery, or request admission.
func historicalOpenAIBillingModel(model string) string {
	if !openai.IsRetiredModel(model) {
		return normalizeKnownOpenAICodexModel(model)
	}
	legacy := canonicalizeOpenAIModelAliasSpelling(model)
	for _, pair := range [][2]string{{"gpt-5-4", "gpt-5.4"}, {"gpt-5-5", "gpt-5.5"}, {"gpt-54", "gpt-5.4"}, {"gpt-55", "gpt-5.5"}} {
		legacy = strings.ReplaceAll(legacy, pair[0], pair[1])
	}
	legacy = canonicalizeOpenAIModelAliasSpelling(legacy)
	for _, family := range []string{"gpt-5.4-mini", "gpt-5.4-nano", "gpt-5.5-pro", "gpt-5.5"} {
		if strings.HasPrefix(legacy, family) {
			return family
		}
	}
	return "gpt-5.4"
}
