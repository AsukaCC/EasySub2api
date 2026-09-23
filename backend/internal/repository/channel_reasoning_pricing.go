package repository

import (
	"encoding/json"
	"fmt"
	"github.com/AsukaCC/EasySub2api/internal/service"
)

func marshalReasoningEffortMultipliers(value map[string]float64) (any, error) {
	if value == nil {
		return nil, nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	if _, err := unmarshalReasoningEffortMultipliers(raw); err != nil {
		return nil, err
	}
	return string(raw), nil
}

func unmarshalReasoningEffortMultipliers(raw []byte) (map[string]float64, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var p service.ChannelModelPricing
	if err := json.Unmarshal(append(append([]byte(`{"reasoning_effort_multipliers":`), raw...), '}'), &p); err != nil {
		return nil, fmt.Errorf("invalid reasoning pricing: %w", err)
	}
	return p.ReasoningEffortMultipliers, nil
}
