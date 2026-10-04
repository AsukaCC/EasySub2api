//go:build unit

package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/pkg/pagination"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type fingerprintKeyRepo struct {
	APIKeyRepository
	keys []APIKey
}

func (r *fingerprintKeyRepo) GetByID(_ context.Context, id string) (*APIKey, error) {
	for _, key := range r.keys {
		if key.ID == id {
			return &key, nil
		}
	}
	return nil, ErrAPIKeyNotFound
}

func (r *fingerprintKeyRepo) ListByUserID(_ context.Context, _ string, _ pagination.PaginationParams, _ APIKeyListFilters) ([]APIKey, *pagination.PaginationResult, error) {
	return r.keys, &pagination.PaginationResult{Pages: 1}, nil
}

type fingerprintGroupRepo struct {
	GroupRepository
	group *Group
}

func (r *fingerprintGroupRepo) GetByID(context.Context, string) (*Group, error) {
	return r.group, nil
}

func TestModelFingerprintLocalAPIKeyValidation(t *testing.T) {
	for _, tc := range []struct {
		name string
		edit func(*APIKey, *Group)
	}{
		{"other owner", func(k *APIKey, _ *Group) { k.UserID = "other" }},
		{"disabled", func(k *APIKey, _ *Group) { k.Status = StatusAPIKeyDisabled }},
		{"expired", func(k *APIKey, _ *Group) { at := time.Now().Add(-time.Minute); k.ExpiresAt = &at }},
		{"exhausted", func(k *APIKey, _ *Group) { k.Quota, k.QuotaUsed = 10, 10 }},
		{"wrong group", func(k *APIKey, _ *Group) { k.GroupIDs = []string{"unrelated"} }},
		{"wrong platform", func(_ *APIKey, g *Group) { g.Platform = PlatformAnthropic }},
		{"disabled group", func(_ *APIKey, g *Group) { g.Status = StatusDisabled }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc, repo, _ := newFingerprintTestService()
			key := &svc.modelFingerprintKeys.(*fingerprintKeyRepo).keys[0]
			tc.edit(key, svc.modelFingerprintGroups.(*fingerprintGroupRepo).group)
			_, err := svc.GetFingerprintModelsForKey(context.Background(), "one", "admin", key.ID)
			require.Error(t, err)
			_, err = svc.StartModelFingerprint(context.Background(), "one", "gpt-6-astra", "admin", ModelFingerprintOptions{APIKeyID: key.ID, Protocol: "chat"})
			require.Error(t, err)
			require.Empty(t, repo.snapshots)
			keys, err := svc.ListFingerprintKeys(context.Background(), "one", "admin", "")
			require.NoError(t, err)
			require.Empty(t, keys)
		})
	}
}

func TestModelFingerprintKeySelectionDoesNotExposeCredentials(t *testing.T) {
	svc, _, _ := newFingerprintTestService()
	key := &svc.modelFingerprintKeys.(*fingerprintKeyRepo).keys[0]
	key.GroupIDs = []string{"another", "fingerprint-group"}
	keys, err := svc.ListFingerprintKeys(context.Background(), "one", "admin", "")
	require.NoError(t, err)
	require.Equal(t, []ModelFingerprintKey{{ID: "local-key", Name: "My key"}}, keys)
	body, err := json.Marshal(keys)
	require.NoError(t, err)
	require.NotContains(t, string(body), key.Key)
	_, err = svc.GetFingerprintModelsForKey(context.Background(), "one", "admin", "")
	require.Error(t, err)
}

func TestModelFingerprintRevokedKeyStopsNextSample(t *testing.T) {
	svc, _, _ := newFingerprintTestService()
	svc.modelFingerprintKeys.(*fingerprintKeyRepo).keys = nil
	probe := &modelFingerprintProbe{apiKeyID: "local-key", userID: "admin", protocol: "chat"}
	ctx := context.WithValue(context.Background(), modelFingerprintContextKey{}, probe)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/test", nil).WithContext(ctx)
	err := svc.TestAccountConnection(c, "one", "gpt-6-astra", "test", AccountTestModeDefault)
	require.ErrorIs(t, err, ErrAPIKeyNotFound)
	require.True(t, probe.fatal)
}

func TestModelFingerprintOldScheduleRequiresKeySelection(t *testing.T) {
	svc, repo, _ := newFingerprintTestService()
	result, err := svc.StartModelFingerprint(context.Background(), "one", "gpt-6-astra", "admin", ModelFingerprintOptions{Protocol: "chat", Source: "scheduled"})
	require.NoError(t, err)
	require.Equal(t, "preparation_failed", result.Error)
	require.Equal(t, 0, result.Completed)
	require.Equal(t, result.ID, waitFingerprint(t, repo.done).ID)
}
