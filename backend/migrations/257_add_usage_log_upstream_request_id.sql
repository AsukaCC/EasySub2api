-- Persist the upstream response request identifier selected by the account
-- extra.upstream_request_id_header setting. Historical rows remain NULL.
ALTER TABLE usage_logs
    ADD COLUMN IF NOT EXISTS upstream_request_id VARCHAR(128);
