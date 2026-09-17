package service

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
