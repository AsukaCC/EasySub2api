-- Use the EasySub2API project identity for installations that still have the
-- upstream default. Custom site names are intentionally preserved.
UPDATE settings
SET value = 'EasySub2API',
    updated_at = NOW()
WHERE key = 'site_name'
  AND LOWER(BTRIM(COALESCE(value, ''))) = 'sub2api';
