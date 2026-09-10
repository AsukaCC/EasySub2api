//go:build unit

package accountproxyupdate_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/handler/admin"
	"github.com/AsukaCC/EasySub2api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type accountProxyUpdateService struct {
	service.AdminService
	updateInput *service.UpdateAccountInput
	bulkInput   *service.BulkUpdateAccountsInput
}

func (s *accountProxyUpdateService) UpdateAccount(_ context.Context, _ string, input *service.UpdateAccountInput) (*service.Account, error) {
	s.updateInput = input
	return &service.Account{ID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"}, nil
}

func (s *accountProxyUpdateService) BulkUpdateAccounts(_ context.Context, input *service.BulkUpdateAccountsInput) (*service.BulkUpdateAccountsResult, error) {
	s.bulkInput = input
	return &service.BulkUpdateAccountsResult{Success: len(input.AccountIDs), Failed: 0, SuccessIDs: input.AccountIDs}, nil
}

func newAccountHandler(svc service.AdminService) *admin.AccountHandler {
	return admin.NewAccountHandler(svc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
}

func TestAccountHandlerUpdateTreatsNullProxyIDAsClear(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const accountID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	proxyID := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"

	for _, tc := range []struct {
		name    string
		body    string
		present bool
		clear   bool
		value   string
	}{
		{name: "omitted", body: `{"name":"keep-proxy"}`},
		{name: "null clears", body: `{"proxy_id":null}`, present: true, clear: true},
		{name: "empty clears", body: `{"proxy_id":""}`, present: true, clear: true},
		{name: "blank clears", body: `{"proxy_id":"  "}`, present: true, clear: true},
		{name: "value", body: `{"proxy_id":"` + proxyID + `"}`, present: true, value: proxyID},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := &accountProxyUpdateService{}
			router := gin.New()
			router.PUT("/accounts/:id", newAccountHandler(svc).Update)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, "/accounts/"+accountID, strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)
			require.Equal(t, http.StatusOK, w.Code, w.Body.String())
			require.NotNil(t, svc.updateInput)
			if !tc.present {
				require.Nil(t, svc.updateInput.ProxyID)
				return
			}
			require.NotNil(t, svc.updateInput.ProxyID)
			if tc.clear {
				require.Empty(t, *svc.updateInput.ProxyID)
				return
			}
			require.Equal(t, tc.value, *svc.updateInput.ProxyID)
		})
	}
}

func TestAccountHandlerBulkUpdateTreatsNullProxyIDAsClear(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &accountProxyUpdateService{}
	router := gin.New()
	router.POST("/accounts/bulk-update", newAccountHandler(svc).BulkUpdate)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/accounts/bulk-update", strings.NewReader(
		`{"account_ids":["aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"],"proxy_id":null}`,
	))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.NotNil(t, svc.bulkInput)
	require.NotNil(t, svc.bulkInput.ProxyID)
	require.Empty(t, *svc.bulkInput.ProxyID)
}
