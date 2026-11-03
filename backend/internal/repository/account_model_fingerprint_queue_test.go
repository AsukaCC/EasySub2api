package repository

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/AsukaCC/EasySub2api/ent"
	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestModelFingerprintQueuePersistenceAndTransitions(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	t.Cleanup(func() { _ = client.Close() })
	repo := &accountRepository{sql: db, client: client}
	snapshot := &service.ModelFingerprintSnapshot{ID: "job", Status: "queued", StartedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour)}
	body, err := json.Marshal(snapshot)
	require.NoError(t, err)
	mock.ExpectQuery(`(?s)SELECT id, extra.*status' = 'queued'.*ORDER BY.*started_at.*id`).WillReturnRows(sqlmock.NewRows([]string{"id", "snapshot"}).AddRow("account", body))
	jobs, err := repo.ListQueuedModelFingerprints(context.Background())
	require.NoError(t, err)
	require.Len(t, jobs, 1)
	require.Equal(t, "job", jobs[0].Snapshot.ID)
	for _, rows := range []int64{1, 0} {
		mock.ExpectExec(`(?s)WITH activated AS.*extra->'model_fingerprint'->>'id' = \$3.*status' = 'queued'.*INSERT INTO model_fingerprint_runs.*ON CONFLICT`).WithArgs(sqlmock.AnyArg(), "account", "job").WillReturnResult(sqlmock.NewResult(0, rows))
		claimed, err := repo.ActivateModelFingerprint(context.Background(), "account", snapshot)
		require.NoError(t, err)
		require.Equal(t, rows == 1, claimed)
		mock.ExpectExec(`(?s)WITH cancelled AS.*extra->'model_fingerprint'->>'id' = \$2.*IN \('queued', 'running', 'waiting'\).*INSERT INTO model_fingerprint_runs`).WithArgs("account", "job", sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, rows))
		cancelled, err := repo.CancelModelFingerprint(context.Background(), "account", "job")
		require.NoError(t, err)
		require.Equal(t, rows == 1, cancelled)
	}
	// A cancelled/completed row cannot be changed by a late worker save.
	mock.ExpectExec(`(?s)WITH saved AS.*id' = \$3.*status' IN \('queued', 'running', 'waiting'\)`).WithArgs(sqlmock.AnyArg(), "account", "job").WillReturnResult(sqlmock.NewResult(0, 0))
	require.Error(t, repo.SaveModelFingerprint(context.Background(), "account", snapshot))
	require.NoError(t, mock.ExpectationsWereMet())
}
