package repository

import (
	"context"
	"github.com/AsukaCC/EasySub2api/internal/service"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSingleUserLevelBatchAssignmentUsesAtomicUpsert(t *testing.T) {
	for _, reset := range []bool{false, true} {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		t.Cleanup(func() { _ = db.Close() })
		user, target := uuid.NewString(), uuid.NewString()
		ids := []string{target}
		mock.ExpectBegin()
		mock.ExpectExec("pg_advisory_xact_lock").WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery("SELECT COUNT.*FROM users").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		if reset {
			ids = nil
			mock.ExpectQuery("required_default_user_level_rule").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(target))
		} else {
			mock.ExpectQuery("SELECT COUNT.*FROM user_level_rules").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		}
		mock.ExpectExec(`(?s)INSERT INTO user_level_rule_assignments.*ON CONFLICT \(user_id\) DO UPDATE`).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()
		n, err := (&userLevelRepository{sql: db}).BatchAssignUserLevelRules(context.Background(), []string{user}, ids, "replace")
		require.NoError(t, err)
		require.EqualValues(t, 1, n)
		require.NoError(t, mock.ExpectationsWereMet())
	}
}

func TestSingleUserLevelDynamicQuotaSplit(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	mock.ExpectBegin()
	tx, err := db.BeginTx(context.Background(), nil)
	require.NoError(t, err)
	expectDynamicRechargeWallet(mock, "user", 10, 0)
	mock.ExpectExec("INSERT INTO user_dynamic_rate_usage").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("(?s)SELECT used_amount.*FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"used_amount"}).AddRow(0.9))
	mock.ExpectExec("UPDATE user_dynamic_rate_usage").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	cmd := &service.UsageBillingCommand{UserID: "user", BalanceCost: .15, APIKeyQuotaCost: .15, APIKeyRateLimitCost: .15,
		DynamicRatePlan: &service.UsageDynamicRatePlan{GroupID: "group", StandardCost: 1, AccountCost: .2, FallbackMultiplier: .15,
			Rules: []service.UsageDynamicRateRule{{RuleID: "discount", QuotaKey: "window", DiscountCoefficient: .5, PersonalQuotaAmount: 1}}}}
	result := &service.UsageBillingApplyResult{}
	require.NoError(t, applyDynamicRateBilling(context.Background(), tx, cmd, result))
	// Half fits the remaining account-side quota: .5*.15*.5 + .5*.15.
	require.InDelta(t, .1125, *result.FinalActualCost, 1e-10)
	require.Equal(t, *result.FinalActualCost, cmd.BalanceCost)
	require.Equal(t, cmd.BalanceCost, cmd.APIKeyQuotaCost)
	require.Equal(t, cmd.BalanceCost, cmd.APIKeyRateLimitCost)
	require.NoError(t, tx.Commit())
	require.NoError(t, mock.ExpectationsWereMet())
}
