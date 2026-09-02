-- Per-group OpenAI Fast billing controls and reasoning-effort ceiling action.
-- Existing groups keep Standard service-tier selection and the historical
-- automatic reasoning-effort downgrade behavior.
ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS force_openai_fast BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS free_openai_fast BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS max_reasoning_effort_over_limit VARCHAR(20) NOT NULL DEFAULT 'downgrade';

COMMENT ON COLUMN groups.force_openai_fast IS
    'Force OpenAI gateway requests in this group to use service_tier=priority before global Fast/Flex policy evaluation';

COMMENT ON COLUMN groups.free_openai_fast IS
    'Whether Fast/priority requests in this OpenAI/Composite group are billed to users at Standard price';

COMMENT ON COLUMN groups.max_reasoning_effort_over_limit IS
    'Action when requested reasoning effort exceeds the group ceiling: downgrade or deny';
