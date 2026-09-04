-- Preserve all existing dynamic-rate usage values as account quota U.
-- A short-lived application build appended |account_usd_v1 to quota keys;
-- merge those rows back into the original absolute interval buckets.
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '10min';

INSERT INTO group_dynamic_rate_usage (
    group_id, rule_id, quota_key, used_amount, created_at, updated_at
)
SELECT
    group_id,
    rule_id,
    LEFT(quota_key, LENGTH(quota_key) - LENGTH('|account_usd_v1')),
    used_amount,
    created_at,
    updated_at
FROM group_dynamic_rate_usage
WHERE quota_key LIKE '%|account_usd_v1'
ON CONFLICT (group_id, rule_id, quota_key) DO UPDATE
SET used_amount = group_dynamic_rate_usage.used_amount + EXCLUDED.used_amount,
    created_at = LEAST(group_dynamic_rate_usage.created_at, EXCLUDED.created_at),
    updated_at = GREATEST(group_dynamic_rate_usage.updated_at, EXCLUDED.updated_at);

DELETE FROM group_dynamic_rate_usage
WHERE quota_key LIKE '%|account_usd_v1';

INSERT INTO user_dynamic_rate_usage (
    user_id, group_id, rule_id, quota_key, used_amount, created_at, updated_at
)
SELECT
    user_id,
    group_id,
    rule_id,
    LEFT(quota_key, LENGTH(quota_key) - LENGTH('|account_usd_v1')),
    used_amount,
    created_at,
    updated_at
FROM user_dynamic_rate_usage
WHERE quota_key LIKE '%|account_usd_v1'
ON CONFLICT (user_id, group_id, rule_id, quota_key) DO UPDATE
SET used_amount = user_dynamic_rate_usage.used_amount + EXCLUDED.used_amount,
    created_at = LEAST(user_dynamic_rate_usage.created_at, EXCLUDED.created_at),
    updated_at = GREATEST(user_dynamic_rate_usage.updated_at, EXCLUDED.updated_at);

DELETE FROM user_dynamic_rate_usage
WHERE quota_key LIKE '%|account_usd_v1';

COMMENT ON COLUMN user_dynamic_rate_usage.quota_key IS 'Normalized UTC RFC3339Nano start_at of the absolute rule interval.';
COMMENT ON COLUMN group_dynamic_rate_usage.quota_key IS 'Normalized UTC RFC3339Nano start_at of the absolute rule interval.';
