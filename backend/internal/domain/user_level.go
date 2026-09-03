package domain

// GroupDynamicRateRule is persisted in groups.dynamic_rate_rules JSONB.
// An empty Levels slice means the rule applies to every user level.
type GroupDynamicRateRule struct {
	ID                  string  `json:"id"`
	Name                string  `json:"name"`
	Enabled             bool    `json:"enabled"`
	StartAt             string  `json:"start_at"`
	EndAt               string  `json:"end_at"`
	SharedQuotaAmount   float64 `json:"shared_quota_amount"`
	PersonalQuotaAmount float64 `json:"personal_quota_amount"`
	Levels              []int   `json:"levels"`
	Multiplier          float64 `json:"multiplier"`
	ActivationSpend     float64 `json:"activation_spend"`

	// Legacy daily-clock fields remain readable so administrators can replace
	// or delete old rules without losing the original configuration.
	Timezone    string  `json:"timezone,omitempty"`
	StartTime   string  `json:"start_time,omitempty"`
	EndTime     string  `json:"end_time,omitempty"`
	QuotaAmount float64 `json:"quota_amount,omitempty"`
}
