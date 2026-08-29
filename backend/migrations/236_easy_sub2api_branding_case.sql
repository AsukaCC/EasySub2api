-- Normalize only the previous project default; custom site names are preserved.
UPDATE settings
SET value = 'EasySub2api',
    updated_at = NOW()
WHERE key = 'site_name'
  AND LOWER(BTRIM(COALESCE(value, ''))) = 'easysub2api';
