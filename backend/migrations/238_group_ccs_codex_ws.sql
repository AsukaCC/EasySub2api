-- Configure CC-Switch Codex WebSocket export independently for each group.
ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS ccs_codex_ws_enabled BOOLEAN NOT NULL DEFAULT FALSE;
