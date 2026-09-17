package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/AsukaCC/EasySub2api/ent"
	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestProtectionLockedSavePreservesCurrentPolicy(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	a := &service.Account{ID: "a", Platform: service.PlatformOpenAI, Type: service.AccountTypeOAuth, Concurrency: 7, Extra: map[string]any{"anti_degradation": false}}
	current := *a
	service.PrepareNewAccountProtection(&current)
	raw, err := json.Marshal(current.Extra)
	require.NoError(t, err)
	mock.ExpectQuery("SELECT extra, updated_at FROM accounts").WithArgs("a").
		WillReturnRows(sqlmock.NewRows([]string{"extra", "updated_at"}).AddRow(raw, time.Now()))
	require.NoError(t, lockAccountProtection(context.Background(), client, a))
	require.True(t, a.AntiDegradationEnabled())
	require.Equal(t, "legacy", a.ProtectionMode())
	require.NoError(t, mock.ExpectationsWereMet())
}

type protectionConflictStore struct {
	client  *dbent.Client
	account *service.Account
}

func (s *protectionConflictStore) GetAccount(context.Context, string) (*service.Account, error) {
	return s.account, nil
}
func (s *protectionConflictStore) UpdateAccount(ctx context.Context, _ string, in *service.UpdateAccountInput) (*service.Account, error) {
	a := *s.account
	a.Extra = in.Extra
	return &a, lockAccountProtection(ctx, s.client, &a)
}
func TestProtectionTransitionRejectsConcurrentRevision(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	now := time.Now()
	a := &service.Account{ID: "a", Platform: service.PlatformOpenAI, Type: service.AccountTypeOAuth, Concurrency: 7, UpdatedAt: now}
	mock.ExpectQuery("SELECT extra, updated_at FROM accounts").WithArgs("a").
		WillReturnRows(sqlmock.NewRows([]string{"extra", "updated_at"}).AddRow([]byte("{}"), now.Add(time.Second)))
	svc := service.NewAntiDegradeService(&protectionConflictStore{client, a})
	_, err = svc.Apply(context.Background(), "a")
	require.ErrorIs(t, err, service.ErrProtectionConflict)
	require.NoError(t, mock.ExpectationsWereMet())
}

// Only the explicitly supplied, isolated test database is used.
func TestProtectionExtraSQLPostgres(t *testing.T) {
	dsn := os.Getenv("PROTECTION_TEST_DSN")
	if dsn == "" {
		t.Skip("isolated PostgreSQL DSN not supplied")
	}
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()
	tx, err := db.Begin()
	require.NoError(t, err)
	defer tx.Rollback()
	_, err = tx.Exec("CREATE TEMP TABLE protection_accounts (extra jsonb)")
	require.NoError(t, err)
	_, err = tx.Exec(`INSERT INTO protection_accounts VALUES ('{"anti_degradation":true,"anti_degrade":{"enabled":true,"mode":"legacy"},"enable_tls_fingerprint":true,"codex_fingerprint_mode":"session","codex_fingerprint_seed":"original","other":1}')`)
	require.NoError(t, err)
	expression := preserveProtectionExtraSQL("COALESCE(extra,'{}'::jsonb) || $1::jsonb")
	var raw []byte
	err = tx.QueryRow("UPDATE protection_accounts SET extra = "+expression+" RETURNING extra",
		`{"anti_degradation":false,"anti_degrade":{"enabled":false},"enable_tls_fingerprint":false,"codex_fingerprint_mode":"off","codex_fingerprint_seed":"forged","other":2,"request_integrity_mode":"enforce"}`).Scan(&raw)
	require.NoError(t, err)
	var result map[string]any
	require.NoError(t, json.Unmarshal(raw, &result))
	require.Equal(t, true, result["anti_degradation"])
	require.Equal(t, true, result["enable_tls_fingerprint"])
	require.Equal(t, "original", result["codex_fingerprint_seed"])
	require.Equal(t, "session", result["codex_fingerprint_mode"])
	require.Equal(t, float64(2), result["other"])
	require.Equal(t, "enforce", result["request_integrity_mode"])
	_, err = tx.Exec("UPDATE protection_accounts SET extra = '{}'::jsonb")
	require.NoError(t, err)
	err = tx.QueryRow("UPDATE protection_accounts SET extra = "+expression+" RETURNING extra", `{"anti_degradation":true,"anti_degrade":{"enabled":true},"codex_fingerprint_mode":"device"}`).Scan(&raw)
	require.NoError(t, err)
	result = nil
	require.NoError(t, json.Unmarshal(raw, &result))
	require.NotContains(t, result, "anti_degradation")
	require.NotContains(t, result, "anti_degrade")
	require.Equal(t, "device", result["codex_fingerprint_mode"])
}
