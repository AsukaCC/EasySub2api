package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"strings"

	"github.com/AsukaCC/EasySub2api/internal/handler/dto"
	"github.com/AsukaCC/EasySub2api/internal/pkg/adminaccess"
	"github.com/AsukaCC/EasySub2api/internal/pkg/pagination"
	"github.com/AsukaCC/EasySub2api/internal/pkg/response"
	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/gin-gonic/gin"
)

func scopedAdminGroupModelsDTO(g *service.Group) gin.H {
	return gin.H{"id": g.ID, "name": g.Name, "platform": g.Platform, "models_list_config": g.ModelsListConfig}
}

// ScopedAdminKeyGroupModels 接管分组模型接口，禁止复用可修改费率等字段的完整处理器。
func (h *GroupHandler) ScopedAdminKeyGroupModels(c *gin.Context) {
	p := adminaccess.FromContext(c.Request.Context())
	if p == nil {
		return
	}
	base := "/api/v1/admin/groups/:id"
	path := c.FullPath()
	fullRead := slices.Contains(p.Policy.GlobalPermissions, "groups.read")
	if c.Request.Method == http.MethodGet && (path == "/api/v1/admin/groups" || path == "/api/v1/admin/groups/all") {
		if fullRead {
			return
		}
		c.Abort()
		h.scopedAdminListGroupModels(c)
		return
	}
	candidates := path == base+"/models-list-candidates"
	if !candidates && path != base {
		return
	}
	c.Abort()
	id := c.Param("id")
	if strings.TrimSpace(id) != id || (id == "" && !candidates) {
		response.BadRequest(c, "分组 ID 无效")
		return
	}
	g, err := h.adminService.GetGroup(c.Request.Context(), id)
	if err != nil || g == nil {
		response.Error(c, 404, "分组不存在")
		return
	}
	permission := "groups.models.read"
	if c.Request.Method == http.MethodPut {
		permission = "groups.models.write"
	}
	allowed := p.Policy.AllowsGroupConfig(permission, g.ID, g.Platform) || (c.Request.Method == http.MethodGet && fullRead)
	if !allowed {
		response.Error(c, 404, "分组不存在")
		return
	}
	if c.Request.Method == http.MethodGet {
		if candidates {
			if platform := strings.TrimSpace(c.Query("platform")); platform != "" && platform != g.Platform {
				response.BadRequest(c, "候选模型平台必须与分组一致")
				return
			}
			h.GetModelsListCandidates(c)
			return
		}
		if fullRead {
			response.Success(c, dto.GroupFromService(g))
		} else {
			response.Success(c, scopedAdminGroupModelsDTO(g))
		}
		return
	}
	if c.Request.Method != http.MethodPut || candidates {
		response.Error(c, 403, "此分组操作未开放给受限管理密钥")
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 65536)
	var fields map[string]json.RawMessage
	if c.ShouldBindJSON(&fields) != nil || len(fields) != 1 {
		response.BadRequest(c, "仅支持单独提交 models_list_config")
		return
	}
	raw := fields["models_list_config"]
	var cfg *service.GroupModelsListConfig
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&cfg) != nil || cfg == nil {
		response.BadRequest(c, "模型展示配置格式无效")
		return
	}
	api, ok := h.adminService.(interface {
		ScopedAdminUpdateGroupModels(context.Context, string, service.GroupModelsListConfig) (*service.Group, error)
	})
	if !ok {
		response.Error(c, 503, "受限分组模型服务不可用")
		return
	}
	updated, err := api.ScopedAdminUpdateGroupModels(c.Request.Context(), id, *cfg)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, scopedAdminGroupModelsDTO(updated))
}

func (h *GroupHandler) scopedAdminListGroupModels(c *gin.Context) {
	api, ok := h.adminService.(interface {
		ScopedAdminListGroupModels(context.Context, pagination.PaginationParams, bool, string, string, string) ([]service.Group, int64, error)
	})
	if !ok {
		response.Error(c, 503, "受限分组模型服务不可用")
		return
	}
	page, size := response.ParsePagination(c)
	all := strings.HasSuffix(c.FullPath(), "/all")
	status := c.Query("status")
	if all {
		status = service.StatusActive
		if c.Query("include_inactive") == "true" {
			status = ""
		}
	}
	search := strings.TrimSpace(c.Query("search"))
	if len(search) > 100 {
		search = search[:100]
	}
	groups, total, err := api.ScopedAdminListGroupModels(c.Request.Context(), pagination.PaginationParams{Page: page, PageSize: size}, all, c.Query("platform"), status, search)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	out := make([]gin.H, 0, len(groups))
	for i := range groups {
		out = append(out, scopedAdminGroupModelsDTO(&groups[i]))
	}
	if all {
		response.Success(c, out)
	} else {
		response.Paginated(c, out, total, page, size)
	}
}
