//go:build unit

package service

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestCalculateCumulativeRefundAmountsHalfRefund(t *testing.T) {
	result, err := CalculateCumulativeRefundAmounts(CumulativeRefundInput{
		PrincipalAmount:    refundDecimal(t, "100"),
		FeeAmount:          refundDecimal(t, "3"),
		RechargePoints:     refundDecimal(t, "100"),
		BonusPoints:        refundDecimal(t, "10"),
		RequestedPrincipal: refundDecimal(t, "50"),
	})
	require.NoError(t, err)
	require.Equal(t, "50", result.PrincipalDelta.String())
	require.Equal(t, "1.5", result.FeeDelta.String())
	require.Equal(t, "51.5", result.GatewayDelta.String())
	require.Equal(t, "50", result.RechargePointsDelta.String())
	require.Equal(t, "5", result.BonusPointsDelta.String())
	require.Equal(t, "55", result.PointsDelta.String())
}

func TestCalculateCumulativeRefundAmountsMultiplePartialRefundsAbsorbRoundingAtFullRefund(t *testing.T) {
	order := CumulativeRefundInput{
		PrincipalAmount: refundDecimal(t, "100"),
		FeeAmount:       refundDecimal(t, "3"),
		RechargePoints:  refundDecimal(t, "100"),
		BonusPoints:     refundDecimal(t, "20"),
	}

	first, err := CalculateCumulativeRefundAmounts(withRefundPrincipal(order, "33.33", t))
	require.NoError(t, err)
	require.Equal(t, "34.33", first.GatewayDelta.String())
	require.Equal(t, "6.666", first.BonusPointsDelta.String())

	secondInput := order
	secondInput.PreviousPrincipal = first.TargetPrincipal
	secondInput.PreviousFee = first.TargetFee
	secondInput.PreviousGateway = first.TargetGateway
	secondInput.PreviousRechargePoints = first.TargetRechargePoints
	secondInput.PreviousBonusPoints = first.TargetBonusPoints
	secondInput.RequestedPrincipal = refundDecimal(t, "33.33")
	second, err := CalculateCumulativeRefundAmounts(secondInput)
	require.NoError(t, err)
	require.Equal(t, "34.33", second.GatewayDelta.String())
	require.Equal(t, "6.666", second.BonusPointsDelta.String())

	finalInput := order
	finalInput.PreviousPrincipal = second.TargetPrincipal
	finalInput.PreviousFee = second.TargetFee
	finalInput.PreviousGateway = second.TargetGateway
	finalInput.PreviousRechargePoints = second.TargetRechargePoints
	finalInput.PreviousBonusPoints = second.TargetBonusPoints
	finalInput.RequestedPrincipal = refundDecimal(t, "33.34")
	final, err := CalculateCumulativeRefundAmounts(finalInput)
	require.NoError(t, err)
	require.Equal(t, "34.34", final.GatewayDelta.String())
	require.Equal(t, "6.668", final.BonusPointsDelta.String())
	require.Equal(t, "103", final.TargetGateway.String())
	require.Equal(t, "100", final.TargetRechargePoints.String())
	require.Equal(t, "20", final.TargetBonusPoints.String())
}

func TestCalculateCumulativeRefundAmountsSupportsDifferentRechargePointTotal(t *testing.T) {
	result, err := CalculateCumulativeRefundAmounts(CumulativeRefundInput{
		PrincipalAmount:    refundDecimal(t, "100"),
		FeeAmount:          refundDecimal(t, "3"),
		RechargePoints:     refundDecimal(t, "125"),
		BonusPoints:        refundDecimal(t, "10"),
		RequestedPrincipal: refundDecimal(t, "20"),
	})
	require.NoError(t, err)
	require.Equal(t, "25", result.RechargePointsDelta.String())
	require.Equal(t, "2", result.BonusPointsDelta.String())
	require.Equal(t, "20.6", result.GatewayDelta.String())
}

func TestCalculateCumulativeRefundAmountsRejectsInvalidState(t *testing.T) {
	tests := []struct {
		name  string
		input CumulativeRefundInput
	}{
		{
			name: "more than remaining principal",
			input: CumulativeRefundInput{
				PrincipalAmount: refundDecimal(t, "100"), PreviousPrincipal: refundDecimal(t, "80"),
				PreviousGateway: refundDecimal(t, "80"), RequestedPrincipal: refundDecimal(t, "20.01"),
			},
		},
		{
			name: "fractional cent",
			input: CumulativeRefundInput{
				PrincipalAmount: refundDecimal(t, "100"), RequestedPrincipal: refundDecimal(t, "0.001"),
			},
		},
		{
			name: "inconsistent previous cumulative state",
			input: CumulativeRefundInput{
				PrincipalAmount: refundDecimal(t, "100"), FeeAmount: refundDecimal(t, "3"),
				PreviousPrincipal: refundDecimal(t, "50"), PreviousFee: refundDecimal(t, "1.49"),
				PreviousGateway: refundDecimal(t, "51.49"), RequestedPrincipal: refundDecimal(t, "10"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := CalculateCumulativeRefundAmounts(test.input)
			require.Error(t, err)
		})
	}
}

func TestMaxAffordableRefundPrincipalClampsToPointCapacity(t *testing.T) {
	input := CumulativeRefundInput{
		PrincipalAmount: refundDecimal(t, "100"),
		FeeAmount:       refundDecimal(t, "3"),
		RechargePoints:  refundDecimal(t, "100"),
		BonusPoints:     refundDecimal(t, "10"),
	}
	remaining := refundDecimal(t, "100")

	require.Equal(t, "50", maxAffordableRefundPrincipal(
		input, remaining, refundDecimal(t, "50"), refundDecimal(t, "5"), decimal.Zero, true,
	).String())
	require.Equal(t, "20", maxAffordableRefundPrincipal(
		input, remaining, refundDecimal(t, "20"), refundDecimal(t, "10"), decimal.Zero, true,
	).String())
	require.Equal(t, "0", maxAffordableRefundPrincipal(
		input, remaining, refundDecimal(t, "100"), refundDecimal(t, "10"), decimal.Zero, false,
	).String())
}

func TestWithinSelfServiceRefundWindowUsesHalfOpenBoundary(t *testing.T) {
	completed := time.Date(2026, time.August, 1, 12, 0, 0, 0, time.UTC)
	deadline := completed.Add(168 * time.Hour)

	require.False(t, withinSelfServiceRefundWindow(completed, deadline, completed.Add(-time.Nanosecond)))
	require.True(t, withinSelfServiceRefundWindow(completed, deadline, completed))
	require.True(t, withinSelfServiceRefundWindow(completed, deadline, deadline.Add(-time.Nanosecond)))
	require.False(t, withinSelfServiceRefundWindow(completed, deadline, deadline))
}

func withRefundPrincipal(input CumulativeRefundInput, requested string, t *testing.T) CumulativeRefundInput {
	t.Helper()
	input.RequestedPrincipal = refundDecimal(t, requested)
	return input
}

func refundDecimal(t *testing.T, value string) decimal.Decimal {
	t.Helper()
	result, err := decimal.NewFromString(value)
	require.NoError(t, err)
	return result
}
