package service

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

const (
	refundMoneyScale  int32 = 2
	refundPointsScale int32 = 8
)

// CumulativeRefundInput contains immutable order totals and the previously
// finalized cumulative refund values. RequestedPrincipal is the incremental
// CNY principal amount requested by the current refund.
type CumulativeRefundInput struct {
	PrincipalAmount decimal.Decimal
	FeeAmount       decimal.Decimal
	RechargePoints  decimal.Decimal
	BonusPoints     decimal.Decimal

	PreviousPrincipal      decimal.Decimal
	PreviousFee            decimal.Decimal
	PreviousGateway        decimal.Decimal
	PreviousRechargePoints decimal.Decimal
	PreviousBonusPoints    decimal.Decimal
	RequestedPrincipal     decimal.Decimal
}

// CumulativeRefundAmounts carries both the new cumulative targets and the
// incremental amounts that must be sent to the gateway and wallet.
type CumulativeRefundAmounts struct {
	TargetPrincipal      decimal.Decimal
	TargetFee            decimal.Decimal
	TargetGateway        decimal.Decimal
	TargetRechargePoints decimal.Decimal
	TargetBonusPoints    decimal.Decimal

	PrincipalDelta      decimal.Decimal
	FeeDelta            decimal.Decimal
	GatewayDelta        decimal.Decimal
	RechargePointsDelta decimal.Decimal
	BonusPointsDelta    decimal.Decimal
	PointsDelta         decimal.Decimal
}

// CalculateCumulativeRefundAmounts calculates refund deltas from cumulative
// targets. The final refund uses the immutable order totals exactly so that
// earlier rounding can never leave a cent or point stranded.
func CalculateCumulativeRefundAmounts(input CumulativeRefundInput) (CumulativeRefundAmounts, error) {
	if err := validateCumulativeRefundInput(input); err != nil {
		return CumulativeRefundAmounts{}, err
	}

	targetPrincipal := input.PreviousPrincipal.Add(input.RequestedPrincipal)
	targetFee := proportionalRefundTarget(input.FeeAmount, targetPrincipal, input.PrincipalAmount, refundMoneyScale)
	targetRecharge := proportionalRefundTarget(input.RechargePoints, targetPrincipal, input.PrincipalAmount, refundPointsScale)
	targetBonus := proportionalRefundTarget(input.BonusPoints, targetPrincipal, input.PrincipalAmount, refundPointsScale)
	targetGateway := targetPrincipal.Add(targetFee).Round(refundMoneyScale)

	result := CumulativeRefundAmounts{
		TargetPrincipal:      targetPrincipal,
		TargetFee:            targetFee,
		TargetGateway:        targetGateway,
		TargetRechargePoints: targetRecharge,
		TargetBonusPoints:    targetBonus,
		PrincipalDelta:       input.RequestedPrincipal,
		FeeDelta:             targetFee.Sub(input.PreviousFee),
		GatewayDelta:         targetGateway.Sub(input.PreviousGateway),
		RechargePointsDelta:  targetRecharge.Sub(input.PreviousRechargePoints),
		BonusPointsDelta:     targetBonus.Sub(input.PreviousBonusPoints),
	}
	result.PointsDelta = result.RechargePointsDelta.Add(result.BonusPointsDelta).Round(refundPointsScale)

	if result.FeeDelta.IsNegative() || result.GatewayDelta.IsNegative() ||
		result.RechargePointsDelta.IsNegative() || result.BonusPointsDelta.IsNegative() {
		return CumulativeRefundAmounts{}, fmt.Errorf("cumulative refund state exceeds the requested target")
	}
	return result, nil
}

func validateCumulativeRefundInput(input CumulativeRefundInput) error {
	if !input.PrincipalAmount.IsPositive() {
		return fmt.Errorf("principal amount must be positive")
	}
	if !input.RequestedPrincipal.IsPositive() {
		return fmt.Errorf("requested principal must be positive")
	}
	for name, value := range map[string]decimal.Decimal{
		"fee amount":               input.FeeAmount,
		"recharge points":          input.RechargePoints,
		"bonus points":             input.BonusPoints,
		"previous principal":       input.PreviousPrincipal,
		"previous fee":             input.PreviousFee,
		"previous gateway":         input.PreviousGateway,
		"previous recharge points": input.PreviousRechargePoints,
		"previous bonus points":    input.PreviousBonusPoints,
	} {
		if value.IsNegative() {
			return fmt.Errorf("%s must be nonnegative", name)
		}
	}
	if !hasAtMostScale(input.PrincipalAmount, refundMoneyScale) ||
		!hasAtMostScale(input.FeeAmount, refundMoneyScale) ||
		!hasAtMostScale(input.PreviousPrincipal, refundMoneyScale) ||
		!hasAtMostScale(input.PreviousFee, refundMoneyScale) ||
		!hasAtMostScale(input.PreviousGateway, refundMoneyScale) ||
		!hasAtMostScale(input.RequestedPrincipal, refundMoneyScale) {
		return fmt.Errorf("money amounts allow at most %d decimal places", refundMoneyScale)
	}
	if !hasAtMostScale(input.RechargePoints, refundPointsScale) ||
		!hasAtMostScale(input.BonusPoints, refundPointsScale) ||
		!hasAtMostScale(input.PreviousRechargePoints, refundPointsScale) ||
		!hasAtMostScale(input.PreviousBonusPoints, refundPointsScale) {
		return fmt.Errorf("point amounts allow at most %d decimal places", refundPointsScale)
	}
	if input.PreviousPrincipal.GreaterThan(input.PrincipalAmount) ||
		input.PreviousPrincipal.Add(input.RequestedPrincipal).GreaterThan(input.PrincipalAmount) {
		return fmt.Errorf("requested principal exceeds the refundable remainder")
	}

	expectedPreviousFee := proportionalRefundTarget(input.FeeAmount, input.PreviousPrincipal, input.PrincipalAmount, refundMoneyScale)
	expectedPreviousGateway := input.PreviousPrincipal.Add(expectedPreviousFee).Round(refundMoneyScale)
	expectedPreviousRecharge := proportionalRefundTarget(input.RechargePoints, input.PreviousPrincipal, input.PrincipalAmount, refundPointsScale)
	expectedPreviousBonus := proportionalRefundTarget(input.BonusPoints, input.PreviousPrincipal, input.PrincipalAmount, refundPointsScale)
	if !input.PreviousFee.Equal(expectedPreviousFee) ||
		!input.PreviousGateway.Equal(expectedPreviousGateway) ||
		!input.PreviousRechargePoints.Equal(expectedPreviousRecharge) ||
		!input.PreviousBonusPoints.Equal(expectedPreviousBonus) {
		return fmt.Errorf("previous cumulative refund values are inconsistent with the order totals")
	}
	return nil
}

func proportionalRefundTarget(total, targetPrincipal, principal decimal.Decimal, scale int32) decimal.Decimal {
	if targetPrincipal.IsZero() || total.IsZero() {
		return decimal.Zero
	}
	if targetPrincipal.Equal(principal) {
		return total
	}
	return total.Mul(targetPrincipal).Div(principal).Round(scale)
}

// maxAffordableRefundPrincipal returns the largest whole-cent principal whose
// point recovery fits the currently available source bonus and recharge
// points. It is monotonic, so a binary search avoids float-based estimates.
func maxAffordableRefundPrincipal(input CumulativeRefundInput, remaining, rechargeAvailable, sourceBonusAvailable, expiredAvailable decimal.Decimal, sourceUsable bool) decimal.Decimal {
	if !sourceUsable {
		return decimal.Zero
	}
	high := remaining.Mul(decimal.NewFromInt(100)).IntPart()
	low := int64(0)
	for low < high {
		mid := (low + high + 1) / 2
		candidate := decimal.New(mid, -refundMoneyScale)
		trial := input
		trial.RequestedPrincipal = candidate
		amounts, err := CalculateCumulativeRefundAmounts(trial)
		if err != nil {
			high = mid - 1
			continue
		}
		bonusCovered := decimal.Min(amounts.BonusPointsDelta, expiredAvailable.Add(sourceBonusAvailable))
		rechargeNeeded := amounts.RechargePointsDelta.Add(amounts.BonusPointsDelta.Sub(bonusCovered)).Round(refundPointsScale)
		if rechargeNeeded.LessThanOrEqual(rechargeAvailable) {
			low = mid
		} else {
			high = mid - 1
		}
	}
	return decimal.New(low, -refundMoneyScale)
}

func withinSelfServiceRefundWindow(completedAt, deadline, now time.Time) bool {
	return !now.Before(completedAt) && now.Before(deadline)
}

func hasAtMostScale(value decimal.Decimal, scale int32) bool {
	return value.Equal(value.Round(scale))
}
