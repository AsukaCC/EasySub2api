package admin

import (
	"testing"

	dbent "github.com/AsukaCC/EasySub2api/ent"
	"github.com/AsukaCC/EasySub2api/internal/service"
)

func TestAdminSubscriptionPlansForResponseIncludesPointAliases(t *testing.T) {
	weekly := 75.0
	originalPrice := 39.0
	plans := []*dbent.SubscriptionPlan{{
		ID: "plan-1", GroupID: "group-1", Name: "Team", Price: 30,
		OriginalPrice: &originalPrice,
	}}
	groupInfo := map[string]service.PlanGroupInfo{
		"group-1": {WeeklyLimitUSD: &weekly},
	}

	got := adminSubscriptionPlansForResponse(plans, groupInfo)
	if len(got) != 1 {
		t.Fatalf("expected one plan, got %d", len(got))
	}
	plan := got[0]
	if plan.WeeklyLimitUSD == nil || plan.WeeklyLimitPoints == nil ||
		*plan.WeeklyLimitUSD != *plan.WeeklyLimitPoints {
		t.Fatalf("weekly aliases differ: usd=%v points=%v", plan.WeeklyLimitUSD, plan.WeeklyLimitPoints)
	}
	if plan.PricePoints != plan.Price {
		t.Fatalf("price_points = %v, want price %v", plan.PricePoints, plan.Price)
	}
	if plan.OriginalPrice == nil || plan.OriginalPricePoints == nil ||
		*plan.OriginalPrice != *plan.OriginalPricePoints {
		t.Fatalf("original price aliases differ: legacy=%v points=%v", plan.OriginalPrice, plan.OriginalPricePoints)
	}
}

func TestAdminPaymentOrderResponseIncludesSuccessfulRefundSplits(t *testing.T) {
	order := &dbent.PaymentOrder{
		ID: "order-1", RefundAmount: 50,
		RefundedPrincipalAmount: 50, RefundedFeeAmount: 1.5, RefundedGatewayAmount: 51.5,
		ReversedBasePoints: 45, ReversedBonusPoints: 10, ReversedAffiliatePoints: 3,
	}
	got := sanitizeAdminPaymentOrderForResponse(order)
	if got == nil {
		t.Fatal("expected admin payment order response")
	}
	if got.RefundAmount != got.RefundedPrincipalAmount {
		t.Fatalf("refund_amount=%v must equal refunded principal=%v", got.RefundAmount, got.RefundedPrincipalAmount)
	}
	if got.RefundedGatewayAmount != 51.5 || got.ReversedBasePoints != 45 ||
		got.ReversedBonusPoints != 10 || got.ReversedAffiliatePoints != 3 {
		t.Fatalf("refund split was not preserved: %+v", got)
	}
}
