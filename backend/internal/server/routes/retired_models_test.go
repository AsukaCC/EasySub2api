package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/server/middleware"
	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestRetiredModelsRejectedBeforeCompositeRewrite(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyAPIKey), &service.APIKey{Group: &service.Group{ID: "group", Platform: service.PlatformComposite}})
		c.Next()
	})
	called := false
	engine.POST("/v1/responses", compositeTargetPlatformMiddleware(nil), func(c *gin.Context) { called = true; c.Status(http.StatusNoContent) })
	request := httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader(`{"model":"openai/gpt5.4-mini","input":"test"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Contains(t, recorder.Body.String(), "model_retired")
	require.False(t, called)
}

func TestRetiredModelsRejectedInGeminiModelURI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	called := false
	engine.POST("/v1beta/models/*modelAction", compositeGeminiTargetPlatformMiddleware(nil), func(c *gin.Context) { called = true; c.Status(http.StatusNoContent) })
	request := httptest.NewRequest(http.MethodPost, "/v1beta/models/gpt-5.5-pro:generateContent", strings.NewReader(`{}`))
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.False(t, called)
}
