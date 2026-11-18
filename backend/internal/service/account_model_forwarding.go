package service

import (
	"fmt"
	"strings"
)

func NormalizeOpenAIModelForwardingCredentials(platform string, credentials map[string]any) error {
	raw, exists := credentials["model_forwarding"]
	if !exists {
		return nil
	}
	if platform != PlatformOpenAI {
		return fmt.Errorf("model forwarding is only supported for OpenAI accounts")
	}
	mapping := stringMappingFromRaw(raw)
	switch typed := raw.(type) {
	case map[string]any:
		if len(mapping) != len(typed) {
			return fmt.Errorf("model forwarding targets must be strings")
		}
	case map[string]string:
	default:
		return fmt.Errorf("model_forwarding must be an object")
	}
	normalized := make(map[string]any, len(mapping))
	for source, target := range mapping {
		source, target = strings.TrimSpace(source), strings.TrimSpace(target)
		if source == "" || target == "" || strings.ContainsAny(source+target, " \t\n\r") ||
			strings.Contains(strings.TrimSuffix(source, "*"), "*") || strings.Contains(target, "*") {
			return fmt.Errorf("invalid model forwarding rule %q -> %q", source, target)
		}
		if _, duplicate := normalized[source]; duplicate {
			return fmt.Errorf("duplicate model forwarding source %q", source)
		}
		if err := CheckActiveModel(source, target); err != nil {
			return err
		}
		normalized[source] = target
	}
	credentials["model_forwarding"] = normalized
	return nil
}

// ResolveForwardedModel applies a single account-local forwarding rule. These
// rules also admit client aliases absent from the account's model whitelist.
// They are independent of legacy model_mapping and remain active in passthrough.
func (a *Account) ResolveForwardedModel(requestedModel string) (string, bool) {
	if a == nil || !a.IsOpenAI() {
		return requestedModel, false
	}
	mapping := stringMappingFromRaw(a.Credentials["model_forwarding"])
	model, matched := resolveRequestedModelInMapping(mapping, requestedModel)
	if !matched {
		normalized := normalizeRequestedModelForLookup(a.Platform, requestedModel)
		model, matched = resolveRequestedModelInMapping(mapping, normalized)
	}
	model = strings.TrimSpace(model)
	if !matched || model == "" || strings.Contains(model, "*") {
		return requestedModel, false
	}
	return model, true
}

// Group forwarding and legacy account rules bill the model actually sent.
// Never fall back to pricing for the client alias or a channel's intermediate model.
func openAIForwardingBillingModel(account *Account, input *OpenAIRecordUsageInput) string {
	if input == nil || input.Result == nil {
		return ""
	}
	result := input.Result
	if target, matched := apiKeyGroup(input.APIKey).ResolveForwardedModel(input.OriginalModel); matched {
		if mapped := strings.TrimSpace(input.ChannelMappedModel); mapped != "" && mapped != input.OriginalModel {
			target = mapped
		}
		return strings.TrimSpace(firstNonEmpty(result.UpstreamModel, target))
	}
	requested := firstNonEmpty(input.ChannelMappedModel, input.OriginalModel, result.Model)
	target, matched := account.ResolveForwardedModel(requested)
	if !matched {
		return ""
	}
	return strings.TrimSpace(firstNonEmpty(result.UpstreamModel, result.BillingModel, target))
}
