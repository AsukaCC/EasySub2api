package model

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/AsukaCC/EasySub2api/internal/domain"
)

// AccountModelRule is a reusable model restriction template. Mapping keys and
// values use the same semantics as account credentials.model_mapping, while
// whitelist contains exact request model names that should be allowed.
type AccountModelRule struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description *string           `json:"description"`
	Platform    string            `json:"platform"`
	Whitelist   []string          `json:"whitelist"`
	Mapping     map[string]string `json:"mapping"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// Validate normalizes and validates a rule before persistence.
func (r *AccountModelRule) Validate() error {
	if r == nil {
		return &ValidationError{Field: "rule", Message: "rule is required"}
	}
	r.Name = strings.TrimSpace(r.Name)
	r.Platform = strings.TrimSpace(strings.ToLower(r.Platform))
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
	normalized := make(map[string]string, len(r.Mapping))
	for rawFrom, rawTo := range r.Mapping {
		from := strings.TrimSpace(rawFrom)
		to := strings.TrimSpace(rawTo)
		if from == "" || to == "" {
			continue
		}
		if !validModelWildcard(from) {
			return &ValidationError{Field: "mapping", Message: "source model wildcard must appear once at the end"}
		}
		if strings.Contains(to, "*") {
			return &ValidationError{Field: "mapping", Message: "target model cannot contain a wildcard"}
		}
		if _, exists := normalized[from]; exists {
			return &ValidationError{Field: "mapping", Message: "duplicate source model"}
		}
		normalized[from] = to
	}
	normalizedWhitelist := make([]string, 0, len(r.Whitelist))
	seenWhitelist := make(map[string]struct{}, len(r.Whitelist))
	for _, rawModel := range r.Whitelist {
		modelName := strings.TrimSpace(rawModel)
		if modelName == "" {
			continue
		}
		if strings.Contains(modelName, "*") {
			return &ValidationError{Field: "whitelist", Message: "whitelist models cannot contain a wildcard"}
		}
		if _, exists := seenWhitelist[modelName]; exists {
			continue
		}
		seenWhitelist[modelName] = struct{}{}
		normalizedWhitelist = append(normalizedWhitelist, modelName)
	}
	if len(normalized) == 0 && len(normalizedWhitelist) == 0 {
		return &ValidationError{Field: "rule", Message: "at least one whitelist model or model mapping is required"}
	}
	r.Whitelist = normalizedWhitelist
	r.Mapping = normalized
	return nil
}

func validModelWildcard(value string) bool {
	star := strings.IndexByte(value, '*')
	if star < 0 {
		return true
	}
	return star == len(value)-1 && strings.LastIndexByte(value, '*') == star
}
