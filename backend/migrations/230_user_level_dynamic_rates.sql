-- User levels and group dynamic rate rules.
-- Global level thresholds are stored in settings.user_level_settings.

ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS level_rate_multipliers JSONB NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN IF NOT EXISTS dynamic_rate_rules JSONB NOT NULL DEFAULT '[]'::jsonb;

COMMENT ON COLUMN groups.level_rate_multipliers IS 'User level multiplier overrides: {"1":1.0,"2":0.9,"3":0.8}.';
COMMENT ON COLUMN groups.dynamic_rate_rules IS 'Time-bound user-level rate rules with rolling-spend activation and per-user daily quota.';

CREATE TABLE IF NOT EXISTS user_dynamic_rate_usage (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    rule_id UUID NOT NULL,
    bucket_date DATE NOT NULL,
    used_amount DECIMAL(20,8) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, group_id, rule_id, bucket_date),
    CONSTRAINT user_dynamic_rate_usage_nonnegative CHECK (used_amount >= 0)
);

CREATE INDEX IF NOT EXISTS idx_user_dynamic_rate_usage_group_date
    ON user_dynamic_rate_usage(group_id, bucket_date);

COMMENT ON TABLE user_dynamic_rate_usage IS 'Per-user daily consumption of group dynamic-rate quotas.';
COMMENT ON COLUMN user_dynamic_rate_usage.used_amount IS 'Discounted USD charged against this rule quota.';
