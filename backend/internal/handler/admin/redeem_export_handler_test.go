package admin

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func setupRedeemExportRouter() (*gin.Engine, *stubAdminService) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	adminSvc := newStubAdminService()

	h := NewRedeemHandler(adminSvc, nil)
	router.GET("/api/v1/admin/redeem-codes/export", h.Export)
	return router, adminSvc
}

func TestRedeemExportReadsEveryPage(t *testing.T) {
	for _, count := range []int{0, 1000, 1001, 2501} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			router, adminSvc := setupRedeemExportRouter()
			adminSvc.redeems = make([]service.RedeemCode, count)
			for i := range adminSvc.redeems {
				adminSvc.redeems[i] = service.RedeemCode{ID: fmt.Sprint(i), Code: fmt.Sprintf("CODE-%d", i)}
			}
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/admin/redeem-codes/export", nil))
			require.Equal(t, http.StatusOK, rec.Code)
			rows, err := csv.NewReader(rec.Body).ReadAll()
			require.NoError(t, err)
			require.Len(t, rows, count+1)
			for i, row := range rows[1:] {
				require.Equal(t, fmt.Sprint(i), row[0])
			}
			require.Equal(t, max(1, (count+999)/1000), adminSvc.lastListRedeemCodes.calls)
		})
	}
}

func TestRedeemExportDoesNotReturnPartialCSVOnPageFailure(t *testing.T) {
	router, adminSvc := setupRedeemExportRouter()
	adminSvc.redeems = make([]service.RedeemCode, 1001)
	for i := range adminSvc.redeems {
		adminSvc.redeems[i].ID = fmt.Sprint(i)
	}
	adminSvc.listRedeemCodesErrorPage = 2
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/admin/redeem-codes/export", nil))
	require.Equal(t, http.StatusNotFound, rec.Code)
	require.NotContains(t, rec.Header().Get("Content-Type"), "text/csv")
	require.Empty(t, rec.Header().Get("Content-Disposition"))
}

func TestRedeemExportRejectsDuplicateRowsAcrossPages(t *testing.T) {
	router, adminSvc := setupRedeemExportRouter()
	adminSvc.redeems = make([]service.RedeemCode, 1001)
	for i := range adminSvc.redeems {
		adminSvc.redeems[i].ID = fmt.Sprint(i)
	}
	adminSvc.redeems[1000].ID = adminSvc.redeems[999].ID
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/admin/redeem-codes/export", nil))
	require.Equal(t, http.StatusConflict, rec.Code)
	require.Empty(t, rec.Header().Get("Content-Disposition"))
	require.Equal(t, 2, adminSvc.lastListRedeemCodes.calls)
}

func TestRedeemExportPassesSearchAndSort(t *testing.T) {
	router, adminSvc := setupRedeemExportRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/redeem-codes/export?type=balance&status=unused&search=ABC&sort_by=value&sort_order=asc", nil)
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	require.Equal(t, 1, adminSvc.lastListRedeemCodes.calls)
	require.Equal(t, "balance", adminSvc.lastListRedeemCodes.codeType)
	require.Equal(t, "unused", adminSvc.lastListRedeemCodes.status)
	require.Equal(t, "ABC", adminSvc.lastListRedeemCodes.search)
	require.Equal(t, "value", adminSvc.lastListRedeemCodes.sortBy)
	require.Equal(t, "asc", adminSvc.lastListRedeemCodes.sortOrder)
}

func TestRedeemExportSortDefaults(t *testing.T) {
	router, adminSvc := setupRedeemExportRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/redeem-codes/export", nil)
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	require.Equal(t, 1, adminSvc.lastListRedeemCodes.calls)
	require.Equal(t, "id", adminSvc.lastListRedeemCodes.sortBy)
	require.Equal(t, "desc", adminSvc.lastListRedeemCodes.sortOrder)
}
