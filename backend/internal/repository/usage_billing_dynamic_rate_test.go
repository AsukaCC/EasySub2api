package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"github.com/AsukaCC/EasySub2api/internal/service"
)

func TestApplyDynamicRateBillingUsesMinimumSharedAndPersonalRemainder(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)

	groupID := "10000000-0000-0000-0000-000000000001"
	userID := "20000000-0000-0000-0000-000000000001"
	ruleID := "30000000-0000-0000-0000-000000000001"
	quotaKey := "2026-09-03T00:00:00Z"

	mock.ExpectExec(`(?s)INSERT INTO group_dynamic_rate_usage.*ON CONFLICT`).
		WithArgs(groupID, ruleID, quotaKey).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`(?s)SELECT used_amount.*FROM group_dynamic_rate_usage.*FOR UPDATE`).
		WithArgs(groupID, ruleID, quotaKey).
		WillReturnRows(sqlmock.NewRows([]string{"used_amount"}).AddRow(0.6))
	mock.ExpectExec(`(?s)INSERT INTO user_dynamic_rate_usage.*ON CONFLICT`).
		WithArgs(userID, groupID, ruleID, quotaKey).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`(?s)SELECT used_amount.*FROM user_dynamic_rate_usage.*FOR UPDATE`).
		WithArgs(userID, groupID, ruleID, quotaKey).
		WillReturnRows(sqlmock.NewRows([]string{"used_amount"}).AddRow(0.75))
	mock.ExpectExec(`(?s)UPDATE group_dynamic_rate_usage.*quota_key = \$5`).
		WithArgs(2.0, 0.25, groupID, ruleID, quotaKey).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`(?s)UPDATE user_dynamic_rate_usage.*quota_key = \$6`).
		WithArgs(1.0, 0.25, userID, groupID, ruleID, quotaKey).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	result := &service.UsageBillingApplyResult{}
	err = applyDynamicRateBilling(ctx, tx, &service.UsageBillingCommand{
		UserID: userID,
		DynamicRatePlan: &service.UsageDynamicRatePlan{
			GroupID:            groupID,
			StandardCost:       2,
			AccountCost:        4,
			FallbackMultiplier: 1,
			Rules: []service.UsageDynamicRateRule{
				{
					RuleID:              ruleID,
					QuotaKey:            quotaKey,
					Multiplier:          0.5,
					SharedQuotaAmount:   2,
					PersonalQuotaAmount: 1,
				},
				{
					RuleID:     "30000000-0000-0000-0000-000000000002",
					QuotaKey:   "2026-09-03T00:00:00Z",
					Multiplier: 0.8,
				},
			},
		},
	}, result)
	require.NoError(t, err)
	require.NotNil(t, result.FinalActualCost)
	require.InDelta(t, 1.5625, *result.FinalActualCost, 0.000000001)
	require.NotNil(t, result.FinalRateMultiplier)
	require.InDelta(t, 0.78125, *result.FinalRateMultiplier, 0.000000001)
	require.NoError(t, tx.Commit())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestApplyDynamicRateBillingFallsBackWhenSharedQuotaIsAlreadyExceeded(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)

	groupID := "10000000-0000-0000-0000-000000000002"
	ruleID := "30000000-0000-0000-0000-000000000003"
	quotaKey := "2026-09-03T01:00:00Z"
	mock.ExpectExec(`(?s)INSERT INTO group_dynamic_rate_usage.*ON CONFLICT`).
		WithArgs(groupID, ruleID, quotaKey).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`(?s)SELECT used_amount.*FROM group_dynamic_rate_usage.*FOR UPDATE`).
		WithArgs(groupID, ruleID, quotaKey).
		WillReturnRows(sqlmock.NewRows([]string{"used_amount"}).AddRow(0.75))
	mock.ExpectExec(`(?s)INSERT INTO user_dynamic_rate_usage.*ON CONFLICT`).
		WithArgs("20000000-0000-0000-0000-000000000002", groupID, ruleID, quotaKey).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`(?s)SELECT used_amount.*FROM user_dynamic_rate_usage.*FOR UPDATE`).
		WithArgs("20000000-0000-0000-0000-000000000002", groupID, ruleID, quotaKey).
		WillReturnRows(sqlmock.NewRows([]string{"used_amount"}).AddRow(0.0))
	mock.ExpectCommit()

	result := &service.UsageBillingApplyResult{}
	err = applyDynamicRateBilling(ctx, tx, &service.UsageBillingCommand{
		UserID: "20000000-0000-0000-0000-000000000002",
		DynamicRatePlan: &service.UsageDynamicRatePlan{
			GroupID:            groupID,
			StandardCost:       2,
			AccountCost:        4,
			FallbackMultiplier: 1,
			Rules: []service.UsageDynamicRateRule{
				{
					RuleID:            ruleID,
					QuotaKey:          quotaKey,
					Multiplier:        0.5,
					SharedQuotaAmount: 0.5,
				},
				{
					RuleID:     "30000000-0000-0000-0000-000000000004",
					QuotaKey:   quotaKey,
					Multiplier: 0.9,
				},
			},
		},
	}, result)
	require.NoError(t, err)
	require.NotNil(t, result.FinalActualCost)
	require.InDelta(t, 1.8, *result.FinalActualCost, 0.000000001)
	require.NoError(t, tx.Commit())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestApplyDynamicRateBillingConsumesAccountQuotaUSDInsteadOfDiscountedPoints(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)

	groupID := "10000000-0000-0000-0000-000000000003"
	userID := "20000000-0000-0000-0000-000000000003"
	ruleID := "30000000-0000-0000-0000-000000000005"
	quotaKey := "2026-09-04T00:00:00Z"

	mock.ExpectExec(`(?s)INSERT INTO group_dynamic_rate_usage.*ON CONFLICT`).
		WithArgs(groupID, ruleID, quotaKey).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`(?s)SELECT used_amount.*FROM group_dynamic_rate_usage.*FOR UPDATE`).
		WithArgs(groupID, ruleID, quotaKey).
		WillReturnRows(sqlmock.NewRows([]string{"used_amount"}).AddRow(0.0))
	mock.ExpectExec(`(?s)INSERT INTO user_dynamic_rate_usage.*ON CONFLICT`).
		WithArgs(userID, groupID, ruleID, quotaKey).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`(?s)SELECT used_amount.*FROM user_dynamic_rate_usage.*FOR UPDATE`).
		WithArgs(userID, groupID, ruleID, quotaKey).
		WillReturnRows(sqlmock.NewRows([]string{"used_amount"}).AddRow(0.0))
	mock.ExpectExec(`(?s)UPDATE group_dynamic_rate_usage.*quota_key = \$5`).
		WithArgs(1.0, 1.0, groupID, ruleID, quotaKey).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`(?s)UPDATE user_dynamic_rate_usage.*quota_key = \$6`).
		WithArgs(0.0, 1.0, userID, groupID, ruleID, quotaKey).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	result := &service.UsageBillingApplyResult{}
	err = applyDynamicRateBilling(ctx, tx, &service.UsageBillingCommand{
		UserID: userID,
		DynamicRatePlan: &service.UsageDynamicRatePlan{
			GroupID:            groupID,
			StandardCost:       2,
			AccountCost:        4,
			FallbackMultiplier: 1,
			Rules: []service.UsageDynamicRateRule{{
				RuleID:            ruleID,
				QuotaKey:          quotaKey,
				Multiplier:        0.5,
				SharedQuotaAmount: 1,
			}},
		},
	}, result)
	require.NoError(t, err)
	require.NotNil(t, result.FinalActualCost)
	require.InDelta(t, 1.75, *result.FinalActualCost, 0.000000001)
	require.NotNil(t, result.FinalRateMultiplier)
	require.InDelta(t, 0.875, *result.FinalRateMultiplier, 0.000000001)
	require.NoError(t, tx.Commit())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestApplyDynamicRateBillingSharesUsageAcrossUsersAndIsolatesPersonalUsage(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	groupID := "10000000-0000-0000-0000-000000000004"
	ruleID := "30000000-0000-0000-0000-000000000006"
	quotaKey := "2026-09-04T01:00:00Z"
	userIDs := []string{
		"20000000-0000-0000-0000-000000000004",
		"20000000-0000-0000-0000-000000000005",
	}

	for index, userID := range userIDs {
		mock.ExpectBegin()
		tx, beginErr := db.BeginTx(ctx, nil)
		require.NoError(t, beginErr)

		mock.ExpectExec(`(?s)INSERT INTO group_dynamic_rate_usage.*ON CONFLICT`).
			WithArgs(groupID, ruleID, quotaKey).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectQuery(`(?s)SELECT used_amount.*FROM group_dynamic_rate_usage.*FOR UPDATE`).
			WithArgs(groupID, ruleID, quotaKey).
			WillReturnRows(sqlmock.NewRows([]string{"used_amount"}).AddRow(float64(index) * 2))
		mock.ExpectExec(`(?s)INSERT INTO user_dynamic_rate_usage.*ON CONFLICT`).
			WithArgs(userID, groupID, ruleID, quotaKey).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectQuery(`(?s)SELECT used_amount.*FROM user_dynamic_rate_usage.*FOR UPDATE`).
			WithArgs(userID, groupID, ruleID, quotaKey).
			WillReturnRows(sqlmock.NewRows([]string{"used_amount"}).AddRow(0.0))
		mock.ExpectExec(`(?s)UPDATE group_dynamic_rate_usage.*quota_key = \$5`).
			WithArgs(10.0, 2.0, groupID, ruleID, quotaKey).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(`(?s)UPDATE user_dynamic_rate_usage.*quota_key = \$6`).
			WithArgs(3.0, 2.0, userID, groupID, ruleID, quotaKey).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		result := &service.UsageBillingApplyResult{}
		err = applyDynamicRateBilling(ctx, tx, &service.UsageBillingCommand{
			UserID: userID,
			DynamicRatePlan: &service.UsageDynamicRatePlan{
				GroupID: groupID, StandardCost: 2, AccountCost: 2, FallbackMultiplier: 1,
				Rules: []service.UsageDynamicRateRule{{
					RuleID: ruleID, QuotaKey: quotaKey, Multiplier: 0.5,
					SharedQuotaAmount: 10, PersonalQuotaAmount: 3,
				}},
			},
		}, result)
		require.NoError(t, err)
		require.InDelta(t, 1.0, *result.FinalActualCost, 0.000000001)
		require.NoError(t, tx.Commit())
	}

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestApplyDynamicRateBillingTracksUnlimitedSharedAndPersonalUsage(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)

	groupID := "10000000-0000-0000-0000-000000000005"
	userID := "20000000-0000-0000-0000-000000000006"
	ruleID := "30000000-0000-0000-0000-000000000007"
	quotaKey := "2026-09-04T02:00:00Z"

	mock.ExpectExec(`(?s)INSERT INTO group_dynamic_rate_usage.*ON CONFLICT`).
		WithArgs(groupID, ruleID, quotaKey).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`(?s)SELECT used_amount.*FROM group_dynamic_rate_usage.*FOR UPDATE`).
		WithArgs(groupID, ruleID, quotaKey).
		WillReturnRows(sqlmock.NewRows([]string{"used_amount"}).AddRow(5.0))
	mock.ExpectExec(`(?s)INSERT INTO user_dynamic_rate_usage.*ON CONFLICT`).
		WithArgs(userID, groupID, ruleID, quotaKey).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`(?s)SELECT used_amount.*FROM user_dynamic_rate_usage.*FOR UPDATE`).
		WithArgs(userID, groupID, ruleID, quotaKey).
		WillReturnRows(sqlmock.NewRows([]string{"used_amount"}).AddRow(1.0))
	mock.ExpectExec(`(?s)UPDATE group_dynamic_rate_usage.*quota_key = \$5`).
		WithArgs(0.0, 3.0, groupID, ruleID, quotaKey).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`(?s)UPDATE user_dynamic_rate_usage.*quota_key = \$6`).
		WithArgs(0.0, 3.0, userID, groupID, ruleID, quotaKey).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	result := &service.UsageBillingApplyResult{}
	err = applyDynamicRateBilling(ctx, tx, &service.UsageBillingCommand{
		UserID: userID,
		DynamicRatePlan: &service.UsageDynamicRatePlan{
			GroupID: groupID, StandardCost: 2, AccountCost: 3, FallbackMultiplier: 1,
			Rules: []service.UsageDynamicRateRule{{RuleID: ruleID, QuotaKey: quotaKey, Multiplier: 0.5}},
		},
	}, result)
	require.NoError(t, err)
	require.InDelta(t, 1.0, *result.FinalActualCost, 0.000000001)
	require.NoError(t, tx.Commit())
	require.NoError(t, mock.ExpectationsWereMet())
}
