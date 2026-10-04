package service

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestWalletSummarySignedBonus(t *testing.T) {
	for _, tc := range []struct{ recharge, bonus, available float64 }{
		{100, -20, 80}, {150, -20, 130}, {100, -10, 90}, {100, 10, 110}, {-10, 20, 10}, {-10, -20, 0},
	} {
		summary := NewWalletSummary(tc.recharge, tc.bonus, 0, 0)
		require.Equal(t, tc.available, summary.AvailableBalance)
		require.Equal(t, tc.bonus, summary.BonusBalance)
		require.Equal(t, tc.recharge+tc.bonus, summary.Balance)
		require.Equal(t, summary.Balance, summary.AvailableBalance-summary.OverdraftAmount)
	}
}
