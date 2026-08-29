-- Separate runtime enablement from user-facing publication for staged feature rollout.
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '10min';

INSERT INTO settings (key, value, updated_at)
SELECT visibility.key, COALESCE(runtime.value, visibility.default_value), NOW()
FROM (VALUES
    ('channel_monitor_user_visible', 'channel_monitor_enabled', 'true'),
    ('available_channels_user_visible', 'available_channels_enabled', 'false'),
    ('model_plaza_user_visible', 'model_plaza_enabled', 'false'),
    ('payment_user_visible', 'payment_enabled', 'false'),
    ('affiliate_user_visible', 'affiliate_enabled', 'false')
) AS visibility(key, runtime_key, default_value)
LEFT JOIN settings AS runtime ON runtime.key = visibility.runtime_key
ON CONFLICT (key) DO NOTHING;
