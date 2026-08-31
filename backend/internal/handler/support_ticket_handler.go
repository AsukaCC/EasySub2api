package handler

import (
	"strings"

	"github.com/AsukaCC/EasySub2api/internal/pkg/response"
	"github.com/AsukaCC/EasySub2api/internal/server/middleware"
	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type SupportTicketHandler struct{ service *service.SupportTicketService }

func NewSupportTicketHandler(ticketService *service.SupportTicketService) *SupportTicketHandler {
	return &SupportTicketHandler{service: ticketService}
}

type createSupportTicketRequest struct {
	Category string `json:"category" binding:"required"`
	Title    string `json:"title"`
	Message  string `json:"message" binding:"required"`
	APIKeyID string `json:"api_key_id"`
	OrderID  string `json:"order_id"`
}

type supportTicketMessageRequest struct {
	Message string `json:"message" binding:"required"`
}

func (h *SupportTicketHandler) List(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	page, pageSize := response.ParsePagination(c)
	result, err := h.service.ListUser(c.Request.Context(), subject.UserID, service.SupportTicketListFilters{
		Category: c.Query("category"), Status: c.Query("status"), Search: c.Query("search"), Unread: parseBoolQuery(c.Query("unread")),
	}, page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, result.Items, int64(result.Total), page, pageSize)
}

func (h *SupportTicketHandler) Summary(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	result, err := h.service.Summary(c.Request.Context(), subject.UserID, false)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *SupportTicketHandler) Create(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	var req createSupportTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.Create(c.Request.Context(), service.CreateSupportTicketInput{UserID: subject.UserID, Category: req.Category, Title: req.Title, Message: req.Message, APIKeyID: req.APIKeyID, OrderID: req.OrderID})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *SupportTicketHandler) Detail(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	result, err := h.service.GetUserDetail(c.Request.Context(), strings.TrimSpace(c.Param("id")), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *SupportTicketHandler) Reply(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	var req supportTicketMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.Reply(c.Request.Context(), strings.TrimSpace(c.Param("id")), subject.UserID, service.SupportTicketRoleUser, req.Message)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *SupportTicketHandler) MarkRead(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	if err := h.service.MarkRead(c.Request.Context(), strings.TrimSpace(c.Param("id")), subject.UserID, false); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "ok"})
}

func (h *SupportTicketHandler) action(c *gin.Context, action string) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	result, err := h.service.UserAction(c.Request.Context(), strings.TrimSpace(c.Param("id")), subject.UserID, action)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *SupportTicketHandler) Cancel(c *gin.Context) { h.action(c, "cancel") }
func (h *SupportTicketHandler) Close(c *gin.Context)  { h.action(c, "close") }
func (h *SupportTicketHandler) Reopen(c *gin.Context) { h.action(c, "reopen") }
