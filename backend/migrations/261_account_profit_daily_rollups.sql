-- Durable per-account daily profit rollups. No foreign key is used so history
-- remains readable after an account is deleted.
CREATE TABLE IF NOT EXISTS account_profit_daily_rollups (
    account_id BIGINT NOT NULL,
    account_name_snapshot TEXT NOT NULL DEFAULT '',
    bucket_date DATE NOT NULL,
    request_count BIGINT NOT NULL DEFAULT 0,
    total_tokens BIGINT NOT NULL DEFAULT 0,
    revenue_points NUMERIC(20, 10) NOT NULL DEFAULT 0,
    cost_usd NUMERIC(20, 10) NOT NULL DEFAULT 0,
    profit_points NUMERIC(20, 10) NOT NULL DEFAULT 0,
    computed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (account_id, bucket_date)
);

CREATE INDEX IF NOT EXISTS idx_account_profit_daily_rollups_account_date
    ON account_profit_daily_rollups (account_id, bucket_date DESC);
