package repository

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAffiliateUserOverviewSQLIncludesMaturedFrozenQuota(t *testing.T) {
	query := strings.Join(strings.Fields(affiliateUserOverviewSQL), " ")

	require.Contains(t, query, "ua.aff_quota + COALESCE(matured.matured_frozen_quota, 0) - COALESCE(reserved.reserved_reversal_quota, 0)")
	require.Contains(t, query, "frozen_until <= NOW()")
	require.Contains(t, query, "GREATEST(amount - reversed_amount, 0)")
	require.Contains(t, query, "SUM(reserved_reversal_amount)")
}

func TestAffiliateRecordQueriesUseLedgerAuditFields(t *testing.T) {
	source, err := os.ReadFile("affiliate_repo.go")
	require.NoError(t, err)
	content := string(source)

	require.Contains(t, content, "JOIN payment_orders po ON po.id = ual.source_order_id")
	require.Contains(t, content, "ual.amount::double precision")
	require.Contains(t, content, "ual.reversed_amount::double precision")
	require.Contains(t, content, "GREATEST(ual.amount - ual.reversed_amount, 0)::double precision")
	require.Contains(t, content, "ual.balance_after::double precision")
	require.NotContains(t, content, "parseAffiliateRebateAmount")
	require.NotContains(t, content, `"current_balance": "u.balance"`)
}

func TestAffiliateThawExcludesSuccessfullyReversedPoints(t *testing.T) {
	source, err := os.ReadFile("affiliate_repo.go")
	require.NoError(t, err)
	query := strings.Join(strings.Fields(string(source)), " ")

	require.Contains(t, query, "RETURNING GREATEST(amount - reversed_amount, 0) AS thaw_amount")
	require.Contains(t, query, "SELECT COALESCE(SUM(thaw_amount), 0) FROM matured")
	require.Contains(t, query, "SUM(GREATEST(amount - reversed_amount, 0))")
}

func TestAffiliateTransferClaimExcludesReservedReversals(t *testing.T) {
	query := strings.Join(strings.Fields(affiliateTransferClaimSQL), " ")

	require.Contains(t, query, "SELECT reserved_reversal_amount FROM user_affiliate_ledger")
	require.Contains(t, query, "FOR UPDATE")
	require.Contains(t, query, "GREATEST(ua.aff_quota - r.amount, 0)")
	require.Contains(t, query, "SET aff_quota = ua.aff_quota - c.amount")
}

func TestAffiliateSummaryAvailableQuotaExcludesReservedReversals(t *testing.T) {
	source, err := os.ReadFile("affiliate_repo.go")
	require.NoError(t, err)
	query := strings.Join(strings.Fields(string(source)), " ")

	require.Contains(t, query, "aff_quota - COALESCE(( SELECT SUM(ual.reserved_reversal_amount) FROM user_affiliate_ledger ual WHERE ual.user_id = user_affiliates.user_id AND ual.action = 'accrue' ), 0)")
}
