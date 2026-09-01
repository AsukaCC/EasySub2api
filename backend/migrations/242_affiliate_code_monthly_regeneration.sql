-- Allow users to regenerate their personal affiliate code at most three times
-- per calendar month. The active application timezone supplies the normalized
-- month boundary; counters reset lazily on the first regeneration in a month.
ALTER TABLE user_affiliates
    ADD COLUMN IF NOT EXISTS aff_code_regeneration_period_start TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS aff_code_regeneration_count INTEGER NOT NULL DEFAULT 0;

COMMENT ON COLUMN user_affiliates.aff_code_regeneration_period_start IS
    '用户自助重置邀请码的自然月起点（使用系统时区）';
COMMENT ON COLUMN user_affiliates.aff_code_regeneration_count IS
    '用户在当前自然月内自助重置邀请码的次数；管理员重置不计入';
