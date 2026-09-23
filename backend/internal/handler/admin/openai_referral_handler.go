package admin

import (
	"context"
	"net/http"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/pkg/response"
	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type openAIReferralService interface {
	QueryReferralEligibility(context.Context, string) (*service.OpenAIReferralEligibility, error)
	CacheReferralSnapshot(context.Context, string, *service.OpenAIReferralEligibility) error
	SendReferralInvite(context.Context, string, service.OpenAIReferralSendRequest) (*service.OpenAIReferralSendResult, error)
}

type openAIReferralRefreshResponse struct {
	Eligibility    *service.OpenAIReferralEligibility `json:"eligibility"`
	CachePersisted bool                               `json:"cache_persisted"`
}

type openAIReferralSendResponse struct {
	service.OpenAIReferralSendResult
	openAIReferralRefreshResponse
	RefreshFailed bool `json:"refresh_failed"`
}

func (h *OpenAIOAuthHandler) referralAccountID(c *gin.Context) (string, bool) {
	id, err := parseEntityID(c.Param("id"))
	if err != nil || id == "" {
		response.BadRequest(c, "Invalid account ID")
		return "", false
	}
	if h.referralService == nil {
		response.BadRequest(c, "OpenAI referral service is not enabled")
		return "", false
	}
	return id, true
}

// RefreshReferrals persists a display snapshot, hence POST and admin audit.
func (h *OpenAIOAuthHandler) RefreshReferrals(c *gin.Context) {
	id, ok := h.referralAccountID(c)
	if !ok {
		return
	}
	eligibility, err := h.referralService.QueryReferralEligibility(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if eligibility == nil {
		response.Error(c, http.StatusBadGateway, "Empty invitation eligibility response")
		return
	}
	cacheErr := h.referralService.CacheReferralSnapshot(c.Request.Context(), id, eligibility)
	response.Success(c, openAIReferralRefreshResponse{Eligibility: eligibility, CachePersisted: cacheErr == nil})
}

func (h *OpenAIOAuthHandler) SendReferralInvite(c *gin.Context) {
	id, ok := h.referralAccountID(c)
	if !ok {
		return
	}
	var input service.OpenAIReferralSendRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid invitation request")
		return
	}
	result, err := h.referralService.SendReferralInvite(c.Request.Context(), id, input)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if result == nil || !result.Sent {
		response.Error(c, http.StatusBadGateway, "Invitation outcome is unknown; check Codex before sending again")
		return
	}
	// The email is already sent. Refresh failure must not turn it into a failed
	// submission and encourage a duplicate send, even if the browser disconnects.
	baseCtx := context.WithoutCancel(c.Request.Context())
	refreshCtx, cancelRefresh := context.WithTimeout(baseCtx, 8*time.Second)
	eligibility, refreshErr := h.referralService.QueryReferralEligibility(refreshCtx, id)
	cancelRefresh()
	if refreshErr != nil {
		eligibility = nil
	}
	// Refresh may exhaust its entire deadline. Give persistence (including
	// invalidating a stale snapshot) a fresh, independent deadline.
	cacheCtx, cancelCache := context.WithTimeout(baseCtx, 3*time.Second)
	defer cancelCache()
	cacheErr := h.referralService.CacheReferralSnapshot(cacheCtx, id, eligibility)
	response.Success(c, openAIReferralSendResponse{
		OpenAIReferralSendResult:      *result,
		openAIReferralRefreshResponse: openAIReferralRefreshResponse{Eligibility: eligibility, CachePersisted: cacheErr == nil},
		RefreshFailed:                 eligibility == nil,
	})
}
