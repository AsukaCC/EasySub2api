package service

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	infraerrors "github.com/AsukaCC/EasySub2api/internal/pkg/errors"
	"github.com/google/uuid"
)

var (
	ErrUserLevelRuleNotFound        = infraerrors.NotFound("USER_LEVEL_RULE_NOT_FOUND", "user level rule not found")
	ErrUserLevelRuleInvalid         = infraerrors.BadRequest("USER_LEVEL_RULE_INVALID", "invalid user level rule")
	ErrUserLevelRuleReferenced      = infraerrors.Conflict("USER_LEVEL_RULE_REFERENCED", "user level rule is still referenced")
	ErrUserLevelRulesUnavailable    = infraerrors.InternalServer("USER_LEVEL_RULES_UNAVAILABLE", "user level rule service unavailable")
	ErrUserLevelRuleAssignmentInput = infraerrors.BadRequest("USER_LEVEL_RULE_ASSIGNMENT_INVALID", "invalid user level rule assignment")
)

// UserLevelRuleTierInput is the editable representation used by the admin API.
// An empty ID creates a new UUIDv7 tier; existing IDs preserve references from
// group and announcement JSON.
type UserLevelRuleTierInput struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	SortOrder         int      `json:"sort_order"`
	MinSpend          float64  `json:"min_spend"`
	DefaultMultiplier *float64 `json:"default_multiplier"`
}

type CreateUserLevelRuleInput struct {
	Name       string `json:"name"`
	WindowDays int    `json:"window_days"`
}

type UpdateUserLevelRuleInput struct {
	Name       string                   `json:"name"`
	WindowDays int                      `json:"window_days"`
	Enabled    *bool                    `json:"enabled"`
	Tiers      []UserLevelRuleTierInput `json:"tiers"`
}

type BatchUserLevelRuleAssignmentInput struct {
	UserIDs   []string `json:"user_ids"`
	RuleIDs   []string `json:"rule_ids"`
	Operation string   `json:"operation"`
}

func (s *UserLevelService) levelRulesRepo() UserLevelRulesRepository {
	if s == nil {
		return nil
	}
	return s.rulesRepo
}

func (s *UserLevelService) ListLevelRules(ctx context.Context) ([]UserLevelRule, error) {
	repo := s.levelRulesRepo()
	if repo == nil {
		return nil, ErrUserLevelRulesUnavailable
	}
	rules, err := repo.ListLevelRules(ctx)
	if err != nil {
		return nil, err
	}
	return rules, nil
}

func (s *UserLevelService) GetLevelRule(ctx context.Context, ruleID string) (*UserLevelRule, error) {
	repo := s.levelRulesRepo()
	if repo == nil {
		return nil, ErrUserLevelRulesUnavailable
	}
	ruleID, err := normalizeUUID(ruleID, ErrUserLevelRuleInvalid)
	if err != nil {
		return nil, err
	}
	rule, err := repo.GetLevelRule(ctx, ruleID)
	if err != nil {
		return nil, err
	}
	if rule == nil {
		return nil, ErrUserLevelRuleNotFound
	}
	return rule, nil
}

// NormalizeDefaultLevelRuleIDs validates the settings reference list while
// preserving rule order. Disabled rules are valid defaults but are skipped by
// assignment until an administrator enables them again.
func (s *UserLevelService) NormalizeDefaultLevelRuleIDs(ctx context.Context, ruleIDs []string) ([]string, error) {
	normalized, err := normalizeUUIDs(ruleIDs, ErrUserLevelRuleInvalid)
	if err != nil {
		return nil, err
	}
	for _, ruleID := range normalized {
		if _, err := s.GetLevelRule(ctx, ruleID); err != nil {
			return nil, err
		}
	}
	return normalized, nil
}

func (s *UserLevelService) CreateLevelRule(ctx context.Context, input CreateUserLevelRuleInput) (*UserLevelRule, error) {
	repo := s.levelRulesRepo()
	if repo == nil {
		return nil, ErrUserLevelRulesUnavailable
	}
	name, err := validateRuleName(input.Name)
	if err != nil {
		return nil, err
	}
	windowDays := input.WindowDays
	if windowDays == 0 {
		windowDays = 7
	}
	if !validLevelRuleWindow(windowDays) {
		return nil, invalidLevelRule("window_days must be 7, 14, or 30")
	}

	rule := &UserLevelRule{
		ID:         uuid.Must(uuid.NewV7()).String(),
		Name:       name,
		WindowDays: windowDays,
		Enabled:    true,
		Tiers: []UserLevelTier{{
			ID:                uuid.Must(uuid.NewV7()).String(),
			Name:              "Base",
			SortOrder:         0,
			MinSpend:          0,
			RuleID:            "",
			DefaultMultiplier: nil,
		}},
	}
	if err := repo.CreateLevelRule(ctx, rule); err != nil {
		return nil, err
	}
	s.invalidateProfiles()
	return rule, nil
}

func (s *UserLevelService) UpdateLevelRule(ctx context.Context, ruleID string, input UpdateUserLevelRuleInput) (*UserLevelRule, error) {
	repo := s.levelRulesRepo()
	if repo == nil {
		return nil, ErrUserLevelRulesUnavailable
	}
	ruleID, err := normalizeUUID(ruleID, ErrUserLevelRuleInvalid)
	if err != nil {
		return nil, err
	}
	existing, err := repo.GetLevelRule(ctx, ruleID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrUserLevelRuleNotFound
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		name = existing.Name
	}
	name, err = validateRuleName(name)
	if err != nil {
		return nil, err
	}
	windowDays := input.WindowDays
	if windowDays == 0 {
		windowDays = existing.WindowDays
	}
	if !validLevelRuleWindow(windowDays) {
		return nil, invalidLevelRule("window_days must be 7, 14, or 30")
	}

	tiers := existing.Tiers
	if input.Tiers != nil {
		tiers, err = normalizeLevelRuleTiers(ruleID, input.Tiers)
		if err != nil {
			return nil, err
		}
		if err := preserveBaseLevelTierID(existing.Tiers, tiers); err != nil {
			return nil, err
		}
	}
	enabled := existing.Enabled
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	updated := &UserLevelRule{
		ID: existing.ID, Name: name, WindowDays: windowDays, Enabled: enabled,
		CreatedAt: existing.CreatedAt, UpdatedAt: existing.UpdatedAt, Tiers: tiers,
	}
	if err := repo.UpdateLevelRule(ctx, updated); err != nil {
		return nil, err
	}
	s.invalidateProfiles()
	return updated, nil
}

func (s *UserLevelService) DeleteLevelRule(ctx context.Context, ruleID string) error {
	repo := s.levelRulesRepo()
	if repo == nil {
		return ErrUserLevelRulesUnavailable
	}
	ruleID, err := normalizeUUID(ruleID, ErrUserLevelRuleInvalid)
	if err != nil {
		return err
	}
	if _, err := repo.GetLevelRule(ctx, ruleID); err != nil {
		return err
	}
	refs, err := repo.GetLevelRuleReferenceCount(ctx, ruleID)
	if err != nil {
		return err
	}
	if refs > 0 {
		return ErrUserLevelRuleReferenced
	}
	if err := repo.DeleteLevelRule(ctx, ruleID); err != nil {
		return err
	}
	s.invalidateProfiles()
	return nil
}

func (s *UserLevelService) GetUserLevelRules(ctx context.Context, userID string) ([]UserLevelRule, error) {
	repo := s.levelRulesRepo()
	if repo == nil {
		return nil, ErrUserLevelRulesUnavailable
	}
	userID, err := normalizeUUID(userID, ErrUserLevelRuleInvalid)
	if err != nil {
		return nil, err
	}
	return repo.ListUserLevelRules(ctx, userID)
}

func (s *UserLevelService) ReplaceUserLevelRules(ctx context.Context, userID string, ruleIDs []string) error {
	repo := s.levelRulesRepo()
	if repo == nil {
		return ErrUserLevelRulesUnavailable
	}
	userID, err := normalizeUUID(userID, ErrUserLevelRuleAssignmentInput)
	if err != nil {
		return err
	}
	ruleIDs, err = normalizeUUIDs(ruleIDs, ErrUserLevelRuleAssignmentInput)
	if err != nil {
		return err
	}
	if err := repo.ReplaceUserLevelRules(ctx, userID, ruleIDs); err != nil {
		return err
	}
	s.invalidateProfiles()
	return nil
}

func (s *UserLevelService) BatchAssignUserLevelRules(ctx context.Context, input BatchUserLevelRuleAssignmentInput) (int64, error) {
	repo := s.levelRulesRepo()
	if repo == nil {
		return 0, ErrUserLevelRulesUnavailable
	}
	users, err := normalizeUUIDs(input.UserIDs, ErrUserLevelRuleAssignmentInput)
	if err != nil || len(users) == 0 {
		return 0, invalidAssignment("user_ids must contain at least one UUID")
	}
	rules, err := normalizeUUIDs(input.RuleIDs, ErrUserLevelRuleAssignmentInput)
	if err != nil {
		return 0, err
	}
	operation := strings.ToLower(strings.TrimSpace(input.Operation))
	if operation != UserLevelRuleAssignmentAdd && operation != UserLevelRuleAssignmentRemove && operation != UserLevelRuleAssignmentReplace {
		return 0, invalidAssignment("operation must be add, remove, or replace")
	}
	if operation == UserLevelRuleAssignmentAdd && len(rules) == 0 {
		return 0, invalidAssignment("add operation requires at least one rule")
	}
	affected, err := repo.BatchAssignUserLevelRules(ctx, users, rules, operation)
	if err != nil {
		return 0, err
	}
	s.invalidateProfiles()
	return affected, nil
}

func (s *UserLevelService) ListLevelRuleMembers(ctx context.Context, ruleID string, page, pageSize int) ([]User, int64, error) {
	repo := s.levelRulesRepo()
	if repo == nil {
		return nil, 0, ErrUserLevelRulesUnavailable
	}
	ruleID, err := normalizeUUID(ruleID, ErrUserLevelRuleInvalid)
	if err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	return repo.ListLevelRuleMembers(ctx, ruleID, page, pageSize)
}

func normalizeLevelRuleTiers(ruleID string, input []UserLevelRuleTierInput) ([]UserLevelTier, error) {
	if len(input) == 0 {
		return nil, invalidLevelRule("at least one tier is required")
	}
	tiers := make([]UserLevelTier, 0, len(input))
	seenIDs := make(map[string]struct{}, len(input))
	seenOrders := make(map[int]struct{}, len(input))
	for _, item := range input {
		name := strings.TrimSpace(item.Name)
		if name == "" || len([]rune(name)) > 100 {
			return nil, invalidLevelRule("tier name is required and must be at most 100 characters")
		}
		if item.SortOrder < 0 {
			return nil, invalidLevelRule("tier sort_order must be nonnegative")
		}
		if _, exists := seenOrders[item.SortOrder]; exists {
			return nil, invalidLevelRule("tier sort_order must be unique")
		}
		seenOrders[item.SortOrder] = struct{}{}
		id := strings.TrimSpace(item.ID)
		if id == "" {
			id = uuid.Must(uuid.NewV7()).String()
		} else {
			parsed, err := uuid.Parse(id)
			if err != nil {
				return nil, invalidLevelRule("tier id must be a UUID")
			}
			id = parsed.String()
		}
		if _, exists := seenIDs[id]; exists {
			return nil, invalidLevelRule("tier id must be unique")
		}
		seenIDs[id] = struct{}{}
		if math.IsNaN(item.MinSpend) || math.IsInf(item.MinSpend, 0) || item.MinSpend < 0 {
			return nil, invalidLevelRule("tier min_spend must be nonnegative")
		}
		var multiplier *float64
		if item.DefaultMultiplier != nil {
			if math.IsNaN(*item.DefaultMultiplier) || math.IsInf(*item.DefaultMultiplier, 0) || *item.DefaultMultiplier < 0.01 || *item.DefaultMultiplier > 100 {
				return nil, invalidLevelRule("tier default_multiplier must be between 0.01 and 100")
			}
			value := QuantizeUsageBillingAmount(*item.DefaultMultiplier)
			multiplier = &value
		}
		tiers = append(tiers, UserLevelTier{
			ID: id, RuleID: ruleID, Name: name, SortOrder: item.SortOrder,
			MinSpend: QuantizeUsageBillingAmount(item.MinSpend), DefaultMultiplier: multiplier,
		})
	}
	sort.SliceStable(tiers, func(i, j int) bool { return tiers[i].SortOrder < tiers[j].SortOrder })
	if tiers[0].SortOrder != 0 || tiers[0].MinSpend != 0 {
		return nil, invalidLevelRule("the base tier must have sort_order 0 and min_spend 0")
	}
	for i := 1; i < len(tiers); i++ {
		if tiers[i].MinSpend <= tiers[i-1].MinSpend {
			return nil, invalidLevelRule("tier min_spend must be strictly increasing")
		}
	}
	return tiers, nil
}

// preserveBaseLevelTierID keeps the existing base-tier UUID when the client
// omits it (create-then-update) or sends a replacement ID. Group and
// announcement JSON store this UUID, so it must stay stable.
func preserveBaseLevelTierID(existing, incoming []UserLevelTier) error {
	baseTierID := ""
	for _, tier := range existing {
		if tier.SortOrder == 0 {
			baseTierID = tier.ID
			break
		}
	}
	if baseTierID == "" {
		return nil
	}
	for i, tier := range incoming {
		if tier.SortOrder == 0 && tier.MinSpend == 0 {
			incoming[i].ID = baseTierID
			return nil
		}
	}
	return invalidLevelRule("the base tier cannot be deleted or moved")
}

func validateRuleName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" || len([]rune(name)) > 100 {
		return "", invalidLevelRule("name is required and must be at most 100 characters")
	}
	return name, nil
}

func validLevelRuleWindow(days int) bool { return days == 7 || days == 14 || days == 30 }

func normalizeUUID(raw string, fallback *infraerrors.ApplicationError) (string, error) {
	parsed, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", fallback
	}
	return parsed.String(), nil
}

func normalizeUUIDs(values []string, fallback *infraerrors.ApplicationError) ([]string, error) {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, raw := range values {
		id, err := normalizeUUID(raw, fallback)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out, nil
}

func invalidLevelRule(message string) error {
	return infraerrors.BadRequest("USER_LEVEL_RULE_INVALID", fmt.Sprintf("invalid user level rule: %s", message))
}

func invalidAssignment(message string) error {
	return infraerrors.BadRequest("USER_LEVEL_RULE_ASSIGNMENT_INVALID", fmt.Sprintf("invalid user level rule assignment: %s", message))
}
