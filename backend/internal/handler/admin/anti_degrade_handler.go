package admin

import (
	"errors"
	"strings"

	"github.com/AsukaCC/EasySub2api/internal/handler/dto"
	"github.com/AsukaCC/EasySub2api/internal/pkg/response"
	"github.com/AsukaCC/EasySub2api/internal/server/middleware"
	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// AntiDegradeHandler 一键防降智：预览 / 应用 / 还原账号保护组合。
type AntiDegradeHandler struct {
	service *service.AntiDegradeService
}

func NewAntiDegradeHandler(svc *service.AntiDegradeService) *AntiDegradeHandler {
	return &AntiDegradeHandler{service: svc}
}

func (h *AntiDegradeHandler) requireService(c *gin.Context) bool {
	role, ok := middleware.GetUserRoleFromContext(c)
	if !ok || role != service.RoleAdmin {
		response.Forbidden(c, "Admin access required")
		return false
	}
	if h == nil || h.service == nil {
		response.ErrorFrom(c, errors.New("anti-degrade service unavailable"))
		return false
	}
	return true
}

func antiDegradeID(c *gin.Context) (string, bool) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		response.BadRequest(c, "Invalid account ID")
		return "", false
	}
	return id, true
}

// Strategies returns the server-owned, data-driven strategy registry.  It is
// read-only and intentionally available only to administrators.
func (h *AntiDegradeHandler) Strategies(c *gin.Context) {
	if !h.requireService(c) {
		return
	}
	response.Success(c, gin.H{"strategies": service.ListAntiDegradeStrategyProfiles()})
}

// Preview 预览改动集（不写库）。
// GET /api/v1/admin/accounts/:id/anti-degrade
func (h *AntiDegradeHandler) Preview(c *gin.Context) {
	if !h.requireService(c) {
		return
	}
	id, ok := antiDegradeID(c)
	if !ok {
		return
	}
	mode := service.AntiDegradeMode(c.Query("mode"))
	account, err := h.service.PreviewMode(c.Request.Context(), id, mode)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, account)
}

// Apply 应用一键防降智并快照旧值。
// POST /api/v1/admin/accounts/:id/anti-degrade/apply
func (h *AntiDegradeHandler) Apply(c *gin.Context) {
	if !h.requireService(c) {
		return
	}
	id, ok := antiDegradeID(c)
	if !ok {
		return
	}
	mode := service.AntiDegradeMode(c.Query("mode"))
	account, err := h.service.ApplyAntiDegradeMode(c.Request.Context(), id, mode)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	middleware.SetAuditAction(c, "account.protection.apply")
	middleware.SetAuditExtra(c, map[string]any{"mode": string(mode)})
	response.Success(c, dto.AccountFromService(account))
}

// Revert 一键还原快照旧值。
// POST /api/v1/admin/accounts/:id/anti-degrade/revert
func (h *AntiDegradeHandler) Revert(c *gin.Context) {
	if !h.requireService(c) {
		return
	}
	id, ok := antiDegradeID(c)
	if !ok {
		return
	}
	var req struct {
		ConfirmDisable bool `json:"confirm_disable"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || !req.ConfirmDisable {
		response.BadRequest(c, "关闭防降智模式需要管理员明确确认")
		return
	}
	middleware.SetAuditAction(c, "account.protection.disable")
	middleware.SetAuditExtra(c, map[string]any{"enabled": false, "confirm": true})
	account, err := h.service.Revert(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.AccountFromService(account))
}

// SetProtection handles the explicit, administrator-confirmed ON/OFF control.
func (h *AntiDegradeHandler) SetProtection(c *gin.Context) {
	if !h.requireService(c) {
		return
	}
	id, ok := antiDegradeID(c)
	if !ok {
		return
	}
	var req struct {
		Enabled        *bool `json:"enabled" binding:"required"`
		ConfirmDisable bool  `json:"confirm_disable"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid protection request")
		return
	}
	middleware.SetAuditAction(c, "account.protection.update")
	middleware.SetAuditExtra(c, map[string]any{"enabled": *req.Enabled, "confirm": req.ConfirmDisable})
	account, err := h.service.SetProtection(c.Request.Context(), id, *req.Enabled, req.ConfirmDisable)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.AccountFromService(account))
}

// EnableBatch deliberately has no disable branch.
func (h *AntiDegradeHandler) EnableBatch(c *gin.Context) {
	if !h.requireService(c) {
		return
	}
	var req struct {
		AccountIDs []string `json:"account_ids" binding:"required,min=1,max=500,dive,required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Select 1–500 valid account IDs")
		return
	}
	middleware.SetAuditAction(c, "account.protection.enable_batch")
	middleware.SetAuditExtra(c, map[string]any{"enabled": true, "requested_count": len(req.AccountIDs)})
	success := []string{}
	failures := map[string]string{}
	seen := map[string]bool{}
	for _, id := range req.AccountIDs {
		if seen[id] {
			continue
		}
		seen[id] = true
		if _, err := h.service.SetProtection(c.Request.Context(), id, true, false); err != nil {
			failures[id] = err.Error()
		} else {
			success = append(success, id)
		}
	}
	response.Success(c, gin.H{"success_ids": success, "failures": failures})
}

func (h *AccountHandler) protectionHandler() *AntiDegradeHandler {
	svc := service.NewAntiDegradeService(h.adminService)
	if h.accountTestService != nil {
		svc.SetRuntimeConfig(h.accountTestService.ProtectionConfig())
	}
	return NewAntiDegradeHandler(svc)
}
func (h *AccountHandler) ProtectionStrategies(c *gin.Context)  { h.protectionHandler().Strategies(c) }
func (h *AccountHandler) PreviewProtection(c *gin.Context)     { h.protectionHandler().Preview(c) }
func (h *AccountHandler) ApplyProtection(c *gin.Context)       { h.protectionHandler().Apply(c) }
func (h *AccountHandler) RevertProtection(c *gin.Context)      { h.protectionHandler().Revert(c) }
func (h *AccountHandler) SetProtection(c *gin.Context)         { h.protectionHandler().SetProtection(c) }
func (h *AccountHandler) EnableProtectionBatch(c *gin.Context) { h.protectionHandler().EnableBatch(c) }

func (h *AccountHandler) SetProtectionIntegrity(c *gin.Context) {
	if !h.protectionHandler().requireService(c) {
		return
	}
	id, ok := antiDegradeID(c)
	if !ok {
		return
	}
	var input struct {
		Mode string `json:"mode" binding:"required,oneof=off observe enforce"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid integrity mode")
		return
	}
	account, err := h.adminService.GetAccount(c.Request.Context(), id)
	if response.ErrorFrom(c, err) {
		return
	}
	if account.Platform != service.PlatformOpenAI {
		response.BadRequest(c, "Integrity checks require an OpenAI account")
		return
	}
	middleware.SetAuditAction(c, "account.protection.integrity")
	if err := h.adminService.UpdateAccountExtra(c.Request.Context(), id, map[string]any{"request_integrity_mode": input.Mode}); response.ErrorFrom(c, err) {
		return
	}
	account, err = h.adminService.GetAccount(c.Request.Context(), id)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, dto.AccountFromService(account))
}
