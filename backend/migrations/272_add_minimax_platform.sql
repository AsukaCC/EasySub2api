-- Migration: 272_add_minimax_platform
-- 把 MiniMax 加入国产供应商平台白名单：
--   1. user_platform_quotas.platform CHECK
--   2. composite_model_routes.target_platform CHECK
--   3. channel_monitors / channel_monitor_request_templates.provider CHECK
--   4. channel_monitor_v2_config.platforms 目录补充 minimax 条目（保留既有模型列表）
--
-- 与 224/226 同型：DROP IF EXISTS 后重建超集约束，存量行瞬时校验通过。

ALTER TABLE user_platform_quotas
    DROP CONSTRAINT IF EXISTS user_platform_quotas_platform_check;

ALTER TABLE user_platform_quotas
    ADD CONSTRAINT user_platform_quotas_platform_check
    CHECK (platform IN ('anthropic', 'openai', 'gemini', 'antigravity', 'grok',
                        'kimi', 'zhipu', 'deepseek', 'minimax'));

ALTER TABLE composite_model_routes
    DROP CONSTRAINT IF EXISTS composite_model_routes_target_platform_check;

ALTER TABLE composite_model_routes
    ADD CONSTRAINT composite_model_routes_target_platform_check
    CHECK (target_platform IN ('anthropic', 'openai', 'gemini', 'antigravity', 'grok',
                               'kimi', 'zhipu', 'deepseek', 'minimax'));

DO $$
DECLARE
    monitor_constraint_def TEXT;
    template_constraint_def TEXT;
BEGIN
    SELECT pg_get_constraintdef(c.oid)
      INTO monitor_constraint_def
      FROM pg_constraint c
      JOIN pg_class t ON t.oid = c.conrelid
     WHERE t.relname = 'channel_monitors'
       AND c.conname = 'channel_monitors_provider_check';

    IF monitor_constraint_def IS NULL OR position('minimax' IN monitor_constraint_def) = 0 THEN
        ALTER TABLE channel_monitors
            DROP CONSTRAINT IF EXISTS channel_monitors_provider_check;
        ALTER TABLE channel_monitors
            ADD CONSTRAINT channel_monitors_provider_check
            CHECK (provider IN ('openai', 'anthropic', 'gemini', 'grok',
                                'antigravity', 'kimi', 'zhipu', 'deepseek', 'minimax'));
    END IF;

    SELECT pg_get_constraintdef(c.oid)
      INTO template_constraint_def
      FROM pg_constraint c
      JOIN pg_class t ON t.oid = c.conrelid
     WHERE t.relname = 'channel_monitor_request_templates'
       AND c.conname = 'channel_monitor_request_templates_provider_check';

    IF template_constraint_def IS NULL OR position('minimax' IN template_constraint_def) = 0 THEN
        ALTER TABLE channel_monitor_request_templates
            DROP CONSTRAINT IF EXISTS channel_monitor_request_templates_provider_check;
        ALTER TABLE channel_monitor_request_templates
            ADD CONSTRAINT channel_monitor_request_templates_provider_check
            CHECK (provider IN ('openai', 'anthropic', 'gemini', 'grok',
                                'antigravity', 'kimi', 'zhipu', 'deepseek', 'minimax'));
    END IF;
END $$;

-- 被动监控平台目录：与 234 同型，补充 minimax；既有平台的模型列表原样保留。
WITH supported(platform, ordinal) AS (
    VALUES
        ('anthropic', 1),
        ('openai', 2),
        ('grok', 3),
        ('kimi', 4),
        ('zhipu', 5),
        ('deepseek', 6),
        ('minimax', 7)
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
