-- Allow one API key to use accounts from multiple groups.
ALTER TABLE api_keys
    ADD COLUMN IF NOT EXISTS group_ids JSONB NOT NULL DEFAULT '[]'::jsonb;

-- Preserve the existing single-group binding as the primary and initial list.
UPDATE api_keys
SET group_ids = jsonb_build_array(group_id)
WHERE group_id IS NOT NULL
  AND (group_ids IS NULL OR group_ids = '[]'::jsonb);

CREATE INDEX IF NOT EXISTS idx_api_keys_group_ids_gin
    ON api_keys USING GIN (group_ids);
