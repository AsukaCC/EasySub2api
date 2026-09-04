package service

import (
	"strings"

	"github.com/AsukaCC/EasySub2api/internal/model"
	"github.com/AsukaCC/EasySub2api/internal/pkg/xai"
)

func ResolveAccountSubscriptionTier(account *Account) string {
	if account == nil {
		return ""
	}
	if account.Platform == PlatformGrok {
		usage := NewGrokQuotaFetcher().BuildUsageInfo(account)
		if usage != nil {
			if tier := xai.NormalizeSubscriptionTier(usage.SubscriptionTier); tier != "" {
				return tier
			}
		}
	}
	if account.Platform == PlatformGemini {
		if tier := firstNonBlankAccountTier(account.GetCredential("tier_id"), account.GetCredential("plan_type"), account.GetExtraString("subscription_tier")); tier != "" {
			return model.NormalizeSubscriptionTier(tier)
		}
	}
	return model.NormalizeSubscriptionTier(firstNonBlankAccountTier(
		account.GetCredential("plan_type"),
		account.GetCredential("chatgpt_plan_type"),
		account.GetCredential("subscription_tier"),
		account.GetExtraString("subscription_tier"),
		account.GetExtraString("plan_type"),
	))
}

func firstNonBlankAccountTier(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}
