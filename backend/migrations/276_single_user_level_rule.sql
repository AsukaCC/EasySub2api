SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '10min';

DO $$
DECLARE
    default_id UUID;
    rule_count BIGINT;
BEGIN
    IF EXISTS (SELECT user_id FROM user_level_rule_assignments GROUP BY user_id HAVING COUNT(*) > 1) THEN
        RAISE EXCEPTION 'single user level upgrade: multiple assignments must be resolved before upgrading';
    END IF;
    IF EXISTS (SELECT 1 FROM user_level_rule_assignments a JOIN user_level_rules r ON r.id = a.rule_id WHERE NOT r.enabled) THEN
        RAISE EXCEPTION 'single user level upgrade: an assigned rule is disabled';
    END IF;
    SELECT COUNT(*) INTO rule_count FROM user_level_rules;
    IF rule_count = 0 THEN
        INSERT INTO user_level_rules (name, window_days, enabled)
        VALUES ('Default user level', 7, TRUE) RETURNING id INTO default_id;
        INSERT INTO user_level_rule_tiers (rule_id, name, sort_order, min_spend, default_multiplier)
        VALUES (default_id, 'Base', 0, 0, 1);
    ELSIF rule_count = 1 THEN
        SELECT id INTO default_id FROM user_level_rules WHERE enabled;
    ELSE
        SELECT r.id INTO default_id FROM settings s JOIN user_level_rules r
          ON r.id::text = (s.value::jsonb ->> 0)
        WHERE s.key = 'default_user_level_rule_ids' AND jsonb_array_length(s.value::jsonb) = 1 AND r.enabled;
    END IF;
    IF default_id IS NULL THEN
        RAISE EXCEPTION 'single user level upgrade: configure exactly one enabled default rule';
    END IF;
    INSERT INTO settings (key, value, updated_at)
    VALUES ('default_user_level_rule_ids', jsonb_build_array(default_id)::text, NOW())
    ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW();
    INSERT INTO user_level_rule_assignments (user_id, rule_id)
    SELECT u.id, default_id FROM users u WHERE u.deleted_at IS NULL
      AND NOT EXISTS (SELECT 1 FROM user_level_rule_assignments a WHERE a.user_id = u.id);
END $$;

ALTER TABLE user_level_rule_assignments ADD CONSTRAINT user_level_rule_one_per_user UNIQUE (user_id);
ALTER TABLE user_level_rule_tiers ALTER COLUMN default_multiplier SET DEFAULT 1;

-- Serialize administrative changes before row locks, including default changes
-- from the existing settings writer. Readers and usage recording do not lock.
CREATE FUNCTION lock_user_level_configuration() RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_advisory_xact_lock(276, 1);
    RETURN NULL;
END $$ LANGUAGE plpgsql;
CREATE TRIGGER user_level_rules_write_lock BEFORE INSERT OR UPDATE OR DELETE ON user_level_rules
FOR EACH STATEMENT EXECUTE FUNCTION lock_user_level_configuration();
CREATE TRIGGER user_level_assignments_write_lock BEFORE INSERT OR UPDATE OR DELETE ON user_level_rule_assignments
FOR EACH STATEMENT EXECUTE FUNCTION lock_user_level_configuration();
CREATE TRIGGER user_level_settings_write_lock BEFORE INSERT OR UPDATE OR DELETE ON settings
FOR EACH STATEMENT EXECUTE FUNCTION lock_user_level_configuration();

CREATE FUNCTION required_default_user_level_rule() RETURNS UUID AS $$
DECLARE
    ids JSONB;
    rule_id UUID;
BEGIN
    SELECT value::jsonb INTO ids FROM settings WHERE key = 'default_user_level_rule_ids';
    IF ids IS NULL OR jsonb_typeof(ids) <> 'array' THEN
        RAISE EXCEPTION 'USER_LEVEL_DEFAULT_INVALID' USING ERRCODE = '23514';
    END IF;
    IF jsonb_array_length(ids) <> 1 THEN
        RAISE EXCEPTION 'USER_LEVEL_DEFAULT_INVALID' USING ERRCODE = '23514';
    END IF;
    SELECT id INTO rule_id FROM user_level_rules WHERE id::text = ids ->> 0 AND enabled;
    IF rule_id IS NULL THEN
        RAISE EXCEPTION 'USER_LEVEL_DEFAULT_INVALID' USING ERRCODE = '23514';
    END IF;
    RETURN rule_id;
END $$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION assign_default_user_level_rules() RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_advisory_xact_lock(276, 1);
    INSERT INTO user_level_rule_assignments (user_id, rule_id)
    VALUES (NEW.id, required_default_user_level_rule());
    RETURN NEW;
END $$ LANGUAGE plpgsql;

CREATE FUNCTION protect_user_level_rule() RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'DELETE' OR (OLD.enabled AND NOT NEW.enabled) THEN
        IF OLD.id = required_default_user_level_rule() THEN
            RAISE EXCEPTION 'USER_LEVEL_DEFAULT_PROTECTED' USING ERRCODE = '23514';
        END IF;
        IF EXISTS (SELECT 1 FROM user_level_rule_assignments WHERE rule_id = OLD.id) THEN
            RAISE EXCEPTION 'USER_LEVEL_RULE_HAS_MEMBERS' USING ERRCODE = '23514';
        END IF;
    END IF;
    IF TG_OP = 'DELETE' THEN RETURN OLD; END IF;
    RETURN NEW;
END $$ LANGUAGE plpgsql;
CREATE TRIGGER user_level_rule_protection BEFORE UPDATE OR DELETE ON user_level_rules
FOR EACH ROW EXECUTE FUNCTION protect_user_level_rule();

CREATE FUNCTION validate_user_level_assignment() RETURNS TRIGGER AS $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM user_level_rules WHERE id = NEW.rule_id AND enabled) THEN
        RAISE EXCEPTION 'USER_LEVEL_RULE_DISABLED' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
END $$ LANGUAGE plpgsql;
CREATE TRIGGER user_level_assignment_enabled BEFORE INSERT OR UPDATE ON user_level_rule_assignments
FOR EACH ROW EXECUTE FUNCTION validate_user_level_assignment();

CREATE FUNCTION validate_default_user_level_setting() RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP <> 'INSERT' AND OLD.key = 'default_user_level_rule_ids' THEN
        PERFORM required_default_user_level_rule();
    ELSIF TG_OP <> 'DELETE' AND NEW.key = 'default_user_level_rule_ids' THEN
        PERFORM required_default_user_level_rule();
    END IF;
    RETURN NULL;
END $$ LANGUAGE plpgsql;
CREATE TRIGGER user_level_default_setting AFTER INSERT OR UPDATE OR DELETE ON settings
FOR EACH ROW EXECUTE FUNCTION validate_default_user_level_setting();

-- Deferred checks permit atomic replacements and user deletion, but never
-- permit an active user to commit without a rule (including restored users).
CREATE FUNCTION require_user_level_assignment() RETURNS TRIGGER AS $$
DECLARE
    target_id UUID;
BEGIN
    IF TG_TABLE_NAME = 'users' THEN
        target_id := NEW.id;
    ELSE
        target_id := OLD.user_id;
    END IF;
    IF EXISTS (SELECT 1 FROM users WHERE id = target_id AND deleted_at IS NULL)
       AND NOT EXISTS (SELECT 1 FROM user_level_rule_assignments WHERE user_id = target_id) THEN
        RAISE EXCEPTION 'USER_LEVEL_ASSIGNMENT_REQUIRED' USING ERRCODE = '23514';
    END IF;
    RETURN NULL;
END $$ LANGUAGE plpgsql;
CREATE CONSTRAINT TRIGGER user_level_assignment_required AFTER DELETE OR UPDATE ON user_level_rule_assignments
DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION require_user_level_assignment();
CREATE CONSTRAINT TRIGGER user_level_user_requires_assignment AFTER INSERT OR UPDATE ON users
DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION require_user_level_assignment();
