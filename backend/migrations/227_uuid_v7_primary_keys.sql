-- Replace sequence-backed entity identifiers with native PostgreSQL UUIDs.
--
-- Deployment contract:
--   * This migration is intended for a freshly reset database. It preserves
--     rows seeded by earlier migrations, but it is not a general online data
--     migration for a populated installation.
--   * Sequence-backed `id` primary keys become UUIDv7 with an application-
--     compatible database default.
--   * Numeric entity references, including denormalized columns without a
--     foreign key, are rewritten through the same legacy-ID mapping.
--   * Fixed singleton/watermark integer IDs are intentionally left unchanged.

-- PostgreSQL 15 is the supported baseline and does not provide a built-in
-- UUIDv7 generator. Construct the RFC 9562 layout from the Unix epoch in
-- milliseconds plus random bits supplied by gen_random_uuid().
CREATE OR REPLACE FUNCTION public.uuid_v7()
RETURNS uuid
LANGUAGE plpgsql
VOLATILE
SET search_path = pg_catalog
AS $function$
DECLARE
    unix_ts_ms bigint;
    random_hex text;
    uuid_hex text;
BEGIN
    unix_ts_ms := floor(extract(epoch FROM clock_timestamp()) * 1000)::bigint;
    random_hex := replace(gen_random_uuid()::text, '-', '');

    uuid_hex := lpad(to_hex(unix_ts_ms), 12, '0')
        || '7'
        || substr(random_hex, 14, 3)
        || substr(
            '89ab',
            (((get_byte(decode(substr(random_hex, 17, 2), 'hex'), 0) >> 4) & 3) + 1),
            1
        )
        || substr(random_hex, 18, 15);

    RETURN (
        substr(uuid_hex, 1, 8) || '-'
        || substr(uuid_hex, 9, 4) || '-'
        || substr(uuid_hex, 13, 4) || '-'
        || substr(uuid_hex, 17, 4) || '-'
        || substr(uuid_hex, 21, 12)
    )::uuid;
END;
$function$;

COMMENT ON FUNCTION public.uuid_v7() IS
    'Generates an RFC 9562 UUIDv7; used by sequence-free entity primary keys.';

CREATE TEMP TABLE _uuid_pk_targets (
    table_schema text NOT NULL,
    table_name text NOT NULL,
    sequence_schema text NOT NULL,
    sequence_name text NOT NULL,
    PRIMARY KEY (table_schema, table_name)
) ON COMMIT DROP;

INSERT INTO _uuid_pk_targets (table_schema, table_name, sequence_schema, sequence_name)
SELECT table_ns.nspname,
       table_class.relname,
       sequence_ns.nspname,
       sequence_class.relname
FROM pg_catalog.pg_class AS table_class
JOIN pg_catalog.pg_namespace AS table_ns
  ON table_ns.oid = table_class.relnamespace
JOIN pg_catalog.pg_attribute AS id_attribute
  ON id_attribute.attrelid = table_class.oid
 AND id_attribute.attname = 'id'
 AND NOT id_attribute.attisdropped
JOIN pg_catalog.pg_constraint AS primary_key
  ON primary_key.conrelid = table_class.oid
 AND primary_key.contype = 'p'
 AND id_attribute.attnum = ANY(primary_key.conkey)
JOIN pg_catalog.pg_class AS sequence_class
  ON sequence_class.oid = pg_catalog.to_regclass(
      pg_catalog.pg_get_serial_sequence(
          pg_catalog.format('%I.%I', table_ns.nspname, table_class.relname),
          'id'
      )
  )
JOIN pg_catalog.pg_namespace AS sequence_ns
  ON sequence_ns.oid = sequence_class.relnamespace
WHERE table_ns.nspname = 'public'
  AND table_class.relkind IN ('r', 'p')
  AND id_attribute.atttypid IN ('int2'::regtype, 'int4'::regtype, 'int8'::regtype);

-- Refuse to treat an existing installation as a fresh database. Earlier
-- migrations intentionally seed these five tables; all other sequence-backed
-- entities must be empty before UUID cutover.
DO $block$
DECLARE
    target record;
    has_rows boolean;
BEGIN
    FOR target IN
        SELECT table_schema, table_name
        FROM _uuid_pk_targets
        WHERE table_name NOT IN (
            'groups',
            'settings',
            'user_attribute_definitions',
            'ops_alert_rules',
            'channel_monitor_request_templates'
        )
        ORDER BY table_schema, table_name
    LOOP
        EXECUTE pg_catalog.format(
            'SELECT EXISTS (SELECT 1 FROM %I.%I LIMIT 1)',
            target.table_schema,
            target.table_name
        ) INTO has_rows;

        IF has_rows THEN
            RAISE EXCEPTION USING
                ERRCODE = '55000',
                MESSAGE = pg_catalog.format(
                    'migration 227 requires a reset database; table %I.%I is not empty',
                    target.table_schema,
                    target.table_name
                );
        END IF;
    END LOOP;
END;
$block$;

-- Trigger functions created by earlier migrations capture entity IDs in local
-- variables. Recompile their final definitions with uuid locals so auth-cache
-- invalidation continues to work after the identifier cutover.
CREATE OR REPLACE FUNCTION public.enqueue_user_auth_cache_invalidation()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $function$
DECLARE
    target_user_id uuid;
BEGIN
    target_user_id := OLD.id;
    IF TG_OP = 'UPDATE'
       AND OLD.status IS NOT DISTINCT FROM NEW.status
       AND OLD.role IS NOT DISTINCT FROM NEW.role
       AND OLD.deleted_at IS NOT DISTINCT FROM NEW.deleted_at THEN
        RETURN NEW;
    END IF;

    INSERT INTO auth_cache_invalidation_outbox (cache_key)
    SELECT encode(sha256(convert_to(k.key, 'UTF8')), 'hex')
    FROM api_keys AS k
    WHERE k.user_id = target_user_id
      AND k.deleted_at IS NULL
      AND k.key <> '';
    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    END IF;
    RETURN NEW;
END;
$function$;

CREATE OR REPLACE FUNCTION public.enqueue_group_auth_cache_invalidation()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $function$
DECLARE
    target_group_id uuid;
BEGIN
    target_group_id := OLD.id;
    IF TG_OP = 'UPDATE'
       AND OLD.status IS NOT DISTINCT FROM NEW.status
       AND OLD.is_exclusive IS NOT DISTINCT FROM NEW.is_exclusive
       AND OLD.allow_image_generation IS NOT DISTINCT FROM NEW.allow_image_generation
       AND OLD.platform IS NOT DISTINCT FROM NEW.platform
       AND OLD.subscription_type IS NOT DISTINCT FROM NEW.subscription_type
       AND OLD.rate_multiplier IS NOT DISTINCT FROM NEW.rate_multiplier
       AND OLD.peak_rate_enabled IS NOT DISTINCT FROM NEW.peak_rate_enabled
       AND OLD.peak_start IS NOT DISTINCT FROM NEW.peak_start
       AND OLD.peak_end IS NOT DISTINCT FROM NEW.peak_end
       AND OLD.peak_rate_multiplier IS NOT DISTINCT FROM NEW.peak_rate_multiplier
       AND OLD.profit_control_enabled IS NOT DISTINCT FROM NEW.profit_control_enabled
       AND OLD.profit_min_margin IS NOT DISTINCT FROM NEW.profit_min_margin
       AND OLD.profit_safety_buffer IS NOT DISTINCT FROM NEW.profit_safety_buffer
       AND OLD.deleted_at IS NOT DISTINCT FROM NEW.deleted_at THEN
        RETURN NEW;
    END IF;

    INSERT INTO auth_cache_invalidation_outbox (cache_key)
    SELECT encode(sha256(convert_to(k.key, 'UTF8')), 'hex')
    FROM api_keys AS k
    WHERE k.group_id = target_group_id
      AND k.deleted_at IS NULL
      AND k.key <> '';
    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    END IF;
    RETURN NEW;
END;
$function$;

CREATE OR REPLACE FUNCTION public.enqueue_allowed_group_auth_cache_invalidation()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $function$
DECLARE
    target_user_id uuid;
    target_group_id uuid;
BEGIN
    IF TG_OP = 'UPDATE'
       AND (OLD.user_id IS DISTINCT FROM NEW.user_id
            OR OLD.group_id IS DISTINCT FROM NEW.group_id) THEN
        IF EXISTS (
            SELECT 1 FROM groups g
            WHERE g.id = OLD.group_id AND g.is_exclusive = TRUE
        ) THEN
            INSERT INTO auth_cache_invalidation_outbox (cache_key)
            SELECT encode(sha256(convert_to(k.key, 'UTF8')), 'hex')
            FROM api_keys AS k
            WHERE k.user_id = OLD.user_id
              AND k.group_id = OLD.group_id
              AND k.deleted_at IS NULL
              AND k.key <> '';
        END IF;
        target_user_id := NEW.user_id;
        target_group_id := NEW.group_id;
    ELSIF TG_OP = 'UPDATE' THEN
        RETURN NEW;
    ELSIF TG_OP = 'INSERT' THEN
        target_user_id := NEW.user_id;
        target_group_id := NEW.group_id;
    ELSE
        target_user_id := OLD.user_id;
        target_group_id := OLD.group_id;
    END IF;

    IF EXISTS (
        SELECT 1 FROM groups g
        WHERE g.id = target_group_id AND g.is_exclusive = TRUE
    ) THEN
        INSERT INTO auth_cache_invalidation_outbox (cache_key)
        SELECT encode(sha256(convert_to(k.key, 'UTF8')), 'hex')
        FROM api_keys AS k
        WHERE k.user_id = target_user_id
          AND k.group_id = target_group_id
          AND k.deleted_at IS NULL
          AND k.key <> '';
    END IF;
    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    END IF;
    RETURN NEW;
END;
$function$;

-- Generate each replacement once so seeded rows and every child reference use
-- exactly the same UUID.
CREATE TEMP TABLE _uuid_pk_mapping (
    table_schema text NOT NULL,
    table_name text NOT NULL,
    legacy_id bigint NOT NULL,
    uuid_id uuid NOT NULL,
    PRIMARY KEY (table_schema, table_name, legacy_id),
    UNIQUE (table_schema, table_name, uuid_id)
) ON COMMIT DROP;

DO $block$
DECLARE
    target record;
BEGIN
    FOR target IN
        SELECT table_schema, table_name
        FROM _uuid_pk_targets
        ORDER BY table_schema, table_name
    LOOP
        EXECUTE pg_catalog.format(
            'INSERT INTO pg_temp._uuid_pk_mapping '
            || '(table_schema, table_name, legacy_id, uuid_id) '
            || 'SELECT %L, %L, id::bigint, public.uuid_v7() FROM %I.%I',
            target.table_schema,
            target.table_name,
            target.table_schema,
            target.table_name
        );
    END LOOP;
END;
$block$;

CREATE FUNCTION pg_temp.uuid_from_legacy(
    parent_schema text,
    parent_table text,
    legacy_id bigint
)
RETURNS uuid
LANGUAGE sql
STABLE
AS $function$
    SELECT CASE
        WHEN $3 IS NULL THEN NULL
        WHEN $3 = 0 THEN '00000000-0000-0000-0000-000000000000'::uuid
        ELSE (
            SELECT mapping.uuid_id
            FROM pg_temp._uuid_pk_mapping AS mapping
            WHERE mapping.table_schema = $1
              AND mapping.table_name = $2
              AND mapping.legacy_id = $3
        )
    END
$function$;

CREATE FUNCTION pg_temp.uuid_array_from_legacy(
    parent_schema text,
    parent_table text,
    legacy_ids bigint[]
)
RETURNS uuid[]
LANGUAGE sql
STABLE
AS $function$
    SELECT CASE
        WHEN $3 IS NULL THEN NULL
        ELSE ARRAY(
            SELECT pg_temp.uuid_from_legacy($1, $2, value)
            FROM unnest($3) WITH ORDINALITY AS item(value, position)
            ORDER BY position
        )
    END
$function$;

-- Save every FK that points at a sequence-backed `id`. Definitions are
-- captured verbatim so ON DELETE/UPDATE, deferrability and validation mode are
-- restored after both sides have changed type.
CREATE TEMP TABLE _uuid_fk_constraints ON COMMIT DROP AS
SELECT child_ns.nspname AS child_schema,
       child_table.relname AS child_table,
       child_attribute.attname AS child_column,
       parent_ns.nspname AS parent_schema,
       parent_table.relname AS parent_table,
       foreign_key.conname AS constraint_name,
       pg_catalog.pg_get_constraintdef(foreign_key.oid, true) AS constraint_definition
FROM pg_catalog.pg_constraint AS foreign_key
JOIN pg_catalog.pg_class AS child_table
  ON child_table.oid = foreign_key.conrelid
JOIN pg_catalog.pg_namespace AS child_ns
  ON child_ns.oid = child_table.relnamespace
JOIN pg_catalog.pg_class AS parent_table
  ON parent_table.oid = foreign_key.confrelid
JOIN pg_catalog.pg_namespace AS parent_ns
  ON parent_ns.oid = parent_table.relnamespace
JOIN _uuid_pk_targets AS target
  ON target.table_schema = parent_ns.nspname
 AND target.table_name = parent_table.relname
JOIN pg_catalog.pg_attribute AS child_attribute
  ON child_attribute.attrelid = child_table.oid
 AND child_attribute.attnum = foreign_key.conkey[1]
JOIN pg_catalog.pg_attribute AS parent_attribute
  ON parent_attribute.attrelid = parent_table.oid
 AND parent_attribute.attnum = foreign_key.confkey[1]
WHERE foreign_key.contype = 'f'
  AND pg_catalog.cardinality(foreign_key.conkey) = 1
  AND pg_catalog.cardinality(foreign_key.confkey) = 1
  AND parent_attribute.attname = 'id';

CREATE TEMP TABLE _uuid_ref_targets (
    child_schema text NOT NULL,
    child_table text NOT NULL,
    child_column text NOT NULL,
    parent_schema text NOT NULL,
    parent_table text NOT NULL,
    zero_default boolean NOT NULL DEFAULT false,
    PRIMARY KEY (child_schema, child_table, child_column)
) ON COMMIT DROP;

INSERT INTO _uuid_ref_targets (
    child_schema,
    child_table,
    child_column,
    parent_schema,
    parent_table
)
SELECT child_schema, child_table, child_column, parent_schema, parent_table
FROM _uuid_fk_constraints
ON CONFLICT (child_schema, child_table, child_column) DO NOTHING;

-- Add denormalized entity references that intentionally have no FK. Ambiguous
-- names (plan_id, rule_id, pricing_id, job_id) retain the FK-derived mapping
-- above when one exists and use this fallback only for unconstrained columns.
WITH inferred AS (
    SELECT columns.table_schema AS child_schema,
           columns.table_name AS child_table,
           columns.column_name AS child_column,
           'public'::text AS parent_schema,
           CASE
               WHEN columns.column_name IN (
                   'user_id', 'actor_user_id', 'resolved_by_user_id',
                   'deleted_key_owner_user_id', 'operator_id', 'created_by',
                   'updated_by', 'used_by', 'assigned_by', 'canceled_by',
				   'source_user_id', 'target_user_id', 'inviter_id',
				   'requested_by_user_id'
               ) THEN 'users'
               WHEN columns.column_name = 'api_key_id' THEN 'api_keys'
			   WHEN columns.column_name IN (
				   'account_id', 'parent_account_id', 'pinned_account_id',
				   'used_account_id'
			   ) THEN 'accounts'
               WHEN columns.column_name IN (
                   'group_id', 'subscription_group_id', 'fallback_group_id',
                   'fallback_group_id_on_invalid_request'
               ) THEN 'groups'
               WHEN columns.column_name IN (
                   'proxy_id', 'proxy_fallback_origin_id', 'backup_proxy_id'
               ) THEN 'proxies'
               WHEN columns.column_name = 'subscription_id' THEN 'user_subscriptions'
               WHEN columns.column_name = 'plan_id' THEN 'subscription_plans'
               WHEN columns.column_name = 'channel_id' THEN 'channels'
               WHEN columns.column_name = 'monitor_id' THEN 'channel_monitors'
               WHEN columns.column_name = 'template_id' THEN 'channel_monitor_request_templates'
               WHEN columns.column_name = 'identity_id' THEN 'auth_identities'
               WHEN columns.column_name = 'pending_auth_session_id' THEN 'pending_auth_sessions'
               WHEN columns.column_name = 'attribute_id' THEN 'user_attribute_definitions'
               WHEN columns.column_name = 'promo_code_id' THEN 'promo_codes'
               WHEN columns.column_name = 'usage_log_id' THEN 'usage_logs'
			   WHEN columns.column_name = 'source_order_id' THEN 'payment_orders'
			   WHEN columns.column_name IN ('source_error_id', 'result_error_id') THEN 'ops_error_logs'
			   WHEN columns.column_name = 'resolved_retry_id' THEN 'ops_retry_attempts'
               WHEN columns.column_name = 'job_id' THEN 'prompt_audit_jobs'
               WHEN columns.column_name = 'rule_id' THEN 'ops_alert_rules'
               WHEN columns.column_name = 'pricing_id' THEN 'channel_model_pricing'
           END AS parent_table,
           columns.column_default
	FROM information_schema.columns AS columns
	WHERE columns.table_schema = 'public'
	  AND columns.data_type IN ('smallint', 'integer', 'bigint')
	  AND NOT EXISTS (
		  SELECT 1
		  FROM pg_catalog.pg_class AS relation
		  JOIN pg_catalog.pg_namespace AS namespace
			ON namespace.oid = relation.relnamespace
		  WHERE namespace.nspname = columns.table_schema
			AND relation.relname = columns.table_name
			AND relation.relispartition
	  )
), valid_inferred AS (
    SELECT inferred.*
    FROM inferred
    JOIN _uuid_pk_targets AS parent
      ON parent.table_schema = inferred.parent_schema
     AND parent.table_name = inferred.parent_table
    WHERE inferred.parent_table IS NOT NULL
)
INSERT INTO _uuid_ref_targets (
    child_schema,
    child_table,
    child_column,
    parent_schema,
    parent_table,
    zero_default
)
SELECT child_schema,
       child_table,
       child_column,
       parent_schema,
       parent_table,
       pg_catalog.btrim(coalesce(column_default, '')) IN (
           '0', '0::smallint', '0::integer', '0::bigint'
       )
FROM valid_inferred
ON CONFLICT (child_schema, child_table, child_column) DO UPDATE
SET zero_default = EXCLUDED.zero_default;

-- Composite/aggregate tables may have no sequence-backed PK of their own. A
-- non-zero legacy reference in any such table still means this is not an empty
-- installation and must stop the migration. Zero-only analytics dimensions
-- are safe and become the nil UUID below.
DO $block$
DECLARE
    reference record;
    has_nonzero_value boolean;
BEGIN
    FOR reference IN
        SELECT child_schema, child_table, child_column
        FROM _uuid_ref_targets
        ORDER BY child_schema, child_table, child_column
    LOOP
        EXECUTE pg_catalog.format(
            'SELECT EXISTS ('
            || 'SELECT 1 FROM %I.%I '
            || 'WHERE %I IS NOT NULL AND %I <> 0 LIMIT 1)',
            reference.child_schema,
            reference.child_table,
            reference.child_column,
            reference.child_column
        ) INTO has_nonzero_value;

        IF has_nonzero_value THEN
            RAISE EXCEPTION USING
                ERRCODE = '55000',
                MESSAGE = pg_catalog.format(
                    'migration 227 requires a reset database; column %I.%I.%I contains legacy IDs',
                    reference.child_schema,
                    reference.child_table,
                    reference.child_column
                );
        END IF;
    END LOOP;

    IF EXISTS (
        SELECT 1
        FROM public.channel_account_stats_pricing_rules
        WHERE pg_catalog.cardinality(account_ids) > 0
           OR pg_catalog.cardinality(group_ids) > 0
    ) OR EXISTS (
        SELECT 1
        FROM public.channel_monitor_v2_config
        WHERE pg_catalog.cardinality(group_ids) > 0
    ) OR EXISTS (
        SELECT 1
        FROM public.groups
        WHERE model_routing IS NOT NULL
          AND model_routing <> '{}'::jsonb
    ) THEN
        RAISE EXCEPTION USING
            ERRCODE = '55000',
            MESSAGE = 'migration 227 requires empty UUID array/JSON routing references';
    END IF;
END;
$block$;

-- PostgreSQL records UPDATE OF and WHEN column dependencies in trigger
-- definitions. Even when the trigger function itself is UUID-safe, those
-- dependencies block ALTER COLUMN TYPE. Preserve every user trigger attached
-- to an affected table and restore the exact server-normalized definition
-- after the UUID cutover.
CREATE TEMP TABLE _uuid_trigger_definitions (
	table_schema text NOT NULL,
	table_name text NOT NULL,
	trigger_name text NOT NULL,
	trigger_definition text NOT NULL,
	PRIMARY KEY (table_schema, table_name, trigger_name)
) ON COMMIT DROP;

INSERT INTO _uuid_trigger_definitions (
	table_schema,
	table_name,
	trigger_name,
	trigger_definition
)
SELECT namespace.nspname,
	relation.relname,
	trigger.tgname,
	pg_catalog.pg_get_triggerdef(trigger.oid, true)
FROM pg_catalog.pg_trigger AS trigger
JOIN pg_catalog.pg_class AS relation
	ON relation.oid = trigger.tgrelid
JOIN pg_catalog.pg_namespace AS namespace
	ON namespace.oid = relation.relnamespace
WHERE NOT trigger.tgisinternal
  AND (
	  EXISTS (
		  SELECT 1
		  FROM _uuid_pk_targets AS target
		  WHERE target.table_schema = namespace.nspname
			AND target.table_name = relation.relname
	  )
	  OR EXISTS (
		  SELECT 1
		  FROM _uuid_ref_targets AS reference
		  WHERE reference.child_schema = namespace.nspname
			AND reference.child_table = relation.relname
	  )
  );

DO $block$
DECLARE
	saved_trigger record;
BEGIN
	FOR saved_trigger IN
		SELECT table_schema, table_name, trigger_name
		FROM _uuid_trigger_definitions
		ORDER BY table_schema, table_name, trigger_name
	LOOP
		EXECUTE pg_catalog.format(
			'DROP TRIGGER %I ON %I.%I',
			saved_trigger.trigger_name,
			saved_trigger.table_schema,
			saved_trigger.table_name
		);
	END LOOP;
END;
$block$;

-- Type-dependent checks and expression indexes must be removed before one
-- side of their expression changes from bigint to uuid. Recreate them after
-- both parent and reference columns have reached their final type.
ALTER TABLE public.accounts
    DROP CONSTRAINT IF EXISTS chk_accounts_parent_not_self;

DROP INDEX IF EXISTS public.idx_ops_metrics_hourly_unique_dim;
DROP INDEX IF EXISTS public.idx_ops_metrics_hourly_group_bucket;
DROP INDEX IF EXISTS public.idx_ops_metrics_daily_unique_dim;
DROP INDEX IF EXISTS public.idx_ops_metrics_daily_group_bucket;

DO $block$
DECLARE
    foreign_key record;
BEGIN
    FOR foreign_key IN
        SELECT child_schema, child_table, constraint_name
        FROM _uuid_fk_constraints
        ORDER BY child_schema, child_table, constraint_name
    LOOP
        EXECUTE pg_catalog.format(
            'ALTER TABLE %I.%I DROP CONSTRAINT %I',
            foreign_key.child_schema,
            foreign_key.child_table,
            foreign_key.constraint_name
        );
    END LOOP;
END;
$block$;

-- Convert FK and denormalized numeric references before changing parent IDs.
DO $block$
DECLARE
    reference record;
BEGIN
    FOR reference IN
        SELECT *
        FROM _uuid_ref_targets
        ORDER BY child_schema, child_table, child_column
    LOOP
        EXECUTE pg_catalog.format(
            'ALTER TABLE %I.%I ALTER COLUMN %I DROP DEFAULT',
            reference.child_schema,
            reference.child_table,
            reference.child_column
        );
        EXECUTE pg_catalog.format(
            'ALTER TABLE %I.%I ALTER COLUMN %I TYPE uuid '
            || 'USING pg_temp.uuid_from_legacy(%L, %L, %I::bigint)',
            reference.child_schema,
            reference.child_table,
            reference.child_column,
            reference.parent_schema,
            reference.parent_table,
            reference.child_column
        );

        IF reference.zero_default THEN
            EXECUTE pg_catalog.format(
                'ALTER TABLE %I.%I ALTER COLUMN %I '
                || 'SET DEFAULT %L::uuid',
                reference.child_schema,
                reference.child_table,
                reference.child_column,
                '00000000-0000-0000-0000-000000000000'
            );
        END IF;
    END LOOP;
END;
$block$;

-- Historical references stored as decimal text are internal
-- identifiers and now use native uuid like every other entity reference.
DO $block$
DECLARE
    reference record;
    current_type text;
BEGIN
    FOR reference IN
        SELECT *
		FROM (VALUES
			('public', 'payment_orders', 'provider_instance_id', 'public', 'payment_provider_instances'),
			('public', 'payment_audit_logs', 'order_id', 'public', 'payment_orders'),
			('public', 'payment_orders', 'refund_requested_by', 'public', 'users')
        ) AS refs(child_schema, child_table, child_column, parent_schema, parent_table)
    LOOP
        SELECT columns.data_type
        INTO current_type
        FROM information_schema.columns AS columns
        WHERE columns.table_schema = reference.child_schema
          AND columns.table_name = reference.child_table
          AND columns.column_name = reference.child_column;

        IF current_type IN ('character varying', 'text')
           AND EXISTS (
               SELECT 1
               FROM _uuid_pk_targets AS parent
               WHERE parent.table_schema = reference.parent_schema
                 AND parent.table_name = reference.parent_table
           ) THEN
            EXECUTE pg_catalog.format(
                'ALTER TABLE %I.%I ALTER COLUMN %I DROP DEFAULT',
                reference.child_schema,
                reference.child_table,
                reference.child_column
            );
            EXECUTE pg_catalog.format(
                'ALTER TABLE %I.%I ALTER COLUMN %I TYPE uuid USING '
                || 'CASE WHEN %I IS NULL OR pg_catalog.btrim(%I) = %L THEN NULL '
                || 'ELSE pg_temp.uuid_from_legacy(%L, %L, %I::bigint) END',
                reference.child_schema,
                reference.child_table,
                reference.child_column,
                reference.child_column,
                reference.child_column,
                '',
                reference.parent_schema,
                reference.parent_table,
                reference.child_column
            );
        END IF;
    END LOOP;
END;
$block$;

-- Array-valued selectors are also internal identifiers. Empty arrays remain
-- empty; the legacy numeric zero sentinel maps to the nil UUID.
DO $block$
DECLARE
    reference record;
    current_udt text;
BEGIN
    FOR reference IN
        SELECT *
        FROM (VALUES
            ('public', 'channel_account_stats_pricing_rules', 'account_ids', 'public', 'accounts'),
            ('public', 'channel_account_stats_pricing_rules', 'group_ids', 'public', 'groups'),
            ('public', 'channel_monitor_v2_config', 'group_ids', 'public', 'groups')
        ) AS refs(child_schema, child_table, child_column, parent_schema, parent_table)
    LOOP
        SELECT columns.udt_name
        INTO current_udt
        FROM information_schema.columns AS columns
        WHERE columns.table_schema = reference.child_schema
          AND columns.table_name = reference.child_table
          AND columns.column_name = reference.child_column;

        IF current_udt = '_int8' THEN
            EXECUTE pg_catalog.format(
                'ALTER TABLE %I.%I ALTER COLUMN %I DROP DEFAULT',
                reference.child_schema,
                reference.child_table,
                reference.child_column
            );
            EXECUTE pg_catalog.format(
                'ALTER TABLE %I.%I ALTER COLUMN %I TYPE uuid[] '
                || 'USING pg_temp.uuid_array_from_legacy(%L, %L, %I)',
                reference.child_schema,
                reference.child_table,
                reference.child_column,
                reference.parent_schema,
                reference.parent_table,
                reference.child_column
            );
            EXECUTE pg_catalog.format(
                'ALTER TABLE %I.%I ALTER COLUMN %I SET DEFAULT %L::uuid[]',
                reference.child_schema,
                reference.child_table,
                reference.child_column,
                '{}'
            );
        END IF;
    END LOOP;
END;
$block$;

-- Finally replace every sequence-backed primary key and install the same UUIDv7
-- default used by Ent.
DO $block$
DECLARE
    target record;
BEGIN
    FOR target IN
        SELECT table_schema, table_name
        FROM _uuid_pk_targets
        ORDER BY table_schema, table_name
    LOOP
        EXECUTE pg_catalog.format(
            'ALTER TABLE %I.%I ALTER COLUMN id DROP DEFAULT',
            target.table_schema,
            target.table_name
        );
        EXECUTE pg_catalog.format(
            'ALTER TABLE %I.%I ALTER COLUMN id TYPE uuid '
            || 'USING pg_temp.uuid_from_legacy(%L, %L, id::bigint)',
            target.table_schema,
            target.table_name,
            target.table_schema,
            target.table_name
        );
        EXECUTE pg_catalog.format(
            'ALTER TABLE %I.%I ALTER COLUMN id SET DEFAULT public.uuid_v7()',
            target.table_schema,
            target.table_name
        );
    END LOOP;
END;
$block$;

ALTER TABLE public.accounts
    ADD CONSTRAINT chk_accounts_parent_not_self
    CHECK (parent_account_id IS NULL OR parent_account_id <> id) NOT VALID;
ALTER TABLE public.accounts
    VALIDATE CONSTRAINT chk_accounts_parent_not_self;

CREATE UNIQUE INDEX idx_ops_metrics_hourly_unique_dim
    ON public.ops_metrics_hourly (
        bucket_start,
        COALESCE(platform, ''),
        COALESCE(group_id, '00000000-0000-0000-0000-000000000000'::uuid)
    );
CREATE INDEX idx_ops_metrics_hourly_group_bucket
    ON public.ops_metrics_hourly (group_id, bucket_start DESC)
    WHERE group_id IS NOT NULL
      AND group_id <> '00000000-0000-0000-0000-000000000000'::uuid;

CREATE UNIQUE INDEX idx_ops_metrics_daily_unique_dim
    ON public.ops_metrics_daily (
        bucket_date,
        COALESCE(platform, ''),
        COALESCE(group_id, '00000000-0000-0000-0000-000000000000'::uuid)
    );
CREATE INDEX idx_ops_metrics_daily_group_bucket
    ON public.ops_metrics_daily (group_id, bucket_date DESC)
    WHERE group_id IS NOT NULL
      AND group_id <> '00000000-0000-0000-0000-000000000000'::uuid;

DO $block$
DECLARE
    foreign_key record;
BEGIN
    FOR foreign_key IN
        SELECT *
        FROM _uuid_fk_constraints
        ORDER BY child_schema, child_table, constraint_name
    LOOP
        EXECUTE pg_catalog.format(
            'ALTER TABLE %I.%I ADD CONSTRAINT %I %s',
            foreign_key.child_schema,
            foreign_key.child_table,
            foreign_key.constraint_name,
            foreign_key.constraint_definition
        );
    END LOOP;
END;
$block$;

DO $block$
DECLARE
	saved_trigger record;
BEGIN
	FOR saved_trigger IN
		SELECT saved.trigger_definition
		FROM _uuid_trigger_definitions AS saved
		ORDER BY table_schema, table_name, trigger_name
	LOOP
		EXECUTE saved_trigger.trigger_definition;
	END LOOP;
END;
$block$;

-- No converted column depends on its legacy sequence after the new default is
-- installed. Remove the sequences so schema inspection cannot accidentally
-- treat UUID entities as serial-backed.
DO $block$
DECLARE
    target record;
BEGIN
    FOR target IN
        SELECT DISTINCT sequence_schema, sequence_name
        FROM _uuid_pk_targets
        ORDER BY sequence_schema, sequence_name
    LOOP
        EXECUTE pg_catalog.format(
            'DROP SEQUENCE IF EXISTS %I.%I',
            target.sequence_schema,
            target.sequence_name
        );
    END LOOP;
END;
$block$;
