package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestMiniMaxPlatformMigration 校验 272 号迁移把 minimax 加入全部平台白名单约束，
// 并把被动监控平台目录补充 minimax 条目。
func TestMiniMaxPlatformMigration(t *testing.T) {
	content, err := FS.ReadFile("272_add_minimax_platform.sql")
	require.NoError(t, err)

	sql := strings.Join(strings.Fields(string(content)), " ")
	require.Contains(t, sql,
		"CHECK (platform IN ('anthropic', 'openai', 'gemini', 'antigravity', 'grok', 'kimi', 'zhipu', 'deepseek', 'minimax'))")
	require.Contains(t, sql,
		"CHECK (target_platform IN ('anthropic', 'openai', 'gemini', 'antigravity', 'grok', 'kimi', 'zhipu', 'deepseek', 'minimax'))")
	require.Contains(t, sql,
		"CHECK (provider IN ('openai', 'anthropic', 'gemini', 'grok', 'antigravity', 'kimi', 'zhipu', 'deepseek', 'minimax'))")
	require.Contains(t, sql, "('minimax', 7)")
	require.Contains(t, sql, "UPDATE channel_monitor_v2_config")
}
