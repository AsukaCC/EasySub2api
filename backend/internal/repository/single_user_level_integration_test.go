//go:build integration

package repository

import (
	"context"
	"sync"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSingleUserLevelAtomicLifecycle(t *testing.T) {
	ctx := context.Background()
	repo := &userLevelRepository{sql: integrationDB}
	var original string
	require.NoError(t, integrationDB.QueryRowContext(ctx, "SELECT required_default_user_level_rule()::text").Scan(&original))
	users := []string{}
	rule := &service.UserLevelRule{ID: uuid.NewString(), Name: "Test level", WindowDays: 7, Enabled: true,
		Tiers: []service.UserLevelTier{{ID: uuid.NewString(), Name: "Base", SortOrder: 0, MinSpend: 0}}}
	require.NoError(t, repo.CreateLevelRule(ctx, rule))
	t.Cleanup(func() {
		_ = repo.SetDefaultLevelRule(ctx, original)
		for _, id := range users {
			_, _ = integrationDB.ExecContext(ctx, "DELETE FROM users WHERE id = $1::uuid", id)
		}
		_ = repo.DeleteLevelRule(ctx, rule.ID)
	})
	user := mustCreateUser(t, testEntClient(t), &service.User{})
	users = append(users, user.ID)
	assigned, err := repo.ListUserLevelRules(ctx, user.ID)
	require.NoError(t, err)
	require.Len(t, assigned, 1)
	require.Equal(t, original, assigned[0].ID)
	require.NoError(t, repo.SetDefaultLevelRule(ctx, rule.ID))
	assigned, err = repo.ListUserLevelRules(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, original, assigned[0].ID, "default switching must not move existing users")
	admin := mustCreateUser(t, testEntClient(t), &service.User{Role: "admin"})
	users = append(users, admin.ID)
	assigned, err = repo.ListUserLevelRules(ctx, admin.ID)
	require.NoError(t, err)
	require.Equal(t, rule.ID, assigned[0].ID)
	require.ErrorIs(t, repo.DeleteLevelRule(ctx, rule.ID), service.ErrUserLevelDefaultProtected)
	require.NoError(t, repo.SetDefaultLevelRule(ctx, original))
	require.ErrorIs(t, repo.DeleteLevelRule(ctx, rule.ID), service.ErrUserLevelRuleHasMembers)
	require.NoError(t, repo.ReplaceUserLevelRules(ctx, admin.ID, nil))

	// Race an assignment against disabling its target. Both operations serialize
	// before row locks; one must reject rather than leave a disabled assigned rule.
	rule.Enabled = false
	var assignmentErr, disableErr error
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		assignmentErr = repo.ReplaceUserLevelRules(ctx, user.ID, []string{rule.ID})
	}()
	go func() { defer wg.Done(); <-start; disableErr = repo.UpdateLevelRule(ctx, rule) }()
	close(start)
	wg.Wait()
	require.True(t, assignmentErr != nil || disableErr != nil)
	var count int
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_level_rule_assignments a JOIN user_level_rules r ON r.id=a.rule_id WHERE a.user_id=$1::uuid AND r.enabled`, user.ID).Scan(&count))
	require.Equal(t, 1, count)

	tx, err := integrationDB.BeginTx(ctx, nil)
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, "DELETE FROM user_level_rule_assignments WHERE user_id=$1::uuid", user.ID)
	require.NoError(t, err)
	require.Error(t, tx.Commit(), "an active user cannot commit without a rule")
	require.NoError(t, repo.ReplaceUserLevelRules(ctx, user.ID, nil))
}
