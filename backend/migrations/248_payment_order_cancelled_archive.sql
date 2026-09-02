ALTER TABLE payment_orders
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_payment_orders_user_visible_created
    ON payment_orders(user_id, created_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_payment_orders_visible_created
    ON payment_orders(created_at DESC)
    WHERE deleted_at IS NULL;

COMMENT ON COLUMN payment_orders.deleted_at IS
    'Time a cancelled order was removed from order lists; retained for delayed payment reconciliation.';
