package admin

import (
	"github.com/AsukaCC/EasySub2api/internal/pkg/response"
	"github.com/AsukaCC/EasySub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
)

func (h *AccountHandler) StartModelFingerprint(c *gin.Context) {
	id, err := parseEntityID(c.Param("id"))
	if err != nil || id == "" {
		response.BadRequest(c, "Invalid account ID")
		return
	}
	var request struct {
		Model string `json:"model_id" binding:"required,max=100"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid model selection")
		return
	}
	if h.accountTestService == nil {
		response.InternalError(c, "Account test service unavailable")
		return
	}
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Admin not authenticated")
		return
	}
	snapshot, err := h.accountTestService.StartModelFingerprint(c.Request.Context(), id, request.Model, subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, snapshot)
}

func (h *AccountHandler) GetModelFingerprint(c *gin.Context) {
	id, err := parseEntityID(c.Param("id"))
	if err != nil || id == "" {
		response.BadRequest(c, "Invalid account ID")
		return
	}
	if h.accountTestService == nil {
		response.InternalError(c, "Account test service unavailable")
		return
	}
	snapshot, err := h.accountTestService.GetModelFingerprint(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, snapshot)
}
