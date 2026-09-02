package service

import "strings"

func optionalTrimmedStringPtr(raw string) *string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// coalesceRequestedReasoningEffort prefers the client-requested value and
// falls back to the effective effort for historical or unmapped requests.
func coalesceRequestedReasoningEffort(requested, forwarded *string) *string {
	if requested != nil {
		if value := strings.TrimSpace(*requested); value != "" {
			return &value
		}
	}
	if forwarded != nil {
		if value := strings.TrimSpace(*forwarded); value != "" {
			return &value
		}
	}
	return nil
}

func forwardResultBillingModel(requestedModel, upstreamModel string) string {
	if trimmed := strings.TrimSpace(requestedModel); trimmed != "" {
		return trimmed
	}
	return strings.TrimSpace(upstreamModel)
}

func optionalInt64Ptr(v int64) *int64 {
	if v == 0 {
		return nil
	}
	return &v
}
