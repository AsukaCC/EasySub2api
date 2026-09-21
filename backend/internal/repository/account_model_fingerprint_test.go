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
		mock.ExpectQuery(query).WithArgs(now.Add(-2 * time.Hour)).WillReturnRows(rows).RowsWillBeClosed()
		removed, err := repo.DeleteExpiredModelFingerprints(context.Background(), now)
		require.NoError(t, err)
		require.Equal(t, int64(len(ids)), removed)
	}
	mock.ExpectQuery(query).WithArgs(now.Add(-2 * time.Hour)).WillReturnError(errors.New("database unavailable"))
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
			WithArgs(sqlmock.AnyArg(), "account", snapshot.StartedAt).WillReturnResult(sqlmock.NewResult(0, rows))
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
