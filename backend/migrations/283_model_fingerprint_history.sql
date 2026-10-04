CREATE TABLE IF NOT EXISTS model_fingerprint_runs (
    id UUID PRIMARY KEY,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    started_at TIMESTAMPTZ NOT NULL,
    snapshot JSONB NOT NULL
);
CREATE INDEX IF NOT EXISTS model_fingerprint_runs_account_time_idx
    ON model_fingerprint_runs (account_id, started_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS model_fingerprint_runs_time_idx ON model_fingerprint_runs (started_at);

CREATE TABLE IF NOT EXISTS model_fingerprint_schedules (
    account_id UUID PRIMARY KEY REFERENCES accounts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    options JSONB NOT NULL,
    next_run_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS model_fingerprint_schedules_due_idx
    ON model_fingerprint_schedules (next_run_at) WHERE enabled;

CREATE INDEX IF NOT EXISTS usage_logs_fingerprint_expiry_idx ON usage_logs (created_at)
    WHERE request_type = 6 AND request_id LIKE 'fingerprint:%';
