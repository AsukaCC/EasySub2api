-- Dynamic-rate quota counters always track account 7-day usage U.
-- A zero configured quota is unlimited but its usage counter still accumulates,
-- so changing the quota later preserves usage from the same absolute interval.
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '10min';

COMMENT ON TABLE group_dynamic_rate_usage IS 'Site-wide account 7-day usage U consumed by all users matching one absolute dynamic-rate rule.';
COMMENT ON COLUMN group_dynamic_rate_usage.used_amount IS 'Account 7-day usage U accumulated across all users while this rule covers their requests.';

COMMENT ON TABLE user_dynamic_rate_usage IS 'Per-user account 7-day usage U consumed while one absolute dynamic-rate rule covers that user requests.';
COMMENT ON COLUMN user_dynamic_rate_usage.used_amount IS 'Account 7-day usage U accumulated independently for one user while this rule covers requests.';
