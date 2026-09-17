package openai

import "testing"

func TestCodexTUIDefaultPreservesExplicitClients(t *testing.T) {
	if CodexDefaultOriginator != "codex-tui" {
		t.Fatal("unexpected default")
	}
	for _, name := range []string{"codex-tui", "codex_cli_rs", "codex_vscode"} {
		originator, _, ok := PairCodexClientIdentity(name + "/0.154.0 (Ubuntu; x86_64)")
		if !ok || originator != name {
			t.Fatalf("explicit client lost: %s", name)
		}
	}
}
