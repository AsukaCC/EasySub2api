package handler

import (
	"encoding/json"
	"testing"

	dbent "github.com/AsukaCC/EasySub2api/ent"
	"github.com/AsukaCC/EasySub2api/internal/service"
)

func TestCheckoutPlansForResponseIncludesPointAliases(t *testing.T) {
	daily, weekly, monthly := 10.0, 50.0, 120.0
	originalPrice := 24.0
	plans := []*dbent.SubscriptionPlan{{
		ID: "plan-1", GroupID: "group-1", Name: "Pro", Price: 20,
		OriginalPrice: &originalPrice, Features: "One\nTwo",
	}}
	groupInfo := map[string]service.PlanGroupInfo{
		"group-1": {
			DailyLimitUSD: &daily, WeeklyLimitUSD: &weekly, MonthlyLimitUSD: &monthly,
		},
	}

	got := checkoutPlansForResponse(plans, groupInfo)
	if len(got) != 1 {
		t.Fatalf("expected one plan, got %d", len(got))
	}
	plan := got[0]
	assertSameOptionalAmount(t, "daily", plan.DailyLimitUSD, plan.DailyLimitPoints)
	assertSameOptionalAmount(t, "weekly", plan.WeeklyLimitUSD, plan.WeeklyLimitPoints)
	assertSameOptionalAmount(t, "monthly", plan.MonthlyLimitUSD, plan.MonthlyLimitPoints)
	if plan.PricePoints != plan.Price {
		t.Fatalf("price_points = %v, want price %v", plan.PricePoints, plan.Price)
	}
	assertSameOptionalAmount(t, "original price", plan.OriginalPrice, plan.OriginalPricePoints)

	body, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("marshal checkout plan: %v", err)
	}
	for _, field := range []string{
		"daily_limit_points", "weekly_limit_points", "monthly_limit_points",
		"price_points", "original_price_points",
	} {
		var payload map[string]json.RawMessage
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("unmarshal checkout plan: %v", err)
		}
		if _, ok := payload[field]; !ok {
			t.Fatalf("expected %q in response: %s", field, body)
		}
	}
}

func TestPaymentOrderResponseIncludesSuccessfulRefundSplits(t *testing.T) {
	order := &dbent.PaymentOrder{
		ID: "order-1", RefundAmount: 25,
		RefundedPrincipalAmount: 25, RefundedFeeAmount: 1.25, RefundedGatewayAmount: 26.25,
		ReversedBasePoints: 20, ReversedBonusPoints: 5, ReversedAffiliatePoints: 2.5,
	}
	got := sanitizePaymentOrderForResponse(order)
	if got == nil {
		t.Fatal("expected payment order response")
	}
	if got.RefundAmount != got.RefundedPrincipalAmount {
		t.Fatalf("refund_amount=%v must equal refunded principal=%v", got.RefundAmount, got.RefundedPrincipalAmount)
	}
	if got.RefundedGatewayAmount != 26.25 || got.ReversedBasePoints != 20 ||
		got.ReversedBonusPoints != 5 || got.ReversedAffiliatePoints != 2.5 {
		t.Fatalf("refund split was not preserved: %+v", got)
	}
}

func assertSameOptionalAmount(t *testing.T, name string, legacy, points *float64) {
	t.Helper()
	if legacy == nil || points == nil || *legacy != *points {
		t.Fatalf("%s aliases differ: legacy=%v points=%v", name, legacy, points)
	}
}
