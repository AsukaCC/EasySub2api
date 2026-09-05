-- Materialize bound account model rules into account credentials and remove
-- the reusable account model rule relation.
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '10min';

DO $$
DECLARE
    invalid_account_id TEXT;
    invalid_rule_id TEXT;
    invalid_route_account_id TEXT;
    invalid_route_rule_id TEXT;
    duplicate_account_id TEXT;
    duplicate_model TEXT;
BEGIN
    IF to_regclass('public.account_model_rules') IS NULL
       OR NOT EXISTS (
           SELECT 1
           FROM information_schema.columns
           WHERE table_schema = 'public'
             AND table_name = 'accounts'
             AND column_name = 'model_rule_id'
       ) THEN
        RETURN;
    END IF;

    SELECT a.id::text, a.model_rule_id::text
    INTO invalid_account_id, invalid_rule_id
    FROM accounts AS a
    LEFT JOIN account_model_rules AS r ON r.id = a.model_rule_id
    WHERE a.model_rule_id IS NOT NULL
      AND r.id IS NULL
    LIMIT 1;

    IF invalid_account_id IS NOT NULL THEN
        RAISE EXCEPTION
            'cannot migrate account model rules: account % references missing rule %',
            invalid_account_id,
            invalid_rule_id;
    END IF;

    SELECT a.id::text, r.id::text
    INTO invalid_route_account_id, invalid_route_rule_id
    FROM accounts AS a
    JOIN account_model_rules AS r ON r.id = a.model_rule_id
    WHERE a.model_rule_id IS NOT NULL
      AND jsonb_typeof(COALESCE(r.model_routes, '[]'::jsonb)) IS DISTINCT FROM 'array'
    LIMIT 1;

    IF invalid_route_account_id IS NOT NULL THEN
        RAISE EXCEPTION
            'cannot migrate account model rules: rule % for account % has non-array model_routes',
            invalid_route_rule_id,
            invalid_route_account_id;
    END IF;

    SELECT a.id::text, r.id::text
    INTO invalid_route_account_id, invalid_route_rule_id
    FROM accounts AS a
    JOIN account_model_rules AS r ON r.id = a.model_rule_id
    CROSS JOIN LATERAL jsonb_array_elements(COALESCE(r.model_routes, '[]'::jsonb)) AS route(value)
    WHERE a.model_rule_id IS NOT NULL
      AND (
          jsonb_typeof(route.value) IS DISTINCT FROM 'object'
          OR NULLIF(BTRIM(route.value->>'request_model'), '') IS NULL
          OR NULLIF(BTRIM(route.value->>'upstream_model'), '') IS NULL
          OR (
              POSITION('*' IN BTRIM(route.value->>'request_model')) > 0
              AND (
                  RIGHT(BTRIM(route.value->>'request_model'), 1) <> '*'
                  OR LENGTH(BTRIM(route.value->>'request_model'))
                     - LENGTH(REPLACE(BTRIM(route.value->>'request_model'), '*', '')) > 1
              )
          )
          OR POSITION('*' IN BTRIM(route.value->>'upstream_model')) > 0
          OR (
              route.value ? 'reasoning_effort'
              AND jsonb_typeof(route.value->'reasoning_effort') IS DISTINCT FROM 'string'
          )
          OR (
              NULLIF(BTRIM(route.value->>'reasoning_effort'), '') IS NOT NULL
              AND (
                  a.platform <> 'openai'
                  OR regexp_replace(
                      LOWER(BTRIM(route.value->>'reasoning_effort')),
                      '[-_[:space:]]',
                      '',
                      'g'
                  ) NOT IN ('minimal', 'low', 'medium', 'high', 'xhigh', 'max')
              )
          )
      )
    LIMIT 1;

    IF invalid_route_account_id IS NOT NULL THEN
        RAISE EXCEPTION
            'cannot migrate account model rules: rule % for account % contains an invalid route',
            invalid_route_rule_id,
            invalid_route_account_id;
    END IF;

    SELECT a.id::text, NULLIF(BTRIM(route.value->>'request_model'), '')
    INTO duplicate_account_id, duplicate_model
    FROM accounts AS a
    JOIN account_model_rules AS r ON r.id = a.model_rule_id
    CROSS JOIN LATERAL jsonb_array_elements(COALESCE(r.model_routes, '[]'::jsonb)) AS route(value)
    WHERE a.model_rule_id IS NOT NULL
    GROUP BY a.id, NULLIF(BTRIM(route.value->>'request_model'), '')
    HAVING COUNT(*) > 1
    LIMIT 1;

    IF duplicate_account_id IS NOT NULL THEN
        RAISE EXCEPTION
            'cannot migrate account model rules: account % has duplicate request model %',
            duplicate_account_id,
            duplicate_model;
    END IF;
END
$$;

WITH route_rows AS (
    SELECT
        a.id AS account_id,
        a.platform,
        NULLIF(BTRIM(route.value->>'request_model'), '') AS request_model,
        NULLIF(BTRIM(route.value->>'upstream_model'), '') AS upstream_model,
        NULLIF(BTRIM(route.value->>'reasoning_effort'), '') AS reasoning_effort
    FROM accounts AS a
    JOIN account_model_rules AS r ON r.id = a.model_rule_id
    CROSS JOIN LATERAL jsonb_array_elements(COALESCE(r.model_routes, '[]'::jsonb)) AS route(value)
    WHERE a.model_rule_id IS NOT NULL
), route_maps AS (
    SELECT
        account_id,
        jsonb_object_agg(request_model, upstream_model) AS model_mapping,
        jsonb_object_agg(request_model, reasoning_effort)
            FILTER (WHERE platform = 'openai' AND reasoning_effort IS NOT NULL) AS model_reasoning_efforts
    FROM route_rows
    GROUP BY account_id
)
UPDATE accounts AS a
SET credentials =
    (COALESCE(a.credentials, '{}'::jsonb) - 'model_mapping' - 'model_reasoning_efforts')
    || COALESCE((
        SELECT jsonb_build_object('model_mapping', rm.model_mapping)
               || CASE
                   WHEN rm.model_reasoning_efforts IS NULL THEN '{}'::jsonb
                   ELSE jsonb_build_object('model_reasoning_efforts', rm.model_reasoning_efforts)
                  END
        FROM route_maps AS rm
        WHERE rm.account_id = a.id
    ), '{}'::jsonb)
WHERE a.model_rule_id IS NOT NULL
  AND EXISTS (
      SELECT 1
      FROM account_model_rules AS r
      WHERE r.id = a.model_rule_id
  );

UPDATE accounts
SET model_rule_id = NULL
WHERE model_rule_id IS NOT NULL;

DROP TRIGGER IF EXISTS accounts_validate_model_rule_binding ON accounts;
DROP FUNCTION IF EXISTS public.validate_account_model_rule_binding();

ALTER TABLE accounts
    DROP CONSTRAINT IF EXISTS accounts_model_rule_id_fkey;

DROP INDEX IF EXISTS accounts_model_rule_id_idx;

ALTER TABLE accounts
    DROP COLUMN IF EXISTS model_rule_id;

DROP TABLE IF EXISTS account_model_rules;
