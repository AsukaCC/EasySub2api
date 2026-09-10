package domain

import "encoding/json"

// GroupDynamicRateRule is persisted in groups.dynamic_rate_rules JSONB.
// Legacy level-target fields are retained only for decoding old rows.
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
	// DiscountCoefficient is the time-window discount coefficient. 1 means
	// original price; values below 1 reduce the user-side charge.
	DiscountCoefficient float64 `json:"discount_coefficient"`
	// Deprecated level scoping fields are retained for decoding old rows only.
	LevelTierIDs []string `json:"level_tier_ids,omitempty"`
	Levels       []int    `json:"levels,omitempty"`
	// Multiplier is retained only to decode pre-migration JSON. It is normalized
	// into DiscountCoefficient and is never emitted for new writes.
	Multiplier      float64 `json:"multiplier,omitempty"`
	ActivationSpend float64 `json:"activation_spend"`

	// Legacy daily-clock fields remain readable so administrators can replace
	// or delete old rules without losing the original configuration.
	Timezone    string  `json:"timezone,omitempty"`
	StartTime   string  `json:"start_time,omitempty"`
	EndTime     string  `json:"end_time,omitempty"`
	QuotaAmount float64 `json:"quota_amount,omitempty"`
}

// MarshalJSON keeps the new wire contract explicit while still accepting the
// legacy multiplier field on read during rolling upgrades.
func (r GroupDynamicRateRule) MarshalJSON() ([]byte, error) {
	type alias GroupDynamicRateRule
	v := alias(r)
	v.Multiplier = 0
	v.LevelTierIDs = nil
	v.Levels = nil
	return json.Marshal(v)
}
