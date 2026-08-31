package handler

import (
	"github.com/AsukaCC/EasySub2api/internal/handler/dto"
	"github.com/AsukaCC/EasySub2api/internal/pkg/response"
	middleware2 "github.com/AsukaCC/EasySub2api/internal/server/middleware"
	"github.com/AsukaCC/EasySub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type activatePendingSubscriptionRequest struct {
	ConfirmForfeitCurrent bool `json:"confirm_forfeit_current"`
}

// SubscriptionSummaryItem represents a subscription item in summary
type SubscriptionSummaryItem struct {
	ID                 string  `json:"id"`
	GroupID            string  `json:"group_id"`
	GroupName          string  `json:"group_name"`
	Status             string  `json:"status"`
	DailyUsedUSD       float64 `json:"daily_used_usd,omitempty"`
	DailyUsedPoints    float64 `json:"daily_used_points,omitempty"`
	DailyLimitUSD      float64 `json:"daily_limit_usd,omitempty"`
	DailyLimitPoints   float64 `json:"daily_limit_points,omitempty"`
	WeeklyUsedUSD      float64 `json:"weekly_used_usd,omitempty"`
	WeeklyUsedPoints   float64 `json:"weekly_used_points,omitempty"`
	WeeklyLimitUSD     float64 `json:"weekly_limit_usd,omitempty"`
	WeeklyLimitPoints  float64 `json:"weekly_limit_points,omitempty"`
	MonthlyUsedUSD     float64 `json:"monthly_used_usd,omitempty"`
	MonthlyUsedPoints  float64 `json:"monthly_used_points,omitempty"`
	MonthlyLimitUSD    float64 `json:"monthly_limit_usd,omitempty"`
	MonthlyLimitPoints float64 `json:"monthly_limit_points,omitempty"`
	ExpiresAt          *string `json:"expires_at,omitempty"`
}

// SubscriptionProgressInfo represents subscription with progress info
type SubscriptionProgressInfo struct {
	Subscription *dto.UserSubscription         `json:"subscription"`
	Progress     *service.SubscriptionProgress `json:"progress"`
}

// SubscriptionHandler handles user subscription operations
type SubscriptionHandler struct {
	subscriptionService *service.SubscriptionService
}

// NewSubscriptionHandler creates a new user subscription handler
func NewSubscriptionHandler(subscriptionService *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionService: subscriptionService,
	}
}

// List handles listing current user's subscriptions
// GET /api/v1/subscriptions
func (h *SubscriptionHandler) List(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	subscriptions, err := h.subscriptionService.ListUserSubscriptions(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]dto.UserSubscription, 0, len(subscriptions))
	for i := range subscriptions {
		out = append(out, *dto.UserSubscriptionFromService(&subscriptions[i]))
	}
	response.Success(c, out)
}

// GetActive handles getting current user's active subscriptions
// GET /api/v1/subscriptions/active
func (h *SubscriptionHandler) GetActive(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	subscriptions, err := h.subscriptionService.ListActiveUserSubscriptions(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]dto.UserSubscription, 0, len(subscriptions))
	for i := range subscriptions {
		out = append(out, *dto.UserSubscriptionFromService(&subscriptions[i]))
	}
	response.Success(c, out)
}

// ListPending returns subscriptions waiting for the current platform term to end.
func (h *SubscriptionHandler) ListPending(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	items, err := h.subscriptionService.ListPendingSubscriptions(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
}

// ActivatePendingNow forfeits the current platform subscription and activates
// the selected pending entitlement. Explicit confirmation is required.
func (h *SubscriptionHandler) ActivatePendingNow(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	var req activatePendingSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	result, err := h.subscriptionService.ActivatePendingNow(c.Request.Context(), subject.UserID, c.Param("id"), req.ConfirmForfeitCurrent)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

// GetProgress handles getting subscription progress for current user
// GET /api/v1/subscriptions/progress
func (h *SubscriptionHandler) GetProgress(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	// Get all active subscriptions with progress
	subscriptions, err := h.subscriptionService.ListActiveUserSubscriptions(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	result := make([]SubscriptionProgressInfo, 0, len(subscriptions))
	for i := range subscriptions {
		sub := &subscriptions[i]
		progress, err := h.subscriptionService.GetSubscriptionProgress(c.Request.Context(), sub.ID)
		if err != nil {
			// Skip subscriptions with errors
			continue
		}
		result = append(result, SubscriptionProgressInfo{
			Subscription: dto.UserSubscriptionFromService(sub),
			Progress:     progress,
		})
	}

	response.Success(c, result)
}

// GetSummary handles getting a summary of current user's subscription status
// GET /api/v1/subscriptions/summary
func (h *SubscriptionHandler) GetSummary(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	// Get all active subscriptions
	subscriptions, err := h.subscriptionService.ListActiveUserSubscriptions(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	pending, err := h.subscriptionService.ListPendingSubscriptions(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	var totalUsed float64
	items := make([]SubscriptionSummaryItem, 0, len(subscriptions))

	for _, sub := range subscriptions {
		item := SubscriptionSummaryItem{
			ID:                sub.ID,
			GroupID:           sub.GroupID,
			Status:            sub.Status,
			DailyUsedUSD:      sub.DailyUsageUSD,
			DailyUsedPoints:   sub.DailyUsageUSD,
			WeeklyUsedUSD:     sub.WeeklyUsageUSD,
			WeeklyUsedPoints:  sub.WeeklyUsageUSD,
			MonthlyUsedUSD:    sub.MonthlyUsageUSD,
			MonthlyUsedPoints: sub.MonthlyUsageUSD,
		}

		// Add group info if preloaded
		if sub.Group != nil {
			item.GroupName = sub.Group.Name
			if sub.Group.DailyLimitUSD != nil {
				item.DailyLimitUSD = *sub.Group.DailyLimitUSD
				item.DailyLimitPoints = *sub.Group.DailyLimitUSD
			}
			if sub.Group.WeeklyLimitUSD != nil {
				item.WeeklyLimitUSD = *sub.Group.WeeklyLimitUSD
				item.WeeklyLimitPoints = *sub.Group.WeeklyLimitUSD
			}
			if sub.Group.MonthlyLimitUSD != nil {
				item.MonthlyLimitUSD = *sub.Group.MonthlyLimitUSD
				item.MonthlyLimitPoints = *sub.Group.MonthlyLimitUSD
			}
		}

		// Format expiration time
		if !sub.ExpiresAt.IsZero() {
			formatted := sub.ExpiresAt.Format("2006-01-02T15:04:05Z07:00")
			item.ExpiresAt = &formatted
		}

		// Track total usage (use monthly as the most comprehensive)
		totalUsed += sub.MonthlyUsageUSD

		items = append(items, item)
	}

	summary := struct {
		ActiveCount     int                           `json:"active_count"`
		TotalUsedUSD    float64                       `json:"total_used_usd"`
		TotalUsedPoints float64                       `json:"total_used_points"`
		Subscriptions   []SubscriptionSummaryItem     `json:"subscriptions"`
		PendingCount    int                           `json:"pending_count"`
		Pending         []service.PendingSubscription `json:"pending"`
	}{
		ActiveCount:     len(subscriptions),
		TotalUsedUSD:    totalUsed,
		TotalUsedPoints: totalUsed,
		Subscriptions:   items,
		PendingCount:    len(pending),
		Pending:         pending,
	}

	response.Success(c, summary)
}
