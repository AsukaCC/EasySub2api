INSERT INTO settings (key, value, updated_at)
VALUES ('default_user_level_rule_ids', '[]', NOW())
ON CONFLICT (key) DO NOTHING;

CREATE OR REPLACE FUNCTION assign_default_user_level_rules()
RETURNS TRIGGER AS $$
DECLARE
    configured_rules JSONB := '[]'::jsonb;
BEGIN
    IF NEW.role <> 'user' THEN
        RETURN NEW;
    END IF;

    BEGIN
        SELECT CASE
                 WHEN jsonb_typeof(COALESCE(NULLIF(value, ''), '[]')::jsonb) = 'array'
                 THEN COALESCE(NULLIF(value, ''), '[]')::jsonb
                 ELSE '[]'::jsonb
               END
          INTO configured_rules
          FROM settings
         WHERE key = 'default_user_level_rule_ids';
    EXCEPTION WHEN OTHERS THEN
        configured_rules := '[]'::jsonb;
    END;

    INSERT INTO user_level_rule_assignments (user_id, rule_id)
    SELECT NEW.id, rules.id
      FROM jsonb_array_elements_text(COALESCE(configured_rules, '[]'::jsonb)) AS selected(rule_id)
      JOIN user_level_rules rules ON rules.id::text = selected.rule_id
     WHERE rules.enabled = TRUE
    ON CONFLICT (user_id, rule_id) DO NOTHING;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_assign_default_user_level_rules ON users;
CREATE TRIGGER trg_assign_default_user_level_rules
AFTER INSERT ON users
FOR EACH ROW
EXECUTE FUNCTION assign_default_user_level_rules();

COMMENT ON FUNCTION assign_default_user_level_rules() IS
'Assigns configured enabled user level rules to newly inserted regular users.';
