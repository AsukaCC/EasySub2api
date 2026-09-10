package openai

import "testing"

func TestIsCodexLoadShedOriginator(t *testing.T) {
	tests := []struct {
		originator string
		want       bool
	}{
		{"codex-tui", true},
		{"  CODEX-TUI  ", true},
		{"codex_cli_rs", false},
		{"codex_vscode", false},
		{"Codex Desktop", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := IsCodexLoadShedOriginator(tt.originator); got != tt.want {
			t.Errorf("IsCodexLoadShedOriginator(%q) = %v, want %v", tt.originator, got, tt.want)
		}
	}
}

func TestNormalizeCodexClientIdentityToCLI(t *testing.T) {
	tests := []struct {
		name           string
		originator     string
		ua             string
		wantOriginator string
		wantUA         string
		wantChanged    bool
	}{
		{
			name:           "full tui identity",
			originator:     "codex-tui",
			ua:             "codex-tui/0.144.1 (Ubuntu 22.4.0; x86_64) xterm-256color (codex-tui; 0.144.1)",
			wantOriginator: "codex_cli_rs",
			wantUA:         "codex_cli_rs/0.144.1 (Ubuntu 22.4.0; x86_64) xterm-256color",
			wantChanged:    true,
		},
		{
			name:           "mac tui identity",
			originator:     "codex-tui",
			ua:             "codex-tui/0.146.0 (Mac OS 26.5.0; arm64) iTerm.app/3.6.10 (codex-tui; 0.153.3)",
			wantOriginator: "codex_cli_rs",
			wantUA:         "codex_cli_rs/0.146.0 (Mac OS 26.5.0; arm64) iTerm.app/3.6.10",
			wantChanged:    true,
		},
		{
			name:           "os group is preserved",
			originator:     "codex-tui",
			ua:             "codex-tui/0.144.1 (Ubuntu 22.4.0; x86_64)",
			wantOriginator: "codex_cli_rs",
			wantUA:         "codex_cli_rs/0.144.1 (Ubuntu 22.4.0; x86_64)",
			wantChanged:    true,
		},
		{
			name:           "healthy identity",
			originator:     "codex_cli_rs",
			ua:             "codex_cli_rs/0.144.1 (Ubuntu 22.4.0; x86_64) xterm-256color",
			wantOriginator: "codex_cli_rs",
			wantUA:         "codex_cli_rs/0.144.1 (Ubuntu 22.4.0; x86_64) xterm-256color",
			wantChanged:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOriginator, gotUA, changed := NormalizeCodexClientIdentityToCLI(tt.originator, tt.ua)
			if gotOriginator != tt.wantOriginator || gotUA != tt.wantUA || changed != tt.wantChanged {
				t.Fatalf("got (%q, %q, %v), want (%q, %q, %v)", gotOriginator, gotUA, changed, tt.wantOriginator, tt.wantUA, tt.wantChanged)
			}
			if changed {
				pairedOriginator, pairedUA, ok := PairCodexClientIdentity(gotUA)
				if !ok || pairedOriginator != gotOriginator || pairedUA != gotUA {
					t.Fatalf("normalized identity is not paired: (%q, %q, %v)", pairedOriginator, pairedUA, ok)
				}
			}
		})
	}
}
