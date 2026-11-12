package repository

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCodexCreditsSnapshotIsSchedulerNeutral(t *testing.T) {
	require.True(t, isSchedulerNeutralExtraKey("codex_credits_snapshot"))
	require.False(t, shouldEnqueueSchedulerOutboxForExtraUpdates(map[string]any{
		"codex_credits_snapshot": map[string]any{"fetched_at": 123},
	}))
}
