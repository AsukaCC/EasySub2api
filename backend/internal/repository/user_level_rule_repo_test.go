package repository

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDeleteLevelRuleRemovesChildRowsBeforeParent(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	ruleID := uuid.Must(uuid.NewV7()).String()
	tierID := uuid.Must(uuid.NewV7()).String()
	repo := &userLevelRepository{sql: db}

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT id::text FROM user_level_rules`).
		WithArgs(ruleID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ruleID))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_level_rule_assignments`).
		WithArgs(ruleID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT id::text FROM user_level_rule_tiers`).
		WithArgs(ruleID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(tierID))
	mock.ExpectQuery(`level_rate_multipliers`).
		WithArgs(tierID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec(`DELETE FROM user_level_rule_assignments`).
		WithArgs(ruleID).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`DELETE FROM user_level_rule_tiers`).
		WithArgs(ruleID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM user_level_rules`).
		WithArgs(ruleID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	require.NoError(t, repo.DeleteLevelRule(context.Background(), ruleID))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteLevelRuleRejectsReferencedRule(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	ruleID := uuid.Must(uuid.NewV7()).String()
	repo := &userLevelRepository{sql: db}

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT id::text FROM user_level_rules`).
		WithArgs(ruleID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ruleID))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_level_rule_assignments`).
		WithArgs(ruleID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT id::text FROM user_level_rule_tiers`).
		WithArgs(ruleID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectRollback()

	err = repo.DeleteLevelRule(context.Background(), ruleID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "still referenced")
	require.NoError(t, mock.ExpectationsWereMet())
}
