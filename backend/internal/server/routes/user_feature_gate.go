package routes

import (
	"context"

	"github.com/AsukaCC/EasySub2api/internal/pkg/response"
	"github.com/AsukaCC/EasySub2api/internal/server/middleware"
	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/gin-gonic/gin"
)

func userFeatureGate(check func(context.Context) bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if role, ok := middleware.GetUserRoleFromContext(c); ok && role == service.RoleAdmin {
			c.Next()
			return
		}
		if check == nil || !check(c.Request.Context()) {
			response.NotFound(c, "Feature is disabled")
			c.Abort()
			return
		}
		c.Next()
	}
}
