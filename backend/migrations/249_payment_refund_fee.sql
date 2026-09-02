ALTER TABLE payment_refunds
    ADD COLUMN IF NOT EXISTS refund_fee_rate decimal(5,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS refund_fee_amount decimal(20,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS target_refund_fee_amount decimal(20,2) NOT NULL DEFAULT 0;

ALTER TABLE payment_refunds
    DROP CONSTRAINT IF EXISTS payment_refunds_amounts_nonnegative,
    DROP CONSTRAINT IF EXISTS payment_refunds_gateway_split_valid,
    DROP CONSTRAINT IF EXISTS payment_refunds_targets_nonnegative;

ALTER TABLE payment_refunds
    ADD CONSTRAINT payment_refunds_amounts_nonnegative CHECK (
        requested_principal_amount >= 0 AND principal_amount >= 0 AND fee_amount >= 0 AND
        refund_fee_rate >= 0 AND refund_fee_amount >= 0 AND gateway_amount >= 0 AND
        base_points >= 0 AND bonus_points >= 0 AND affiliate_rebate_points >= 0 AND bonus_expired_offset >= 0
    ),
    ADD CONSTRAINT payment_refunds_gateway_split_valid CHECK (
        gateway_amount + refund_fee_amount = principal_amount + fee_amount
    ),
    ADD CONSTRAINT payment_refunds_targets_nonnegative CHECK (
        target_principal_amount >= 0 AND target_fee_amount >= 0 AND target_refund_fee_amount >= 0 AND
        target_base_points >= 0 AND target_bonus_points >= 0 AND target_affiliate_points >= 0
    );
