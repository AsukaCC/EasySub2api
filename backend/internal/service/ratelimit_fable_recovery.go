package service

import (
	"context"
	"log/slog"
	"maps"
	"math"
)

func (a *Account) modelRateLimitReason(scope string) string {
	if a == nil {
		return ""
	}
	limits, _ := a.Extra[modelRateLimitsKey].(map[string]any)
	limit, _ := limits[scope].(map[string]any)
	reason, _ := limit["reason"].(string)
	return reason
}

func clearAccountModelRateLimitSnapshot(account *Account, scope string) {
	if account == nil {
		return
	}
	old, _ := account.Extra[modelRateLimitsKey].(map[string]any)
	limits := maps.Clone(old)
	delete(limits, scope)
	extra := maps.Clone(account.Extra)
	if extra == nil {
		return
	}
	extra[modelRateLimitsKey] = limits
	account.Extra = extra
}

func recoverAnthropicFableWindowLimit(ctx context.Context, repo AccountRepository, account *Account, utilization float64) bool {
	if repo == nil || account == nil || account.ID == "" {
		return false
	}
	if math.IsNaN(utilization) || math.IsInf(utilization, 0) || utilization < 0 || utilization >= 0.95 {
		return false
	}
	if !account.isRateLimitActiveForKey(anthropicFableRateLimitKey) {
		return false
	}
	if account.modelRateLimitReason(anthropicFableRateLimitKey) != anthropicFableWindowReason {
		return false
	}
	clearer, ok := repo.(interface {
		ClearModelRateLimitIfReason(context.Context, string, string, string) (bool, error)
	})
	if !ok {
		return false
	}
	cleared, err := clearer.ClearModelRateLimitIfReason(ctx, account.ID, anthropicFableRateLimitKey, anthropicFableWindowReason)
	if err != nil {
		slog.Warn("anthropic_fable_window_recovery_clear_failed",
			"account_id", account.ID,
			"utilization", utilization,
			"error", err)
		return false
	}
	if !cleared {
		return false
	}
	clearAccountModelRateLimitSnapshot(account, anthropicFableRateLimitKey)
	slog.Info("anthropic_fable_window_recovered",
		"account_id", account.ID,
		"scope", anthropicFableRateLimitKey,
		"utilization", utilization)
	return true
}
