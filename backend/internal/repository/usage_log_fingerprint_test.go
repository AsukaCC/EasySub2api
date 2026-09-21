package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"strings"
	"testing"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestModelFingerprintKeylessUsagePersistence(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := &usageLogRepository{sql: db, db: db}
	log := &service.UsageLog{
		UserID: "admin", AccountID: "account", RequestID: "fingerprint:job:1",
		Model: "gpt-6-astra", RequestType: service.RequestTypeTest, Stream: true, CreatedAt: time.Now(),
	}
	prepared := prepareUsageLogInsert(log)
	require.Equal(t, sql.NullString{}, prepared.args[1])
	args := make([]driver.Value, len(prepared.args))
	for i := range args {
		args[i] = sqlmock.AnyArg()
	}
	args[1] = nil
	insert := `(?s)INSERT INTO usage_logs.*ON CONFLICT DO NOTHING.*RETURNING id, created_at`
	mock.ExpectQuery(insert).WithArgs(args...).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow("log", log.CreatedAt))
	inserted, err := repo.Create(context.Background(), log)
	require.NoError(t, err)
	require.True(t, inserted)
	mock.ExpectQuery(insert).WithArgs(args...).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}))
	mock.ExpectQuery(`SELECT id, created_at FROM usage_logs WHERE request_id = \$1 AND api_key_id IS NOT DISTINCT FROM \$2`).
		WithArgs(log.RequestID, nil).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow("log", log.CreatedAt))
	inserted, err = repo.Create(context.Background(), log)
	require.NoError(t, err)
	require.False(t, inserted)
	require.Empty(t, collectUsageLogIDs([]service.UsageLog{*log}).apiKeyIDs)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestModelFingerprintKeylessUsageScan(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	log := &service.UsageLog{UserID: "admin", AccountID: "account", Model: "gpt-6-astra", RequestType: service.RequestTypeTest, Stream: true}
	prepared := prepareUsageLogInsert(log)
	values := []driver.Value{"log"}
	for _, arg := range prepared.args {
		v, err := driver.DefaultParameterConverter.ConvertValue(arg)
		require.NoError(t, err)
		values = append(values, v)
	}
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(strings.Split(usageLogSelectColumns, ", ")).AddRow(values...))
	rows, err := db.Query("SELECT")
	require.NoError(t, err)
	defer rows.Close()
	require.True(t, rows.Next())
	got, err := scanUsageLog(rows)
	require.NoError(t, err)
	require.Empty(t, got.APIKeyID)
	require.Equal(t, service.RequestTypeTest, got.RequestType)
	require.True(t, got.Stream)
	require.NoError(t, mock.ExpectationsWereMet())
}
