package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestUsageLogRepositoryGetModelStatsAccountFilterKeepsPointsAndAccountCostSeparate(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := &usageLogRepository{sql: db}

	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	mock.ExpectQuery("SUM\\(actual_cost\\).*as actual_cost.*account_stats_cost.*as account_cost").
		WithArgs(start, end, "account-1").
		WillReturnRows(sqlmock.NewRows([]string{
			"model", "requests", "input_tokens", "output_tokens",
			"cache_creation_tokens", "cache_read_tokens", "total_tokens",
			"cost", "actual_cost", "account_cost",
		}).AddRow("gpt-5", int64(1), int64(10), int64(20), int64(0), int64(0), int64(30), 0.1, 0.24, 0.07))

	results, err := repo.GetModelStatsWithFilters(
		context.Background(), start, end,
		"", "", "account-1", "", nil, nil, nil,
	)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, 0.24, results[0].ActualCost)
	require.Equal(t, 0.07, results[0].AccountCost)
	require.NoError(t, mock.ExpectationsWereMet())
}
