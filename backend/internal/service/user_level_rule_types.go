package service

import (
	"context"
	"time"
)

// UserLevelRule is an administrator-created spend rule profile. A rule owns
// one rolling window and any number of ordered tiers.
type UserLevelRule struct {
	ID                string          `json:"id"`
	Name              string          `json:"name"`
	WindowDays        int             `json:"window_days"`
	Enabled           bool            `json:"enabled"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	Tiers             []UserLevelTier `json:"tiers"`
	AssignedUserCount int64           `json:"assigned_user_count"`
	ReferenceCount    int64           `json:"reference_count"`
}

// UserLevelTier is ordered by SortOrder. The tier at SortOrder 0 is the
// immutable-threshold base tier; its name and multiplier remain editable.
type UserLevelTier struct {
	ID                string    `json:"id"`
	RuleID            string    `json:"rule_id"`
	Name              string    `json:"name"`
	SortOrder         int       `json:"sort_order"`
	MinSpend          float64   `json:"min_spend"`
	DefaultMultiplier *float64  `json:"default_multiplier,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// UserLevelRuleProfile is the calculated state of one assigned rule for a
// user at a point in time.
type UserLevelRuleProfile struct {
	RuleID            string          `json:"rule_id"`
	RuleName          string          `json:"rule_name"`
	WindowDays        int             `json:"window_days"`
	Enabled           bool            `json:"enabled"`
	Spend             float64         `json:"spend"`
	WindowFrom        time.Time       `json:"window_from"`
	CalculatedAt      time.Time       `json:"calculated_at"`
	CurrentTierID     string          `json:"current_tier_id,omitempty"`
	CurrentTierName   string          `json:"current_tier_name,omitempty"`
	CurrentTierOrder  int             `json:"current_tier_order"`
	MinSpend          float64         `json:"min_spend"`
	DefaultMultiplier *float64        `json:"default_multiplier,omitempty"`
	Tiers             []UserLevelTier `json:"tiers,omitempty"`
}

// UserLevelRulesRepository is deliberately separate from UserLevelRepository.
// Existing gateway test doubles and alternate implementations only need the
// rolling-spend/dynamic-quota methods; the SQL repository implements both.
type UserLevelRulesRepository interface {
	ListLevelRules(ctx context.Context) ([]UserLevelRule, error)
	GetLevelRule(ctx context.Context, ruleID string) (*UserLevelRule, error)
	CreateLevelRule(ctx context.Context, rule *UserLevelRule) error
	UpdateLevelRule(ctx context.Context, rule *UserLevelRule) error
	DeleteLevelRule(ctx context.Context, ruleID string) error

	ListUserLevelRules(ctx context.Context, userID string) ([]UserLevelRule, error)
	GetAssignedLevelRulesBatch(ctx context.Context, userIDs []string) (map[string][]UserLevelRule, error)
	ReplaceUserLevelRules(ctx context.Context, userID string, ruleIDs []string) error
	BatchAssignUserLevelRules(ctx context.Context, userIDs, ruleIDs []string, operation string) (int64, error)
	ListLevelRuleMembers(ctx context.Context, ruleID string, page, pageSize int) ([]User, int64, error)
	GetLevelRuleReferenceCount(ctx context.Context, ruleID string) (int64, error)
}

// UserLevelRuleAssignmentOperation is the stable API spelling for atomic
// multi-user assignment changes.
const (
	UserLevelRuleAssignmentAdd     = "add"
	UserLevelRuleAssignmentRemove  = "remove"
	UserLevelRuleAssignmentReplace = "replace"
)
