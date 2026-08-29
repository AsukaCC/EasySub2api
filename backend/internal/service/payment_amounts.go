package service

import (
	"math"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/payment"
	"github.com/shopspring/decimal"
)

const defaultBalanceRechargeMultiplier = 1.0
const rechargeBonusValidity = 168 * time.Hour
const rechargeRefundWindow = 168 * time.Hour

type rechargeOrderPricing struct {
	PrincipalAmount float64
	BasePoints      float64
	BonusPoints     float64
	CreditedPoints  float64
	BonusTier       *RechargeBonusTier
	BonusExpiresAt  *time.Time
}

// normalizeSubscriptionUSDToCNYRate keeps the legacy configuration field
// readable without allowing invalid values. It is not used for checkout.
func normalizeSubscriptionUSDToCNYRate(rate float64) float64 {
	if math.IsNaN(rate) || math.IsInf(rate, 0) || rate < 0 {
		return 0
	}
	return rate
}

func buildRechargeOrderPricing(principal float64, tiers []RechargeBonusTier) *rechargeOrderPricing {
	basePoints := decimal.NewFromFloat(principal).Round(2).InexactFloat64()
	var selected *RechargeBonusTier
	for i := range tiers {
		if basePoints+1e-9 < tiers[i].MinAmount {
			continue
		}
		candidate := tiers[i]
		if selected == nil || candidate.MinAmount > selected.MinAmount {
			selected = &candidate
		}
	}
	bonusPoints := 0.0
	if selected != nil {
		bonusPoints = selected.BonusPoints
	}
	creditedPoints := decimal.NewFromFloat(basePoints).
		Add(decimal.NewFromFloat(bonusPoints)).
		Round(8).
		InexactFloat64()
	return &rechargeOrderPricing{
		PrincipalAmount: basePoints,
		BasePoints:      basePoints,
		BonusPoints:     bonusPoints,
		CreditedPoints:  creditedPoints,
		BonusTier:       selected,
	}
}

func rechargeFeeAmount(payAmount, principal float64) float64 {
	fee := decimal.NewFromFloat(payAmount).
		Sub(decimal.NewFromFloat(principal)).
		Round(2).
		InexactFloat64()
	if fee < 0 {
		return 0
	}
	return fee
}

func calculateGatewayRefundAmount(orderAmount, payAmount, refundAmount float64, currency string) float64 {
	if orderAmount <= 0 || payAmount <= 0 || refundAmount <= 0 {
		return 0
	}
	fractionDigits := int32(payment.CurrencyMaxFractionDigits(currency))
	if math.Abs(refundAmount-orderAmount) <= paymentAmountToleranceForCurrency(currency) {
		return decimal.NewFromFloat(payAmount).Round(fractionDigits).InexactFloat64()
	}
	return decimal.NewFromFloat(payAmount).
		Mul(decimal.NewFromFloat(refundAmount)).
		Div(decimal.NewFromFloat(orderAmount)).
		Round(fractionDigits).
		InexactFloat64()
}
