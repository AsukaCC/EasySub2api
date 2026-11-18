ALTER TABLE groups ADD COLUMN IF NOT EXISTS model_forwarding JSONB NOT NULL DEFAULT '{}';

-- Lock and migrate the same set of account rules. Unassigned accounts keep their
-- legacy rules. The group membership priority wins conflicting sources, followed
-- by the account priority and ID as deterministic tie breakers.
WITH memberships AS (
    SELECT ag.group_id, ag.account_id, ag.priority
    FROM account_groups ag JOIN groups g ON g.id = ag.group_id
    WHERE g.deleted_at IS NULL
), eligible AS (
    SELECT a.id, a.priority, a.credentials->'model_forwarding' AS forwarding
    FROM accounts a
    WHERE a.platform = 'openai' AND a.deleted_at IS NULL
      AND jsonb_typeof(a.credentials->'model_forwarding') = 'object'
      AND EXISTS (SELECT 1 FROM memberships m WHERE m.account_id = a.id)
    FOR UPDATE OF a
), rules AS (
    SELECT DISTINCT ON (m.group_id, rule.key) m.group_id, rule.key, rule.value
    FROM eligible a JOIN memberships m ON m.account_id = a.id
    CROSS JOIN LATERAL jsonb_each(a.forwarding) rule
    WHERE jsonb_typeof(rule.value) = 'string'
    ORDER BY m.group_id, rule.key, m.priority, a.priority, a.id
), merged AS (
    SELECT group_id, jsonb_object_agg(key, value) AS mapping FROM rules GROUP BY group_id
), migrated_groups AS (
    UPDATE groups g SET model_forwarding = merged.mapping || g.model_forwarding, updated_at = NOW()
    FROM merged WHERE g.id = merged.group_id
), migrated_accounts AS (
    UPDATE accounts a SET credentials = a.credentials - 'model_forwarding', updated_at = NOW()
    FROM eligible WHERE a.id = eligible.id
    RETURNING a.id
)
INSERT INTO scheduler_outbox (event_type, account_id)
SELECT 'account_changed', id FROM migrated_accounts;
