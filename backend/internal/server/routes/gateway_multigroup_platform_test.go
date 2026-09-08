package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	servermiddleware "github.com/AsukaCC/EasySub2api/internal/server/middleware"
	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestMultiGroupTargetPlatformMiddlewareResolvesGrokModel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	openaiGroupID := "g-openai"
	router.Use(gin.HandlerFunc(servermiddleware.APIKeyAuthMiddleware(func(c *gin.Context) {
		c.Set(string(servermiddleware.ContextKeyAPIKey), &service.APIKey{
			GroupID:  &openaiGroupID,
			GroupIDs: []string{"g-openai", "g-grok"},
			Group:    &service.Group{ID: openaiGroupID, Platform: service.PlatformOpenAI},
		})
		c.Next()
	})))
	router.Use(compositeTargetPlatformMiddleware(nil))
	router.POST("/v1/responses", func(c *gin.Context) {
		require.Equal(t, service.PlatformGrok, getGroupPlatform(c))
		platform, ok := service.ResolvedTargetPlatformFromContext(c.Request.Context())
		require.True(t, ok)
		require.Equal(t, service.PlatformGrok, platform)
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader(`{"model":"grok-4.5"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
}
