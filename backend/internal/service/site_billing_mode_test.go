package service

import "testing"

func TestNormalizeSiteBillingMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "combined", in: SiteBillingModeRechargeAndSubscription, want: SiteBillingModeRechargeAndSubscription},
		{name: "recharge only", in: SiteBillingModeRechargeOnly, want: SiteBillingModeRechargeOnly},
		{name: "subscription only", in: SiteBillingModeSubscriptionOnly, want: SiteBillingModeSubscriptionOnly},
		{name: "empty defaults combined", want: SiteBillingModeRechargeAndSubscription},
		{name: "unknown defaults combined", in: "disabled", want: SiteBillingModeRechargeAndSubscription},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := NormalizeSiteBillingMode(tt.in); got != tt.want {
				t.Fatalf("NormalizeSiteBillingMode(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestSiteBillingModeSettings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		mode                string
		subscriptionEnabled bool
		balanceDisabled     bool
	}{
		{mode: SiteBillingModeRechargeAndSubscription, subscriptionEnabled: true},
		{mode: SiteBillingModeRechargeOnly},
		{mode: SiteBillingModeSubscriptionOnly, subscriptionEnabled: true, balanceDisabled: true},
		{mode: "invalid", subscriptionEnabled: true},
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			t.Parallel()
			subscriptionEnabled, balanceDisabled := SiteBillingModeSettings(tt.mode)
			if subscriptionEnabled != tt.subscriptionEnabled || balanceDisabled != tt.balanceDisabled {
				t.Fatalf("SiteBillingModeSettings(%q) = (%v, %v), want (%v, %v)", tt.mode, subscriptionEnabled, balanceDisabled, tt.subscriptionEnabled, tt.balanceDisabled)
			}
		})
	}
}
