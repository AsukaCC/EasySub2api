-- Normalize account subscription tiers and replace split whitelist/mapping rules
-- with persistent, platform/tier-scoped model routing rules.
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '10min';

ALTER TABLE accounts
    ADD COLUMN IF NOT EXISTS subscription_tier VARCHAR(100),
    ADD COLUMN IF NOT EXISTS model_rule_id UUID;

ALTER TABLE account_model_rules
    ADD COLUMN IF NOT EXISTS subscription_tier VARCHAR(100),
    ADD COLUMN IF NOT EXISTS model_routes JSONB NOT NULL DEFAULT '[]'::jsonb;

CREATE OR REPLACE FUNCTION public.normalize_account_subscription_tier(value TEXT)
RETURNS TEXT
LANGUAGE sql
IMMUTABLE
AS $$
    SELECT NULLIF(
        trim(BOTH '_' FROM regexp_replace(lower(btrim(COALESCE(value, ''))), '[[:space:]_/-]+', '_', 'g')),
        ''
    );
$$;

CREATE OR REPLACE FUNCTION public.resolve_account_subscription_tier(
    account_platform TEXT,
    account_credentials JSONB,
    account_extra JSONB
)
RETURNS TEXT
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    raw_tier TEXT;
    normalized_credential TEXT;
    quota_snapshot JSONB;
    observed_at TIMESTAMPTZ;
    request_limit NUMERIC;
    token_limit NUMERIC;
    observed_model TEXT;
BEGIN
    account_credentials := COALESCE(account_credentials, '{}'::jsonb);
    account_extra := COALESCE(account_extra, '{}'::jsonb);

    IF account_platform = 'grok' THEN
        raw_tier := NULLIF(btrim(account_credentials->>'subscription_tier'), '');
        normalized_credential := public.normalize_account_subscription_tier(raw_tier);
        quota_snapshot := COALESCE(account_extra->'grok_usage_snapshot', account_extra->'grok_quota_snapshot');

        IF normalized_credential = 'supergrokpro' THEN
            observed_model := lower(COALESCE(quota_snapshot->>'model', ''));
            BEGIN
                observed_at := COALESCE(
                    NULLIF(quota_snapshot->>'plan_from_45_responses_at', '')::timestamptz,
                    NULLIF(quota_snapshot->>'last_headers_seen_at', '')::timestamptz,
                    NULLIF(quota_snapshot->>'updated_at', '')::timestamptz
                );
            EXCEPTION WHEN invalid_datetime_format THEN
                observed_at := NULL;
            END;
            request_limit := CASE WHEN COALESCE(quota_snapshot#>>'{requests,limit}', '') ~ '^[0-9]+([.][0-9]+)?$'
                THEN (quota_snapshot#>>'{requests,limit}')::numeric ELSE 0 END;
            token_limit := CASE WHEN COALESCE(quota_snapshot#>>'{tokens,limit}', '') ~ '^[0-9]+([.][0-9]+)?$'
                THEN (quota_snapshot#>>'{tokens,limit}')::numeric ELSE 0 END;
            IF observed_at IS NOT NULL
                AND observed_at >= NOW() - INTERVAL '24 hours'
                AND observed_at <= NOW() + INTERVAL '5 minutes'
                AND (
                    public.normalize_account_subscription_tier(quota_snapshot->>'plan_from_45_responses') = 'supergrok_heavy'
                    OR ((observed_model = 'grok-4.5' OR observed_model LIKE 'grok-4.5-%')
                        AND (request_limit >= 8300 OR token_limit >= 53000000))
                ) THEN
                RETURN 'supergrok_heavy';
            END IF;
            raw_tier := COALESCE(NULLIF(account_extra#>>'{grok_billing_snapshot,plan}', ''), 'supergrok');
        ELSIF normalized_credential IS NOT NULL THEN
            RETURN normalized_credential;
        ELSE
            raw_tier := COALESCE(
                NULLIF(account_extra#>>'{grok_billing_snapshot,plan}', ''),
                NULLIF(account_extra#>>'{grok_usage_snapshot,subscription_tier}', ''),
                NULLIF(account_extra#>>'{grok_quota_snapshot,subscription_tier}', ''),
                NULLIF(account_extra->>'subscription_tier', ''),
                NULLIF(account_credentials->>'plan_type', '')
            );
        END IF;
    ELSIF account_platform = 'gemini' THEN
        raw_tier := COALESCE(
            NULLIF(account_credentials->>'tier_id', ''),
            NULLIF(account_credentials->>'plan_type', ''),
            NULLIF(account_extra->>'subscription_tier', '')
        );
    ELSE
        raw_tier := COALESCE(
            NULLIF(account_credentials->>'plan_type', ''),
            NULLIF(account_credentials->>'chatgpt_plan_type', ''),
            NULLIF(account_credentials->>'subscription_tier', ''),
            NULLIF(account_extra->>'subscription_tier', ''),
            NULLIF(account_extra->>'plan_type', '')
        );
    END IF;

    RETURN public.normalize_account_subscription_tier(raw_tier);
END;
$$;

CREATE OR REPLACE FUNCTION public.enforce_account_subscription_tier()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF NEW.parent_account_id IS NOT NULL THEN
        SELECT parent.subscription_tier
        INTO NEW.subscription_tier
        FROM accounts AS parent
        WHERE parent.id = NEW.parent_account_id;
    ELSE
        NEW.subscription_tier := public.resolve_account_subscription_tier(NEW.platform, NEW.credentials, NEW.extra);
    END IF;
    RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION public.propagate_account_subscription_tier_to_shadows()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    WITH updated_shadows AS (
        UPDATE accounts AS shadow
        SET subscription_tier = NEW.subscription_tier,
            updated_at = NOW()
        WHERE shadow.parent_account_id = NEW.id
          AND shadow.subscription_tier IS DISTINCT FROM NEW.subscription_tier
        RETURNING shadow.id
    )
    INSERT INTO scheduler_outbox (event_type, account_id)
    SELECT 'account_changed', id FROM updated_shadows;
    RETURN NULL;
END;
$$;

DROP TRIGGER IF EXISTS accounts_enforce_subscription_tier ON accounts;
CREATE TRIGGER accounts_enforce_subscription_tier
BEFORE INSERT OR UPDATE OF platform, credentials, extra, parent_account_id
ON accounts
FOR EACH ROW
EXECUTE FUNCTION public.enforce_account_subscription_tier();

CREATE OR REPLACE FUNCTION public.validate_account_model_rule_binding()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    rule_platform TEXT;
    rule_tier TEXT;
BEGIN
    IF NEW.model_rule_id IS NULL OR (TG_OP = 'UPDATE' AND NEW.model_rule_id IS NOT DISTINCT FROM OLD.model_rule_id) THEN
        RETURN NEW;
    END IF;
    SELECT platform, subscription_tier
    INTO rule_platform, rule_tier
    FROM account_model_rules
    WHERE id = NEW.model_rule_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'account model rule not found' USING ERRCODE = '23503';
    END IF;
    IF rule_platform IS DISTINCT FROM NEW.platform THEN
        RAISE EXCEPTION 'account model rule platform mismatch' USING ERRCODE = '23514';
    END IF;
    IF rule_tier IS NOT NULL AND rule_tier IS DISTINCT FROM NEW.subscription_tier THEN
        RAISE EXCEPTION 'account model rule subscription tier mismatch' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS accounts_validate_model_rule_binding ON accounts;
CREATE TRIGGER accounts_validate_model_rule_binding
BEFORE INSERT OR UPDATE OF model_rule_id
ON accounts
FOR EACH ROW
EXECUTE FUNCTION public.validate_account_model_rule_binding();

DROP TRIGGER IF EXISTS accounts_propagate_subscription_tier ON accounts;
CREATE TRIGGER accounts_propagate_subscription_tier
AFTER UPDATE OF subscription_tier
ON accounts
FOR EACH ROW
WHEN (NEW.parent_account_id IS NULL AND OLD.subscription_tier IS DISTINCT FROM NEW.subscription_tier)
EXECUTE FUNCTION public.propagate_account_subscription_tier_to_shadows();

UPDATE accounts
SET subscription_tier = public.resolve_account_subscription_tier(platform, credentials, extra)
WHERE parent_account_id IS NULL;

UPDATE accounts AS shadow
SET subscription_tier = parent.subscription_tier
FROM accounts AS parent
WHERE shadow.parent_account_id = parent.id;

WITH legacy_routes AS (
    SELECT
        rule.id,
        COALESCE((
            SELECT jsonb_agg(route ORDER BY request_model)
            FROM (
                SELECT mapping_entry.key AS request_model,
                    jsonb_build_object(
                        'request_model', mapping_entry.key,
                        'upstream_model', mapping_entry.value,
                        'reasoning_effort', COALESCE(rule.reasoning_efforts->>mapping_entry.key, '')
                    ) AS route
                FROM jsonb_each_text(COALESCE(rule.mapping, '{}'::jsonb)) AS mapping_entry
                UNION ALL
                SELECT whitelist_entry.value AS request_model,
                    jsonb_build_object(
                        'request_model', whitelist_entry.value,
                        'upstream_model', whitelist_entry.value
                    ) AS route
                FROM jsonb_array_elements_text(COALESCE(rule.whitelist, '[]'::jsonb)) AS whitelist_entry
                WHERE NOT (COALESCE(rule.mapping, '{}'::jsonb) ? whitelist_entry.value)
            ) AS converted
        ), '[]'::jsonb) AS routes
    FROM account_model_rules AS rule
)
UPDATE account_model_rules AS rule
SET model_routes = legacy_routes.routes
FROM legacy_routes
WHERE rule.id = legacy_routes.id
  AND rule.model_routes = '[]'::jsonb;

ALTER TABLE accounts DROP CONSTRAINT IF EXISTS accounts_model_rule_id_fkey;
ALTER TABLE accounts
    ADD CONSTRAINT accounts_model_rule_id_fkey
    FOREIGN KEY (model_rule_id) REFERENCES account_model_rules(id) ON DELETE RESTRICT;

ALTER TABLE account_model_rules
    DROP CONSTRAINT IF EXISTS account_model_rules_platform_name_key;

CREATE UNIQUE INDEX IF NOT EXISTS account_model_rules_platform_tier_name_key
    ON account_model_rules (platform, subscription_tier, name)
    WHERE subscription_tier IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS account_model_rules_platform_generic_name_key
    ON account_model_rules (platform, name)
    WHERE subscription_tier IS NULL;

CREATE INDEX IF NOT EXISTS accounts_subscription_tier_idx ON accounts (subscription_tier);
CREATE INDEX IF NOT EXISTS accounts_model_rule_id_idx ON accounts (model_rule_id);
CREATE INDEX IF NOT EXISTS account_model_rules_scope_idx ON account_model_rules (platform, subscription_tier);
