-- Dynamic-rate rules now use one absolute UTC interval and two independent
-- point quotas. Existing daily bucket rows are retained for history but are
-- not read by the new billing key format.

CREATE TABLE IF NOT EXISTS user_dynamic_rate_usage (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    rule_id UUID NOT NULL,
    quota_key TEXT NOT NULL,
    used_amount DECIMAL(20,8) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, group_id, rule_id, quota_key),
    CONSTRAINT user_dynamic_rate_usage_nonnegative CHECK (used_amount >= 0)
);

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'user_dynamic_rate_usage'
          AND column_name = 'bucket_date'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'user_dynamic_rate_usage'
          AND column_name = 'quota_key'
    ) THEN
        ALTER TABLE user_dynamic_rate_usage RENAME COLUMN bucket_date TO quota_key;
    END IF;
END $$;

ALTER TABLE user_dynamic_rate_usage
    ALTER COLUMN quota_key TYPE TEXT USING quota_key::text;

DO $$
DECLARE
    existing_primary_key TEXT;
BEGIN
    SELECT conname
    INTO existing_primary_key
    FROM pg_constraint
    WHERE conrelid = 'public.user_dynamic_rate_usage'::regclass
      AND contype = 'p';

    IF existing_primary_key IS NOT NULL THEN
        EXECUTE format('ALTER TABLE user_dynamic_rate_usage DROP CONSTRAINT %I', existing_primary_key);
    END IF;

    ALTER TABLE user_dynamic_rate_usage
        ADD CONSTRAINT user_dynamic_rate_usage_pkey
        PRIMARY KEY (user_id, group_id, rule_id, quota_key);
END $$;

DROP INDEX IF EXISTS idx_user_dynamic_rate_usage_group_date;
CREATE INDEX IF NOT EXISTS idx_user_dynamic_rate_usage_group_key
    ON user_dynamic_rate_usage(group_id, quota_key);

COMMENT ON TABLE user_dynamic_rate_usage IS 'Per-user consumption of absolute group dynamic-rate quotas.';
COMMENT ON COLUMN user_dynamic_rate_usage.quota_key IS 'Normalized UTC RFC3339Nano start_at of the absolute rule interval.';
COMMENT ON COLUMN user_dynamic_rate_usage.used_amount IS 'Charged platform points consumed from the personal interval quota.';

CREATE TABLE IF NOT EXISTS group_dynamic_rate_usage (
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    rule_id UUID NOT NULL,
    quota_key TEXT NOT NULL,
    used_amount DECIMAL(20,8) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (group_id, rule_id, quota_key),
    CONSTRAINT group_dynamic_rate_usage_nonnegative CHECK (used_amount >= 0)
);

CREATE INDEX IF NOT EXISTS idx_group_dynamic_rate_usage_group_key
    ON group_dynamic_rate_usage(group_id, quota_key);

COMMENT ON TABLE group_dynamic_rate_usage IS 'Shared consumption of absolute group dynamic-rate quotas.';
COMMENT ON COLUMN group_dynamic_rate_usage.quota_key IS 'Normalized UTC RFC3339Nano start_at of the absolute rule interval.';
COMMENT ON COLUMN group_dynamic_rate_usage.used_amount IS 'Charged platform points consumed from the shared interval quota.';
