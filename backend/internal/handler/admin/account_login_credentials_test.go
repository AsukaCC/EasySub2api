//go:build unit

package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/handler/dto"
	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type loginCredentialsAdminStub struct {
	service.AdminService
	account *service.Account
}

func (s *loginCredentialsAdminStub) GetAccount(context.Context, string) (*service.Account, error) {
	return s.account, nil
}

func TestAccountLoginCredentialsDedicatedReadAndRedaction(t *testing.T) {
	account := &service.Account{ID: "42000000-0000-0000-0000-000000000042", Credentials: map[string]any{
		"access_token": "upstream-token", "login_username": "user@example.com", "login_password": " password ", "login_totp_secret": "JBSWY3DPEHPK3PXP",
	}}
	redacted, status := dto.RedactCredentials(account.Credentials)
	require.Empty(t, redacted)
	require.True(t, status["has_login_password"])
	require.True(t, status["has_login_totp_secret"])
	h := &AccountHandler{adminService: &loginCredentialsAdminStub{account: account}}
	router := gin.New()
	router.GET("/accounts/:id/login-credentials", h.GetLoginCredentials)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/accounts/"+account.ID+"/login-credentials", nil))
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "no-store", recorder.Header().Get("Cache-Control"))
	require.Equal(t, " password ", gjson.GetBytes(recorder.Body.Bytes(), "data.login_password").String())
	require.Equal(t, "JBSWY3DPEHPK3PXP", gjson.GetBytes(recorder.Body.Bytes(), "data.login_totp_secret").String())
	require.NotContains(t, recorder.Body.String(), "upstream-token")
}
