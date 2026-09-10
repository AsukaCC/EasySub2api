-- 将现有用户价格切换为 groups.rate_multiplier * user_group_rate_multipliers.rate_multiplier。
-- 仅初始化现有记录；后续新增用户/分组组合仍按未配置用户系数 1 处理。
UPDATE groups
SET rate_multiplier = 0.2,
    updated_at = NOW()
WHERE deleted_at IS NULL;

INSERT INTO user_group_rate_multipliers (user_id, group_id, rate_multiplier, created_at, updated_at)
SELECT u.id, g.id, 0.75, NOW(), NOW()
FROM users u
CROSS JOIN groups g
WHERE u.deleted_at IS NULL
  AND g.deleted_at IS NULL
ON CONFLICT (user_id, group_id)
DO UPDATE SET rate_multiplier = 0.75, updated_at = NOW();
