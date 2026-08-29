-- Wallet balance buckets, expiring bonus grants, holds, and immutable ledger.

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS bonus_balance DECIMAL(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS frozen_bonus_balance DECIMAL(20,8) NOT NULL DEFAULT 0;

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_bonus_balance_nonnegative;
ALTER TABLE users ADD CONSTRAINT users_bonus_balance_nonnegative
    CHECK (bonus_balance >= 0 AND bonus_balance <= GREATEST(balance, 0));
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_frozen_bonus_balance_valid;
ALTER TABLE users ADD CONSTRAINT users_frozen_bonus_balance_valid
    CHECK (frozen_bonus_balance >= 0 AND frozen_bonus_balance <= frozen_balance);

ALTER TABLE payment_orders
    ADD COLUMN IF NOT EXISTS wallet_amount DECIMAL(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS wallet_bonus_amount DECIMAL(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS wallet_recharge_amount DECIMAL(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS wallet_hold_id UUID,
    ADD COLUMN IF NOT EXISTS wallet_only BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS gateway_base_amount DECIMAL(20,8) NOT NULL DEFAULT 0;

ALTER TABLE redeem_codes
    ADD COLUMN IF NOT EXISTS bonus_validity_days INT NOT NULL DEFAULT 90;
ALTER TABLE promo_codes
    ADD COLUMN IF NOT EXISTS bonus_validity_days INT NOT NULL DEFAULT 90;

CREATE TABLE IF NOT EXISTS wallet_bonus_grants (
    id UUID PRIMARY KEY DEFAULT public.uuid_v7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    original_amount DECIMAL(20,8) NOT NULL CHECK (original_amount > 0),
    remaining_amount DECIMAL(20,8) NOT NULL DEFAULT 0 CHECK (remaining_amount >= 0),
    frozen_amount DECIMAL(20,8) NOT NULL DEFAULT 0 CHECK (frozen_amount >= 0),
    expires_at TIMESTAMPTZ NOT NULL,
    source_type VARCHAR(40) NOT NULL,
    source_id VARCHAR(128) NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (remaining_amount + frozen_amount <= original_amount)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_wallet_bonus_grants_source
    ON wallet_bonus_grants(user_id, source_type, source_id)
    WHERE source_id <> '';
CREATE INDEX IF NOT EXISTS idx_wallet_bonus_grants_spend
    ON wallet_bonus_grants(user_id, expires_at, created_at)
    WHERE remaining_amount > 0;
CREATE INDEX IF NOT EXISTS idx_wallet_bonus_grants_expiry
    ON wallet_bonus_grants(expires_at)
    WHERE remaining_amount > 0;

CREATE TABLE IF NOT EXISTS wallet_holds (
    id UUID PRIMARY KEY DEFAULT public.uuid_v7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    purpose VARCHAR(32) NOT NULL,
    reference_id VARCHAR(128) NOT NULL,
    amount DECIMAL(20,8) NOT NULL CHECK (amount > 0),
    bonus_amount DECIMAL(20,8) NOT NULL DEFAULT 0 CHECK (bonus_amount >= 0),
    recharge_amount DECIMAL(20,8) NOT NULL DEFAULT 0 CHECK (recharge_amount >= 0),
    refunded_amount DECIMAL(20,8) NOT NULL DEFAULT 0 CHECK (refunded_amount >= 0),
    refunded_bonus_amount DECIMAL(20,8) NOT NULL DEFAULT 0 CHECK (refunded_bonus_amount >= 0),
    refunded_recharge_amount DECIMAL(20,8) NOT NULL DEFAULT 0 CHECK (refunded_recharge_amount >= 0),
    status VARCHAR(20) NOT NULL DEFAULT 'held',
    request_fingerprint VARCHAR(128) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (purpose, reference_id),
    CHECK (bonus_amount + recharge_amount = amount),
    CHECK (refunded_amount <= amount)
);

CREATE INDEX IF NOT EXISTS idx_wallet_holds_user_status ON wallet_holds(user_id, status);

CREATE TABLE IF NOT EXISTS wallet_hold_allocations (
    id UUID PRIMARY KEY DEFAULT public.uuid_v7(),
    hold_id UUID NOT NULL REFERENCES wallet_holds(id) ON DELETE CASCADE,
    bonus_grant_id UUID NOT NULL REFERENCES wallet_bonus_grants(id) ON DELETE RESTRICT,
    amount DECIMAL(20,8) NOT NULL CHECK (amount > 0),
    refunded_amount DECIMAL(20,8) NOT NULL DEFAULT 0 CHECK (refunded_amount >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (hold_id, bonus_grant_id),
    CHECK (refunded_amount <= amount)
);

CREATE TABLE IF NOT EXISTS wallet_transactions (
    id UUID PRIMARY KEY DEFAULT public.uuid_v7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action VARCHAR(32) NOT NULL,
    amount DECIMAL(20,8) NOT NULL,
    bonus_amount DECIMAL(20,8) NOT NULL DEFAULT 0,
    recharge_amount DECIMAL(20,8) NOT NULL DEFAULT 0,
    frozen_amount DECIMAL(20,8) NOT NULL DEFAULT 0,
    balance_before DECIMAL(20,8) NOT NULL,
    balance_after DECIMAL(20,8) NOT NULL,
    bonus_before DECIMAL(20,8) NOT NULL,
    bonus_after DECIMAL(20,8) NOT NULL,
    source_type VARCHAR(40) NOT NULL,
    source_id VARCHAR(128) NOT NULL DEFAULT '',
    idempotency_key VARCHAR(160) NOT NULL,
    notes TEXT NOT NULL DEFAULT '',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (idempotency_key)
);

CREATE INDEX IF NOT EXISTS idx_wallet_transactions_user_created
    ON wallet_transactions(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_source
    ON wallet_transactions(source_type, source_id);

INSERT INTO settings (key, value, updated_at)
VALUES ('bonus_balance_default_validity_days', '90', NOW())
ON CONFLICT (key) DO NOTHING;
