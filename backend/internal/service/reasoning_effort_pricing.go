package service

import (
	"encoding/json"
	"fmt"
)

// These defaults preserve existing prices; channel/group maps replace the whole
// default map, including an explicitly empty map.
func defaultReasoningEffortMultipliers(model string) map[string]float64 {
	if isClaudeFable51Model(model) {
		return map[string]float64{"max": 3}
	}
	return nil
}

func effectiveReasoningEffortMultipliers(model string, configured map[string]float64) map[string]float64 {
	if configured != nil {
		return configured
	}
	return defaultReasoningEffortMultipliers(model)
}

func RejectLegacyReasoningPricing(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	if _, exists := fields["max_reasoning_effort_multiplier"]; exists {
		return fmt.Errorf("max_reasoning_effort_multiplier is no longer supported; use reasoning_effort_multipliers")
	}
	return nil
}

func (p *ChannelModelPricing) UnmarshalJSON(data []byte) error {
	if err := RejectLegacyReasoningPricing(data); err != nil {
		return err
	}
	type plain ChannelModelPricing
	var value plain
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if err := validateReasoningEffortMultipliers([]ChannelModelPricing{ChannelModelPricing(value)}); err != nil {
		return err
	}
	*p = ChannelModelPricing(value)
	return nil
}
