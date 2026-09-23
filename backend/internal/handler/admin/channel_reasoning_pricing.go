package admin

import (
	"encoding/json"
	"github.com/AsukaCC/EasySub2api/internal/service"
)

func (p *channelModelPricingRequest) UnmarshalJSON(data []byte) error {
	if err := service.RejectLegacyReasoningPricing(data); err != nil {
		return err
	}
	type plain channelModelPricingRequest
	var value plain
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*p = channelModelPricingRequest(value)
	return nil
}
