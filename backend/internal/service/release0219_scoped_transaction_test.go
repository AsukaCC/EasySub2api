//go:build unit

package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/AsukaCC/EasySub2api/ent"
	"github.com/AsukaCC/EasySub2api/internal/pkg/adminaccess"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestRelease0219ScopedAccountFieldPermissionsAndGroups(t *testing.T) {
	for _, fields := range []map[string]json.RawMessage{
		{"extra": json.RawMessage(`{}`)},
		{"rate_multiplier": json.RawMessage(`0`)},
		{"credentials": json.RawMessage(`{"unknown":"value"}`)},
		{"credentials": json.RawMessage(`{"model_mapping":{"alias":42}}`)},
	} {
		_, err := scopedAccountFieldPermissions(fields, false)
		require.Error(t, err)
	}
	permissions, err := scopedAccountFieldPermissions(map[string]json.RawMessage{"credentials": json.RawMessage(`{"model_mapping":{"alias":"upstream"}}`)}, false)
	require.NoError(t, err)
	require.Equal(t, []string{"accounts.models.write"}, permissions)
	policy := adminaccess.Policy{Rules: []adminaccess.Rule{{Permissions: []string{"accounts.groups"}, GroupIDs: []string{"allowed"}, AccountTypes: []string{"apikey"}, Platforms: []string{"openai"}}}}
	resource := adminaccess.Account{Type: "apikey", Platform: "openai", GroupIDs: []string{"allowed", "hidden"}}
	groups, err := scopedAdminMergeGroups(policy, resource, resource.GroupIDs, nil)
	require.NoError(t, err)
	require.Equal(t, []string{"hidden"}, groups)
	_, err = scopedAdminMergeGroups(policy, resource, resource.GroupIDs, []string{"new-unapproved"})
	require.Error(t, err)
	// A separate rule matching a different platform cannot grant group access.
	policy.Rules = append(policy.Rules, adminaccess.Rule{Permissions: []string{"accounts.groups"}, GroupIDs: []string{"new-unapproved"}, AccountTypes: []string{"apikey"}, Platforms: []string{"grok"}})
	_, err = scopedAdminMergeGroups(policy, resource, resource.GroupIDs, []string{"new-unapproved"})
	require.Error(t, err)
}

func TestRelease0219AccountSideEffectsWaitForCommit(t *testing.T) {
	for _, result := range []string{"commit", "rollback", "failed commit"} {
		t.Run(result, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			t.Cleanup(func() { _ = db.Close() })
			client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, db)))
			t.Cleanup(func() { _ = client.Close() })
			mock.ExpectBegin()
			tx, err := client.Tx(context.Background())
			require.NoError(t, err)
			ran := false
			AfterAccountCommit(dbent.NewTxContext(context.Background(), tx), func(ctx context.Context) {
				require.Nil(t, dbent.TxFromContext(ctx))
				ran = true
			})
			require.False(t, ran)
			switch result {
			case "commit":
				mock.ExpectCommit()
				require.NoError(t, tx.Commit())
				require.True(t, ran)
			case "rollback":
				mock.ExpectRollback()
				require.NoError(t, tx.Rollback())
				require.False(t, ran)
			case "failed commit":
				mock.ExpectCommit().WillReturnError(errors.New("commit failed"))
				require.Error(t, tx.Commit())
				require.False(t, ran)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
