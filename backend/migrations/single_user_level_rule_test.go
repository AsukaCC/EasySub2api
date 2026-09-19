package migrations

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSingleUserLevelMigrationPreservesExistingPrices(t *testing.T) {
	data, err := FS.ReadFile("276_single_user_level_rule.sql")
	require.NoError(t, err)
	sql := string(data)
	require.NotContains(t, sql, "UPDATE groups")
	require.NotContains(t, sql, "UPDATE user_level_rule_tiers")
	require.Contains(t, sql, "HAVING COUNT(*) > 1")
	require.Contains(t, sql, "UNIQUE (user_id)")
	require.Contains(t, sql, "DEFERRABLE INITIALLY DEFERRED")
	require.Contains(t, sql, "USER_LEVEL_RULE_HAS_MEMBERS")
	require.Contains(t, sql, "USER_LEVEL_DEFAULT_PROTECTED")
	require.Contains(t, sql, "required_default_user_level_rule()")
	require.NotContains(t, sql, "NEW.role <>")
}
