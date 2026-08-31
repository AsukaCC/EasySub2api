SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '5min';

ALTER TABLE subscription_plans
    ADD COLUMN IF NOT EXISTS stock_quantity INTEGER NULL,
    ADD COLUMN IF NOT EXISTS stock_frozen INTEGER NOT NULL DEFAULT 0;

ALTER TABLE subscription_plans DROP CONSTRAINT IF EXISTS subscription_plans_stock_nonnegative;
ALTER TABLE subscription_plans ADD CONSTRAINT subscription_plans_stock_nonnegative
    CHECK (stock_quantity IS NULL OR stock_quantity >= 0);
ALTER TABLE subscription_plans DROP CONSTRAINT IF EXISTS subscription_plans_frozen_nonnegative;
ALTER TABLE subscription_plans ADD CONSTRAINT subscription_plans_frozen_nonnegative
    CHECK (stock_frozen >= 0);
ALTER TABLE subscription_plans DROP CONSTRAINT IF EXISTS subscription_plans_stock_covers_frozen;
ALTER TABLE subscription_plans ADD CONSTRAINT subscription_plans_stock_covers_frozen
    CHECK (stock_quantity IS NULL OR stock_quantity >= stock_frozen);

ALTER TABLE payment_orders
    ADD COLUMN IF NOT EXISTS inventory_status VARCHAR(16) NOT NULL DEFAULT 'NONE',
    ADD COLUMN IF NOT EXISTS inventory_reserved_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS inventory_consumed_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS inventory_released_at TIMESTAMPTZ NULL;

ALTER TABLE payment_orders DROP CONSTRAINT IF EXISTS payment_orders_inventory_status_valid;
ALTER TABLE payment_orders ADD CONSTRAINT payment_orders_inventory_status_valid
    CHECK (inventory_status IN ('NONE', 'RESERVED', 'CONSUMED', 'RELEASED'));
CREATE INDEX IF NOT EXISTS idx_payment_orders_plan_inventory
    ON payment_orders(plan_id, inventory_status);

CREATE TABLE IF NOT EXISTS pending_subscriptions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE RESTRICT,
    platform VARCHAR(50) NOT NULL,
    validity_days INTEGER NOT NULL,
    source_type VARCHAR(32) NOT NULL,
    source_id VARCHAR(128) NOT NULL DEFAULT '',
    blocked_by_subscription_id UUID NULL REFERENCES user_subscriptions(id) ON DELETE SET NULL,
    expected_activation_at TIMESTAMPTZ NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'PENDING',
    activated_subscription_id UUID NULL REFERENCES user_subscriptions(id) ON DELETE SET NULL,
    activation_mode VARCHAR(16) NOT NULL DEFAULT '',
    forfeited_subscription_ids JSONB NULL,
    activated_at TIMESTAMPTZ NULL,
    cancelled_at TIMESTAMPTZ NULL,
    last_error TEXT NOT NULL DEFAULT '',
    assigned_by UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT pending_subscriptions_status_valid CHECK (status IN ('PENDING', 'ACTIVATED', 'CANCELLED')),
    CONSTRAINT pending_subscriptions_validity_positive CHECK (validity_days > 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_pending_subscriptions_user_platform_active
    ON pending_subscriptions(user_id, platform) WHERE status = 'PENDING';
CREATE UNIQUE INDEX IF NOT EXISTS idx_pending_subscriptions_source
    ON pending_subscriptions(source_type, source_id) WHERE source_id <> '';
CREATE INDEX IF NOT EXISTS idx_pending_subscriptions_due
    ON pending_subscriptions(status, expected_activation_at);
CREATE INDEX IF NOT EXISTS idx_pending_subscriptions_group
    ON pending_subscriptions(group_id);
