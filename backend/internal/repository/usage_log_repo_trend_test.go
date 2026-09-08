//go:build unit

package repository

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func newTrendSQLMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return db, mock
}

func TestGetUserUsageTrendWithRoleScope_RegularUsesExistsNotJoin(t *testing.T) {
	db, mock := newTrendSQLMock(t)
	repo := &usageLogRepository{sql: db}
	loc, err := time.LoadLocation("Asia/Shanghai")
	require.NoError(t, err)
	start := time.Date(2026, 9, 7, 0, 0, 0, 0, loc)
	end := time.Date(2026, 9, 9, 0, 0, 0, 0, loc)

	// Dashboard users-trend defaults to role_scope=regular. JOIN users made
	// created_at ambiguous; the query must use EXISTS and qualify usage_logs.created_at.
	mock.ExpectQuery(regexp.QuoteMeta("EXISTS (SELECT 1 FROM users usage_scope_user WHERE usage_scope_user.id = usage_logs.user_id AND usage_scope_user.role <> 'admin')")).
		WithArgs(start, end, 12, "Asia/Shanghai").
		WillReturnRows(sqlmock.NewRows([]string{
			"date", "user_id", "email", "username", "requests", "tokens", "cost", "actual_cost",
		}))

	rows, err := repo.GetUserUsageTrendWithRoleScope(context.Background(), start, end, "hour", 12, "tokens", "regular")
	require.NoError(t, err)
	require.Empty(t, rows)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserUsageTrendWithRoleScope_AdminUsesExists(t *testing.T) {
	db, mock := newTrendSQLMock(t)
	repo := &usageLogRepository{sql: db}
	start := time.Date(2026, 9, 7, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 9, 9, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta("EXISTS (SELECT 1 FROM users usage_scope_user WHERE usage_scope_user.id = usage_logs.user_id AND usage_scope_user.role = 'admin')")).
		WithArgs(start, end, 12, "UTC").
		WillReturnRows(sqlmock.NewRows([]string{
			"date", "user_id", "email", "username", "requests", "tokens", "cost", "actual_cost",
		}))

	rows, err := repo.GetUserUsageTrendWithRoleScope(context.Background(), start, end, "hour", 12, "tokens", "admin")
	require.NoError(t, err)
	require.Empty(t, rows)
	require.NoError(t, mock.ExpectationsWereMet())
}
