package repository

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInactiveUserFilterUsesSuccessfulUsageAndAvailableBalance(t *testing.T) {
	require.Contains(t, inactiveUserFilterSQL, "MAX(ul.created_at) FILTER (WHERE ul.actual_cost > 0)")
	require.Contains(t, inactiveUserFilterSQL, "GREATEST(u.balance, 0) <= $1")
	require.Contains(t, inactiveUserFilterSQL, "activity.last_used_at < $2")
	require.Contains(t, inactiveUserFilterSQL, "activity.usage_7d <= $3")
	require.Contains(t, inactiveUserCandidateSelectSQL, "GREATEST(u.balance, 0)::double precision")
}
