ALTER TABLE groups ADD COLUMN IF NOT EXISTS i18n JSONB NOT NULL DEFAULT '{}'::jsonb;
COMMENT ON COLUMN groups.i18n IS 'Localized group name and description; missing translations use original fields';
