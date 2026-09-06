-- Dynamic-rate quotas are independent per user and per effective rule.
-- Convert legacy group-wide values to the same amount for each user so old
-- configurations keep a finite quota without sharing consumption.
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '10min';

UPDATE groups AS g
SET dynamic_rate_rules = COALESCE((
    SELECT jsonb_agg(
        jsonb_set(
            jsonb_set(
                item.rule,
                '{personal_quota_amount}',
                CASE
                    WHEN jsonb_typeof(item.rule->'personal_quota_amount') = 'number'
                         AND (item.rule->>'personal_quota_amount')::numeric > 0
                        THEN item.rule->'personal_quota_amount'
                    WHEN jsonb_typeof(item.rule->'shared_quota_amount') = 'number'
                         AND (item.rule->>'shared_quota_amount')::numeric > 0
                        THEN item.rule->'shared_quota_amount'
                    ELSE '0'::jsonb
                END,
                true
            ),
            '{shared_quota_amount}',
            '0'::jsonb,
            true
        ) ORDER BY item.ordinality
    )
    FROM jsonb_array_elements(g.dynamic_rate_rules) WITH ORDINALITY AS item(rule, ordinality)
), '[]'::jsonb)
WHERE jsonb_typeof(g.dynamic_rate_rules) = 'array';

COMMENT ON TABLE group_dynamic_rate_usage IS
    'Legacy group-wide dynamic-rate counters retained for history; live quota selection and billing use user_dynamic_rate_usage only.';
COMMENT ON COLUMN group_dynamic_rate_usage.used_amount IS
    'Legacy group-wide counter; no longer read or written by live dynamic-rate billing.';
COMMENT ON TABLE user_dynamic_rate_usage IS
    'Independent dynamic-rate consumption for one user, group, rule, and absolute interval.';
COMMENT ON COLUMN user_dynamic_rate_usage.used_amount IS
    'Account-side billed USD consumed independently by the effective rule for one user.';
