SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '5min';

ALTER TABLE subscription_plans
    ADD COLUMN IF NOT EXISTS reset_card_count INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS reset_card_validity_days INTEGER NOT NULL DEFAULT 30;

ALTER TABLE subscription_plans DROP CONSTRAINT IF EXISTS subscription_plans_reset_card_count_nonnegative;
ALTER TABLE subscription_plans ADD CONSTRAINT subscription_plans_reset_card_count_nonnegative
    CHECK (reset_card_count >= 0);
ALTER TABLE subscription_plans DROP CONSTRAINT IF EXISTS subscription_plans_reset_card_validity_positive;
ALTER TABLE subscription_plans ADD CONSTRAINT subscription_plans_reset_card_validity_positive
    CHECK (reset_card_validity_days > 0);

ALTER TABLE payment_orders
    ADD COLUMN IF NOT EXISTS subscription_reset_card_count INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS subscription_reset_card_validity_days INTEGER NOT NULL DEFAULT 30;

ALTER TABLE payment_orders DROP CONSTRAINT IF EXISTS payment_orders_reset_card_count_nonnegative;
ALTER TABLE payment_orders ADD CONSTRAINT payment_orders_reset_card_count_nonnegative
    CHECK (subscription_reset_card_count >= 0);
ALTER TABLE payment_orders DROP CONSTRAINT IF EXISTS payment_orders_reset_card_validity_positive;
ALTER TABLE payment_orders ADD CONSTRAINT payment_orders_reset_card_validity_positive
    CHECK (subscription_reset_card_validity_days > 0);

ALTER TABLE pending_subscriptions
    ADD COLUMN IF NOT EXISTS reset_card_count INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS reset_card_validity_days INTEGER NOT NULL DEFAULT 30;

ALTER TABLE pending_subscriptions DROP CONSTRAINT IF EXISTS pending_subscriptions_reset_card_count_nonnegative;
ALTER TABLE pending_subscriptions ADD CONSTRAINT pending_subscriptions_reset_card_count_nonnegative
    CHECK (reset_card_count >= 0);
ALTER TABLE pending_subscriptions DROP CONSTRAINT IF EXISTS pending_subscriptions_reset_card_validity_positive;
ALTER TABLE pending_subscriptions ADD CONSTRAINT pending_subscriptions_reset_card_validity_positive
    CHECK (reset_card_validity_days > 0);

CREATE TABLE IF NOT EXISTS subscription_reset_cards (
    id UUID PRIMARY KEY DEFAULT public.uuid_v7(),
    subscription_id UUID NOT NULL REFERENCES user_subscriptions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE RESTRICT,
    status VARCHAR(16) NOT NULL DEFAULT 'available',
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ NULL,
    issued_by UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    consumed_by UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    source_type VARCHAR(40) NOT NULL DEFAULT 'manual',
    source_id VARCHAR(128) NOT NULL DEFAULT '',
    grant_batch_id UUID NOT NULL,
    grant_index INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT subscription_reset_cards_status_valid
        CHECK (status IN ('available', 'consumed', 'expired', 'revoked')),
    CONSTRAINT subscription_reset_cards_grant_index_nonnegative
        CHECK (grant_index >= 0)
);

CREATE INDEX IF NOT EXISTS idx_subscription_reset_cards_subscription_status_expiry
    ON subscription_reset_cards(subscription_id, status, expires_at);
CREATE INDEX IF NOT EXISTS idx_subscription_reset_cards_user_status_expiry
    ON subscription_reset_cards(user_id, status, expires_at);
CREATE INDEX IF NOT EXISTS idx_subscription_reset_cards_expiry_status
    ON subscription_reset_cards(expires_at, status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_subscription_reset_cards_source_sequence
    ON subscription_reset_cards(source_type, source_id, subscription_id, grant_index);
CREATE UNIQUE INDEX IF NOT EXISTS idx_subscription_reset_cards_batch_sequence
    ON subscription_reset_cards(grant_batch_id, grant_index);
