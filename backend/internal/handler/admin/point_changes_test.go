package admin

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type pointChangeServiceStub struct {
	service.AdminService
	filter service.PointChangeFilter
}

func (s *pointChangeServiceStub) ListPointChanges(_ context.Context, filter service.PointChangeFilter) (service.PointChangePage, error) {
	s.filter = filter
	return service.PointChangePage{Items: []service.PointChange{}, Page: filter.Page, PageSize: filter.PageSize}, nil
}

func TestPointChangeQueryValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, query := range []string{"point_type=invalid", "direction=invalid", "user_id=bad", "start_time=2026-10-03", "start_time=2026-10-04T00:00:00Z&end_time=2026-10-03T00:00:00Z"} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/?"+query, nil)
		(&UserHandler{}).ListPointChanges(c)
		require.Equal(t, 400, w.Code, query)
	}
	svc := &pointChangeServiceStub{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/?keyword=person&point_type=bonus&direction=decrease&start_time=2026-10-03T00:00:00%2B08:00&end_time=2026-10-04T00:00:00%2B08:00&page=2&page_size=20", nil)
	(&UserHandler{adminService: svc}).ListPointChanges(c)
	require.Equal(t, 200, w.Code)
	require.Equal(t, "person", svc.filter.Keyword)
	require.Equal(t, "bonus", svc.filter.PointType)
	require.Equal(t, "decrease", svc.filter.Direction)
	require.Equal(t, 2, svc.filter.Page)
	require.Equal(t, "2026-10-02T16:00:00Z", svc.filter.StartTime.UTC().Format(time.RFC3339))
}
