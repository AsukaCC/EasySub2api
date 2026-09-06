-- Keep administrator usage available as a separate reporting scope while
-- allowing the regular admin dashboard to continue using pre-aggregated data.
ALTER TABLE usage_dashboard_hourly
    ADD COLUMN IF NOT EXISTS admin_requests BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS admin_input_tokens BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS admin_output_tokens BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS admin_cache_creation_tokens BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS admin_cache_read_tokens BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS admin_total_cost DECIMAL(20, 10) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS admin_actual_cost DECIMAL(20, 10) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS admin_account_cost DECIMAL(20, 10) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS admin_total_duration_ms BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS admin_active_users BIGINT NOT NULL DEFAULT 0;

ALTER TABLE usage_dashboard_daily
    ADD COLUMN IF NOT EXISTS admin_requests BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS admin_input_tokens BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS admin_output_tokens BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS admin_cache_creation_tokens BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS admin_cache_read_tokens BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS admin_total_cost DECIMAL(20, 10) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS admin_actual_cost DECIMAL(20, 10) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS admin_account_cost DECIMAL(20, 10) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS admin_total_duration_ms BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS admin_active_users BIGINT NOT NULL DEFAULT 0;

-- These tables are derived data. Rebuilding them prevents old all-user rows
-- from being interpreted as regular-user-only rows before the next aggregate
-- cycle runs. Raw usage logs and billing data are untouched.
TRUNCATE TABLE
    usage_dashboard_hourly_users,
    usage_dashboard_daily_users,
    usage_dashboard_hourly,
    usage_dashboard_daily;

UPDATE usage_dashboard_aggregation_watermark
SET last_aggregated_at = TIMESTAMPTZ '1970-01-01 00:00:00+00',
    updated_at = NOW()
WHERE id = 1;
