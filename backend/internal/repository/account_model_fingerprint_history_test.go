package repository

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestModelFingerprintHistoryPaginationAndExpiry(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := &accountRepository{sql: db}
	snapshot := service.ModelFingerprintSnapshot{ID: "job", StartedAt: time.Now(), ExpiresAt: time.Now().Add(time.Minute), Status: "running"}
	body, _ := json.Marshal([]service.ModelFingerprintSnapshot{snapshot})
	mock.ExpectQuery(`(?s)WITH filtered AS.*started_at > \$2.*ORDER BY started_at DESC, id DESC LIMIT \$3 OFFSET \$4.*count\(\*\).*jsonb_agg`).WithArgs("account", sqlmock.AnyArg(), 10, 10).
		WillReturnRows(sqlmock.NewRows([]string{"count", "snapshots"}).AddRow(11, body))
	page, err := repo.ListModelFingerprintHistory(context.Background(), "account", 2, 10)
	require.NoError(t, err)
	require.Equal(t, 11, page.Total)
	require.Len(t, page.Items, 1)
	require.Equal(t, "job", page.Items[0].ID)
	mock.ExpectQuery(`(?s)WITH filtered AS`).WithArgs("account", sqlmock.AnyArg(), 100, 0).
		WillReturnRows(sqlmock.NewRows([]string{"count", "snapshots"}).AddRow(0, []byte(`[]`)))
	page, err = repo.ListModelFingerprintHistory(context.Background(), "account", -1, 1000)
	require.NoError(t, err)
	require.Empty(t, page.Items)
	require.Equal(t, 0, page.Total)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestModelFingerprintHistoryCleanupOnlyTargetsFingerprintRecords(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := &accountRepository{sql: db}
	now := time.Now()
	mock.ExpectExec(`(?s)DELETE FROM usage_logs.*request_type=6 AND request_id LIKE 'fingerprint:%' AND created_at <= \$1.*LIMIT 5000`).WithArgs(now.Add(-24 * time.Hour)).WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectExec(`(?s)DELETE FROM model_fingerprint_runs.*started_at <= \$1.*LIMIT 1000`).WithArgs(now.Add(-24 * time.Hour)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.CleanupModelFingerprintHistory(context.Background(), now))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestModelFingerprintScheduleOnlyLoadsActiveAdminPlans(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := &accountRepository{sql: db}
	now := time.Now()
	mock.ExpectQuery(`(?s)SELECT s.account_id.*s.enabled AND s.next_run_at <= \$1.*a.deleted_at IS NULL AND a.status='active'.*u.deleted_at IS NULL AND u.status='active' AND u.role='admin'.*ORDER BY s.next_run_at, s.account_id LIMIT 100`).WithArgs(now).
		WillReturnRows(sqlmock.NewRows([]string{"account_id", "user_id", "options", "next_run_at"}).AddRow("account", "admin", []byte(`{"model_id":"gpt-6-astra","protocol":"chat","reasoning_effort":"low"}`), now))
	plans, err := repo.ListDueModelFingerprints(context.Background(), now)
	require.NoError(t, err)
	require.Len(t, plans, 1)
	require.Equal(t, "admin", plans[0].UserID)
	require.Equal(t, "low", plans[0].Options.ReasoningEffort)
	require.NoError(t, mock.ExpectationsWereMet())
}
