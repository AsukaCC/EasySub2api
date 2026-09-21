package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestModelFingerprintDirectModelCatalog(t *testing.T) {
	for _, platform := range []string{service.PlatformOpenAI, service.PlatformAnthropic, service.PlatformGemini, service.PlatformGrok} {
		t.Run(platform, func(t *testing.T) {
			const id = "01995000-0000-7000-8000-000000000001"
			mapping := map[string]any{"routed-alias": "gpt-6-astra"}
			svc := &availableModelsAdminService{stubAdminService: newStubAdminService(), account: service.Account{
				ID: id, Platform: platform, Type: service.AccountTypeAPIKey,
				Credentials: map[string]any{"model_mapping": mapping},
			}}
			router := setupAvailableModelsRouter(svc)
			for _, direct := range []bool{false, true} {
				url := "/api/v1/admin/accounts/" + id + "/models"
				if direct {
					url += "?direct=true"
				}
				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, url, nil))
				require.Equal(t, http.StatusOK, rec.Code)
				var response struct {
					Data []struct {
						ID string `json:"id"`
					} `json:"data"`
				}
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
				require.NotEmpty(t, response.Data)
				if direct {
					for _, model := range response.Data {
						require.NotEqual(t, "routed-alias", model.ID)
					}
				} else {
					require.Len(t, response.Data, 1)
					require.Equal(t, "routed-alias", response.Data[0].ID)
				}
			}
			require.Equal(t, mapping, svc.account.Credentials["model_mapping"])
		})
	}
}

func TestModelFingerprintRequiresAuditActor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := &AccountHandler{accountTestService: &service.AccountTestService{}}
	router.POST("/accounts/:id/model-fingerprint", handler.StartModelFingerprint)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/accounts/01995000-0000-7000-8000-000000000001/model-fingerprint", strings.NewReader(`{"model_id":"gpt-6-astra"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestModelFingerprintUsageFilter(t *testing.T) {
	repo := &adminUsageRepoCapture{}
	router := newAdminUsageRequestTypeTestRouter(repo)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/admin/usage?request_type=test", nil))
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, repo.listFilters.RequestType)
	require.Equal(t, int16(service.RequestTypeTest), *repo.listFilters.RequestType)
	require.Nil(t, repo.listFilters.Stream)
}
