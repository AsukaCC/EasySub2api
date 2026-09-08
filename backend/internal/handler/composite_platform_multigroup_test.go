package handler

import (
	"net/http/httptest"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestEnsureCompositeTargetPlatformResolvesMultiGroupGrokModel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/v1/responses", nil)
	apiKey := &service.APIKey{
		GroupIDs: []string{"g-openai", "g-grok"},
		Group:    &service.Group{ID: "g-openai", Platform: service.PlatformOpenAI},
	}

	ensureCompositeTargetPlatform(c, apiKey, "grok-4.5")
	platform, ok := service.ResolvedTargetPlatformFromContext(c.Request.Context())
	require.True(t, ok)
	require.Equal(t, service.PlatformGrok, platform)
}

func TestEnsureCompositeTargetPlatformSkipsSingleGroupKeys(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/v1/responses", nil)
	apiKey := &service.APIKey{
		GroupIDs: []string{"g-openai"},
		Group:    &service.Group{ID: "g-openai", Platform: service.PlatformOpenAI},
	}

	ensureCompositeTargetPlatform(c, apiKey, "grok-4.5")
	_, ok := service.ResolvedTargetPlatformFromContext(c.Request.Context())
	require.False(t, ok)
}
