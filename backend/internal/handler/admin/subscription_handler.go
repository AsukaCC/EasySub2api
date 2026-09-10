package admin

import (
	"context"
	"strings"

	"github.com/AsukaCC/EasySub2api/internal/handler/dto"
	"github.com/AsukaCC/EasySub2api/internal/pkg/pagination"
	"github.com/AsukaCC/EasySub2api/internal/pkg/response"
	middleware2 "github.com/AsukaCC/EasySub2api/internal/server/middleware"
	"github.com/AsukaCC/EasySub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// toResponsePagination converts pagination.PaginationResult to response.PaginationResult
func toResponsePagination(p *pagination.PaginationResult) *response.PaginationResult {
	if p == nil {
		return nil
	}
	return &response.PaginationResult{
		Total:    p.Total,
		Page:     p.Page,
		PageSize: p.PageSize,
		Pages:    p.Pages,
	}
}

// SubscriptionHandler handles admin subscription management
type SubscriptionHandler struct {
	subscriptionService *service.SubscriptionService
}

// NewSubscriptionHandler creates a new admin subscription handler
func NewSubscriptionHandler(subscriptionService *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionService: subscriptionService,
	}
}

// AssignSubscriptionRequest represents assign subscription request
type AssignSubscriptionRequest struct {
	UserID       string `json:"user_id" binding:"required"`
	GroupID      string `json:"group_id" binding:"required"`
	ValidityDays int    `json:"validity_days" binding:"omitempty,max=36500"` // max 100 years
	Notes        string `json:"notes"`
}

// BulkAssignSubscriptionRequest represents bulk assign subscription request
type BulkAssignSubscriptionRequest struct {
	UserIDs      []string `json:"user_ids" binding:"required,min=1"`
	GroupID      string   `json:"group_id" binding:"required"`
	ValidityDays int      `json:"validity_days" binding:"omitempty,max=36500"` // max 100 years
	Notes        string   `json:"notes"`
}

// AdjustSubscriptionRequest represents adjust subscription request (extend or shorten)
type AdjustSubscriptionRequest struct {
	Days int `json:"days" binding:"required,min=-36500,max=36500"` // negative to shorten, positive to extend
}

// List handles listing all subscriptions with pagination and filters
// GET /api/v1/admin/subscriptions
func (h *SubscriptionHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)

	// Parse optional filters
	var userID, groupID *string
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if id, err := parseEntityID(userIDStr); err == nil {
			userID = &id
		}
	}
	if groupIDStr := c.Query("group_id"); groupIDStr != "" {
		if id, err := parseEntityID(groupIDStr); err == nil {
			groupID = &id
		}
	}
	status := c.Query("status")
	platform := c.Query("platform")

	// Parse sorting parameters
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")

	subscriptions, pagination, err := h.subscriptionService.List(c.Request.Context(), page, pageSize, userID, groupID, status, platform, sortBy, sortOrder)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]dto.AdminUserSubscription, 0, len(subscriptions))
	for i := range subscriptions {
		out = append(out, *dto.UserSubscriptionFromServiceAdmin(&subscriptions[i]))
	}
	response.PaginatedWithResult(c, out, toResponsePagination(pagination))
}

// ListPending returns read-only pending subscription activations for admins.
func (h *SubscriptionHandler) ListPending(c *gin.Context) {
	items, err := h.subscriptionService.ListPendingSubscriptionsAdmin(
		c.Request.Context(), c.Query("user_id"), c.Query("platform"), c.Query("group_id"),
	)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
}

// GetByID handles getting a subscription by ID
// GET /api/v1/admin/subscriptions/:id
func (h *SubscriptionHandler) GetByID(c *gin.Context) {
	subscriptionID, err := parseEntityID(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID")
		return
	}

	subscription, err := h.subscriptionService.GetByID(c.Request.Context(), subscriptionID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.UserSubscriptionFromServiceAdmin(subscription))
}

// GetProgress handles getting subscription usage progress
// GET /api/v1/admin/subscriptions/:id/progress
func (h *SubscriptionHandler) GetProgress(c *gin.Context) {
	subscriptionID, err := parseEntityID(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID")
		return
	}

	progress, err := h.subscriptionService.GetSubscriptionProgress(c.Request.Context(), subscriptionID)
	if err != nil {
		response.NotFound(c, "Subscription not found")
		return
	}

	response.Success(c, progress)
}

// Assign handles assigning a subscription to a user
// POST /api/v1/admin/subscriptions/assign
func (h *SubscriptionHandler) Assign(c *gin.Context) {
	var req AssignSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	// Get admin user ID from context
	adminID := getAdminIDFromContext(c)

	grant, err := h.subscriptionService.GrantOrQueueSubscription(c.Request.Context(), &service.SubscriptionGrantInput{
		AssignSubscriptionInput: service.AssignSubscriptionInput{
			UserID: req.UserID, GroupID: req.GroupID, ValidityDays: req.ValidityDays,
			AssignedBy: adminID, Notes: req.Notes,
		},
		SourceType: "admin_assignment",
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, grant)
}

// BulkAssign handles bulk assigning subscriptions to multiple users
// POST /api/v1/admin/subscriptions/bulk-assign
func (h *SubscriptionHandler) BulkAssign(c *gin.Context) {
	var req BulkAssignSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	// Get admin user ID from context
	adminID := getAdminIDFromContext(c)

	result, err := h.subscriptionService.BulkAssignSubscription(c.Request.Context(), &service.BulkAssignSubscriptionInput{
		UserIDs:      req.UserIDs,
		GroupID:      req.GroupID,
		ValidityDays: req.ValidityDays,
		AssignedBy:   adminID,
		Notes:        req.Notes,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.BulkAssignResultFromService(result))
}

// Extend handles adjusting a subscription (extend or shorten)
// POST /api/v1/admin/subscriptions/:id/extend
func (h *SubscriptionHandler) Extend(c *gin.Context) {
	subscriptionID, err := parseEntityID(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID")
		return
	}

	var req AdjustSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	idempotencyPayload := struct {
		SubscriptionID string                    `json:"subscription_id"`
		Body           AdjustSubscriptionRequest `json:"body"`
	}{
		SubscriptionID: subscriptionID,
		Body:           req,
	}
	executeAdminIdempotentJSON(c, "admin.subscriptions.extend", idempotencyPayload, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		subscription, execErr := h.subscriptionService.ExtendSubscription(ctx, subscriptionID, req.Days)
		if execErr != nil {
			return nil, execErr
		}
		return dto.UserSubscriptionFromServiceAdmin(subscription), nil
	})
}

// ResetSubscriptionQuotaRequest represents the reset quota request
type ResetSubscriptionQuotaRequest struct {
	Daily   bool `json:"daily"`
	Weekly  bool `json:"weekly"`
	Monthly bool `json:"monthly"`
}

type IssueResetCardsRequest struct {
	SubscriptionIDs []string `json:"subscription_ids" binding:"required,min=1"`
	Quantity        int      `json:"quantity" binding:"required,min=1,max=1000"`
	ValidityDays    int      `json:"validity_days" binding:"omitempty,min=1,max=3650"`
}

type ResetWeeklySubscriptionsRequest struct {
	SubscriptionIDs []string `json:"subscription_ids" binding:"required,min=1"`
}

// IssueResetCards grants independently expiring reset cards to one or more
// subscriptions. Each subscription is committed independently so a single
// invalid row does not roll back successful rows.
// POST /api/v1/admin/subscriptions/reset-cards/issue
func (h *SubscriptionHandler) IssueResetCards(c *gin.Context) {
	var req IssueResetCardsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	for _, id := range req.SubscriptionIDs {
		if strings.TrimSpace(id) == "" {
			response.BadRequest(c, "subscription_ids must contain non-empty subscription IDs")
			return
		}
	}
	payload := struct {
		SubscriptionIDs []string `json:"subscription_ids"`
		Quantity        int      `json:"quantity"`
		ValidityDays    int      `json:"validity_days"`
	}{req.SubscriptionIDs, req.Quantity, req.ValidityDays}
	adminID := getAdminIDFromContext(c)
	executeAdminIdempotentJSON(c, "admin.subscriptions.issue_reset_cards", payload, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		result := h.subscriptionService.IssueResetCards(ctx, req.SubscriptionIDs, req.Quantity, req.ValidityDays, adminID, c.GetHeader("Idempotency-Key"))
		middleware2.SetAuditExtra(c, map[string]any{
			"subscription_count": len(req.SubscriptionIDs),
			"quantity":           req.Quantity,
			"validity_days":      req.ValidityDays,
			"success_count":      result.SuccessCount,
			"failed_count":       result.FailedCount,
			"total_issued":       result.TotalIssued,
		})
		return result, nil
	})
}

// ResetWeekly resets only the rolling seven-day quota window and never
// consumes reset cards.
// POST /api/v1/admin/subscriptions/reset-weekly
func (h *SubscriptionHandler) ResetWeekly(c *gin.Context) {
	var req ResetWeeklySubscriptionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	for _, id := range req.SubscriptionIDs {
		if strings.TrimSpace(id) == "" {
			response.BadRequest(c, "subscription_ids must contain non-empty subscription IDs")
			return
		}
	}
	payload := struct {
		SubscriptionIDs []string `json:"subscription_ids"`
	}{req.SubscriptionIDs}
	executeAdminIdempotentJSON(c, "admin.subscriptions.reset_weekly", payload, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		result := h.subscriptionService.ResetWeeklyQuotas(ctx, req.SubscriptionIDs)
		middleware2.SetAuditExtra(c, map[string]any{
			"subscription_count": len(req.SubscriptionIDs),
			"success_count":      result.SuccessCount,
			"failed_count":       result.FailedCount,
		})
		return result, nil
	})
}

// ResetQuota resets daily, weekly, and/or monthly usage for a subscription.
// POST /api/v1/admin/subscriptions/:id/reset-quota
func (h *SubscriptionHandler) ResetQuota(c *gin.Context) {
	subscriptionID, err := parseEntityID(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID")
		return
	}
	var req ResetSubscriptionQuotaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if !req.Daily && !req.Weekly && !req.Monthly {
		response.BadRequest(c, "At least one of 'daily', 'weekly', or 'monthly' must be true")
		return
	}
	sub, err := h.subscriptionService.AdminResetQuota(c.Request.Context(), subscriptionID, req.Daily, req.Weekly, req.Monthly)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.UserSubscriptionFromServiceAdmin(sub))
}

// Revoke handles revoking a subscription.
// POST /api/v1/admin/subscriptions/:id/revoke
// DELETE /api/v1/admin/subscriptions/:id is kept for backward compatibility.
func (h *SubscriptionHandler) Revoke(c *gin.Context) {
	subscriptionID, err := parseEntityID(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID")
		return
	}

	err = h.subscriptionService.RevokeSubscription(c.Request.Context(), subscriptionID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Subscription revoked successfully"})
}

// Restore handles restoring a revoked subscription.
// POST /api/v1/admin/subscriptions/:id/restore
func (h *SubscriptionHandler) Restore(c *gin.Context) {
	subscriptionID, err := parseEntityID(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID")
		return
	}

	subscription, err := h.subscriptionService.RestoreSubscription(c.Request.Context(), subscriptionID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.UserSubscriptionFromServiceAdmin(subscription))
}

// ListByGroup handles listing subscriptions for a specific group
// GET /api/v1/admin/groups/:id/subscriptions
func (h *SubscriptionHandler) ListByGroup(c *gin.Context) {
	groupID, err := parseEntityID(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid group ID")
		return
	}

	page, pageSize := response.ParsePagination(c)

	subscriptions, pagination, err := h.subscriptionService.ListGroupSubscriptions(c.Request.Context(), groupID, page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]dto.AdminUserSubscription, 0, len(subscriptions))
	for i := range subscriptions {
		out = append(out, *dto.UserSubscriptionFromServiceAdmin(&subscriptions[i]))
	}
	response.PaginatedWithResult(c, out, toResponsePagination(pagination))
}

// ListByUser handles listing subscriptions for a specific user
// GET /api/v1/admin/users/:id/subscriptions
func (h *SubscriptionHandler) ListByUser(c *gin.Context) {
	userID, err := parseEntityID(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	subscriptions, err := h.subscriptionService.ListUserSubscriptions(c.Request.Context(), userID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]dto.AdminUserSubscription, 0, len(subscriptions))
	for i := range subscriptions {
		out = append(out, *dto.UserSubscriptionFromServiceAdmin(&subscriptions[i]))
	}
	response.Success(c, out)
}

// Helper function to get admin ID from context
func getAdminIDFromContext(c *gin.Context) string {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		return ""
	}
	return subject.UserID
}
