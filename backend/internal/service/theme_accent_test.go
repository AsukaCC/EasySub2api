package service

import "testing"

func TestNormalizeThemeAccent(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"#0A84FF": DefaultThemeAccent,
		"#abc":    "#aabbcc",
		"blue":    DefaultThemeAccent,
		"":        DefaultThemeAccent,
		"#059669": "#059669",
	}
	for input, want := range cases {
		if got := normalizeThemeAccent(input); got != want {
			t.Fatalf("normalizeThemeAccent(%q) = %q, want %q", input, got, want)
		}
	}
}
