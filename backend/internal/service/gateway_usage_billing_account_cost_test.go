package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveAccountUsageCostMatchesAccountSevenDayStatsBasis(t *testing.T) {
	accountStatsCost := 4.0
	params := &postUsageBillingParams{
		Cost:                  &CostBreakdown{TotalCost: 2},
		AccountRateMultiplier: 1.5,
	}

	require.InDelta(t, 3.0, resolveAccountUsageCost(nil, params), 0.000000001)
	require.InDelta(t, 6.0, resolveAccountUsageCost(&UsageLog{AccountStatsCost: &accountStatsCost}, params), 0.000000001)
}

func TestBuildUsageBillingCommandUsesAccountSevenDayCostForDynamicQuota(t *testing.T) {
	accountStatsCost := 4.0
	params := &postUsageBillingParams{
		Cost:                  &CostBreakdown{TotalCost: 2, ActualCost: 1, BillingMode: string(BillingModeToken)},
		APIKey:                &APIKey{ID: "key-1"},
		User:                  &User{ID: "user-1"},
		Account:               &Account{ID: "account-1"},
		AccountRateMultiplier: 1.5,
		RatePlan: &UserRatePlan{
			GroupID:        "group-1",
			BaseMultiplier: 1,
			PeakMultiplier: 1,
			DynamicCandidates: []DynamicRateCandidate{{
				RuleID: "rule-1", QuotaKey: "2026-09-04T00:00:00Z", Multiplier: 0.5, SharedQuotaAmount: 10, PersonalQuotaAmount: 5,
			}},
		},
	}

	command := buildUsageBillingCommand("request-1", &UsageLog{AccountStatsCost: &accountStatsCost}, params)
	require.NotNil(t, command)
	require.NotNil(t, command.DynamicRatePlan)
	require.InDelta(t, 6.0, command.DynamicRatePlan.AccountCost, 0.000000001)
}
