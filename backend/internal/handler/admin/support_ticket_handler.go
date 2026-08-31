package admin

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

type adminTicketMessageRequest struct {
	Message string `json:"message" binding:"required"`
}
type adminTicketStatusRequest struct {
	Status  string `json:"status" binding:"required"`
	Message string `json:"message"`
}
type adminTicketRefundRequest struct {
	Decision                string   `json:"decision" binding:"required"`
	ApprovedPrincipalAmount *float64 `json:"approved_principal_amount"`
	Message                 string   `json:"message"`
}
type adminCreateRefundTicketRequest struct {
	OrderID                 string   `json:"order_id" binding:"required"`
	ApprovedPrincipalAmount *float64 `json:"approved_principal_amount"`
	Message                 string   `json:"message" binding:"required"`
}

func (h *SupportTicketHandler) CreateRefund(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	var req adminCreateRefundTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.CreateAdminRefund(c.Request.Context(), service.CreateAdminRefundTicketInput{
		OrderID: req.OrderID, ReviewerID: subject.UserID, Amount: req.ApprovedPrincipalAmount, Message: req.Message,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *SupportTicketHandler) List(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	page, pageSize := response.ParsePagination(c)
	result, err := h.service.ListAdmin(c.Request.Context(), subject.UserID, service.SupportTicketListFilters{Category: c.Query("category"), Status: c.Query("status"), Search: c.Query("search"), Unread: parseAdminTicketBool(c.Query("unread"))}, page, pageSize)
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
	result, err := h.service.Summary(c.Request.Context(), subject.UserID, true)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *SupportTicketHandler) Detail(c *gin.Context) {
	result, err := h.service.GetAdminDetail(c.Request.Context(), strings.TrimSpace(c.Param("id")))
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
	var req adminTicketMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.Reply(c.Request.Context(), strings.TrimSpace(c.Param("id")), subject.UserID, service.SupportTicketRoleAdmin, req.Message)
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
	if err := h.service.MarkRead(c.Request.Context(), strings.TrimSpace(c.Param("id")), subject.UserID, true); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "ok"})
}

func (h *SupportTicketHandler) SetStatus(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	var req adminTicketStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.AdminSetStatus(c.Request.Context(), strings.TrimSpace(c.Param("id")), subject.UserID, req.Status, req.Message)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *SupportTicketHandler) ReviewRefund(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	var req adminTicketRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.ReviewRefund(c.Request.Context(), service.ReviewSupportTicketRefundInput{TicketID: strings.TrimSpace(c.Param("id")), ReviewerID: subject.UserID, Decision: req.Decision, Amount: req.ApprovedPrincipalAmount, Message: req.Message})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func parseAdminTicketBool(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return value == "1" || value == "true" || value == "yes"
}
