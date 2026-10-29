//go:build unit

package adminaccess

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScopedAdminPolicyDefaultDeny(t *testing.T) {
	selectedGroupsOnly := Policy{Rules: []Rule{{GroupIDs: []string{"00000000-0000-4000-8000-000000000001"}, AccountTypes: []string{"oauth"}}}}
	for _, policy := range []Policy{{}, selectedGroupsOnly} {
		for _, route := range []struct{ method, path string }{
			{"GET", "/api/v1/admin/accounts"}, {"GET", "/api/v1/admin/accounts/:id"},
			{"PUT", "/api/v1/admin/settings"}, {"GET", "/api/v1/admin/users"},
			{"POST", "/api/v1/admin/payment/orders/:id/refund"},
			{"POST", "/api/v1/admin/settings/admin-api-keys"},
		} {
			require.False(t, RouteAllowed(policy, route.method, route.path), route.path)
		}
	}
}

func TestScopedAdminModelPermissionsAreIndependent(t *testing.T) {
	write := Policy{Rules: []Rule{{Permissions: []string{"accounts.models.write"}, GroupIDs: []string{"00000000-0000-4000-8000-000000000001"}, AccountTypes: []string{"oauth"}}}}
	read := Policy{GlobalPermissions: []string{"models.defaults.read"}}
	require.NoError(t, ValidatePolicy(write))
	require.NoError(t, ValidatePolicy(read))
	require.True(t, RouteAllowed(write, "PUT", "/api/v1/admin/accounts/:id"))
	for _, path := range []string{"/api/v1/admin/accounts/antigravity/default-model-mapping"} {
		require.True(t, RouteAllowed(read, "GET", path))
		require.False(t, RouteAllowed(write, "GET", path))
		require.False(t, RouteAllowed(Policy{}, "GET", path))
	}
	for _, policy := range []Policy{write, read} {
		require.False(t, RouteAllowed(policy, "PUT", "/api/v1/admin/model-catalog/:platform"))
		require.False(t, RouteAllowed(policy, "DELETE", "/api/v1/admin/model-catalog/:platform"))
		require.False(t, RouteAllowed(policy, "GET", "/api/v1/admin/accounts/:id/models"))
		require.False(t, RouteAllowed(policy, "POST", "/api/v1/admin/accounts/:id/models/sync-upstream"))
	}
}

func TestScopedAdminPolicyAccountRulesNeverCrossProduct(t *testing.T) {
	p := Policy{Rules: []Rule{
		{Permissions: []string{"accounts.read", "accounts.refresh"}, GroupIDs: []string{"00000000-0000-4000-8000-000000000001"}, AccountTypes: []string{"oauth"}, Platforms: []string{"openai"}},
		{Permissions: []string{"accounts.read"}, GroupIDs: []string{"00000000-0000-4000-8000-000000000002"}, AccountTypes: []string{"apikey"}},
	}}
	for _, tc := range []struct {
		name, permission string
		account          Account
		allowed          bool
	}{
		{"组内OAuth", "accounts.refresh", Account{GroupIDs: []string{"00000000-0000-4000-8000-000000000001"}, Type: "oauth", Platform: "openai"}, true},
		{"共享账户任一组匹配", "accounts.refresh", Account{GroupIDs: []string{"00000000-0000-4000-8000-000000000001", "00000000-0000-4000-8000-000000000009"}, Type: "oauth", Platform: "openai"}, true},
		{"其他类型拒绝", "accounts.refresh", Account{GroupIDs: []string{"00000000-0000-4000-8000-000000000001"}, Type: "apikey", Platform: "openai"}, false},
		{"其他组拒绝", "accounts.refresh", Account{GroupIDs: []string{"00000000-0000-4000-8000-000000000009"}, Type: "oauth", Platform: "openai"}, false},
		{"其他平台拒绝", "accounts.refresh", Account{GroupIDs: []string{"00000000-0000-4000-8000-000000000001"}, Type: "oauth", Platform: "anthropic"}, false},
		{"不能借另一规则提权", "accounts.refresh", Account{GroupIDs: []string{"00000000-0000-4000-8000-000000000002"}, Type: "apikey", Platform: "openai"}, false},
		{"setup-token不能冒充OAuth", "accounts.read", Account{GroupIDs: []string{"00000000-0000-4000-8000-000000000001"}, Type: "setup-token", Platform: "openai"}, false},
		{"未分组默认拒绝", "accounts.read", Account{Type: "oauth", Platform: "openai"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) { require.Equal(t, tc.allowed, p.AllowsAccount(tc.permission, tc.account)) })
	}
	require.True(t, RouteAllowed(p, "GET", "/api/v1/admin/accounts"))
	require.False(t, RouteAllowed(p, "DELETE", "/api/v1/admin/accounts/:id"))
	require.False(t, RouteAllowed(p, "GET", "/api/v1/admin/groups"))
	require.False(t, RouteAllowed(p, "POST", "/api/v1/admin/openai/refresh-token"))
}

func TestScopedAdminPolicyGroupChanges(t *testing.T) {
	a := Account{GroupIDs: []string{"00000000-0000-4000-8000-000000000001", "00000000-0000-4000-8000-000000000002"}, Type: "oauth", Platform: "openai"}
	p := Policy{Rules: []Rule{{Permissions: []string{"accounts.groups"}, GroupIDs: []string{"00000000-0000-4000-8000-000000000001", "00000000-0000-4000-8000-000000000003"}, AccountTypes: []string{"oauth"}}}}
	require.True(t, p.AllowsGroup("accounts.groups", a, "00000000-0000-4000-8000-000000000001"))
	require.True(t, p.AllowsGroup("accounts.groups", a, "00000000-0000-4000-8000-000000000003"))
	require.False(t, p.AllowsGroup("accounts.groups", a, "00000000-0000-4000-8000-000000000002"))
	require.False(t, p.AllowsGroup("accounts.groups", Account{GroupIDs: []string{"00000000-0000-4000-8000-000000000002"}, Type: "oauth"}, "00000000-0000-4000-8000-000000000001"))
	require.False(t, p.AllowsGroup("accounts.groups", a, "00000000-0000-4000-8000-000000000000"))
}

func TestScopedAdminPolicyExplicitAllAndUngrouped(t *testing.T) {
	p := Policy{Rules: []Rule{{Permissions: []string{"accounts.read"}, AllGroups: true, AccountTypes: []string{"oauth"}}}}
	require.True(t, p.AllowsAccount("accounts.read", Account{GroupIDs: []string{"00000000-0000-4000-8000-000000000999"}, Type: "oauth"}))
	require.False(t, p.AllowsAccount("accounts.read", Account{Type: "oauth"}))
	p.Rules[0].IncludeUngrouped = true
	require.True(t, p.AllowsAccount("accounts.read", Account{Type: "oauth"}))
	p.Rules[0].AccountTypes = nil
	require.False(t, p.AllowsAccount("accounts.read", Account{Type: "oauth"}))
}

func TestScopedAdminPolicyValidationAndGlobalIsolation(t *testing.T) {
	require.NoError(t, ValidatePolicy(Policy{}))
	for _, p := range []Policy{
		{GlobalPermissions: []string{"*"}}, {GlobalPermissions: []string{"accounts.read"}},
		{Rules: []Rule{{Permissions: []string{"users.read"}}}},
		{Rules: []Rule{{Permissions: []string{"accounts.read"}, GroupIDs: []string{"invalid"}}}},
		{Rules: []Rule{{AccountTypes: []string{"future_type"}}}},
	} {
		require.Error(t, ValidatePolicy(p))
	}
	p := Policy{GlobalPermissions: []string{"users.read"}}
	require.NoError(t, ValidatePolicy(p))
	require.True(t, RouteAllowed(p, "GET", "/api/v1/admin/users"))
	require.False(t, RouteAllowed(p, "PUT", "/api/v1/admin/users/:id"))
	require.False(t, RouteAllowed(p, "POST", "/api/v1/admin/users/:id/balance"))
	require.False(t, RouteAllowed(p, "GET", "/api/v1/admin/accounts/:id"))
	all := Policy{}
	for _, perm := range PermissionCatalog() {
		if perm.Scope == "global" {
			all.GlobalPermissions = append(all.GlobalPermissions, perm.ID)
		} else {
			all.Rules = append(all.Rules, Rule{Permissions: []string{perm.ID}, AllGroups: true, IncludeUngrouped: true, AccountTypes: []string{"oauth", "apikey"}})
		}
	}
	for _, path := range []string{"/api/v1/admin/future-endpoint", "/api/v1/admin/settings/admin-api-key/regenerate", "/api/v1/admin/settings/admin-api-keys", "/api/v1/admin/system/restart", "/api/v1/admin/accounts/data", "/api/v1/admin/backups/:id/download-url"} {
		for _, method := range []string{"GET", "POST", "PUT", "DELETE"} {
			require.False(t, RouteAllowed(all, method, path), method+" "+path)
		}
	}
}

func TestScopedAdminPrincipalContext(t *testing.T) {
	require.Nil(t, FromContext(context.Background()))
	p := &Principal{KeyID: "key-1", KeyName: "只读"}
	require.Same(t, p, FromContext(WithPrincipal(context.Background(), p)))
}

func TestScopedAdminGroupModelsAreExplicitAndScoped(t *testing.T) {
	p := Policy{Rules: []Rule{
		{Permissions: []string{"groups.models.read"}, GroupIDs: []string{"00000000-0000-4000-8000-000000000002"}, Platforms: []string{"openai"}},
		{Permissions: []string{"groups.models.write"}, GroupIDs: []string{"00000000-0000-4000-8000-000000000003"}, Platforms: []string{"anthropic"}},
	}}
	require.NoError(t, ValidatePolicy(p))
	require.True(t, p.AllowsGroupConfig("groups.models.read", "00000000-0000-4000-8000-000000000002", "openai"))
	require.False(t, p.AllowsGroupConfig("groups.models.write", "00000000-0000-4000-8000-000000000002", "openai"))
	require.False(t, p.AllowsGroupConfig("groups.models.read", "00000000-0000-4000-8000-000000000003", "openai"))
	require.False(t, p.AllowsGroupConfig("groups.models.read", "00000000-0000-4000-8000-000000000002", "anthropic"))
	require.False(t, p.AllowsGroupConfig("groups.models.read", "00000000-0000-4000-8000-000000000000", "openai"))
	require.False(t, p.AllowsAccount("accounts.read", Account{GroupIDs: []string{"00000000-0000-4000-8000-000000000002"}, Type: "oauth", Platform: "openai"}))
	require.True(t, RouteAllowed(p, "GET", "/api/v1/admin/groups"))
	require.False(t, RouteAllowed(p, "DELETE", "/api/v1/admin/groups/:id"))
	require.Error(t, ValidatePolicy(Policy{GlobalPermissions: []string{"groups.models.write"}}))
}
