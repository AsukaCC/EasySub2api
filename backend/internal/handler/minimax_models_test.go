package handler

import (
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/stretchr/testify/require"
)

// MiniMax 走 Codex 默认模型清单：M 系列用于 /v1/models 兜底列表。
func TestDefaultCodexModelIDsForPlatform_MiniMax(t *testing.T) {
	ids := defaultCodexModelIDsForPlatform(service.PlatformMiniMax)
	require.Equal(t, []string{"MiniMax-M3", "MiniMax-M2.7", "MiniMax-M2.5"}, ids)
}
