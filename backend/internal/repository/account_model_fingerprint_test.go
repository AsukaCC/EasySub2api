package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/AsukaCC/EasySub2api/ent"
	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestModelFingerprintCleanup(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := &accountRepository{sql: db}
	now := time.Now()
	query := `(?s)UPDATE accounts\s+SET extra = extra - 'model_fingerprint'.*WHERE extra \? 'model_fingerprint' AND COALESCE\(.*finished_at.*expires_at.*started_at.*\) <= \$1\s+RETURNING id`
	for _, ids := range [][]string{{"expired-one", "expired-two"}, {}} {
		rows := sqlmock.NewRows([]string{"id"})
		for _, id := range ids {
			rows.AddRow(id)
		}
		mock.ExpectQuery(query).WithArgs(now.Add(-24 * time.Hour)).WillReturnRows(rows).RowsWillBeClosed()
		removed, err := repo.DeleteExpiredModelFingerprints(context.Background(), now)
		require.NoError(t, err)
		require.Equal(t, int64(len(ids)), removed)
	}
	mock.ExpectQuery(query).WithArgs(now.Add(-24 * time.Hour)).WillReturnError(errors.New("database unavailable"))
	_, err = repo.DeleteExpiredModelFingerprints(context.Background(), now)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestModelFingerprintLeaseConflictAndLostWorker(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	t.Cleanup(func() { _ = client.Close() })
	repo := &accountRepository{client: client}
	snapshot := &service.ModelFingerprintSnapshot{ID: "job", Status: "running", StartedAt: time.Now(), ExpiresAt: time.Now().Add(5 * time.Minute)}
	for _, rows := range []int64{0, 1} {
		mock.ExpectExec(`(?s)UPDATE accounts.*deleted_at IS NULL.*status.*running.*expires_at.*timestamptz`).
			WithArgs(sqlmock.AnyArg(), "account", snapshot.StartedAt, nil, nil).WillReturnResult(sqlmock.NewResult(0, rows))
		claimed, err := repo.ClaimModelFingerprint(context.Background(), "account", snapshot)
		require.NoError(t, err)
		require.Equal(t, rows == 1, claimed)
	}
	mock.ExpectExec(`(?s)UPDATE accounts.*deleted_at IS NULL AND extra->'model_fingerprint'->>'id' = \$3`).
		WithArgs(sqlmock.AnyArg(), "account", "job").WillReturnResult(sqlmock.NewResult(0, 0))
	require.ErrorContains(t, repo.SaveModelFingerprint(context.Background(), "account", snapshot), "lease lost")
	require.False(t, shouldEnqueueSchedulerOutboxForExtraUpdates(map[string]any{service.ModelFingerprintExtraKey: snapshot}))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestModelFingerprintScheduledClaimWindowAndAtomicAdvance(t *testing.T) {
	for _, tc := range []struct {
		name, now, due, next string
		window               string
		rows                 int64
	}{
		{"hourly", "2026-10-09T01:00:00Z", "2026-10-09T01:00:00Z", "2026-10-09T02:00:00Z", `,"time_window":{"start_hour":9,"end_hour":18,"timezone":"Asia/Shanghai"}`, 1},
		{"next day", "2026-10-09T09:00:00Z", "2026-10-09T09:00:00Z", "2026-10-10T01:00:00Z", `,"time_window":{"start_hour":9,"end_hour":18,"timezone":"Asia/Shanghai"}`, 1},
		{"closed window", "2026-10-09T10:00:00Z", "2026-10-09T09:00:00Z", "", `,"time_window":{"start_hour":9,"end_hour":18,"timezone":"Asia/Shanghai"}`, 0},
		{"legacy half hour waits", "2026-10-09T01:30:00Z", "2026-10-09T01:30:00Z", "", "", 0},
		{"legacy runs hourly", "2026-10-09T02:00:00Z", "2026-10-09T01:30:00Z", "2026-10-09T03:00:00Z", "", 1},
		{"concurrent edit or claim", "2026-10-09T02:00:00Z", "2026-10-09T02:00:00Z", "2026-10-09T03:00:00Z", "", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, db)))
			t.Cleanup(func() { _ = client.Close() })
			repo := &accountRepository{sql: db, client: client}
			now, err := time.Parse(time.RFC3339, tc.now)
			require.NoError(t, err)
			due, err := time.Parse(time.RFC3339, tc.due)
			require.NoError(t, err)
			raw := `{"model_id":"model","protocol":"chat"` + tc.window + `}`
			mock.ExpectQuery(`SELECT user_id, enabled, options, next_run_at FROM model_fingerprint_schedules`).WithArgs("account").
				WillReturnRows(sqlmock.NewRows([]string{"user_id", "enabled", "options", "next_run_at"}).AddRow("admin", true, []byte(raw), due))
			if tc.next != "" {
				next, err := time.Parse(time.RFC3339, tc.next)
				require.NoError(t, err)
				mock.ExpectExec(`(?s)WITH claimed AS.*s.options = \$4::jsonb.*FOR UPDATE OF s.*SET next_run_at = \$5::timestamptz.*INSERT INTO model_fingerprint_runs`).
					WithArgs(sqlmock.AnyArg(), "account", now, raw, next).WillReturnResult(sqlmock.NewResult(0, tc.rows))
			}
			claimed, err := repo.ClaimModelFingerprint(context.Background(), "account", &service.ModelFingerprintSnapshot{
				ID: "job", UserID: "admin", Model: "model", Protocol: "chat", Source: "scheduled", StartedAt: now,
			})
			require.NoError(t, err)
			require.Equal(t, tc.rows > 0, claimed)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
