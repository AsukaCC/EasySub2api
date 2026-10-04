-- Existing expired deductions remain historical; outstanding bonus never expires.
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_bonus_balance_nonnegative;
ALTER TABLE wallet_bonus_grants ALTER COLUMN expires_at DROP NOT NULL;
UPDATE wallet_bonus_grants SET expires_at = NULL WHERE expires_at IS NOT NULL;

-- Old balance snapshots used both aggregate and split-bucket conventions.
ALTER TABLE wallet_transactions ADD COLUMN snapshot_version SMALLINT NOT NULL DEFAULT 0;
ALTER TABLE wallet_transactions ALTER COLUMN snapshot_version SET DEFAULT 1;
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_created_id
    ON wallet_transactions (created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_user_created_id
    ON wallet_transactions (user_id, created_at DESC, id DESC);
