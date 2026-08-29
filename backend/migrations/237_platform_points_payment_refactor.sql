-- Split recharge fiat from platform points and add idempotent refund records.

-- amount is the compatibility points total; gateway_base_amount is CNY fiat.
ALTER TABLE payment_orders
    ALTER COLUMN amount TYPE DECIMAL(20,8) USING amount::DECIMAL(20,8),
    ALTER COLUMN gateway_base_amount TYPE DECIMAL(20,2) USING gateway_base_amount::DECIMAL(20,2);

ALTER TABLE payment_orders
    ADD COLUMN IF NOT EXISTS principal_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS fee_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS currency VARCHAR(3) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS base_points DECIMAL(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS bonus_points DECIMAL(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS credited_points DECIMAL(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS bonus_tier_snapshot JSONB,
    ADD COLUMN IF NOT EXISTS bonus_expires_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS bonus_grant_id UUID,
    ADD COLUMN IF NOT EXISTS affiliate_rebate_points DECIMAL(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS refund_deadline TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS refunded_principal_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS refunded_fee_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS refunded_gateway_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS reversed_base_points DECIMAL(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS reversed_bonus_points DECIMAL(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS reversed_affiliate_points DECIMAL(20,8) NOT NULL DEFAULT 0;

ALTER TABLE payment_orders DROP CONSTRAINT IF EXISTS payment_orders_order_type_valid;
ALTER TABLE payment_orders ADD CONSTRAINT payment_orders_order_type_valid CHECK (
    order_type IN ('balance', 'subscription')
);
ALTER TABLE payment_orders DROP CONSTRAINT IF EXISTS payment_orders_points_nonnegative;
ALTER TABLE payment_orders ADD CONSTRAINT payment_orders_points_nonnegative CHECK (
    principal_amount >= 0 AND fee_amount >= 0 AND base_points >= 0 AND bonus_points >= 0
    AND credited_points >= 0 AND affiliate_rebate_points >= 0
    AND refunded_principal_amount >= 0 AND refunded_fee_amount >= 0 AND refunded_gateway_amount >= 0
    AND reversed_base_points >= 0 AND reversed_bonus_points >= 0 AND reversed_affiliate_points >= 0
);
ALTER TABLE payment_orders DROP CONSTRAINT IF EXISTS payment_orders_recharge_split_valid;
ALTER TABLE payment_orders ADD CONSTRAINT payment_orders_recharge_split_valid CHECK (
    order_type <> 'balance' OR credited_points = base_points + bonus_points
);
ALTER TABLE payment_orders DROP CONSTRAINT IF EXISTS payment_orders_recharge_contract_valid;
ALTER TABLE payment_orders ADD CONSTRAINT payment_orders_recharge_contract_valid CHECK (
    order_type <> 'balance' OR (
        currency = 'CNY'
        AND amount = credited_points
        AND gateway_base_amount = principal_amount
        AND base_points = principal_amount
        AND pay_amount = principal_amount + fee_amount
    )
);
ALTER TABLE payment_orders DROP CONSTRAINT IF EXISTS payment_orders_refund_totals_valid;
ALTER TABLE payment_orders ADD CONSTRAINT payment_orders_refund_totals_valid CHECK (
    order_type <> 'balance' OR (
        refund_amount = refunded_principal_amount
        AND refunded_principal_amount <= principal_amount
        AND refunded_fee_amount <= fee_amount
        AND refunded_gateway_amount = refunded_principal_amount + refunded_fee_amount
        AND reversed_base_points <= base_points
        AND reversed_bonus_points <= bonus_points
        AND reversed_affiliate_points <= affiliate_rebate_points
    )
);

CREATE TABLE IF NOT EXISTS payment_refunds (
    id UUID PRIMARY KEY DEFAULT public.uuid_v7(),
    order_id UUID NOT NULL REFERENCES payment_orders(id) ON DELETE RESTRICT,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    ticket_id UUID,
    source VARCHAR(24) NOT NULL DEFAULT 'SELF_SERVICE',
    status VARCHAR(30) NOT NULL DEFAULT 'REQUESTED',
    idempotency_key VARCHAR(160) NOT NULL,
    request_fingerprint VARCHAR(160) NOT NULL DEFAULT '',
    provider_request_id VARCHAR(128) NOT NULL,
    provider_refund_id VARCHAR(128),
    currency VARCHAR(3) NOT NULL DEFAULT 'CNY',
    requested_principal_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    principal_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    fee_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    gateway_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    base_points DECIMAL(20,8) NOT NULL DEFAULT 0,
    bonus_points DECIMAL(20,8) NOT NULL DEFAULT 0,
    affiliate_rebate_points DECIMAL(20,8) NOT NULL DEFAULT 0,
    bonus_expired_offset DECIMAL(20,8) NOT NULL DEFAULT 0,
    target_principal_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    target_fee_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    target_base_points DECIMAL(20,8) NOT NULL DEFAULT 0,
    target_bonus_points DECIMAL(20,8) NOT NULL DEFAULT 0,
    target_affiliate_points DECIMAL(20,8) NOT NULL DEFAULT 0,
    wallet_hold_id UUID,
    requested_by UUID,
    reason TEXT NOT NULL DEFAULT '',
    error_code VARCHAR(80) NOT NULL DEFAULT '',
    error_message TEXT NOT NULL DEFAULT '',
    submitted_at TIMESTAMPTZ,
    settled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(order_id, idempotency_key)
);

ALTER TABLE payment_refunds DROP CONSTRAINT IF EXISTS payment_refunds_currency_valid;
ALTER TABLE payment_refunds ADD CONSTRAINT payment_refunds_currency_valid CHECK (
    currency = 'CNY'
);
ALTER TABLE payment_refunds DROP CONSTRAINT IF EXISTS payment_refunds_amounts_nonnegative;
ALTER TABLE payment_refunds ADD CONSTRAINT payment_refunds_amounts_nonnegative CHECK (
    requested_principal_amount >= 0
    AND principal_amount >= 0 AND fee_amount >= 0 AND gateway_amount >= 0
    AND base_points >= 0 AND bonus_points >= 0 AND affiliate_rebate_points >= 0
    AND bonus_expired_offset >= 0
);
ALTER TABLE payment_refunds DROP CONSTRAINT IF EXISTS payment_refunds_gateway_split_valid;
ALTER TABLE payment_refunds ADD CONSTRAINT payment_refunds_gateway_split_valid CHECK (
    gateway_amount = principal_amount + fee_amount
);
ALTER TABLE payment_refunds DROP CONSTRAINT IF EXISTS payment_refunds_bonus_expired_offset_valid;
ALTER TABLE payment_refunds ADD CONSTRAINT payment_refunds_bonus_expired_offset_valid CHECK (
    bonus_expired_offset <= bonus_points
);
ALTER TABLE payment_refunds DROP CONSTRAINT IF EXISTS payment_refunds_targets_nonnegative;
ALTER TABLE payment_refunds ADD CONSTRAINT payment_refunds_targets_nonnegative CHECK (
    target_principal_amount >= 0 AND target_fee_amount >= 0
    AND target_base_points >= 0 AND target_bonus_points >= 0 AND target_affiliate_points >= 0
);
ALTER TABLE payment_refunds DROP CONSTRAINT IF EXISTS payment_refunds_status_valid;
ALTER TABLE payment_refunds ADD CONSTRAINT payment_refunds_status_valid CHECK (
    status IN ('REQUESTED', 'RESERVED', 'SUBMITTING', 'PENDING', 'SUCCEEDED', 'FAILED')
);
ALTER TABLE payment_refunds DROP CONSTRAINT IF EXISTS payment_refunds_source_valid;
ALTER TABLE payment_refunds ADD CONSTRAINT payment_refunds_source_valid CHECK (
    source IN ('SELF_SERVICE', 'TICKET', 'ADMIN')
);

CREATE INDEX IF NOT EXISTS idx_payment_refunds_order ON payment_refunds(order_id);
CREATE INDEX IF NOT EXISTS idx_payment_refunds_user_created ON payment_refunds(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_payment_refunds_status ON payment_refunds(status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_payment_refunds_one_active_order
    ON payment_refunds(order_id)
    WHERE status IN ('REQUESTED', 'RESERVED', 'SUBMITTING', 'PENDING');

CREATE TABLE IF NOT EXISTS refund_tickets (
    id UUID PRIMARY KEY DEFAULT public.uuid_v7(),
    order_id UUID NOT NULL REFERENCES payment_orders(id) ON DELETE RESTRICT,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    refund_id UUID REFERENCES payment_refunds(id) ON DELETE SET NULL,
    status VARCHAR(24) NOT NULL DEFAULT 'PENDING',
    comment TEXT NOT NULL DEFAULT '',
    approved_principal_amount DECIMAL(20,2),
    reviewer_id UUID,
    review_note TEXT NOT NULL DEFAULT '',
    affiliate_action VARCHAR(24) NOT NULL DEFAULT 'MANUAL',
    reviewed_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE refund_tickets DROP CONSTRAINT IF EXISTS refund_tickets_status_valid;
ALTER TABLE refund_tickets ADD CONSTRAINT refund_tickets_status_valid CHECK (
    status IN ('PENDING', 'APPROVED', 'PROCESSING', 'COMPLETED', 'REJECTED', 'CANCELLED', 'FAILED')
);
ALTER TABLE refund_tickets DROP CONSTRAINT IF EXISTS refund_tickets_approved_principal_nonnegative;
ALTER TABLE refund_tickets ADD CONSTRAINT refund_tickets_approved_principal_nonnegative CHECK (
    approved_principal_amount IS NULL OR approved_principal_amount >= 0
);
ALTER TABLE refund_tickets DROP CONSTRAINT IF EXISTS refund_tickets_affiliate_action_valid;
ALTER TABLE refund_tickets ADD CONSTRAINT refund_tickets_affiliate_action_valid CHECK (
    affiliate_action = 'MANUAL'
);

ALTER TABLE payment_refunds DROP CONSTRAINT IF EXISTS payment_refunds_ticket_id_fkey;
ALTER TABLE payment_refunds ADD CONSTRAINT payment_refunds_ticket_id_fkey
    FOREIGN KEY (ticket_id) REFERENCES refund_tickets(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_refund_tickets_order ON refund_tickets(order_id);
CREATE INDEX IF NOT EXISTS idx_refund_tickets_user_created ON refund_tickets(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_refund_tickets_status_created ON refund_tickets(status, created_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_refund_tickets_one_active_order
    ON refund_tickets(order_id)
    WHERE status IN ('PENDING', 'APPROVED', 'PROCESSING');

ALTER TABLE wallet_bonus_grants
    ADD COLUMN IF NOT EXISTS spent_amount DECIMAL(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS expired_amount DECIMAL(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS reversed_amount DECIMAL(20,8) NOT NULL DEFAULT 0;

ALTER TABLE wallet_bonus_grants DROP CONSTRAINT IF EXISTS wallet_bonus_grants_lifecycle_amounts_nonnegative;
ALTER TABLE wallet_bonus_grants ADD CONSTRAINT wallet_bonus_grants_lifecycle_amounts_nonnegative CHECK (
    spent_amount >= 0 AND expired_amount >= 0 AND reversed_amount >= 0
);
ALTER TABLE wallet_bonus_grants DROP CONSTRAINT IF EXISTS wallet_bonus_grants_lifecycle_total_valid;
ALTER TABLE wallet_bonus_grants ADD CONSTRAINT wallet_bonus_grants_lifecycle_total_valid CHECK (
    remaining_amount + frozen_amount + spent_amount + expired_amount + reversed_amount <= original_amount
);

ALTER TABLE user_affiliate_ledger
    ADD COLUMN IF NOT EXISTS reserved_reversal_amount DECIMAL(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS reversed_amount DECIMAL(20,8) NOT NULL DEFAULT 0;

ALTER TABLE user_affiliate_ledger DROP CONSTRAINT IF EXISTS user_affiliate_ledger_reversal_amounts_nonnegative;
ALTER TABLE user_affiliate_ledger ADD CONSTRAINT user_affiliate_ledger_reversal_amounts_nonnegative CHECK (
    reserved_reversal_amount >= 0 AND reversed_amount >= 0
);
ALTER TABLE user_affiliate_ledger DROP CONSTRAINT IF EXISTS user_affiliate_ledger_accrue_reversal_valid;
ALTER TABLE user_affiliate_ledger ADD CONSTRAINT user_affiliate_ledger_accrue_reversal_valid CHECK (
    action <> 'accrue' OR reserved_reversal_amount + reversed_amount <= amount
);

INSERT INTO settings (key, value, updated_at)
VALUES
    ('BALANCE_RECHARGE_MULTIPLIER', '1', NOW()),
    ('affiliate_rebate_freeze_hours', '168', NOW()),
    ('recharge_bonus_tiers', '[]', NOW())
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value, updated_at = NOW();
