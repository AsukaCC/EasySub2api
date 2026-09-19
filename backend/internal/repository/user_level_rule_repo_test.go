package repository

import (
	"context"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/service"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDeleteLevelRuleRemovesChildRowsBeforeParent(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	ruleID, tierID := uuid.NewString(), uuid.NewString()
	repo := &userLevelRepository{sql: db}
	mock.ExpectBegin()
	mock.ExpectExec("pg_advisory_xact_lock").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("required_default_user_level_rule").WithArgs(ruleID).WillReturnRows(sqlmock.NewRows([]string{"default", "members"}).AddRow(false, false))
	mock.ExpectQuery("SELECT id::text FROM user_level_rules").WithArgs(ruleID).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ruleID))
	mock.ExpectQuery("SELECT COUNT").WithArgs(ruleID).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("SELECT id::text FROM user_level_rule_tiers").WithArgs(ruleID).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(tierID))
	mock.ExpectQuery("FROM settings").WithArgs(ruleID).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("level_rate_multipliers").WithArgs(tierID).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec("DELETE FROM user_level_rule_assignments").WithArgs(ruleID).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM user_level_rule_tiers").WithArgs(ruleID).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM user_level_rules").WithArgs(ruleID).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	require.NoError(t, repo.DeleteLevelRule(context.Background(), ruleID))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteLevelRuleRejectsReferencedRule(t *testing.T) {
	for _, isDefault := range []bool{true, false} {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		t.Cleanup(func() { _ = db.Close() })
		id := uuid.NewString()
		mock.ExpectBegin()
		mock.ExpectExec("pg_advisory_xact_lock").WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery("required_default_user_level_rule").WithArgs(id).WillReturnRows(sqlmock.NewRows([]string{"default", "members"}).AddRow(isDefault, true))
		mock.ExpectRollback()
		want := service.ErrUserLevelRuleHasMembers
		if isDefault {
			want = service.ErrUserLevelDefaultProtected
		}
		require.ErrorIs(t, (&userLevelRepository{sql: db}).DeleteLevelRule(context.Background(), id), want)
		require.NoError(t, mock.ExpectationsWereMet())
	}
}
