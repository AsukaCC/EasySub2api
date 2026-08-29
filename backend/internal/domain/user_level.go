package domain

// GroupDynamicRateRule is persisted in groups.dynamic_rate_rules JSONB.
// An empty Levels slice means the rule applies to every user level.
type GroupDynamicRateRule struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Enabled         bool    `json:"enabled"`
	Timezone        string  `json:"timezone"`
	StartTime       string  `json:"start_time"`
	EndTime         string  `json:"end_time"`
	Levels          []int   `json:"levels"`
	Multiplier      float64 `json:"multiplier"`
	ActivationSpend float64 `json:"activation_spend"`
	QuotaAmount     float64 `json:"quota_amount"`
}
