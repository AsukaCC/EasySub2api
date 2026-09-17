package service

import "strings"

const (
	SiteBillingModeRechargeAndSubscription = "recharge_and_subscription"
	SiteBillingModeRechargeOnly            = "recharge_only"
	SiteBillingModeSubscriptionOnly        = "subscription_only"
)

func siteBillingMode(subscriptionEnabled, balanceDisabled bool) string {
	if !subscriptionEnabled {
		return SiteBillingModeRechargeOnly
	}
	if balanceDisabled {
		return SiteBillingModeSubscriptionOnly
	}
	return SiteBillingModeRechargeAndSubscription
}

// NormalizeSiteBillingMode returns the canonical site billing mode. Empty and
// unknown values preserve the historical default where both purchase paths are
// available.
func NormalizeSiteBillingMode(mode string) string {
	switch strings.TrimSpace(mode) {
	case SiteBillingModeRechargeOnly:
		return SiteBillingModeRechargeOnly
	case SiteBillingModeSubscriptionOnly:
		return SiteBillingModeSubscriptionOnly
	case SiteBillingModeRechargeAndSubscription:
		return SiteBillingModeRechargeAndSubscription
	default:
		return SiteBillingModeRechargeAndSubscription
	}
}

// SiteBillingModeSettings translates the public three-state value into the two
// compatibility settings that existing deployments already persist.
func SiteBillingModeSettings(mode string) (subscriptionEnabled, balanceDisabled bool) {
	switch NormalizeSiteBillingMode(mode) {
	case SiteBillingModeRechargeOnly:
		return false, false
	case SiteBillingModeSubscriptionOnly:
		return true, true
	default:
		return true, false
	}
}
