package model

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/AsukaCC/EasySub2api/internal/domain"
)

// AccountModelRule is a reusable model-routing configuration scoped by
// platform and, optionally, normalized subscription tier.
type AccountModelRule struct {
	ID                string                     `json:"id"`
	Name              string                     `json:"name"`
	Description       *string                    `json:"description"`
	Platform          string                     `json:"platform"`
	SubscriptionTier  *string                    `json:"subscription_tier"`
	ModelRoutes       []domain.AccountModelRoute `json:"model_routes"`
	BoundAccountCount int                        `json:"bound_account_count,omitempty"`
	CreatedAt         time.Time                  `json:"created_at"`
	UpdatedAt         time.Time                  `json:"updated_at"`
}

// Validate normalizes and validates a rule before persistence.
func (r *AccountModelRule) Validate() error {
	if r == nil {
		return &ValidationError{Field: "rule", Message: "rule is required"}
	}
	r.Name = strings.TrimSpace(r.Name)
	r.Platform = strings.TrimSpace(strings.ToLower(r.Platform))
	if r.SubscriptionTier != nil {
		tier := NormalizeSubscriptionTier(*r.SubscriptionTier)
		if tier == "" {
			r.SubscriptionTier = nil
		} else {
			r.SubscriptionTier = &tier
		}
	}
	if r.Description != nil {
		description := strings.TrimSpace(*r.Description)
		if description == "" {
			r.Description = nil
		} else {
			r.Description = &description
		}
	}
	if r.Name == "" {
		return &ValidationError{Field: "name", Message: "name is required"}
	}
	if utf8.RuneCountInString(r.Name) > 100 {
		return &ValidationError{Field: "name", Message: "name must be at most 100 characters"}
	}
	if !domain.IsAccountPlatform(r.Platform) {
		return &ValidationError{Field: "platform", Message: "unsupported account platform"}
	}
	normalizedRoutes := make([]domain.AccountModelRoute, 0, len(r.ModelRoutes))
	seenSources := make(map[string]struct{}, len(r.ModelRoutes))
	for _, rawRoute := range r.ModelRoutes {
		from := strings.TrimSpace(rawRoute.RequestModel)
		to := strings.TrimSpace(rawRoute.UpstreamModel)
		if from == "" || to == "" {
			return &ValidationError{Field: "model_routes", Message: "request_model and upstream_model are required"}
		}
		if !validModelWildcard(from) {
			return &ValidationError{Field: "model_routes", Message: "request model wildcard must appear once at the end"}
		}
		if strings.Contains(to, "*") {
			return &ValidationError{Field: "model_routes", Message: "upstream model cannot contain a wildcard"}
		}
		if _, exists := seenSources[from]; exists {
			return &ValidationError{Field: "model_routes", Message: "duplicate request model"}
		}
		seenSources[from] = struct{}{}
		effort := ""
		if strings.TrimSpace(rawRoute.ReasoningEffort) != "" {
			if r.Platform != domain.PlatformOpenAI {
				return &ValidationError{Field: "model_routes", Message: "reasoning effort is only supported for OpenAI rules"}
			}
			effort = normalizeRuleReasoningEffort(rawRoute.ReasoningEffort)
			if effort == "" {
				return &ValidationError{Field: "model_routes", Message: "unsupported reasoning effort"}
			}
		}
		normalizedRoutes = append(normalizedRoutes, domain.AccountModelRoute{RequestModel: from, UpstreamModel: to, ReasoningEffort: effort})
	}
	r.ModelRoutes = normalizedRoutes
	return nil
}

// NormalizeSubscriptionTier produces the stable key used by account filters
// and model-rule scope matching.
func NormalizeSubscriptionTier(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}
	var builder strings.Builder
	lastUnderscore := false
	for _, ch := range value {
		if ch == ' ' || ch == '-' || ch == '_' || ch == '/' {
			if builder.Len() > 0 && !lastUnderscore {
				_ = builder.WriteByte('_')
				lastUnderscore = true
			}
			continue
		}
		_, _ = builder.WriteRune(ch)
		lastUnderscore = false
	}
	return strings.Trim(builder.String(), "_")
}

func normalizeRuleReasoningEffort(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.NewReplacer("-", "", "_", "", " ", "").Replace(normalized)
	switch normalized {
	case "minimal", "low", "medium", "high", "max":
		return normalized
	case "xhigh", "extrahigh":
		return "xhigh"
	default:
		return ""
	}
}

func validModelWildcard(value string) bool {
	star := strings.IndexByte(value, '*')
	if star < 0 {
		return true
	}
	return star == len(value)-1 && strings.LastIndexByte(value, '*') == star
}
