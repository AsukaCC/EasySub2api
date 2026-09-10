-- Replace the user-level/group dynamic multiplier semantics with an explicit
-- time-window discount coefficient.  Legacy JSON remains readable during a
-- rolling deployment; the runtime treats multiplier as a compatibility alias.
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '10min';

UPDATE groups AS g
SET dynamic_rate_rules = COALESCE((
    SELECT jsonb_agg(
        (item.rule - 'multiplier' - 'level_tier_ids' - 'levels')
        || jsonb_build_object(
            'discount_coefficient',
            COALESCE(item.rule->'discount_coefficient', item.rule->'multiplier', '1'::jsonb)
        )
        ORDER BY item.ordinality
    )
    FROM jsonb_array_elements(COALESCE(g.dynamic_rate_rules, '[]'::jsonb))
        WITH ORDINALITY AS item(rule, ordinality)
), '[]'::jsonb)
WHERE jsonb_typeof(COALESCE(g.dynamic_rate_rules, '[]'::jsonb)) = 'array';

COMMENT ON COLUMN groups.dynamic_rate_rules IS
    'Time-window discount rules. Effective user rate is group rate multiplied by user rate and discount_coefficient (0.01..1).';
