-- Persist the client-requested reasoning effort before policy/model rewriting,
-- and track native remote compaction v2 independently from transport type.
ALTER TABLE usage_logs
    ADD COLUMN IF NOT EXISTS requested_reasoning_effort VARCHAR(20),
    ADD COLUMN IF NOT EXISTS native_compaction_v2 BOOLEAN NOT NULL DEFAULT FALSE;

COMMENT ON COLUMN usage_logs.requested_reasoning_effort IS
    'Client-requested reasoning effort before group policy and model-family remapping';
COMMENT ON COLUMN usage_logs.native_compaction_v2 IS
    'Whether this usage row was positively identified as native OpenAI remote compaction v2';
