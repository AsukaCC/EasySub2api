package handler

import (
	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestSingleUserLevelBillingResponse(t *testing.T) {
	level := .75
	response := buildKeyBillingInfoWithPlan(&service.APIKey{Group: &service.Group{RateMultiplier: .2, PeakRateEnabled: true, PeakRateMultiplier: 3}}, &service.UserRatePlan{
		EffectiveBaseMultiplier: .15, EffectiveMultiplier: .075, UserLevelMultiplier: &level, SelectedDynamicRuleID: "discount",
		DynamicCandidates: []service.DynamicRateCandidate{{RuleID: "discount", DiscountCoefficient: .5}},
	}, time.Now())
	require.Equal(t, .75, *response.UserLevelMultiplier)
	require.Equal(t, .75, *response.UserRateMultiplier)
	require.Equal(t, .5, *response.DynamicRateMultiplier)
	require.Equal(t, .15, response.EffectiveBaseMultiplier)
	require.Equal(t, .075, response.EffectiveRateMultiplier)
	require.False(t, response.PeakRateEnabled)
	require.Nil(t, response.PeakRateMultiplier)
}
