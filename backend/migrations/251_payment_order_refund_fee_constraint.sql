ALTER TABLE payment_orders
    DROP CONSTRAINT IF EXISTS payment_orders_refund_totals_valid;

ALTER TABLE payment_orders
    ADD CONSTRAINT payment_orders_refund_totals_valid CHECK (
        order_type <> 'balance' OR (
            refund_amount = refunded_principal_amount
            AND refunded_principal_amount <= principal_amount
            AND refunded_fee_amount <= fee_amount
            AND refunded_gateway_amount <= refunded_principal_amount + refunded_fee_amount
            AND reversed_base_points <= base_points
            AND reversed_bonus_points <= bonus_points
            AND reversed_affiliate_points <= affiliate_rebate_points
        )
    );
