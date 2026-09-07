package handler

import (
	"net/http"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/config"
	"github.com/AsukaCC/EasySub2api/internal/pkg/timezone"
	middleware2 "github.com/AsukaCC/EasySub2api/internal/server/middleware"
	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/gin-gonic/gin"
)

const keyBillingInfoSchemaVersion = 1

type keyBillingInfoResponse struct {
	Object                  string    `json:"object"`
	SchemaVersion           int       `json:"schema_version"`
	BillingScope            string    `json:"billing_scope"`
	GroupRateMultiplier     float64   `json:"group_rate_multiplier"`
	UserRateMultiplier      *float64  `json:"user_rate_multiplier,omitempty"`
	UserLevelMultiplier     *float64  `json:"user_level_multiplier,omitempty"`
	GroupRuleMultiplier     *float64  `json:"group_rule_multiplier,omitempty"`
	ResolvedRateMultiplier  float64   `json:"resolved_rate_multiplier"`
	EffectiveBaseMultiplier float64   `json:"effective_base_multiplier"`
	EffectiveSource         string    `json:"effective_source,omitempty"`
	PeakRateEnabled         bool      `json:"peak_rate_enabled"`
	PeakStart               *string   `json:"peak_start,omitempty"`
	PeakEnd                 *string   `json:"peak_end,omitempty"`
	PeakRateMultiplier      *float64  `json:"peak_rate_multiplier,omitempty"`
	AppliedPeakMultiplier   *float64  `json:"applied_peak_multiplier,omitempty"`
	EffectiveRateMultiplier float64   `json:"effective_rate_multiplier"`
	Timezone                *string   `json:"timezone,omitempty"`
	ObservedAt              time.Time `json:"observed_at"`
}

// KeyBillingInfo returns the token billing multiplier effective for the authenticated API key.
// GET /v1/easysub2api/billing
//
// The legacy /v1/sub2api/billing route remains available for one migration
// release. It keeps its original object discriminator so older clients can
// complete the transition without silently changing their wire contract.
func (h *GatewayHandler) KeyBillingInfo(c *gin.Context) {
	apiKey, ok := middleware2.GetAPIKeyFromContext(c)
	if !ok {
		h.errorResponse(c, http.StatusUnauthorized, "authentication_error", "Invalid API key")
		return
	}
	if h.cfg != nil && h.cfg.RunMode == config.RunModeSimple {
		h.errorResponse(c, http.StatusNotFound, "not_found_error", "Billing information is not supported in simple mode")
		return
	}
	if apiKey.GroupID == nil {
		h.errorResponse(c, http.StatusForbidden, "permission_error", "API key is not assigned to a group")
		return
	}
	if apiKey.Group == nil {
		h.errorResponse(c, http.StatusInternalServerError, "api_error", "Billing information is unavailable")
		return
	}

	c.Header("Cache-Control", "no-store")
	now := timezone.Now()
	response, ok := h.buildKeyBillingResponse(c, apiKey, now)
	if !ok {
		h.errorResponse(c, http.StatusInternalServerError, "api_error", "Billing information is unavailable")
		return
	}
	if c.Request.URL.Path == "/v1/sub2api/billing" {
		c.Header("Deprecation", "true")
		c.Header("Link", "</v1/easysub2api/billing>; rel=\"successor-version\"")
		response.Object = "sub2api.key_billing"
	}
	c.JSON(http.StatusOK, response)
}

func (h *GatewayHandler) buildKeyBillingResponse(c *gin.Context, apiKey *service.APIKey, now time.Time) (keyBillingInfoResponse, bool) {
	if plan, ok := h.resolveKeyBillingPlan(c, apiKey, now); ok {
		return buildKeyBillingInfoWithPlan(apiKey, plan, now), true
	}
	resolvedRate, ok := h.resolveKeyBillingRate(c, apiKey)
	if !ok {
		return keyBillingInfoResponse{}, false
	}
	return buildKeyBillingInfo(apiKey, resolvedRate, now), true
}

func (h *GatewayHandler) resolveKeyBillingPlan(c *gin.Context, apiKey *service.APIKey, at time.Time) (*service.UserRatePlan, bool) {
	if h == nil || apiKey == nil || apiKey.Group == nil {
		return nil, false
	}
	ctx := c.Request.Context()
	switch apiKey.Group.Platform {
	case service.PlatformOpenAI, service.PlatformGrok:
		if h.openAIGatewayService == nil {
			return nil, false
		}
		return h.openAIGatewayService.ResolveUserRatePlan(ctx, apiKey.UserID, apiKey.Group, at)
	default:
		if h.gatewayService == nil {
			return nil, false
		}
		return h.gatewayService.ResolveUserRatePlan(ctx, apiKey.UserID, apiKey.Group, at)
	}
}

func (h *GatewayHandler) resolveKeyBillingRate(c *gin.Context, apiKey *service.APIKey) (float64, bool) {
	groupRate := apiKey.Group.RateMultiplier
	switch apiKey.Group.Platform {
	case service.PlatformOpenAI, service.PlatformGrok:
		if h.openAIGatewayService == nil {
			return 0, false
		}
		return h.openAIGatewayService.ResolveUserGroupRateMultiplier(c.Request.Context(), apiKey.UserID, *apiKey.GroupID, groupRate), true
	default:
		if h.gatewayService == nil {
			return 0, false
		}
		return h.gatewayService.ResolveUserGroupRateMultiplier(c.Request.Context(), apiKey.UserID, *apiKey.GroupID, groupRate), true
	}
}

func buildKeyBillingInfo(apiKey *service.APIKey, resolvedRate float64, now time.Time) keyBillingInfoResponse {
	groupRate := apiKey.Group.RateMultiplier
	var userRate *float64
	if resolvedRate != groupRate {
		userRate = &resolvedRate
	}
	appliedPeak := apiKey.Group.PeakMultiplierAt(now)

	response := keyBillingInfoResponse{
		Object:                  "easysub2api.key_billing",
		SchemaVersion:           keyBillingInfoSchemaVersion,
		BillingScope:            "token",
		GroupRateMultiplier:     groupRate,
		UserRateMultiplier:      userRate,
		ResolvedRateMultiplier:  resolvedRate,
		EffectiveBaseMultiplier: resolvedRate,
		PeakRateEnabled:         apiKey.Group.PeakRateEnabled,
		EffectiveRateMultiplier: resolvedRate * appliedPeak,
		ObservedAt:              now.UTC(),
	}
	if apiKey.Group.PeakRateEnabled {
		response.PeakStart = &apiKey.Group.PeakStart
		response.PeakEnd = &apiKey.Group.PeakEnd
		response.PeakRateMultiplier = &apiKey.Group.PeakRateMultiplier
		response.AppliedPeakMultiplier = &appliedPeak
		tz := timezone.Location().String()
		response.Timezone = &tz
	}
	return response
}

func buildKeyBillingInfoWithPlan(apiKey *service.APIKey, plan *service.UserRatePlan, now time.Time) keyBillingInfoResponse {
	if plan == nil {
		return buildKeyBillingInfo(apiKey, apiKey.Group.RateMultiplier, now)
	}
	response := buildKeyBillingInfo(apiKey, plan.EffectiveBaseMultiplier, now)
	// Keep the legacy field's meaning stable: it is the configured group base
	// rate, while GroupRuleMultiplier explains the complete group-side minimum.
	response.GroupRateMultiplier = apiKey.Group.RateMultiplier
	response.UserRateMultiplier = cloneBillingFloat(plan.UserLevelMultiplier)
	response.UserLevelMultiplier = cloneBillingFloat(plan.UserLevelMultiplier)
	response.GroupRuleMultiplier = cloneBillingFloat(plan.GroupRuleMultiplier)
	response.ResolvedRateMultiplier = plan.EffectiveBaseMultiplier
	response.EffectiveBaseMultiplier = plan.EffectiveBaseMultiplier
	response.EffectiveSource = plan.EffectiveSource
	response.EffectiveRateMultiplier = plan.EffectiveMultiplier
	if apiKey.Group.PeakRateEnabled {
		response.AppliedPeakMultiplier = &plan.PeakMultiplier
	}
	return response
}

func cloneBillingFloat(value *float64) *float64 {
	if value == nil {
		return nil
	}
	out := *value
	return &out
}
