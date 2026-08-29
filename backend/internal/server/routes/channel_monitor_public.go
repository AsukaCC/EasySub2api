package routes

import (
	"github.com/AsukaCC/EasySub2api/internal/handler"
	"github.com/AsukaCC/EasySub2api/internal/server/middleware"
	"github.com/AsukaCC/EasySub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// RegisterChannelMonitorPublicRoutes exposes the compact group status view
// without requiring a user session. The endpoint is still protected by the
// user-visible feature switch and the passive monitor runtime switch.
func RegisterChannelMonitorPublicRoutes(
	v1 *gin.RouterGroup,
	h *handler.Handlers,
	settingService *service.SettingService,
	panelRateLimiter *middleware.PanelRateLimiter,
) {
	monitor := v1.Group("/channel-monitor-v2/public")
	monitor.Use(panelRateLimiter.PublicIP())
	if settingService != nil {
		monitor.Use(userFeatureGate(settingService.IsChannelMonitorUserAvailable))
	}
	monitor.Use(channelMonitorModeV2Guard(settingService))
	monitor.GET("/matrix", h.ChannelMonitorV2.Matrix)
}
