-- Align account profit rollups with the UUID entity identifiers introduced by
-- migration 227. The table intentionally has no foreign key, so TEXT keeps
-- old numeric history readable while allowing current UUID account IDs.
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '10min';

DO $$
DECLARE
    account_id_type TEXT;
BEGIN
    IF to_regclass('public.account_profit_daily_rollups') IS NULL THEN
        RETURN;
    END IF;

    SELECT c.data_type
    INTO account_id_type
    FROM information_schema.columns AS c
    WHERE c.table_schema = 'public'
      AND c.table_name = 'account_profit_daily_rollups'
      AND c.column_name = 'account_id';

    IF account_id_type = 'bigint' THEN
        ALTER TABLE account_profit_daily_rollups
            ALTER COLUMN account_id TYPE TEXT
            USING account_id::text;
    ELSIF account_id_type IS DISTINCT FROM 'text' THEN
        RAISE EXCEPTION
            'account_profit_daily_rollups.account_id has unsupported type %',
            account_id_type;
    END IF;
END
$$;

COMMENT ON COLUMN account_profit_daily_rollups.account_id IS
    'Account entity identifier as text; current values are UUIDs and legacy numeric history is preserved.';
