//go:build unit

package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type proxyPartialUpdateService struct {
	service.AdminService
	input *service.UpdateProxyInput
}

func (s *proxyPartialUpdateService) UpdateProxy(_ context.Context, _ string, input *service.UpdateProxyInput) (*service.Proxy, error) {
	s.input = input
	return &service.Proxy{ID: "90000000-0000-0000-0000-000000000009"}, nil
}

func TestProxyHandlerUpdatePreservesFieldPresence(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, body       string
		set, clear, warn bool
	}{
		{name: "omitted", body: `{"status":"inactive"}`},
		{name: "null clears", body: `{"expires_at":null,"backup_proxy_id":null,"fallback_mode":"none","expiry_warn_days":0}`, set: true, clear: true, warn: true},
		{name: "zero clears expiry", body: `{"expires_at":0,"backup_proxy_id":null}`, set: true, clear: true},
		{name: "values", body: `{"expires_at":1800000000,"backup_proxy_id":"10000000-0000-0000-0000-000000000010","fallback_mode":"proxy","expiry_warn_days":7}`, set: true, warn: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := &proxyPartialUpdateService{}
			router := gin.New()
			router.PUT("/proxies/:id", NewProxyHandler(svc).Update)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, "/proxies/90000000-0000-0000-0000-000000000009", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)
			require.Equal(t, http.StatusOK, w.Code, w.Body.String())
			require.NotNil(t, svc.input)
			require.Equal(t, tc.set, svc.input.ExpiresAt != nil || svc.input.ClearExpiresAt)
			require.Equal(t, tc.set, svc.input.BackupProxyID != nil || svc.input.ClearBackupID)
			require.Equal(t, tc.warn, svc.input.ExpiryWarnDays != nil)
			if !tc.set || tc.clear {
				require.Nil(t, svc.input.ExpiresAt)
				require.Nil(t, svc.input.BackupProxyID)
			} else {
				require.Equal(t, int64(1800000000), svc.input.ExpiresAt.Unix())
				require.Equal(t, "10000000-0000-0000-0000-000000000010", *svc.input.BackupProxyID)
			}
		})
	}
}
