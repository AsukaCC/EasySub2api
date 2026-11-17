package openai

import (
	"fmt"
	"strings"
	"time"
)

// GPT62BaseModel recognizes published IDs and local effort/compact aliases.
func GPT62BaseModel(model string) string {
	model = strings.ToLower(strings.TrimSpace(model))
	if i := strings.LastIndexByte(model, '/'); i >= 0 {
		model = model[i+1:]
	}
	model = strings.Join(strings.Fields(strings.ReplaceAll(model, "_", "-")), "-")
	model = strings.ReplaceAll(model, "gpt62", "gpt-6.2")
	model = strings.ReplaceAll(model, "gpt6.2", "gpt-6.2")
	for _, base := range []string{"gpt-6.2-astra", "gpt-6.2-sol"} {
		if model == base {
			return base
		}
		if suffix, ok := strings.CutPrefix(model, base+"-"); ok {
			switch suffix {
			case "none", "minimal", "low", "medium", "high", "xhigh", "max", "openai-compact":
				return base
			default:
				if _, err := time.Parse("2006-01-02", suffix); err == nil {
					return base
				}
			}
		}
	}
	return ""
}

func IsGPT62ModelSpelling(model string) bool { return GPT62BaseModel(model) != "" }

func ValidateGPT62ReasoningEffort(model, effort string) error {
	base := GPT62BaseModel(model)
	if base == "" {
		return nil
	}
	canonical := strings.ToLower(strings.TrimSpace(model))
	canonical = strings.Join(strings.Fields(strings.ReplaceAll(canonical, "_", "-")), "-")
	for _, disabled := range []string{"none", "minimal"} {
		if strings.EqualFold(strings.TrimSpace(effort), disabled) || strings.HasSuffix(canonical, "-"+disabled) {
			return fmt.Errorf("%s does not support reasoning effort %q; use low, medium, high, xhigh or max", base, disabled)
		}
	}
	return nil
}
