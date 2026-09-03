package service

import (
	"testing"

	appTimezone "github.com/AsukaCC/EasySub2api/internal/pkg/timezone"
	"github.com/stretchr/testify/require"
)

func TestNormalizeInactiveUserFilterUsesServerTimezoneBoundary(t *testing.T) {
	require.NoError(t, appTimezone.Init("Asia/Shanghai"))
	t.Cleanup(func() { _ = appTimezone.Init("UTC") })

	today := appTimezone.StartOfDay(appTimezone.Now())
	normalized, err := normalizeInactiveUserFilter(InactiveUserFilter{
		LastUsedBefore: today,
	})
	require.NoError(t, err)
	require.Equal(t, today.UTC(), normalized.LastUsedBefore)
	require.False(t, normalized.EvaluationTime.IsZero())

	_, err = normalizeInactiveUserFilter(InactiveUserFilter{
		LastUsedBefore: today.AddDate(0, 0, 1),
	})
	require.Error(t, err)
}
