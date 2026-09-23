-- Nullable maps inherit model defaults; an explicit empty map means 1x.
-- Only backfill when introducing each column, so replay cannot revive cleared values.
DO $$
DECLARE relation_name text;
BEGIN
    FOREACH relation_name IN ARRAY ARRAY['channel_model_pricing', 'channel_account_stats_model_pricing'] LOOP
        IF NOT EXISTS (
            SELECT 1 FROM pg_attribute
            WHERE attrelid = relation_name::regclass
              AND attname = 'reasoning_effort_multipliers' AND NOT attisdropped
        ) THEN
            EXECUTE format('ALTER TABLE %I ADD COLUMN reasoning_effort_multipliers JSONB', relation_name);
            EXECUTE format('UPDATE %I SET reasoning_effort_multipliers = jsonb_build_object(''max'', max_reasoning_effort_multiplier) WHERE max_reasoning_effort_multiplier IS NOT NULL', relation_name);
            EXECUTE format('ALTER TABLE %I ADD CONSTRAINT %I CHECK (reasoning_effort_multipliers IS NULL OR jsonb_typeof(reasoning_effort_multipliers) = ''object'')', relation_name, relation_name || '_reasoning_map');
        END IF;
    END LOOP;
END $$;

UPDATE groups AS g
SET model_pricing = (
    SELECT jsonb_agg(
        CASE WHEN jsonb_typeof(entry) = 'object' AND entry ? 'max_reasoning_effort_multiplier' THEN
            (entry - 'max_reasoning_effort_multiplier') ||
            CASE WHEN NOT (entry ? 'reasoning_effort_multipliers')
                      AND jsonb_typeof(entry->'max_reasoning_effort_multiplier') = 'number'
                THEN jsonb_build_object('reasoning_effort_multipliers', jsonb_build_object('max', entry->'max_reasoning_effort_multiplier'))
                ELSE '{}'::jsonb END
        ELSE entry END ORDER BY ordinal
    )
    FROM jsonb_array_elements(g.model_pricing) WITH ORDINALITY AS pricing(entry, ordinal)
)
WHERE jsonb_typeof(g.model_pricing) = 'array'
  AND EXISTS (
      SELECT 1 FROM jsonb_array_elements(CASE WHEN jsonb_typeof(g.model_pricing) = 'array' THEN g.model_pricing ELSE '[]'::jsonb END) AS pricing(entry)
      WHERE jsonb_typeof(entry) = 'object' AND entry ? 'max_reasoning_effort_multiplier'
  );
