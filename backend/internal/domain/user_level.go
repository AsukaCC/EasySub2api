package domain

// GroupDynamicRateRule is persisted in groups.dynamic_rate_rules JSONB.
// An empty LevelTierIDs slice means the rule is not scoped to a user tier.
type GroupDynamicRateRule struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
	StartAt string `json:"start_at"`
	EndAt   string `json:"end_at"`
	// SharedQuotaAmount is retained only to decode pre-263 configurations.
	// Live selection and billing use PersonalQuotaAmount per user.
	SharedQuotaAmount   float64 `json:"shared_quota_amount"`
	PersonalQuotaAmount float64 `json:"personal_quota_amount"`
	// LevelTierIDs limits the rule to the currently reached tiers. An empty
	// slice means all users, while legacy numeric Levels are intentionally inert.
	LevelTierIDs []string `json:"level_tier_ids,omitempty"`
	// Levels is retained only so old JSON can be decoded and re-saved safely.
	// It is never used for live matching after migration 265.
	Levels          []int   `json:"levels,omitempty"`
	Multiplier      float64 `json:"multiplier"`
	ActivationSpend float64 `json:"activation_spend"`

	// Legacy daily-clock fields remain readable so administrators can replace
	// or delete old rules without losing the original configuration.
	Timezone    string  `json:"timezone,omitempty"`
	StartTime   string  `json:"start_time,omitempty"`
	EndTime     string  `json:"end_time,omitempty"`
	QuotaAmount float64 `json:"quota_amount,omitempty"`
}
