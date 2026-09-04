-- Dynamic-rate shared and personal quotas are measured in account billed USD
-- (account quota U), rather than user-facing charged platform points.
-- Existing point-denominated rows remain for audit history. The application
-- uses a versioned quota_key so active intervals start a clean USD bucket.
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '10min';

COMMENT ON TABLE user_dynamic_rate_usage IS 'Per-user account billed USD consumed from absolute group dynamic-rate quotas.';
COMMENT ON COLUMN user_dynamic_rate_usage.quota_key IS 'Normalized UTC rule start_at plus quota unit version.';
COMMENT ON COLUMN user_dynamic_rate_usage.used_amount IS 'Account billed USD consumed from the personal interval quota.';

COMMENT ON TABLE group_dynamic_rate_usage IS 'Shared account billed USD consumed from absolute group dynamic-rate quotas.';
COMMENT ON COLUMN group_dynamic_rate_usage.quota_key IS 'Normalized UTC rule start_at plus quota unit version.';
COMMENT ON COLUMN group_dynamic_rate_usage.used_amount IS 'Account billed USD consumed from the shared interval quota.';
