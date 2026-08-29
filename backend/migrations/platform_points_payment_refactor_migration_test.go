package migrations

import (
	"strings"
	"testing"

	entsqlschema "entgo.io/ent/dialect/sql/schema"
	entmigrate "github.com/AsukaCC/EasySub2api/ent/migrate"
	"github.com/stretchr/testify/require"
)

func TestPlatformPointsPaymentRefactorMigration(t *testing.T) {
	content, err := FS.ReadFile("237_platform_points_payment_refactor.sql")
	require.NoError(t, err)

	sql := strings.Join(strings.Fields(string(content)), " ")
	require.Contains(t, sql, "ALTER COLUMN amount TYPE DECIMAL(20,8)")
	require.Contains(t, sql, "ALTER COLUMN gateway_base_amount TYPE DECIMAL(20,2)")
	require.Contains(t, sql, "ADD COLUMN IF NOT EXISTS principal_amount DECIMAL(20,2)")
	require.Contains(t, sql, "ADD COLUMN IF NOT EXISTS credited_points DECIMAL(20,8)")
	require.Contains(t, sql, "CONSTRAINT payment_orders_order_type_valid CHECK ( order_type IN ('balance', 'subscription') )")
	require.Contains(t, sql, "CREATE TABLE IF NOT EXISTS payment_refunds")
	require.Contains(t, sql, "UNIQUE(order_id, idempotency_key)")
	require.Contains(t, sql, "CONSTRAINT payment_orders_refund_totals_valid CHECK ( order_type <> 'balance' OR ( refund_amount = refunded_principal_amount")
	require.Contains(t, sql, "CONSTRAINT payment_refunds_currency_valid CHECK ( currency = 'CNY' )")
	require.Contains(t, sql, "CONSTRAINT payment_refunds_gateway_split_valid CHECK ( gateway_amount = principal_amount + fee_amount )")
	require.Contains(t, sql, "CONSTRAINT payment_refunds_bonus_expired_offset_valid CHECK ( bonus_expired_offset <= bonus_points )")
	require.Contains(t, sql, "CONSTRAINT payment_refunds_targets_nonnegative CHECK")
	require.Contains(t, sql, "CONSTRAINT payment_refunds_status_valid CHECK ( status IN ('REQUESTED', 'RESERVED', 'SUBMITTING', 'PENDING', 'SUCCEEDED', 'FAILED') )")
	require.Contains(t, sql, "CONSTRAINT payment_refunds_source_valid CHECK ( source IN ('SELF_SERVICE', 'TICKET', 'ADMIN') )")
	require.Contains(t, sql, "CREATE TABLE IF NOT EXISTS refund_tickets")
	require.Contains(t, sql, "CONSTRAINT refund_tickets_status_valid CHECK ( status IN ('PENDING', 'APPROVED', 'PROCESSING', 'COMPLETED', 'REJECTED', 'CANCELLED', 'FAILED') )")
	require.Contains(t, sql, "CREATE UNIQUE INDEX IF NOT EXISTS idx_payment_refunds_one_active_order ON payment_refunds(order_id) WHERE status IN ('REQUESTED', 'RESERVED', 'SUBMITTING', 'PENDING')")
	require.Contains(t, sql, "CREATE UNIQUE INDEX IF NOT EXISTS idx_refund_tickets_one_active_order ON refund_tickets(order_id) WHERE status IN ('PENDING', 'APPROVED', 'PROCESSING')")
	require.Contains(t, sql, "CONSTRAINT wallet_bonus_grants_lifecycle_amounts_nonnegative CHECK ( spent_amount >= 0 AND expired_amount >= 0 AND reversed_amount >= 0 )")
	require.Contains(t, sql, "CONSTRAINT wallet_bonus_grants_lifecycle_total_valid CHECK ( remaining_amount + frozen_amount + spent_amount + expired_amount + reversed_amount <= original_amount )")
	require.Contains(t, sql, "CONSTRAINT user_affiliate_ledger_reversal_amounts_nonnegative CHECK ( reserved_reversal_amount >= 0 AND reversed_amount >= 0 )")
	require.Contains(t, sql, "CONSTRAINT user_affiliate_ledger_accrue_reversal_valid CHECK ( action <> 'accrue' OR reserved_reversal_amount + reversed_amount <= amount )")
	require.Contains(t, sql, "('BALANCE_RECHARGE_MULTIPLIER', '1', NOW())")
	require.Contains(t, sql, "('affiliate_rebate_freeze_hours', '168', NOW())")
	require.Contains(t, sql, "('recharge_bonus_tiers', '[]', NOW())")
}

func TestPlatformPointsRefundIndexesAreInEntMigrationSchema(t *testing.T) {
	require.Equal(t, "order_type IN ('balance', 'subscription')", entmigrate.PaymentOrdersTable.Annotation.Checks["payment_orders_order_type_valid"])

	assertPartialUniqueIndex(t, entmigrate.PaymentRefundsTable,
		"idx_payment_refunds_one_active_order",
		"status IN ('REQUESTED', 'RESERVED', 'SUBMITTING', 'PENDING')")
	assertPartialUniqueIndex(t, entmigrate.RefundTicketsTable,
		"idx_refund_tickets_one_active_order",
		"status IN ('PENDING', 'APPROVED', 'PROCESSING')")

	require.Equal(t, "currency = 'CNY'", entmigrate.PaymentRefundsTable.Annotation.Checks["payment_refunds_currency_valid"])
	require.Equal(t, "gateway_amount = principal_amount + fee_amount", entmigrate.PaymentRefundsTable.Annotation.Checks["payment_refunds_gateway_split_valid"])
	require.Equal(t, "bonus_expired_offset <= bonus_points", entmigrate.PaymentRefundsTable.Annotation.Checks["payment_refunds_bonus_expired_offset_valid"])
}

func assertPartialUniqueIndex(t *testing.T, table *entsqlschema.Table, name, predicate string) {
	t.Helper()

	idx, ok := table.Index(name)
	require.True(t, ok, "missing Ent migration index %s", name)
	require.True(t, idx.Unique, "index %s must be unique", name)
	require.Len(t, idx.Columns, 1)
	require.Equal(t, "order_id", idx.Columns[0].Name)
	require.NotNil(t, idx.Annotation, "index %s must retain its partial predicate", name)
	require.Equal(t, predicate, idx.Annotation.Where)
}
