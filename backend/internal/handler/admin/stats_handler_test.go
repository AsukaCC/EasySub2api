package admin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestAdminStatsHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const id = "01990000-0000-7000-8000-000000000001"
	for _, route := range []string{"groups", "proxies", "redeem-codes"} {
		t.Run(route, func(t *testing.T) {
			svc := newStubAdminService()
			svc.groupStats = &service.AdminGroupStats{TotalAPIKeys: 3, ActiveAPIKeys: 2, TotalRequests: 12, TotalCost: 4.5}
			svc.proxyStats = &service.AdminProxyStats{TotalAccounts: 3, ActiveAccounts: 2}
			svc.redeemStats = &service.AdminRedeemStats{TotalCodes: 8, UsedCodes: 3, ByType: map[string]int{"balance": 8}}
			r := gin.New()
			var handler gin.HandlerFunc
			switch route {
			case "groups":
				handler = (&GroupHandler{adminService: svc}).GetStats
			case "proxies":
				handler = (&ProxyHandler{adminService: svc}).GetStats
			default:
				handler = (&RedeemHandler{adminService: svc}).GetStats
			}
			r.GET("/:id/stats", handler)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/"+id+"/stats", nil))
			require.Equal(t, http.StatusOK, rec.Code)
			data := gjson.GetBytes(rec.Body.Bytes(), "data")
			switch route {
			case "groups":
				require.Equal(t, int64(12), data.Get("total_requests").Int())
				require.Equal(t, 4.5, data.Get("total_cost").Float())
			case "proxies":
				require.Equal(t, int64(3), data.Get("total_accounts").Int())
				for _, field := range []string{"success_rate", "total_requests", "average_latency"} {
					require.Equal(t, "null", data.Get(field).Raw)
				}
			default:
				require.Equal(t, int64(8), data.Get("total_codes").Int())
			}
			svc.statsErr = service.ErrGroupNotFound
			rec = httptest.NewRecorder()
			r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/"+id+"/stats", nil))
			require.Equal(t, http.StatusNotFound, rec.Code)
			if route != "redeem-codes" {
				rec = httptest.NewRecorder()
				r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/invalid/stats", nil))
				require.Equal(t, http.StatusBadRequest, rec.Code)
			}
		})
	}
}
