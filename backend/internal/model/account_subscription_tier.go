package model

import "strings"

// NormalizeSubscriptionTier produces the stable key used by account filters
// and subscription-tier resolution.
func NormalizeSubscriptionTier(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}
	var builder strings.Builder
	lastUnderscore := false
	for _, ch := range value {
		if ch == ' ' || ch == '-' || ch == '_' || ch == '/' {
			if builder.Len() > 0 && !lastUnderscore {
				_ = builder.WriteByte('_')
				lastUnderscore = true
			}
			continue
		}
		_, _ = builder.WriteRune(ch)
		lastUnderscore = false
	}
	return strings.Trim(builder.String(), "_")
}
