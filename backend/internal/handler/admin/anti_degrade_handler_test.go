package admin

import (
	"bytes"
	"github.com/AsukaCC/EasySub2api/internal/server/middleware"
	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProtectionHandlerRequiresAdminAndDisableConfirmation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAntiDegradeHandler(service.NewAntiDegradeService(nil))
	for _, check := range []struct {
		role, body string
		status     int
	}{
		{"user", `{"enabled":false,"confirm_disable":true}`, http.StatusForbidden},
		{"admin", `{"enabled":false}`, http.StatusBadRequest},
	} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/accounts/a/anti-degrade/revert", bytes.NewBufferString(check.body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set(string(middleware.ContextKeyUserRole), check.role)
		c.Params = gin.Params{{Key: "id", Value: "a"}}
		h.Revert(c)
		require.Equal(t, check.status, w.Code)
	}
}
