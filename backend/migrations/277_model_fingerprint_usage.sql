ALTER TABLE usage_logs ALTER COLUMN api_key_id DROP NOT NULL;

ALTER TABLE usage_logs DROP CONSTRAINT IF EXISTS usage_logs_request_type_check;
ALTER TABLE usage_logs ADD CONSTRAINT usage_logs_request_type_check
    CHECK (request_type >= 0 AND request_type <= 6);
ALTER TABLE usage_logs ADD CONSTRAINT usage_logs_keyless_test_check
    CHECK (api_key_id IS NOT NULL OR request_type = 6);

-- Keyless tests use a stable job/turn request ID for idempotent logging.
CREATE UNIQUE INDEX IF NOT EXISTS usage_logs_keyless_request_id_idx
    ON usage_logs (request_id) WHERE api_key_id IS NULL;
