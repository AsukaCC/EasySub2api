//go:build unit

package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/config"
	"github.com/AsukaCC/EasySub2api/internal/pkg/ctxkey"
	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestKeySchedulingAuthenticatedIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, google := range []bool{false, true} {
		t.Run(map[bool]string{false: "standard", true: "google"}[google], func(t *testing.T) {
			key := &service.APIKey{ID: "stored-key-id", UserID: "same-user", Key: "secret-credential", Status: service.StatusActive, User: &service.User{ID: "same-user", Status: service.StatusActive, Role: service.RoleUser}}
			repo := &stubApiKeyRepo{getByKey: func(context.Context, string) (*service.APIKey, error) { return key, nil }}
			cfg := &config.Config{RunMode: config.RunModeSimple}
			svc := service.NewAPIKeyService(repo, nil, nil, nil, nil, nil, cfg)
			router := gin.New()
			if google {
				router.Use(APIKeyAuthGoogle(svc, cfg))
			} else {
				router.Use(gin.HandlerFunc(NewAPIKeyAuthMiddleware(svc, nil, cfg)))
			}
			router.GET("/t", func(c *gin.Context) {
				require.Equal(t, key.ID, c.Request.Context().Value(ctxkey.APIKeyID))
				c.Status(http.StatusNoContent)
			})
			request := httptest.NewRequest(http.MethodGet, "/t", nil)
			request.Header.Set("x-api-key", key.Key)
			request.Header.Set("x-goog-api-key", key.Key)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			require.Equal(t, http.StatusNoContent, response.Code, response.Body.String())
		})
	}
}
