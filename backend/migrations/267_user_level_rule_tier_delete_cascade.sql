-- Tiers are owned by the rule and always exist (a base tier is created with
-- every rule). ON DELETE RESTRICT made unreferenced-rule deletion fail with
-- a raw FK 500. Cascade the owned child rows; assignments stay RESTRICT so
-- a still-assigned rule cannot be removed by a raw DELETE.
ALTER TABLE user_level_rule_tiers
    DROP CONSTRAINT IF EXISTS user_level_rule_tiers_rule_id_fkey;

ALTER TABLE user_level_rule_tiers
    ADD CONSTRAINT user_level_rule_tiers_rule_id_fkey
    FOREIGN KEY (rule_id) REFERENCES user_level_rules(id) ON DELETE CASCADE;
