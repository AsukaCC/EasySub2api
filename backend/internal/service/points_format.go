package service

import (
	"strings"

	"github.com/shopspring/decimal"
)

// formatPlatformPoints renders point amounts with 2-8 decimal places.
func formatPlatformPoints(value float64) string {
	formatted := decimal.NewFromFloat(value).Round(8).StringFixed(8)
	formatted = strings.TrimRight(formatted, "0")
	if strings.HasSuffix(formatted, ".") {
		return formatted + "00"
	}
	if dot := strings.LastIndexByte(formatted, '.'); dot >= 0 {
		if decimals := len(formatted) - dot - 1; decimals < 2 {
			return formatted + strings.Repeat("0", 2-decimals)
		}
		return formatted
	}
	return formatted + ".00"
}
