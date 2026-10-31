package service

import (
	"context"
	"strings"
)

type accountTestPromptKey struct{}

func accountTestPrompt(ctx context.Context) string {
	prompt, _ := ctx.Value(accountTestPromptKey{}).(string)
	return strings.TrimSpace(prompt)
}

func applyAccountTestPrompt(ctx context.Context, payload map[string]any, protocol string) {
	prompt := accountTestPrompt(ctx)
	if prompt == "" {
		return
	}
	switch protocol {
	case "anthropic", "chat":
		payload["messages"] = []map[string]any{{"role": "user", "content": prompt}}
	case "responses":
		payload["input"] = []map[string]any{{"role": "user", "content": []map[string]any{{"type": "input_text", "text": prompt}}}}
	case "gemini":
		payload["contents"] = []map[string]any{{"role": "user", "parts": []map[string]any{{"text": prompt}}}}
	}
}
