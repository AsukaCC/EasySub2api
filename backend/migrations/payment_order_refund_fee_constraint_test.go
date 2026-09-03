package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPaymentOrderRefundFeeConstraintMigration(t *testing.T) {
	content, err := FS.ReadFile("251_payment_order_refund_fee_constraint.sql")
	require.NoError(t, err)

	sql := strings.Join(strings.Fields(string(content)), " ")
	require.Contains(t, sql, "DROP CONSTRAINT IF EXISTS payment_orders_refund_totals_valid")
	require.Contains(t, sql, "refunded_gateway_amount <= refunded_principal_amount + refunded_fee_amount")
	require.NotContains(t, sql, "refunded_gateway_amount = refunded_principal_amount + refunded_fee_amount")
}
