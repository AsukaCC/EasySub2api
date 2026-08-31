SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '10min';

ALTER TABLE refund_tickets RENAME TO support_tickets;

ALTER TABLE support_tickets DROP CONSTRAINT IF EXISTS refund_tickets_status_valid;
ALTER TABLE support_tickets DROP CONSTRAINT IF EXISTS refund_tickets_approved_principal_nonnegative;
ALTER TABLE support_tickets DROP CONSTRAINT IF EXISTS refund_tickets_affiliate_action_valid;
ALTER TABLE support_tickets ALTER COLUMN order_id DROP NOT NULL;
ALTER TABLE support_tickets
    ADD COLUMN category VARCHAR(24) NOT NULL DEFAULT 'REFUND',
    ADD COLUMN origin VARCHAR(16) NOT NULL DEFAULT 'MIGRATED',
    ADD COLUMN title VARCHAR(120) NOT NULL DEFAULT '',
    ADD COLUMN api_key_id UUID,
    ADD COLUMN api_key_name_snapshot VARCHAR(160) NOT NULL DEFAULT '',
    ADD COLUMN group_id UUID,
    ADD COLUMN group_name_snapshot VARCHAR(160) NOT NULL DEFAULT '',
    ADD COLUMN refund_decision VARCHAR(16) NOT NULL DEFAULT 'PENDING',
    ADD COLUMN reopen_count INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN resolved_at TIMESTAMPTZ,
    ADD COLUMN closed_at TIMESTAMPTZ,
    ADD COLUMN last_user_activity_at TIMESTAMPTZ,
    ADD COLUMN last_admin_activity_at TIMESTAMPTZ;

UPDATE support_tickets
SET title = 'Refund request for order #' || order_id,
    refund_decision = CASE
        WHEN status IN ('APPROVED', 'PROCESSING', 'COMPLETED', 'FAILED') THEN 'APPROVED'
        WHEN status = 'REJECTED' THEN 'REJECTED'
        ELSE 'PENDING'
    END,
    resolved_at = CASE WHEN status IN ('COMPLETED', 'REJECTED') THEN COALESCE(completed_at, updated_at) END,
    closed_at = CASE WHEN status = 'CANCELLED' THEN COALESCE(completed_at, updated_at) END,
    last_user_activity_at = created_at,
    last_admin_activity_at = reviewed_at;

UPDATE support_tickets SET status = CASE status
    WHEN 'PENDING' THEN 'PENDING_ADMIN'
    WHEN 'APPROVED' THEN 'IN_PROGRESS'
    WHEN 'PROCESSING' THEN 'IN_PROGRESS'
    WHEN 'COMPLETED' THEN 'RESOLVED'
    WHEN 'REJECTED' THEN 'RESOLVED'
    WHEN 'FAILED' THEN 'PENDING_ADMIN'
    WHEN 'CANCELLED' THEN 'CANCELLED'
    ELSE 'PENDING_ADMIN'
END;

CREATE TABLE support_ticket_messages (
    id UUID PRIMARY KEY DEFAULT public.uuid_v7(),
    ticket_id UUID NOT NULL REFERENCES support_tickets(id) ON DELETE CASCADE,
    author_id UUID,
    author_role VARCHAR(16) NOT NULL,
    kind VARCHAR(16) NOT NULL DEFAULT 'COMMENT',
    body TEXT NOT NULL DEFAULT '',
    event_type VARCHAR(64) NOT NULL DEFAULT '',
    event_data JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT support_ticket_messages_author_role_valid CHECK (author_role IN ('USER', 'ADMIN', 'SYSTEM')),
    CONSTRAINT support_ticket_messages_kind_valid CHECK (kind IN ('COMMENT', 'SYSTEM'))
);

INSERT INTO support_ticket_messages (ticket_id, author_id, author_role, kind, body, created_at)
SELECT id, user_id, 'USER', 'COMMENT', comment, created_at
FROM support_tickets WHERE BTRIM(comment) <> '';

INSERT INTO support_ticket_messages (ticket_id, author_id, author_role, kind, body, created_at)
SELECT id, reviewer_id, CASE WHEN reviewer_id IS NULL THEN 'SYSTEM' ELSE 'ADMIN' END,
       CASE WHEN reviewer_id IS NULL THEN 'SYSTEM' ELSE 'COMMENT' END,
       review_note, COALESCE(reviewed_at, updated_at)
FROM support_tickets WHERE BTRIM(review_note) <> '';

CREATE INDEX idx_support_ticket_messages_ticket_created ON support_ticket_messages(ticket_id, created_at);

CREATE TABLE support_ticket_reads (
    id UUID PRIMARY KEY DEFAULT public.uuid_v7(),
    ticket_id UUID NOT NULL REFERENCES support_tickets(id) ON DELETE CASCADE,
    reader_id UUID NOT NULL,
    read_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT support_ticket_reads_unique UNIQUE (ticket_id, reader_id)
);
CREATE INDEX idx_support_ticket_reads_reader_read ON support_ticket_reads(reader_id, read_at);

DROP INDEX IF EXISTS idx_refund_tickets_order;
DROP INDEX IF EXISTS idx_refund_tickets_user_created;
DROP INDEX IF EXISTS idx_refund_tickets_status_created;
DROP INDEX IF EXISTS idx_refund_tickets_one_active_order;

ALTER TABLE support_tickets DROP COLUMN comment;
ALTER TABLE support_tickets DROP COLUMN review_note;
ALTER TABLE support_tickets DROP COLUMN affiliate_action;
ALTER TABLE support_tickets DROP COLUMN completed_at;

ALTER TABLE support_tickets ADD CONSTRAINT support_tickets_category_valid CHECK (category IN ('ACCOUNT', 'REFUND'));
ALTER TABLE support_tickets ADD CONSTRAINT support_tickets_status_valid CHECK (status IN ('PENDING_ADMIN', 'PENDING_USER', 'IN_PROGRESS', 'RESOLVED', 'CLOSED', 'CANCELLED'));
ALTER TABLE support_tickets ADD CONSTRAINT support_tickets_origin_valid CHECK (origin IN ('USER', 'ADMIN', 'MIGRATED'));
ALTER TABLE support_tickets ADD CONSTRAINT support_tickets_refund_decision_valid CHECK (refund_decision IN ('NONE', 'PENDING', 'APPROVED', 'REJECTED'));
ALTER TABLE support_tickets ADD CONSTRAINT support_tickets_reopen_count_valid CHECK (reopen_count >= 0 AND reopen_count <= 1);
ALTER TABLE support_tickets ADD CONSTRAINT support_tickets_approved_nonnegative CHECK (approved_principal_amount IS NULL OR approved_principal_amount >= 0);

CREATE INDEX idx_support_tickets_user_created ON support_tickets(user_id, created_at DESC);
CREATE INDEX idx_support_tickets_status_created ON support_tickets(status, created_at DESC);
CREATE INDEX idx_support_tickets_category_created ON support_tickets(category, created_at DESC);
CREATE INDEX idx_support_tickets_order ON support_tickets(order_id);
CREATE UNIQUE INDEX idx_support_tickets_one_active_refund ON support_tickets(order_id)
    WHERE category = 'REFUND' AND order_id IS NOT NULL AND status IN ('PENDING_ADMIN', 'PENDING_USER', 'IN_PROGRESS');

ALTER TABLE payment_refunds DROP CONSTRAINT IF EXISTS payment_refunds_ticket_id_fkey;
ALTER TABLE payment_refunds ADD CONSTRAINT payment_refunds_ticket_id_fkey
    FOREIGN KEY (ticket_id) REFERENCES support_tickets(id) ON DELETE SET NULL;

INSERT INTO settings (key, value, updated_at) VALUES
    ('support_tickets_enabled', 'false', NOW()),
    ('support_ticket_account_enabled', 'true', NOW()),
    ('support_ticket_refund_enabled', 'true', NOW())
ON CONFLICT (key) DO NOTHING;
