package admin

import (
	"bytes"
	"encoding/json"
	"sort"
	"strings"

	"github.com/AsukaCC/EasySub2api/internal/domain"
	"github.com/AsukaCC/EasySub2api/internal/model"
	"github.com/AsukaCC/EasySub2api/internal/pkg/response"
	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type AccountModelRuleHandler struct {
	service *service.AccountModelRuleService
}

func NewAccountModelRuleHandler(svc *service.AccountModelRuleService) *AccountModelRuleHandler {
	return &AccountModelRuleHandler{service: svc}
}

type createAccountModelRuleRequest struct {
	Name             string                      `json:"name" binding:"required"`
	Description      *string                     `json:"description"`
	Platform         string                      `json:"platform" binding:"required"`
	SubscriptionTier *string                     `json:"subscription_tier"`
	ModelRoutes      *[]domain.AccountModelRoute `json:"model_routes"`
	Whitelist        []string                    `json:"whitelist"`
	Mapping          map[string]string           `json:"mapping"`
	ReasoningEfforts map[string]string           `json:"reasoning_efforts"`
}

type updateAccountModelRuleRequest struct {
	Name             *string                     `json:"name"`
	Description      json.RawMessage             `json:"description"`
	Platform         *string                     `json:"platform"`
	SubscriptionTier json.RawMessage             `json:"subscription_tier"`
	ModelRoutes      *[]domain.AccountModelRoute `json:"model_routes"`
	Whitelist        []string                    `json:"whitelist"`
	Mapping          map[string]string           `json:"mapping"`
	ReasoningEfforts map[string]string           `json:"reasoning_efforts"`
}

func (h *AccountModelRuleHandler) List(c *gin.Context) {
	rules, err := h.service.List(c.Request.Context(), c.Query("platform"), c.Query("subscription_tier"))
	if err != nil {
		if _, ok := err.(*model.ValidationError); ok {
			response.BadRequest(c, err.Error())
			return
		}
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, rules)
}

func (h *AccountModelRuleHandler) GetByID(c *gin.Context) {
	id, err := parseEntityID(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid rule ID")
		return
	}
	rule, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if rule == nil {
		response.NotFound(c, "Account model rule not found")
		return
	}
	response.Success(c, rule)
}

func (h *AccountModelRuleHandler) Create(c *gin.Context) {
	var req createAccountModelRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	routes := legacyAccountModelRoutes(req.ModelRoutes, req.Whitelist, req.Mapping, req.ReasoningEfforts)
	rule := &model.AccountModelRule{
		Name:             req.Name,
		Description:      req.Description,
		Platform:         req.Platform,
		SubscriptionTier: req.SubscriptionTier,
		ModelRoutes:      routes,
	}
	created, err := h.service.Create(c.Request.Context(), rule)
	if err != nil {
		if _, ok := err.(*model.ValidationError); ok {
			response.BadRequest(c, err.Error())
			return
		}
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, created)
}

func (h *AccountModelRuleHandler) Update(c *gin.Context) {
	id, err := parseEntityID(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid rule ID")
		return
	}
	var req updateAccountModelRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	existing, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if existing == nil {
		response.NotFound(c, "Account model rule not found")
		return
	}
	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Description != nil {
		if bytes.Equal(bytes.TrimSpace(req.Description), []byte("null")) {
			existing.Description = nil
		} else {
			var description string
			if err := json.Unmarshal(req.Description, &description); err != nil {
				response.BadRequest(c, "Invalid description")
				return
			}
			existing.Description = &description
		}
	}
	if req.Platform != nil {
		existing.Platform = *req.Platform
	}
	if req.SubscriptionTier != nil {
		if bytes.Equal(bytes.TrimSpace(req.SubscriptionTier), []byte("null")) {
			existing.SubscriptionTier = nil
		} else {
			var subscriptionTier string
			if err := json.Unmarshal(req.SubscriptionTier, &subscriptionTier); err != nil {
				response.BadRequest(c, "Invalid subscription tier")
				return
			}
			existing.SubscriptionTier = &subscriptionTier
		}
	}
	if req.ModelRoutes != nil || req.Whitelist != nil || req.Mapping != nil || req.ReasoningEfforts != nil {
		existing.ModelRoutes = legacyAccountModelRoutes(req.ModelRoutes, req.Whitelist, req.Mapping, req.ReasoningEfforts)
	}
	updated, err := h.service.Update(c.Request.Context(), existing)
	if err != nil {
		if _, ok := err.(*model.ValidationError); ok {
			response.BadRequest(c, err.Error())
			return
		}
		response.ErrorFrom(c, err)
		return
	}
	if updated == nil {
		response.NotFound(c, "Account model rule not found")
		return
	}
	response.Success(c, updated)
}

func (h *AccountModelRuleHandler) Delete(c *gin.Context) {
	id, err := parseEntityID(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid rule ID")
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "Account model rule deleted successfully"})
}

func legacyAccountModelRoutes(
	provided *[]domain.AccountModelRoute,
	whitelist []string,
	mapping map[string]string,
	reasoningEfforts map[string]string,
) []domain.AccountModelRoute {
	if provided != nil {
		return append([]domain.AccountModelRoute(nil), (*provided)...)
	}
	routes := make(map[string]domain.AccountModelRoute, len(mapping)+len(whitelist))
	for requestModel, upstreamModel := range mapping {
		requestModel = strings.TrimSpace(requestModel)
		if requestModel == "" {
			continue
		}
		routes[requestModel] = domain.AccountModelRoute{
			RequestModel:    requestModel,
			UpstreamModel:   strings.TrimSpace(upstreamModel),
			ReasoningEffort: reasoningEfforts[requestModel],
		}
	}
	for _, requestModel := range whitelist {
		requestModel = strings.TrimSpace(requestModel)
		if requestModel == "" {
			continue
		}
		if _, exists := routes[requestModel]; !exists {
			routes[requestModel] = domain.AccountModelRoute{
				RequestModel:    requestModel,
				UpstreamModel:   requestModel,
				ReasoningEffort: reasoningEfforts[requestModel],
			}
		}
	}
	keys := make([]string, 0, len(routes))
	for key := range routes {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]domain.AccountModelRoute, 0, len(keys))
	for _, key := range keys {
		result = append(result, routes[key])
	}
	return result
}
