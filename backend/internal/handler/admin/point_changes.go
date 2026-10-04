package admin

import (
	"strings"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/pkg/response"
	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/gin-gonic/gin"
)

func (h *UserHandler) ListPointChanges(c *gin.Context) {
	page, size := response.ParsePagination(c)
	f := service.PointChangeFilter{Keyword: strings.TrimSpace(c.Query("keyword")),
		UserID: c.Query("user_id"), PointType: c.Query("point_type"), Direction: c.Query("direction"), Page: page, PageSize: size}
	if f.PointType != "" && f.PointType != service.WalletKindRecharge && f.PointType != service.WalletKindBonus {
		response.BadRequest(c, "Invalid point_type")
		return
	}
	if f.Direction != "" && f.Direction != "increase" && f.Direction != "decrease" {
		response.BadRequest(c, "Invalid direction")
		return
	}
	if f.UserID != "" {
		id, err := parseEntityID(f.UserID)
		if err != nil {
			response.BadRequest(c, "Invalid user ID")
			return
		}
		f.UserID = id
	}
	for name, target := range map[string]**time.Time{"start_time": &f.StartTime, "end_time": &f.EndTime} {
		if raw := c.Query(name); raw != "" {
			value, err := time.Parse(time.RFC3339Nano, raw)
			if err != nil {
				response.BadRequest(c, "Invalid "+name)
				return
			}
			*target = &value
		}
	}
	if f.StartTime != nil && f.EndTime != nil && !f.EndTime.After(*f.StartTime) {
		response.BadRequest(c, "end_time must be after start_time")
		return
	}
	svc, ok := h.adminService.(service.PointChangeReader)
	if !ok {
		response.Error(c, 503, "Point change query is unavailable")
		return
	}
	result, err := svc.ListPointChanges(c.Request.Context(), f)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}
