package service

import (
	"fmt"
	"strings"

	"github.com/AsukaCC/EasySub2api/internal/pkg/openai"
	"github.com/tidwall/gjson"
)

func validateGPT61SolChatOnlyRequest(body []byte, upstreamModel string, hasTools bool) error {
	if err := validateGPT61SolReasoningEffort(body, upstreamModel); err != nil {
		return err
	}
	if openai.IsGPT61SolModelSpelling(upstreamModel) && hasTools {
		return fmt.Errorf("gpt-6.1-sol requires Responses for tool calls; this account only supports Chat Completions")
	}
	return nil
}

// validateGPT61SolReasoningEffort rejects legacy disabled reasoning aliases at
// every compatibility boundary before they can be silently normalized away.
func validateGPT61SolReasoningEffort(body []byte, models ...string) error {
	models = append(models, gjson.GetBytes(body, "model").String())
	for _, model := range models {
		if !openai.IsGPT61SolModelSpelling(model) {
			continue
		}
		for _, path := range []string{"reasoning.effort", "reasoning_effort", "output_config.effort"} {
			effort := strings.TrimSpace(gjson.GetBytes(body, path).String())
			if strings.EqualFold(gjson.GetBytes(body, "thinking.type").String(), "disabled") {
				effort = "none"
			}
			if err := openai.ValidateGPT61SolReasoningEffort(model, effort); err != nil {
				return fmt.Errorf("invalid reasoning effort for %s: %w", model, err)
			}
		}
	}
	return nil
}
