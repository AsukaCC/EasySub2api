-- Affiliate invitation binding rewards and resumable legacy backfill runs.

ALTER TABLE user_affiliates
    ADD COLUMN IF NOT EXISTS inviter_bound_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS binding_reward_version SMALLINT NOT NULL DEFAULT 0;

COMMENT ON COLUMN user_affiliates.inviter_bound_at IS
    'Timestamp when the inviter relationship was established; NULL for legacy rows without a reliable timestamp';
COMMENT ON COLUMN user_affiliates.binding_reward_version IS
    '0=legacy relationship eligible for explicit backfill, 1=current binding reward policy processed';

CREATE TABLE IF NOT EXISTS affiliate_reward_backfill_runs (
    id UUID PRIMARY KEY DEFAULT public.uuid_v7(),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    preview_token VARCHAR(128) NOT NULL,
    inviter_points DECIMAL(20,8) NOT NULL DEFAULT 0,
    inviter_validity_days INT NOT NULL DEFAULT 90,
    invitee_points DECIMAL(20,8) NOT NULL DEFAULT 0,
    invitee_validity_days INT NOT NULL DEFAULT 90,
    eligible_relations INT NOT NULL DEFAULT 0,
    processed_relations INT NOT NULL DEFAULT 0,
    inviter_grants INT NOT NULL DEFAULT 0,
    invitee_grants INT NOT NULL DEFAULT 0,
    inviter_points_granted DECIMAL(20,8) NOT NULL DEFAULT 0,
    invitee_points_granted DECIMAL(20,8) NOT NULL DEFAULT 0,
    error_message TEXT NOT NULL DEFAULT '',
    created_by UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ NULL,
    completed_at TIMESTAMPTZ NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (status IN ('pending', 'running', 'completed', 'failed')),
    CHECK (inviter_points >= 0 AND invitee_points >= 0),
    CHECK (inviter_validity_days BETWEEN 1 AND 3650),
    CHECK (invitee_validity_days BETWEEN 1 AND 3650)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_affiliate_reward_backfill_active
    ON affiliate_reward_backfill_runs ((1))
    WHERE status IN ('pending', 'running');

INSERT INTO settings (key, value, updated_at)
VALUES
    ('affiliate_inviter_binding_reward_points', '0', NOW()),
    ('affiliate_inviter_binding_reward_validity_days', '90', NOW()),
    ('affiliate_invitee_binding_reward_points', '0', NOW()),
    ('affiliate_invitee_binding_reward_validity_days', '90', NOW())
ON CONFLICT (key) DO NOTHING;

INSERT INTO settings (key, value, updated_at)
VALUES ('affiliate_rebate_recipient', 'inviter', NOW())
ON CONFLICT (key) DO UPDATE SET value = 'inviter', updated_at = NOW();
