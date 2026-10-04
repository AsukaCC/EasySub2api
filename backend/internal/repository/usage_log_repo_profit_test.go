package repository

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/pkg/timezone"
	"github.com/AsukaCC/EasySub2api/internal/pkg/usagestats"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestAccountProfit30DayWindow(t *testing.T) {
	for _, tc := range []struct{ day, start, end string }{
		{"2026-10-04", "2026-09-05", "2026-10-05"},
		{"2026-01-01", "2025-12-03", "2026-01-02"},
		{"2024-03-01", "2024-02-01", "2024-03-02"},
	} {
		t.Run(tc.day, func(t *testing.T) {
			now, err := time.ParseInLocation(time.DateOnly, tc.day, timezone.Location())
			require.NoError(t, err)
			for _, clock := range []time.Time{now, now.Add(23*time.Hour + 59*time.Minute)} {
				from, to := accountProfit30DayWindow(clock)
				require.Equal(t, tc.start, from.Format(time.DateOnly))
				require.Equal(t, tc.end, to.Format(time.DateOnly))
				require.Equal(t, 0, from.Hour())
				require.Equal(t, 0, to.Hour())
			}
		})
	}
}

func TestGetAccountProfit30DaysIndependentOfHistoryAndExpiry(t *testing.T) {
	for _, cost := range []driver.Value{nil, float64(25)} {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		t.Cleanup(func() { _ = db.Close() })
		repo := &usageLogRepository{sql: db}
		today := timezone.Today()
		tomorrow := today.AddDate(0, 0, 1)
		created := today.AddDate(0, 0, -60)
		mock.ExpectQuery(`SELECT a.created_at, aps.subscription_cost_points`).WithArgs("account").
			WillReturnRows(sqlmock.NewRows([]string{"created_at", "cost"}).AddRow(created, cost))
		mock.ExpectExec(`(?s)WITH aggregated AS.*INSERT INTO account_profit_daily_rollups`).
			WithArgs(sqlmock.AnyArg(), resolveUsageStatsTimezone(), created, tomorrow).WillReturnResult(sqlmock.NewResult(0, 3))
		for _, from := range []time.Time{today, timezone.StartOfWeek(today), timezone.StartOfMonth(today), today.AddDate(0, 0, -6), created, today.AddDate(0, 0, -29)} {
			mock.ExpectQuery(`(?s)SELECT COALESCE\(SUM\(revenue_points\).*bucket_date >= \$2::date AND bucket_date < \$3::date`).
				WithArgs("account", from.Format(time.DateOnly), tomorrow.Format(time.DateOnly)).
				WillReturnRows(sqlmock.NewRows([]string{"revenue", "cost", "profit", "requests", "tokens"}).AddRow(100, 30, 70, 2, 400))
		}
		mock.ExpectQuery(`(?s)SELECT TO_CHAR.*generate_series`).
			WithArgs("account", today.AddDate(0, 0, -3).Format(time.DateOnly), tomorrow.Format(time.DateOnly), 30, 0).
			WillReturnRows(sqlmock.NewRows([]string{"date", "revenue", "cost", "profit", "requests", "tokens"}))
		result, err := repo.GetAccountProfit(context.Background(), "account", today.AddDate(0, 0, -3), tomorrow, 1, 30)
		require.NoError(t, err)
		require.Equal(t, 100.0, result.Period30d.RevenuePoints)
		require.EqualValues(t, 2, result.Period30d.Requests)
		if cost == nil {
			require.Equal(t, 70.0, result.Period30d.ProfitPoints)
			require.Nil(t, result.Period30d.SubscriptionCostPoints)
		} else {
			require.Equal(t, 75.0, result.Period30d.ProfitPoints)
			require.Equal(t, 25.0, *result.Period30d.SubscriptionCostPoints)
		}
		body, err := json.Marshal(result)
		require.NoError(t, err)
		require.Contains(t, string(body), `"period_30d"`)
		require.NoError(t, mock.ExpectationsWereMet())
	}
}

func TestListAccountProfit30DaysIncludesAccountsWithoutExpiry(t *testing.T) {
	today := timezone.Today()
	for _, expires := range []driver.Value{nil, today.AddDate(0, 0, -60), today.AddDate(0, 0, 100)} {
		matcher := sqlmock.QueryMatcherFunc(func(expected, actual string) error {
			if expected == "profit-list" {
				require.Contains(t, actual, "WHERE r.bucket_date >= ($2::date - 29)")
				require.Contains(t, actual, "AND r.bucket_date < $3::date")
				require.NotContains(t, actual, "a.expires_at AT TIME ZONE")
				require.NotContains(t, actual, "a.expires_at IS NULL")
				require.Contains(t, actual, "ORDER BY period_30d_revenue desc NULLS LAST")
				return nil
			}
			return sqlmock.QueryMatcherRegexp.Match(expected, actual)
		})
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(matcher))
		require.NoError(t, err)
		t.Cleanup(func() { _ = db.Close() })
		repo := &usageLogRepository{sql: db}
		created := today.AddDate(0, 0, -60)
		mock.ExpectQuery(`SELECT a.id::text, a.created_at`).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow("account", created))
		mock.ExpectExec(`(?s)WITH aggregated AS.*INSERT INTO account_profit_daily_rollups`).WillReturnResult(sqlmock.NewResult(0, 3))
		columns := strings.Split("id,name,platform,tier,status,created,expires,updated,extra,subscription,quota,weekRevenue,weekCost,weekProfit,weekRequests,weekTokens,monthRevenue,monthCost,monthProfit,monthRequests,monthTokens,lifetimeRevenue,lifetimeCost,lifetimeProfit,lifetimeRequests,lifetimeTokens", ",")
		mock.ExpectQuery("profit-list").WithArgs(resolveUsageStatsTimezone(), today.Format(time.DateOnly), today.AddDate(0, 0, 1).Format(time.DateOnly), 20, 0).
			WillReturnRows(sqlmock.NewRows(columns).AddRow("account", "Fixture", "openai", "", "active", created, expires, today, []byte(`{}`), 25, nil, 10, 3, 7, 1, 40, 100, 30, 75, 2, 400, 200, 60, 140, 5, 800))
		result, err := repo.ListAccountProfit(context.Background(), usagestats.AccountProfitListParams{Page: 1, PageSize: 20, SortBy: "period_30d_revenue"})
		require.NoError(t, err)
		require.Len(t, result.Items, 1)
		require.Equal(t, 100.0, result.Items[0].Period30d.RevenuePoints)
		require.Equal(t, 25.0, *result.Items[0].Period30d.SubscriptionCostPoints)
		require.NoError(t, mock.ExpectationsWereMet())
	}
}
