//go:build integration

package repository

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"

	dbent "github.com/AsukaCC/EasySub2api/ent"
	"github.com/AsukaCC/EasySub2api/internal/pkg/adminaccess"
	"github.com/AsukaCC/EasySub2api/internal/pkg/pagination"
	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestRelease0219AdminKeyLifecycleWithUUIDCreator(t *testing.T) {
	ctx := context.Background()
	creator := mustCreateUser(t, integrationEntClient, &service.User{Role: service.RoleAdmin})
	t.Cleanup(func() { _, _ = integrationDB.ExecContext(ctx, "DELETE FROM users WHERE id=$1", creator.ID) })
	repo := NewScopedAdminKeyRepository(integrationDB)
	settings := &service.SettingService{}
	settings.SetScopedAdminKeyRepository(repo)
	expires := time.Now().Add(time.Hour)
	input := service.ScopedAdminKeyInput{
		Name: "release0219 integration", Enabled: true, ExpiresAt: &expires,
		AllowedIPs: []string{"192.0.2.0/24"},
		Policy:     adminaccess.Policy{GlobalPermissions: []string{"ops.read"}},
	}
	secret, created, err := settings.CreateScopedAdminKey(ctx, input, creator.ID)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = integrationDB.ExecContext(ctx, "DELETE FROM scoped_admin_api_keys WHERE id=$1", created.ID)
	})
	stored, err := repo.GetScopedAdminKey(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, creator.ID, stored.CreatedBy)
	require.Len(t, stored.Hash, 64)
	require.NotEqual(t, secret, stored.Hash)
	require.Equal(t, input.Policy, stored.Policy)
	require.Equal(t, input.AllowedIPs, stored.AllowedIPs)
	_, err = settings.AuthenticateScopedAdminKey(ctx, secret, "198.51.100.1")
	require.ErrorIs(t, err, service.ErrScopedAdminKeyInvalid)
	_, err = settings.AuthenticateScopedAdminKey(ctx, secret, "192.0.2.1")
	require.NoError(t, err)
	used, err := repo.GetScopedAdminKey(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, used.LastUsedAt)
	require.Equal(t, "192.0.2.1", used.LastUsedIP)
	require.ErrorIs(t, repo.DeleteScopedAdminKeyRecord(ctx, created.ID, time.Now()), service.ErrScopedAdminKeyNotFound)

	// A policy change invalidates an authentication already read by another request.
	input.Policy.GlobalPermissions = []string{"system.read"}
	_, err = settings.UpdateScopedAdminKey(ctx, created.ID, input)
	require.NoError(t, err)
	require.ErrorIs(t, repo.TouchScopedAdminKey(ctx, used.ID, used.Hash, used.UpdatedAt, "192.0.2.2", time.Now()), service.ErrScopedAdminKeyNotFound)
	rotatedSecret, rotated, err := settings.RotateScopedAdminKey(ctx, created.ID)
	require.NoError(t, err)
	require.NotEqual(t, secret, rotatedSecret)
	_, err = settings.AuthenticateScopedAdminKey(ctx, secret, "192.0.2.1")
	require.ErrorIs(t, err, service.ErrScopedAdminKeyInvalid)
	_, err = settings.AuthenticateScopedAdminKey(ctx, rotatedSecret, "192.0.2.1")
	require.NoError(t, err)
	current, err := repo.GetScopedAdminKey(ctx, rotated.ID)
	require.NoError(t, err)
	require.ErrorIs(t, repo.TouchScopedAdminKey(ctx, current.ID, current.Hash, current.UpdatedAt, "192.0.2.1", expires.Add(time.Second)), service.ErrScopedAdminKeyNotFound)

	// Two rotations of the same observed version must have one winner.
	var wg sync.WaitGroup
	start := make(chan struct{})
	errors := make(chan error, 2)
	for _, hash := range []string{strings.Repeat("a", 64), strings.Repeat("b", 64)} {
		wg.Add(1)
		go func(hash string) {
			defer wg.Done()
			<-start
			errors <- repo.RotateScopedAdminKeyHash(ctx, current.ID, current.Hash, hash, time.Now())
		}(hash)
	}
	close(start)
	wg.Wait()
	close(errors)
	winners := 0
	for err := range errors {
		if err == nil {
			winners++
		} else {
			require.ErrorIs(t, err, service.ErrScopedAdminKeyConflict)
		}
	}
	require.Equal(t, 1, winners)
	require.NoError(t, settings.RevokeScopedAdminKey(ctx, current.ID))
	_, err = settings.AuthenticateScopedAdminKey(ctx, rotatedSecret, "192.0.2.1")
	require.ErrorIs(t, err, service.ErrScopedAdminKeyInvalid)
	require.NoError(t, settings.DeleteScopedAdminKeyRecord(ctx, current.ID))
	listed, err := repo.ListScopedAdminKeys(ctx)
	require.NoError(t, err)
	for _, key := range listed {
		require.NotEqual(t, current.ID, key.ID)
	}
	retained, err := repo.GetScopedAdminKey(ctx, current.ID)
	require.NoError(t, err)
	require.NotNil(t, retained.RevokedAt)
	require.NotNil(t, retained.LastUsedAt)
	require.Equal(t, creator.ID, retained.CreatedBy)
	var hidden bool
	require.NoError(t, integrationDB.QueryRowContext(ctx, "SELECT deleted_at IS NOT NULL FROM scoped_admin_api_keys WHERE id=$1", current.ID).Scan(&hidden))
	require.True(t, hidden)
}

func TestRelease0219ImageCountBillingMigrationAndRepository(t *testing.T) {
	tx := testEntTx(t)
	ctx := dbent.NewTxContext(context.Background(), tx)
	client := tx.Client()
	// Omit the new column so this assertion exercises the database default.
	inserted, err := client.QueryContext(ctx, "INSERT INTO groups (name,platform) VALUES ($1,$2) RETURNING id,image_count_billing", "release0219-image-default", service.PlatformOpenAI)
	require.NoError(t, err)
	var groupID string
	var defaultValue bool
	require.True(t, inserted.Next())
	require.NoError(t, inserted.Scan(&groupID, &defaultValue))
	require.NoError(t, inserted.Close())
	require.False(t, defaultValue)
	repo := newGroupRepositoryWithSQL(client, client)
	loaded, err := repo.GetByID(ctx, groupID)
	require.NoError(t, err)
	require.False(t, loaded.ImageCountBilling)
	loaded.AllowImageGeneration = true
	loaded.ImageCountBilling = true
	require.NoError(t, repo.Update(ctx, loaded))
	roundtrip, err := repo.GetByID(ctx, groupID)
	require.NoError(t, err)
	require.True(t, roundtrip.ImageCountBilling)
	var actual bool
	rows, err := client.QueryContext(ctx, "SELECT image_count_billing FROM groups WHERE id=$1", groupID)
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()
	require.True(t, rows.Next())
	require.NoError(t, rows.Scan(&actual))
	require.True(t, actual)
}

func TestRelease0219GrokCatalogConditionalWritesAndRollback(t *testing.T) {
	ctx := context.Background()
	account := mustCreateAccount(t, integrationEntClient, &service.Account{
		Name: "release0219-grok-cas", Platform: service.PlatformGrok,
		Extra: map[string]any{"unrelated": "preserved"},
	})
	t.Cleanup(func() { _, _ = integrationDB.ExecContext(ctx, "DELETE FROM accounts WHERE id=$1", account.ID) })
	repo := newAccountRepositoryWithSQL(integrationEntClient, integrationDB, nil)
	first := json.RawMessage(`{"models":["grok-4.6"],"success_started_at":"2026-10-29T01:00:00Z"}`)
	second := json.RawMessage(`{"models":["grok-4.7"],"success_started_at":"2026-10-29T02:00:00Z"}`)
	applied, err := repo.CompareAndSwapGrokObservedModels(ctx, account.ID, nil, first)
	require.NoError(t, err)
	require.True(t, applied)
	updates := map[string]any{
		service.UpstreamModelCatalogExtraKey:  map[string]any{"models": []string{"grok-4.6"}},
		service.UpstreamModelMetadataExtraKey: map[string]any{"source": "first"},
		"unrelated":                           "must-not-overwrite", "grok_observed_models": "must-not-overwrite",
	}
	applied, err = repo.UpdateGrokCatalogExtraIfObserved(ctx, account.ID, first, updates)
	require.NoError(t, err)
	require.True(t, applied)
	applied, err = repo.CompareAndSwapGrokObservedModels(ctx, account.ID, first, second)
	require.NoError(t, err)
	require.True(t, applied)
	updates[service.UpstreamModelCatalogExtraKey] = map[string]any{"models": []string{"grok-4.7"}}
	updates[service.UpstreamModelMetadataExtraKey] = map[string]any{"source": "second"}
	applied, err = repo.UpdateGrokCatalogExtraIfObserved(ctx, account.ID, second, updates)
	require.NoError(t, err)
	require.True(t, applied)
	applied, err = repo.UpdateGrokCatalogExtraIfObserved(ctx, account.ID, first, map[string]any{
		service.UpstreamModelCatalogExtraKey: map[string]any{"models": []string{"stale"}},
	})
	require.NoError(t, err)
	require.False(t, applied)
	loaded, err := repo.GetByID(ctx, account.ID)
	require.NoError(t, err)
	require.Equal(t, "preserved", loaded.Extra["unrelated"])
	require.Equal(t, "second", loaded.Extra[service.UpstreamModelMetadataExtraKey].(map[string]any)["source"])
	require.Equal(t, []any{"grok-4.7"}, loaded.Extra[service.UpstreamModelCatalogExtraKey].(map[string]any)["models"])

	// A request-scoped transaction must include both raw SQL and ORM operations.
	tx := testEntTx(t)
	txCtx := dbent.NewTxContext(ctx, tx)
	applied, err = repo.CompareAndSwapGrokObservedModels(txCtx, account.ID, second, first)
	require.NoError(t, err)
	require.True(t, applied)
	applied, err = repo.UpdateGrokCatalogExtraIfObserved(txCtx, account.ID, first, map[string]any{
		service.UpstreamModelCatalogExtraKey: map[string]any{"models": []string{"rolled-back"}},
	})
	require.NoError(t, err)
	require.True(t, applied)
	inTransaction, err := repo.GetByID(txCtx, account.ID)
	require.NoError(t, err)
	inTransaction.Name = "must-rollback"
	require.NoError(t, repo.Update(txCtx, inTransaction))
	require.NoError(t, tx.Rollback())
	loaded, err = repo.GetByID(ctx, account.ID)
	require.NoError(t, err)
	require.Equal(t, account.Name, loaded.Name)
	require.Equal(t, []any{"grok-4.7"}, loaded.Extra[service.UpstreamModelCatalogExtraKey].(map[string]any)["models"])
	require.Equal(t, []any{"grok-4.7"}, loaded.Extra["grok_observed_models"].(map[string]any)["models"])
}

func TestRelease0219ScopedAccountQueriesKeepRulesSeparate(t *testing.T) {
	tx := testEntTx(t)
	ctx := dbent.NewTxContext(context.Background(), tx)
	client := tx.Client()
	repo := newAccountRepositoryWithSQL(client, client, nil)
	first := mustCreateGroup(t, client, &service.Group{Name: "release0219-scope-first"})
	second := mustCreateGroup(t, client, &service.Group{Name: "release0219-scope-second"})
	create := func(name, platform, kind, groupID string) *service.Account {
		account := mustCreateAccount(t, client, &service.Account{
			Name: name, Platform: platform, Type: kind,
			Credentials: map[string]any{
				"api_key": "must-not-return", "access_token": "must-not-return",
				"model_mapping": map[string]any{"visible": "upstream"},
			},
		})
		mustBindAccountToGroup(t, client, account.ID, groupID, 19)
		return account
	}
	allowedFirst := create("release0219-allowed-first", service.PlatformOpenAI, service.AccountTypeAPIKey, first.ID)
	allowedSecond := create("release0219-allowed-second", service.PlatformGemini, service.AccountTypeOAuth, second.ID)
	create("release0219-denied-cross-rule", service.PlatformGemini, service.AccountTypeOAuth, first.ID)
	create("release0219-denied-cross-group", service.PlatformOpenAI, service.AccountTypeAPIKey, second.ID)
	principal := &adminaccess.Principal{Policy: adminaccess.Policy{Rules: []adminaccess.Rule{
		{Permissions: []string{"accounts.read"}, GroupIDs: []string{first.ID}, AccountTypes: []string{service.AccountTypeAPIKey}, Platforms: []string{service.PlatformOpenAI}},
		{Permissions: []string{"accounts.read"}, GroupIDs: []string{second.ID}, AccountTypes: []string{service.AccountTypeOAuth}, Platforms: []string{service.PlatformGemini}},
	}}}
	ctx = adminaccess.WithPrincipal(ctx, principal)
	accounts, total, err := repo.ScopedAdminList(ctx, pagination.PaginationParams{Page: 1, PageSize: 20}, "", "", "", "")
	require.NoError(t, err)
	require.EqualValues(t, 2, total)
	require.Len(t, accounts, 2)
	ids := []string{}
	for _, account := range accounts {
		ids = append(ids, account.ID)
		require.Equal(t, map[string]any{"model_mapping": map[string]string{"visible": "upstream"}}, account.Credentials)
		require.Nil(t, account.Extra)
	}
	require.ElementsMatch(t, []string{allowedFirst.ID, allowedSecond.ID}, ids)

	// The import candidate SQL uses the same UUID resource scope and trims identity values.
	candidate := mustCreateAccount(t, client, &service.Account{
		Name: "release0219-codex-candidate", Platform: service.PlatformOpenAI, Type: service.AccountTypeOAuth,
		Credentials: map[string]any{"chatgpt_account_id": " account ", "chatgpt_user_id": " user ", "refresh_token": "must-not-load"},
	})
	mustBindAccountToGroup(t, client, candidate.ID, first.ID, 7)
	principal.Policy.Rules = append(principal.Policy.Rules, adminaccess.Rule{
		Permissions: []string{"accounts.credentials"}, GroupIDs: []string{first.ID},
		AccountTypes: []string{service.AccountTypeOAuth}, Platforms: []string{service.PlatformOpenAI},
	})
	candidates, err := repo.ScopedAdminCodexCandidates(ctx, service.ScopedCodexIdentity{AccountID: "account", UserID: "user"})
	require.NoError(t, err)
	require.Len(t, candidates, 1)
	require.Equal(t, candidate.ID, candidates[0].ID)
	require.NotContains(t, candidates[0].Credentials, "refresh_token")
}
