package admin

import (
	"github.com/AsukaCC/EasySub2api/internal/pkg/response"
	"github.com/AsukaCC/EasySub2api/internal/server/middleware"
	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/gin-gonic/gin"
	"strconv"
)

func (h *AccountHandler) StartModelFingerprint(c *gin.Context) {
	id, err := parseEntityID(c.Param("id"))
	if err != nil || id == "" {
		response.BadRequest(c, "Invalid account ID")
		return
	}
	var request struct {
		Model           string `json:"model_id" binding:"required,max=100"`
		Protocol        string `json:"protocol"`
		ReasoningEffort string `json:"reasoning_effort"`
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
	snapshot, err := h.accountTestService.StartModelFingerprint(c.Request.Context(), id, request.Model, subject.UserID, service.ModelFingerprintOptions{Protocol: request.Protocol, ReasoningEffort: request.ReasoningEffort})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, snapshot)
}

func (h *AccountHandler) GetFingerprintModels(c *gin.Context) {
	id, err := parseEntityID(c.Param("id"))
	if err != nil || id == "" {
		response.BadRequest(c, "Invalid account ID")
		return
	}
	models, err := h.accountTestService.GetFingerprintModels(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, models)
}

func (h *AccountHandler) GetFingerprintHistory(c *gin.Context) {
	id, err := parseEntityID(c.Param("id"))
	if err != nil || id == "" {
		response.BadRequest(c, "Invalid account ID")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	result, err := h.accountTestService.GetFingerprintHistory(c.Request.Context(), id, page, size)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) GetFingerprintSchedule(c *gin.Context) {
	id, err := parseEntityID(c.Param("id"))
	if err != nil || id == "" {
		response.BadRequest(c, "Invalid account ID")
		return
	}
	result, err := h.accountTestService.GetFingerprintSchedule(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) SetFingerprintSchedule(c *gin.Context) {
	id, err := parseEntityID(c.Param("id"))
	if err != nil || id == "" {
		response.BadRequest(c, "Invalid account ID")
		return
	}
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Admin not authenticated")
		return
	}
	var request service.ModelFingerprintSchedule
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid schedule")
		return
	}
	request.AccountID, request.UserID = id, subject.UserID
	if err := h.accountTestService.SetFingerprintSchedule(c.Request.Context(), &request); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, request)
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
