package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type stubLevelRulesRepo struct {
	rule      *UserLevelRule
	refs      int64
	deleted   string
	updated   *UserLevelRule
	createErr error
	updateErr error
	deleteErr error
}

func (s *stubLevelRulesRepo) ListLevelRules(context.Context) ([]UserLevelRule, error) {
	if s.rule == nil {
		return nil, nil
	}
	return []UserLevelRule{*s.rule}, nil
}
func (s *stubLevelRulesRepo) GetLevelRule(context.Context, string) (*UserLevelRule, error) {
	if s.rule == nil {
		return nil, ErrUserLevelRuleNotFound
	}
	cloned := *s.rule
	cloned.Tiers = append([]UserLevelTier(nil), s.rule.Tiers...)
	return &cloned, nil
}
func (s *stubLevelRulesRepo) CreateLevelRule(_ context.Context, rule *UserLevelRule) error {
	if s.createErr != nil {
		return s.createErr
	}
	s.rule = rule
	return nil
}
func (s *stubLevelRulesRepo) UpdateLevelRule(_ context.Context, rule *UserLevelRule) error {
	if s.updateErr != nil {
		return s.updateErr
	}
	s.updated = rule
	s.rule = rule
	return nil
}
func (s *stubLevelRulesRepo) DeleteLevelRule(_ context.Context, ruleID string) error {
	if s.deleteErr != nil {
		return s.deleteErr
	}
	s.deleted = ruleID
	return nil
}
func (s *stubLevelRulesRepo) ListUserLevelRules(context.Context, string) ([]UserLevelRule, error) {
	return nil, nil
}
func (s *stubLevelRulesRepo) GetAssignedLevelRulesBatch(context.Context, []string) (map[string][]UserLevelRule, error) {
	return map[string][]UserLevelRule{}, nil
}
func (s *stubLevelRulesRepo) ReplaceUserLevelRules(context.Context, string, []string) error {
	return nil
}
func (s *stubLevelRulesRepo) BatchAssignUserLevelRules(context.Context, []string, []string, string) (int64, error) {
	return 0, nil
}
func (s *stubLevelRulesRepo) ListLevelRuleMembers(context.Context, string, int, int) ([]User, int64, error) {
	return nil, 0, nil
}
func (s *stubLevelRulesRepo) GetLevelRuleReferenceCount(context.Context, string) (int64, error) {
	return s.refs, nil
}

func newLevelRuleService(repo *stubLevelRulesRepo) *UserLevelService {
	svc := NewUserLevelService(nil, nil, nil, nil, nil)
	svc.rulesRepo = repo
	return svc
}

func TestPreserveBaseLevelTierIDRewritesOmittedBaseID(t *testing.T) {
	baseID := uuid.Must(uuid.NewV7()).String()
	incoming := []UserLevelTier{
		{ID: uuid.Must(uuid.NewV7()).String(), Name: "Base", SortOrder: 0, MinSpend: 0},
		{ID: uuid.Must(uuid.NewV7()).String(), Name: "Gold", SortOrder: 1, MinSpend: 10},
	}

	require.NoError(t, preserveBaseLevelTierID([]UserLevelTier{{
		ID: baseID, Name: "Base", SortOrder: 0, MinSpend: 0,
	}}, incoming))
	require.Equal(t, baseID, incoming[0].ID)
}

func TestPreserveBaseLevelTierIDRejectsMissingBase(t *testing.T) {
	err := preserveBaseLevelTierID([]UserLevelTier{{
		ID: uuid.Must(uuid.NewV7()).String(), Name: "Base", SortOrder: 0, MinSpend: 0,
	}}, []UserLevelTier{{
		ID: uuid.Must(uuid.NewV7()).String(), Name: "Gold", SortOrder: 1, MinSpend: 10,
	}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "base tier cannot be deleted or moved")
}

func TestUpdateLevelRulePreservesBaseTierIDWhenClientOmitsIt(t *testing.T) {
	ruleID := uuid.Must(uuid.NewV7()).String()
	baseID := uuid.Must(uuid.NewV7()).String()
	repo := &stubLevelRulesRepo{rule: &UserLevelRule{
		ID: ruleID, Name: "VIP", WindowDays: 7, Enabled: true, CreatedAt: time.Now(),
		Tiers: []UserLevelTier{{ID: baseID, RuleID: ruleID, Name: "Base", SortOrder: 0, MinSpend: 0}},
	}}
	svc := newLevelRuleService(repo)

	updated, err := svc.UpdateLevelRule(context.Background(), ruleID, UpdateUserLevelRuleInput{
		Name: "VIP", WindowDays: 7, Tiers: []UserLevelRuleTierInput{
			{Name: "Base", SortOrder: 0, MinSpend: 0},
			{Name: "Gold", SortOrder: 1, MinSpend: 50},
		},
	})
	require.NoError(t, err)
	require.Equal(t, baseID, updated.Tiers[0].ID)
	require.Equal(t, baseID, repo.updated.Tiers[0].ID)
	require.Equal(t, "Gold", updated.Tiers[1].Name)
}

func TestDeleteLevelRuleRefusesReferencedRules(t *testing.T) {
	ruleID := uuid.Must(uuid.NewV7()).String()
	repo := &stubLevelRulesRepo{
		rule: &UserLevelRule{ID: ruleID, Name: "VIP", WindowDays: 7, Enabled: true},
		refs: 2,
	}
	err := newLevelRuleService(repo).DeleteLevelRule(context.Background(), ruleID)
	require.ErrorIs(t, err, ErrUserLevelRuleReferenced)
	require.Empty(t, repo.deleted)
}

func TestDeleteLevelRuleDeletesUnreferencedRule(t *testing.T) {
	ruleID := uuid.Must(uuid.NewV7()).String()
	repo := &stubLevelRulesRepo{
		rule: &UserLevelRule{ID: ruleID, Name: "VIP", WindowDays: 7, Enabled: true},
	}
	require.NoError(t, newLevelRuleService(repo).DeleteLevelRule(context.Background(), ruleID))
	require.Equal(t, ruleID, repo.deleted)
}
