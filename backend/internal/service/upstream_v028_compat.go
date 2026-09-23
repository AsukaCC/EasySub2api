package service

import (
	"github.com/tidwall/gjson"
	"strings"
	"time"
)

func setAccountModelRateLimitSnapshot(account *Account, scope string, resetAt time.Time, reason string, now time.Time) {
	if account == nil || strings.TrimSpace(scope) == "" {
		return
	}
	if account.Extra == nil {
		account.Extra = make(map[string]any)
	}
	limits, ok := account.Extra[modelRateLimitsKey].(map[string]any)
	if !ok {
		limits = make(map[string]any)
		account.Extra[modelRateLimitsKey] = limits
	}
	limits[scope] = map[string]any{"rate_limited_at": now.UTC().Format(time.RFC3339), "rate_limit_reset_at": resetAt.UTC().Format(time.RFC3339), "reason": reason}
}

func isOpenAICompatibleModelNotFoundBody(body []byte) bool {
	code := strings.TrimSpace(extractUpstreamErrorCode(body))
	if code != "" {
		return strings.EqualFold(code, "model_not_found")
	}
	message := strings.ToLower(strings.TrimSpace(extractUpstreamErrorMessage(body)))
	if message == "" && !gjson.ValidBytes(body) {
		message = strings.ToLower(strings.TrimSpace(string(body)))
	}
	return strings.Contains(message, "unknown provider for model") || strings.Contains(message, "unknown model") || strings.Contains(message, "model not found") || strings.Contains(message, "model is not supported")
}
