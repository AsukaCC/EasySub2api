-- Keep the hot usage log table writable while adding the optional lookup index.
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_usage_logs_upstream_request_id_created_at
    ON usage_logs (upstream_request_id, created_at DESC)
    WHERE upstream_request_id IS NOT NULL;
