package service

import "testing"

func TestFormatPlatformPoints(t *testing.T) {
	tests := map[float64]string{
		0:           "0.00",
		1:           "1.00",
		1.2:         "1.20",
		1.234:       "1.234",
		1.23456789:  "1.23456789",
		1.234567895: "1.2345679",
	}
	for input, want := range tests {
		if got := formatPlatformPoints(input); got != want {
			t.Errorf("formatPlatformPoints(%v) = %q, want %q", input, got, want)
		}
	}
}
