CREATE INDEX IF NOT EXISTS idx_accounts_model_fingerprint_queue
ON accounts ((extra->'model_fingerprint'->>'started_at'), id)
WHERE deleted_at IS NULL AND extra->'model_fingerprint'->>'status' = 'queued';
