//go:build unit

package repository

import (
	"context"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"errors"
	dbent "github.com/AsukaCC/EasySub2api/ent"
	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
)

// The transaction lock must be taken before counting, and failures must roll back.
func TestAPIKeyCreateLimitTransaction(t *testing.T) {
	for _, tc := range []struct {
		name        string
		count       int64
		lockError   bool
		insertError bool
	}{
		{"at_limit", 200, false, false}, {"lock_failure", 0, true, false}, {"insert_failure", 199, false, true}, {"commit", 199, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()
			client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, db)))
			repo := &apiKeyRepository{client: client, sql: db}
			mock.ExpectBegin()
			lock := mock.ExpectExec("SELECT pg_advisory_xact_lock").WithArgs("apikey-create:user-1")
			if tc.lockError {
				lock.WillReturnError(errors.New("lock unavailable"))
			} else {
				lock.WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectQuery(`SELECT COUNT.*FROM "api_keys".*"deleted_at" IS NULL.*"user_id" = \$1`).WithArgs("user-1").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(tc.count))
				if tc.count < 200 {
					insert := mock.ExpectExec(`INSERT INTO "api_keys"`)
					if tc.insertError {
						insert.WillReturnError(errors.New("insert failed"))
					} else {
						insert.WillReturnResult(sqlmock.NewResult(0, 1))
						mock.ExpectExec("UPDATE api_keys SET group_ids").WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 1))
					}
				}
			}
			if tc.lockError || tc.insertError || tc.count >= 200 {
				mock.ExpectRollback()
			} else {
				mock.ExpectCommit()
			}
			err = repo.CreateWithLimit(context.Background(), &service.APIKey{UserID: "user-1", Name: "test", Key: "synthetic-key", Status: service.StatusActive}, 200)
			if tc.count >= 200 {
				require.ErrorIs(t, err, service.ErrAPIKeyCountExceeded)
			} else if tc.lockError || tc.insertError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
func TestAPIKeyCreateCounterAtomicFixedWindow(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	cache := &apiKeyCache{rdb: rdb}
	ctx := context.Background()
	var wg sync.WaitGroup
	for i := 0; i < 60; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := cache.IncrementCreateCount(ctx, "user", time.Hour)
			if err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	value, err := mr.Get(apiKeyCreateCountKey("user"))
	require.NoError(t, err)
	require.Equal(t, "60", value)
	mr.FastForward(30 * time.Minute)
	n, err := cache.IncrementCreateCount(ctx, "user", time.Hour)
	require.NoError(t, err)
	require.Equal(t, int64(61), n)
	require.Equal(t, 30*time.Minute, mr.TTL(apiKeyCreateCountKey("user")))
	other, err := cache.IncrementCreateCount(ctx, "other", time.Hour)
	require.NoError(t, err)
	require.Equal(t, int64(1), other)
	mr.FastForward(31 * time.Minute)
	n, err = cache.IncrementCreateCount(ctx, "user", time.Hour)
	require.NoError(t, err)
	require.Equal(t, int64(1), n)
}
