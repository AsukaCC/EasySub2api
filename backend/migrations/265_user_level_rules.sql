-- Custom user level rule profiles.
-- There is intentionally no seed row, default tier, or automatic assignment.
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '10min';

-- Validate legacy JSON before the new runtime starts ignoring numeric level keys.
-- A malformed configuration must stop the migration rather than silently becoming
-- an all-users rule.
DO $$
DECLARE
    bad_group_id TEXT;
    bad_reason TEXT;
BEGIN
    IF to_regclass('public.groups') IS NOT NULL THEN
        IF EXISTS (
            SELECT 1 FROM groups
            WHERE jsonb_typeof(COALESCE(level_rate_multipliers, '{}'::jsonb)) IS DISTINCT FROM 'object'
        ) THEN
            RAISE EXCEPTION 'cannot migrate user level rules: groups.level_rate_multipliers must be a JSON object';
        END IF;

        SELECT g.id::text, e.key
        INTO bad_group_id, bad_reason
        FROM groups AS g
        CROSS JOIN LATERAL jsonb_each(COALESCE(g.level_rate_multipliers, '{}'::jsonb)) AS e(key, value)
        WHERE e.key !~ '^[123]$'
          AND e.key !~* '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'
        LIMIT 1;
        IF bad_group_id IS NOT NULL THEN
            RAISE EXCEPTION 'cannot migrate user level rules: group % has unsupported level key %', bad_group_id, bad_reason;
        END IF;

        SELECT g.id::text, e.key
        INTO bad_group_id, bad_reason
        FROM groups AS g
        CROSS JOIN LATERAL jsonb_each(COALESCE(g.level_rate_multipliers, '{}'::jsonb)) AS e(key, value)
        WHERE jsonb_typeof(e.value) IS DISTINCT FROM 'number'
           OR (e.value::text)::numeric < 0.01
           OR (e.value::text)::numeric > 100
        LIMIT 1;
        IF bad_group_id IS NOT NULL THEN
            RAISE EXCEPTION 'cannot migrate user level rules: group % has invalid multiplier for key %', bad_group_id, bad_reason;
        END IF;

        IF EXISTS (
            SELECT 1 FROM groups
            WHERE jsonb_typeof(COALESCE(dynamic_rate_rules, '[]'::jsonb)) IS DISTINCT FROM 'array'
        ) THEN
            RAISE EXCEPTION 'cannot migrate user level rules: groups.dynamic_rate_rules must be a JSON array';
        END IF;

        SELECT g.id::text, 'dynamic rule'
        INTO bad_group_id, bad_reason
        FROM groups AS g
        CROSS JOIN LATERAL jsonb_array_elements(COALESCE(g.dynamic_rate_rules, '[]'::jsonb)) AS item(rule)
        WHERE jsonb_typeof(item.rule) IS DISTINCT FROM 'object'
           OR (item.rule ? 'levels' AND jsonb_typeof(item.rule->'levels') IS DISTINCT FROM 'array')
           OR (item.rule ? 'level_tier_ids' AND jsonb_typeof(item.rule->'level_tier_ids') IS DISTINCT FROM 'array')
           OR (
               item.rule ? 'levels'
               AND jsonb_typeof(item.rule->'levels') = 'array'
               AND EXISTS (
                   SELECT 1
                   FROM jsonb_array_elements(item.rule->'levels') AS level(value)
                   WHERE jsonb_typeof(level.value) IS DISTINCT FROM 'number'
                      OR (level.value::text)::numeric NOT IN (1, 2, 3)
               )
           )
           OR (
               item.rule ? 'level_tier_ids'
               AND jsonb_typeof(item.rule->'level_tier_ids') = 'array'
               AND EXISTS (
                   SELECT 1
                   FROM jsonb_array_elements(item.rule->'level_tier_ids') AS tier(value)
                   WHERE jsonb_typeof(tier.value) IS DISTINCT FROM 'string'
                      OR tier.value::text !~* '"[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}"'
               )
           )
        LIMIT 1;
        IF bad_group_id IS NOT NULL THEN
            RAISE EXCEPTION 'cannot migrate user level rules: group % has malformed dynamic rule', bad_group_id;
        END IF;

        -- Numeric level targets have no safe UUID equivalent. Keep the JSON for
        -- audit/reconfiguration, but make those rules inert. Empty legacy levels
        -- remain unscoped rules and continue to work.
        UPDATE groups AS g
        SET dynamic_rate_rules = COALESCE((
            SELECT jsonb_agg(
                CASE
                    WHEN (item.rule ? 'levels')
                         AND jsonb_array_length(item.rule->'levels') > 0
                         AND NOT (item.rule ? 'level_tier_ids')
                    THEN jsonb_set(item.rule, '{enabled}', 'false'::jsonb, true)
                    ELSE item.rule
                END
                ORDER BY item.ordinality
            )
            FROM jsonb_array_elements(COALESCE(g.dynamic_rate_rules, '[]'::jsonb))
                WITH ORDINALITY AS item(rule, ordinality)
        ), '[]'::jsonb);
    END IF;

    IF to_regclass('public.announcements') IS NOT NULL THEN
        IF EXISTS (
            SELECT 1 FROM announcements
            WHERE jsonb_typeof(COALESCE(targeting, '{}'::jsonb)) IS DISTINCT FROM 'object'
        ) THEN
            RAISE EXCEPTION 'cannot migrate user level rules: announcements.targeting must be a JSON object';
        END IF;
        IF EXISTS (
            SELECT 1
            FROM announcements AS a
            CROSS JOIN LATERAL jsonb_array_elements(COALESCE(a.targeting->'any_of', '[]'::jsonb)) AS og(value)
            WHERE jsonb_typeof(og.value) IS DISTINCT FROM 'object'
               OR (og.value ? 'all_of' AND jsonb_typeof(og.value->'all_of') IS DISTINCT FROM 'array')
        ) THEN
            RAISE EXCEPTION 'cannot migrate user level rules: announcements.targeting has malformed condition groups';
        END IF;
        IF EXISTS (
            SELECT 1
            FROM announcements AS a
            CROSS JOIN LATERAL jsonb_array_elements(COALESCE(a.targeting->'any_of', '[]'::jsonb)) AS og(value)
            CROSS JOIN LATERAL jsonb_array_elements(COALESCE(og.value->'all_of', '[]'::jsonb)) AS cond(value)
            WHERE (cond.value ? 'levels' AND jsonb_typeof(cond.value->'levels') IS DISTINCT FROM 'array')
               OR (
                   cond.value ? 'levels'
                   AND jsonb_typeof(cond.value->'levels') = 'array'
                   AND EXISTS (
                       SELECT 1
                       FROM jsonb_array_elements(cond.value->'levels') AS level(value)
                       WHERE jsonb_typeof(level.value) IS DISTINCT FROM 'number'
                          OR (level.value::text)::numeric NOT IN (1, 2, 3)
                   )
               )
               OR (cond.value ? 'level_tier_ids' AND jsonb_typeof(cond.value->'level_tier_ids') IS DISTINCT FROM 'array')
        ) THEN
            RAISE EXCEPTION 'cannot migrate user level rules: announcements.targeting has malformed level condition';
        END IF;
    END IF;
END
$$;

CREATE TABLE IF NOT EXISTS user_level_rules (
    id UUID PRIMARY KEY DEFAULT public.uuid_v7(),
    name VARCHAR(100) NOT NULL,
    window_days SMALLINT NOT NULL DEFAULT 7,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT user_level_rules_window_days_check CHECK (window_days IN (7, 14, 30))
);

CREATE TABLE IF NOT EXISTS user_level_rule_tiers (
    id UUID PRIMARY KEY DEFAULT public.uuid_v7(),
    rule_id UUID NOT NULL REFERENCES user_level_rules(id) ON DELETE RESTRICT,
    name VARCHAR(100) NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    min_spend NUMERIC(20, 8) NOT NULL DEFAULT 0,
    default_multiplier NUMERIC(10, 4),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT user_level_rule_tiers_sort_order_check CHECK (sort_order >= 0),
    CONSTRAINT user_level_rule_tiers_min_spend_check CHECK (min_spend >= 0),
    CONSTRAINT user_level_rule_tiers_multiplier_check CHECK (
        default_multiplier IS NULL OR (default_multiplier >= 0.01 AND default_multiplier <= 100)
    ),
    CONSTRAINT user_level_rule_tiers_rule_order_key UNIQUE (rule_id, sort_order)
);

CREATE TABLE IF NOT EXISTS user_level_rule_assignments (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rule_id UUID NOT NULL REFERENCES user_level_rules(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, rule_id)
);

CREATE INDEX IF NOT EXISTS idx_user_level_rule_tiers_rule_order
    ON user_level_rule_tiers (rule_id, sort_order);
CREATE INDEX IF NOT EXISTS idx_user_level_rule_assignments_rule_user
    ON user_level_rule_assignments (rule_id, user_id);

COMMENT ON TABLE user_level_rules IS 'Administrator-created user level rule profiles; no implicit defaults.';
COMMENT ON COLUMN user_level_rules.window_days IS 'Rolling spend window shared by all tiers in this rule.';
COMMENT ON COLUMN user_level_rule_tiers.default_multiplier IS 'Optional user-side multiplier candidate; NULL means no candidate.';
