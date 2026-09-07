package service

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

const defaultDynamicRateTimezone = "Asia/Shanghai"

// dynamicRatePersonalQuotaAmount returns the only quota that participates in
// live selection. SharedQuotaAmount is retained solely for reading legacy
// configurations; until migration 263 rewrites those rows, an old shared
// value is treated as the same quota for each individual user.
func dynamicRatePersonalQuotaAmount(rule GroupDynamicRateRule) float64 {
	if rule.PersonalQuotaAmount > 0 {
		return rule.PersonalQuotaAmount
	}
	if rule.SharedQuotaAmount > 0 {
		return rule.SharedQuotaAmount
	}
	return 0
}

func hasLegacyDynamicRateFields(rule GroupDynamicRateRule) bool {
	return strings.TrimSpace(rule.Timezone) != "" ||
		strings.TrimSpace(rule.StartTime) != "" ||
		strings.TrimSpace(rule.EndTime) != "" ||
		rule.QuotaAmount != 0
}

func isLegacyDynamicRateRule(rule GroupDynamicRateRule) bool {
	return strings.TrimSpace(rule.StartAt) == "" &&
		strings.TrimSpace(rule.EndAt) == "" &&
		hasLegacyDynamicRateFields(rule)
}

func parseDynamicRateWindow(rule GroupDynamicRateRule) (start, end time.Time, quotaKey string, ok bool) {
	startText := strings.TrimSpace(rule.StartAt)
	endText := strings.TrimSpace(rule.EndAt)
	if startText == "" || endText == "" {
		return time.Time{}, time.Time{}, "", false
	}
	start, err := time.Parse(time.RFC3339Nano, startText)
	if err != nil {
		return time.Time{}, time.Time{}, "", false
	}
	end, err = time.Parse(time.RFC3339Nano, endText)
	if err != nil || !start.Before(end) {
		return time.Time{}, time.Time{}, "", false
	}
	start = start.UTC()
	end = end.UTC()
	return start, end, start.Format(time.RFC3339Nano), true
}

func NormalizeLevelRateMultipliers(input map[string]float64) (map[string]float64, error) {
	out := make(map[string]float64, len(input))
	for key, value := range input {
		key = strings.TrimSpace(key)
		// Numeric keys belong to the removed fixed L1/L2/L3 model. They are
		// intentionally discarded instead of being interpreted as all-users
		// settings. New keys must identify a real user-level tier UUID.
		if key == "1" || key == "2" || key == "3" {
			continue
		}
		parsed, err := uuid.Parse(key)
		if err != nil {
			return nil, fmt.Errorf("level_rate_multipliers key must be a user level tier UUID")
		}
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0.01 || value > 100 {
			return nil, fmt.Errorf("level %s multiplier must be between 0.01 and 100", key)
		}
		out[parsed.String()] = math.Round(value*10000) / 10000
	}
	return out, nil
}

func NormalizeDynamicRateRules(input []GroupDynamicRateRule) ([]GroupDynamicRateRule, error) {
	out := make([]GroupDynamicRateRule, 0, len(input))
	seenIDs := make(map[string]struct{}, len(input))
	for i := range input {
		rule := input[i]
		rule.Name = strings.TrimSpace(rule.Name)
		if rule.Name == "" {
			return nil, fmt.Errorf("dynamic_rate_rules[%d].name is required", i)
		}
		if len([]rune(rule.Name)) > 100 {
			return nil, fmt.Errorf("dynamic_rate_rules[%d].name is too long", i)
		}
		if strings.TrimSpace(rule.ID) == "" {
			rule.ID = uuid.Must(uuid.NewV7()).String()
		} else if _, err := uuid.Parse(rule.ID); err != nil {
			return nil, fmt.Errorf("dynamic_rate_rules[%d].id must be a UUID", i)
		}
		if _, exists := seenIDs[rule.ID]; exists {
			return nil, fmt.Errorf("dynamic_rate_rules contains duplicate id %s", rule.ID)
		}
		seenIDs[rule.ID] = struct{}{}

		rule.Timezone = strings.TrimSpace(rule.Timezone)
		rule.StartTime = strings.TrimSpace(rule.StartTime)
		rule.EndTime = strings.TrimSpace(rule.EndTime)
		rule.StartAt = strings.TrimSpace(rule.StartAt)
		rule.EndAt = strings.TrimSpace(rule.EndAt)
		legacy := isLegacyDynamicRateRule(rule)
		if legacy {
			if rule.Timezone == "" {
				rule.Timezone = defaultDynamicRateTimezone
			}
			if rule.Timezone == "Local" {
				return nil, fmt.Errorf("dynamic_rate_rules[%d].timezone must be an IANA timezone", i)
			}
			if _, err := time.LoadLocation(rule.Timezone); err != nil {
				return nil, fmt.Errorf("dynamic_rate_rules[%d].timezone: %w", i, err)
			}
			start, ok := parseMinutes(rule.StartTime)
			if !ok {
				return nil, fmt.Errorf("dynamic_rate_rules[%d].start_time must use HH:MM", i)
			}
			end, ok := parseMinutes(rule.EndTime)
			if !ok {
				return nil, fmt.Errorf("dynamic_rate_rules[%d].end_time must use HH:MM", i)
			}
			if start >= end {
				return nil, fmt.Errorf("dynamic_rate_rules[%d] does not support cross-day intervals", i)
			}
		} else {
			if rule.StartAt == "" || rule.EndAt == "" {
				return nil, fmt.Errorf("dynamic_rate_rules[%d].start_at and end_at are required", i)
			}
			start, end, _, ok := parseDynamicRateWindow(rule)
			if !ok {
				return nil, fmt.Errorf("dynamic_rate_rules[%d].start_at and end_at must be valid RFC3339 timestamps with start_at before end_at", i)
			}
			rule.StartAt = start.Format(time.RFC3339Nano)
			rule.EndAt = end.Format(time.RFC3339Nano)
			// New absolute rules never consult the legacy daily-clock fields.
			rule.Timezone = ""
			rule.StartTime = ""
			rule.EndTime = ""
			rule.QuotaAmount = 0
		}
		if math.IsNaN(rule.Multiplier) || math.IsInf(rule.Multiplier, 0) || rule.Multiplier < 0.01 || rule.Multiplier > 100 {
			return nil, fmt.Errorf("dynamic_rate_rules[%d].multiplier must be between 0.01 and 100", i)
		}
		if math.IsNaN(rule.ActivationSpend) || math.IsInf(rule.ActivationSpend, 0) || rule.ActivationSpend < 0 {
			return nil, fmt.Errorf("dynamic_rate_rules[%d].activation_spend must be nonnegative", i)
		}
		if math.IsNaN(rule.QuotaAmount) || math.IsInf(rule.QuotaAmount, 0) || rule.QuotaAmount < 0 {
			return nil, fmt.Errorf("dynamic_rate_rules[%d].quota_amount must be nonnegative", i)
		}
		if math.IsNaN(rule.SharedQuotaAmount) || math.IsInf(rule.SharedQuotaAmount, 0) || rule.SharedQuotaAmount < 0 {
			return nil, fmt.Errorf("dynamic_rate_rules[%d].shared_quota_amount must be nonnegative", i)
		}
		if math.IsNaN(rule.PersonalQuotaAmount) || math.IsInf(rule.PersonalQuotaAmount, 0) || rule.PersonalQuotaAmount < 0 {
			return nil, fmt.Errorf("dynamic_rate_rules[%d].personal_quota_amount must be nonnegative", i)
		}
		// Shared quotas were removed from the live model. If an older client
		// still submits one, preserve its finite limit as an independent quota
		// for every user instead of restoring group-wide consumption.
		if rule.PersonalQuotaAmount <= 0 && rule.SharedQuotaAmount > 0 {
			rule.PersonalQuotaAmount = rule.SharedQuotaAmount
		}
		rule.SharedQuotaAmount = 0
		rule.Multiplier = math.Round(rule.Multiplier*10000) / 10000
		rule.ActivationSpend = QuantizeUsageBillingAmount(rule.ActivationSpend)
		rule.QuotaAmount = QuantizeUsageBillingAmount(rule.QuotaAmount)
		rule.SharedQuotaAmount = QuantizeUsageBillingAmount(rule.SharedQuotaAmount)
		rule.PersonalQuotaAmount = QuantizeUsageBillingAmount(rule.PersonalQuotaAmount)

		// Legacy numeric targets cannot be mapped safely to a tier UUID. Keep
		// the payload for an administrator to inspect, but make the rule inert.
		if len(rule.Levels) > 0 {
			rule.Enabled = false
		}
		seenTierIDs := make(map[string]struct{}, len(rule.LevelTierIDs))
		tierIDs := make([]string, 0, len(rule.LevelTierIDs))
		for _, tierID := range rule.LevelTierIDs {
			tierID = strings.TrimSpace(tierID)
			if tierID == "" {
				return nil, fmt.Errorf("dynamic_rate_rules[%d].level_tier_ids cannot contain empty values", i)
			}
			if _, err := uuid.Parse(tierID); err != nil {
				return nil, fmt.Errorf("dynamic_rate_rules[%d].level_tier_ids must contain UUIDs", i)
			}
			if _, exists := seenTierIDs[tierID]; exists {
				continue
			}
			seenTierIDs[tierID] = struct{}{}
			tierIDs = append(tierIDs, tierID)
		}
		sort.Strings(tierIDs)
		rule.LevelTierIDs = tierIDs
		out = append(out, rule)
	}
	return out, nil
}
