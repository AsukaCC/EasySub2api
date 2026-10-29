//go:build unit

package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/pkg/adminaccess"
	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func scopedReleaseContext(policy adminaccess.Policy) *gin.Context {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request = c.Request.WithContext(adminaccess.WithPrincipal(c.Request.Context(), &adminaccess.Principal{KeyID: "key-one", Policy: policy}))
	return c
}

func TestRelease0219ScopedAdminProjections(t *testing.T) {
	const visibleGroup = "11111111-1111-4111-8111-111111111111"
	const hiddenGroup = "22222222-2222-4222-8222-222222222222"
	policy := adminaccess.Policy{Rules: []adminaccess.Rule{{Permissions: []string{"accounts.read"}, GroupIDs: []string{visibleGroup}, AccountTypes: []string{"apikey"}, Platforms: []string{"openai"}}}, GlobalPermissions: []string{"users.read"}}
	c := scopedReleaseContext(policy)
	account := &service.Account{ID: "account", Platform: "openai", Type: "apikey", GroupIDs: []string{visibleGroup, hiddenGroup}, Credentials: map[string]any{"api_key": "secret-token", "base_url": "secret-endpoint", "model_mapping": map[string]any{"public": "upstream"}}, Extra: map[string]any{"private": "secret-extra"}}
	projected := scopedAdminAccountDTO(account, policy)
	require.Equal(t, []string{visibleGroup}, projected["group_ids"])
	encoded, err := json.Marshal(projected)
	require.NoError(t, err)
	require.Contains(t, string(encoded), "upstream")
	for _, hidden := range []string{hiddenGroup, "secret-token", "secret-endpoint", "secret-extra"} {
		require.NotContains(t, string(encoded), hidden)
	}
	user := &service.User{ID: "user", Email: "visible@example.com", Balance: 42, AllowedGroups: []string{hiddenGroup}}
	profile := adminUserResponse(c, user).(gin.H)
	require.Equal(t, user.Email, profile["email"])
	require.NotContains(t, profile, "balance")
	require.NotContains(t, profile, "allowed_groups")
	usage := scopedAdminUsageLogResponse(&service.UsageLog{ID: "usage", UserID: user.ID, User: user, APIKey: &service.APIKey{Key: "secret-key"}, Account: account, InputTokens: 123})
	require.Equal(t, 123, usage["input_tokens"])
	encoded, err = json.Marshal(usage)
	require.NoError(t, err)
	for _, hidden := range []string{"secret-key", "secret-token", "visible@example.com", hiddenGroup} {
		require.NotContains(t, string(encoded), hidden)
	}
	grant := adminSubscriptionGrantResponse(c, &service.SubscriptionGrantResult{ActivationStatus: "active", Subscription: &service.UserSubscription{ID: "subscription", User: user, Group: &service.Group{ID: hiddenGroup}}, PendingSubscription: &service.PendingSubscription{ID: "pending", LastError: "private-error", Group: &service.Group{ID: hiddenGroup}}})
	encoded, err = json.Marshal(grant)
	require.NoError(t, err)
	require.Contains(t, string(encoded), "subscription")
	for _, hidden := range []string{"private-error", "visible@example.com", hiddenGroup, "daily_usage_usd"} {
		require.NotContains(t, string(encoded), hidden)
	}
}

func TestRelease0219ScopedAdminIdempotencyIsolationAndReplayProjection(t *testing.T) {
	service.SetDefaultIdempotencyCoordinator(service.NewIdempotencyCoordinator(newMemoryIdempotencyRepoStub(), service.DefaultIdempotencyConfig()))
	t.Cleanup(func() { service.SetDefaultIdempotencyCoordinator(nil) })
	policy := adminaccess.Policy{GlobalPermissions: []string{"subscriptions.extend"}}
	keyID := "key-one"
	executions := 0
	router := gin.New()
	router.POST("/extend", func(c *gin.Context) {
		c.Request = c.Request.WithContext(adminaccess.WithPrincipal(c.Request.Context(), &adminaccess.Principal{KeyID: keyID, Policy: policy}))
		executeAdminIdempotentJSONWithResponse(c, "release0219.subscription", gin.H{"days": 1}, time.Minute, func(context.Context) (any, error) {
			executions++
			return gin.H{"id": "subscription", "user_id": "user", "daily_usage_usd": 12, "user": gin.H{"email": "private@example.com"}, "group": gin.H{"name": "private-group"}}, nil
		}, func(data any) any { return adminSubscriptionResponse(c, data) })
	})
	call := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/extend", nil)
		req.Header.Set("Idempotency-Key", "same-request")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		require.NotContains(t, rec.Body.String(), "private")
		return rec
	}
	first := call()
	require.NotContains(t, first.Body.String(), "daily_usage_usd")
	second := call()
	require.Equal(t, "true", second.Header().Get("X-Idempotency-Replayed"))
	require.JSONEq(t, first.Body.String(), second.Body.String())
	require.Equal(t, 1, executions)
	keyID = "key-two"
	require.Empty(t, call().Header().Get("X-Idempotency-Replayed"))
	require.Equal(t, 2, executions)
	policy.GlobalPermissions = append(policy.GlobalPermissions, "subscriptions.read")
	third := call()
	require.Empty(t, third.Header().Get("X-Idempotency-Replayed"))
	require.Contains(t, third.Body.String(), "daily_usage_usd")
	require.Equal(t, 3, executions)
	require.Equal(t, "true", call().Header().Get("X-Idempotency-Replayed"))
	require.Equal(t, 3, executions)
}
