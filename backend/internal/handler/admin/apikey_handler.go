package admin

import (
	"context"
	"fmt"

	"github.com/AsukaCC/EasySub2api/internal/handler/dto"
	"github.com/AsukaCC/EasySub2api/internal/pkg/response"
	"github.com/AsukaCC/EasySub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// AdminAPIKeyHandler handles admin API key management
type AdminAPIKeyHandler struct {
	adminService service.AdminService
}

// NewAdminAPIKeyHandler creates a new admin API key handler
func NewAdminAPIKeyHandler(adminService service.AdminService) *AdminAPIKeyHandler {
	return &AdminAPIKeyHandler{
		adminService: adminService,
	}
}

// AdminUpdateAPIKeyGroupRequest represents the request to update an API key.
type AdminUpdateAPIKeyGroupRequest struct {
	GroupID             *string   `json:"group_id"`               // nil=不修改, 0=解绑, >0=绑定到目标分组
	GroupIDs            *[]string `json:"group_ids"`              // 非 nil 时替换为多分组绑定，空数组表示解绑
	ResetRateLimitUsage *bool     `json:"reset_rate_limit_usage"` // true=重置 5h/1d/7d 限速用量
}

type adminAPIKeyGroupsUpdater interface {
	AdminUpdateAPIKeyGroups(context.Context, string, []string) (*service.AdminUpdateAPIKeyGroupIDResult, error)
}

// UpdateGroup handles updating an API key's admin-managed fields.
// PUT /api/v1/admin/api-keys/:id
func (h *AdminAPIKeyHandler) UpdateGroup(c *gin.Context) {
	keyID, err := parseEntityID(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid API key ID")
		return
	}

	var req AdminUpdateAPIKeyGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	var resetKey *service.APIKey
	if req.ResetRateLimitUsage != nil && *req.ResetRateLimitUsage {
		resetKey, err = h.adminService.AdminResetAPIKeyRateLimitUsage(c.Request.Context(), keyID)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
	}

	var result *service.AdminUpdateAPIKeyGroupIDResult
	if req.GroupIDs != nil {
		updater, ok := h.adminService.(adminAPIKeyGroupsUpdater)
		if !ok {
			response.ErrorFrom(c, fmt.Errorf("admin service does not support multi-group API key updates"))
			return
		}
		result, err = updater.AdminUpdateAPIKeyGroups(c.Request.Context(), keyID, *req.GroupIDs)
	} else {
		result, err = h.adminService.AdminUpdateAPIKeyGroupID(c.Request.Context(), keyID, req.GroupID)
	}
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if resetKey != nil && req.GroupID == nil {
		result.APIKey = resetKey
	}

	resp := struct {
		APIKey                 *dto.APIKey `json:"api_key"`
		AutoGrantedGroupAccess bool        `json:"auto_granted_group_access"`
		GrantedGroupID         *string     `json:"granted_group_id,omitempty"`
		GrantedGroupName       string      `json:"granted_group_name,omitempty"`
	}{
		APIKey:                 dto.APIKeyFromService(result.APIKey),
		AutoGrantedGroupAccess: result.AutoGrantedGroupAccess,
		GrantedGroupID:         result.GrantedGroupID,
		GrantedGroupName:       result.GrantedGroupName,
	}
	response.Success(c, resp)
}
