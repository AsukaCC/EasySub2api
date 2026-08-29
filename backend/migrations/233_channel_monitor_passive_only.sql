-- Channel monitoring now uses gateway traffic only. Keep legacy probe data for
-- history, but retire the active scheduler and its selectable mode.
INSERT INTO settings (key, value)
VALUES ('channel_monitor_mode', 'v2')
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value,
    updated_at = NOW();

UPDATE channel_monitor_v2_config
SET enabled = TRUE,
    updated_at = NOW()
WHERE enabled = FALSE;
