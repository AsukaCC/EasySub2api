//go:build unit

package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"github.com/AsukaCC/EasySub2api/internal/service"
)

func TestUsageBillingApply_DeletedKeyStillSettles(t *testing.T) {
	for _, tc := range []struct {
		name         string
		quotaMissing bool
		rateMissing  bool
	}{
		{"deleted_before_settlement", true, true},
		{"quota_only_missing", true, false},
		{"deleted_before_rate_update", false, true},
		{"active_key", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()
			subscriptionID := "50000000-0000-0000-0000-000000000005"
			cmd := &service.UsageBillingCommand{
				RequestID:      "req-deleted-key",
				APIKeyID:       "70000000-0000-0000-0000-000000000007",
				UserID:         "20000000-0000-0000-0000-000000000042",
				AccountID:      "30000000-0000-0000-0000-000000000003",
				AccountType:    service.AccountTypeAPIKey,
				SubscriptionID: &subscriptionID, SubscriptionCost: 2.5,
				BalanceCost: 1.25, APIKeyQuotaCost: 1.25, APIKeyRateLimitCost: 1.25,
				AccountQuotaCost: 1.25,
			}
			cmd.Normalize()
			mock.ExpectBegin()
			expectDeletedKeyBillingClaim(mock, cmd)
			mock.ExpectExec(`(?s)UPDATE user_subscriptions us.*WHERE us.id = \$2.*AND us.user_id = \$3`).
				WithArgs(2.5, subscriptionID, cmd.UserID, service.SubscriptionStatusActive).
				WillReturnResult(sqlmock.NewResult(0, 1))
			expectWalletDebitQueries(mock, cmd.UserID, "wallet-usage:"+cmd.APIKeyID+":"+cmd.RequestID, 100, 98.75)
			quota := mock.ExpectQuery(`(?s)UPDATE api_keys.*SET quota_used.*WHERE id = \$2 AND deleted_at IS NULL`).
				WithArgs(1.25, cmd.APIKeyID, service.StatusAPIKeyActive, service.StatusAPIKeyQuotaExhausted)
			if tc.quotaMissing {
				quota.WillReturnError(sql.ErrNoRows)
			} else {
				quota.WillReturnRows(sqlmock.NewRows([]string{"exhausted"}).AddRow(true))
			}
			var affected int64 = 1
			if tc.rateMissing {
				affected = 0
			}
			mock.ExpectExec(`(?s)UPDATE api_keys SET.*usage_5h.*WHERE id = \$2 AND deleted_at IS NULL`).
				WithArgs(1.25, cmd.APIKeyID).WillReturnResult(sqlmock.NewResult(0, affected))
			mock.ExpectQuery(`(?s)UPDATE accounts SET extra.*WHERE id = \$2 AND deleted_at IS NULL`).
				WithArgs(1.25, cmd.AccountID).
				WillReturnRows(sqlmock.NewRows([]string{"total_used", "total_limit", "daily_used", "daily_limit", "weekly_used", "weekly_limit"}).
					AddRow(1.25, 100, 0, 0, 0, 0))
			mock.ExpectCommit()

			repo := &usageBillingRepository{db: db}
			result, err := repo.Apply(context.Background(), cmd)
			require.NoError(t, err)
			require.True(t, result.Applied)
			require.Equal(t, !tc.quotaMissing, result.APIKeyQuotaExhausted)
			require.NotNil(t, result.NewBalance)
			require.InDelta(t, 98.75, *result.NewBalance, 0.000001)
			require.False(t, result.BalanceOverdrafted)
			require.NotNil(t, result.QuotaState)
			require.Equal(t, 1.25, result.QuotaState.TotalUsed)

			// A retry must stop at the dedup claim, without repeating any charge.
			mock.ExpectBegin()
			mock.ExpectQuery(`INSERT INTO usage_billing_dedup`).
				WithArgs(cmd.RequestID, cmd.APIKeyID, cmd.RequestFingerprint).WillReturnError(sql.ErrNoRows)
			mock.ExpectQuery(`(?s)SELECT request_fingerprint.*FROM usage_billing_dedup`).
				WithArgs(cmd.RequestID, cmd.APIKeyID).
				WillReturnRows(sqlmock.NewRows([]string{"request_fingerprint"}).AddRow(cmd.RequestFingerprint))
			mock.ExpectRollback()
			result, err = repo.Apply(context.Background(), cmd)
			require.NoError(t, err)
			require.False(t, result.Applied)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUsageBillingApply_KeyDatabaseErrorsRollBack(t *testing.T) {
	for _, stage := range []string{"quota", "rate_limit"} {
		t.Run(stage, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()
			cmd := &service.UsageBillingCommand{
				RequestID:   "req-key-db-error",
				APIKeyID:    "70000000-0000-0000-0000-000000000007",
				UserID:      "20000000-0000-0000-0000-000000000042",
				BalanceCost: 1.25, APIKeyQuotaCost: 1.25, APIKeyRateLimitCost: 1.25,
			}
			cmd.Normalize()
			mock.ExpectBegin()
			expectDeletedKeyBillingClaim(mock, cmd)
			expectWalletDebitQueries(mock, cmd.UserID, "wallet-usage:"+cmd.APIKeyID+":"+cmd.RequestID, 100, 98.75)
			quota := mock.ExpectQuery(`(?s)UPDATE api_keys.*SET quota_used`).
				WithArgs(1.25, cmd.APIKeyID, service.StatusAPIKeyActive, service.StatusAPIKeyQuotaExhausted)
			if stage == "quota" {
				quota.WillReturnError(sql.ErrConnDone)
			} else {
				quota.WillReturnError(sql.ErrNoRows)
				mock.ExpectExec(`(?s)UPDATE api_keys SET.*usage_5h`).
					WithArgs(1.25, cmd.APIKeyID).WillReturnError(sql.ErrConnDone)
			}
			mock.ExpectRollback()
			result, err := (&usageBillingRepository{db: db}).Apply(context.Background(), cmd)
			require.ErrorIs(t, err, sql.ErrConnDone)
			require.Nil(t, result)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func expectDeletedKeyBillingClaim(mock sqlmock.Sqlmock, cmd *service.UsageBillingCommand) {
	mock.ExpectQuery(`INSERT INTO usage_billing_dedup`).
		WithArgs(cmd.RequestID, cmd.APIKeyID, cmd.RequestFingerprint).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("80000000-0000-0000-0000-000000000008"))
	mock.ExpectQuery(`(?s)SELECT request_fingerprint.*FROM usage_billing_dedup_archive`).
		WithArgs(cmd.RequestID, cmd.APIKeyID).WillReturnError(sql.ErrNoRows)
}
