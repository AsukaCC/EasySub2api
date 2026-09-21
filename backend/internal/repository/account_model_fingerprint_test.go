package repository

import (
	"context"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/AsukaCC/EasySub2api/ent"
	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

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
