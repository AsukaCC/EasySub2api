-- Keep passive monitoring aligned with the production account platform catalog.
-- Existing model lists are preserved for supported platforms; retired and unknown
-- platform entries are removed.
WITH supported(platform, ordinal) AS (
    VALUES
        ('anthropic', 1),
        ('openai', 2),
        ('grok', 3),
        ('kimi', 4),
        ('zhipu', 5),
        ('deepseek', 6)
), current_platforms AS (
    SELECT entry.value, entry.value ->> 'platform' AS platform
    FROM channel_monitor_v2_config AS config
    CROSS JOIN LATERAL jsonb_array_elements(config.platforms) AS entry(value)
    WHERE config.id = 1
), rebuilt AS (
    SELECT jsonb_agg(
        COALESCE(
            current_platforms.value,
            jsonb_build_object(
                'platform', supported.platform,
                'enabled', TRUE,
                'models', '[]'::jsonb
            )
        )
        ORDER BY supported.ordinal
    ) AS platforms
    FROM supported
    LEFT JOIN current_platforms USING (platform)
)
UPDATE channel_monitor_v2_config AS config
SET platforms = rebuilt.platforms,
    version = config.version + 1,
    updated_at = NOW()
FROM rebuilt
WHERE config.id = 1
  AND config.platforms IS DISTINCT FROM rebuilt.platforms;
