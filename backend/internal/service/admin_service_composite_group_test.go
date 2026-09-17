//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type accountRepoStubForCompositeModelsList struct {
	accountRepoStub
	accounts []Account
}

func (s *accountRepoStubForCompositeModelsList) ListSchedulableByGroupID(_ context.Context, _ string) ([]Account, error) {
	return s.accounts, nil
}

func TestAdminService_CreateCompositeGroupCopiesAccountsFromConcreteGroups(t *testing.T) {
	var copiedFrom []string
	var boundGroupID string
	var boundAccountIDs []string
	groupRepo := &groupRepoStubForAdmin{
		createID: "99",
		getByIDByID: map[string]*Group{
			"10": {ID: "10", Platform: PlatformOpenAI},
			"20": {ID: "20", Platform: PlatformGemini},
		},
		getAccountIDsByGroupIDsFn: func(groupIDs []string) ([]string, error) {
			copiedFrom = append([]string{}, groupIDs...)
			return []string{"101", "202"}, nil
		},
		bindAccountsToGroupFn: func(groupID string, accountIDs []string) error {
			boundGroupID = groupID
			boundAccountIDs = append([]string{}, accountIDs...)
			return nil
		},
	}
	svc := &adminServiceImpl{groupRepo: groupRepo}

	group, err := svc.CreateGroup(context.Background(), &CreateGroupInput{
		Name:               "Composite",
		Platform:           PlatformComposite,
		RateMultiplier:     1,
		MaxReasoningEffort: "medium",
		ReasoningEffortMappings: []ReasoningEffortMapping{
			{From: "max", To: "xhigh"},
		},
		CopyAccountsFromGroupIDs: []string{"10", "20", "10"},
	})

	require.NoError(t, err)
	require.Equal(t, PlatformComposite, groupRepo.created.Platform)
	require.Equal(t, "medium", groupRepo.created.MaxReasoningEffort)
	require.Equal(t, []ReasoningEffortMapping{{From: "max", To: "xhigh"}}, groupRepo.created.ReasoningEffortMappings)
	require.Equal(t, "99", group.ID)
	require.Equal(t, int64(2), group.AccountCount)
	require.ElementsMatch(t, []string{"10", "20"}, copiedFrom)
	require.Equal(t, "99", boundGroupID)
	require.ElementsMatch(t, []string{"101", "202"}, boundAccountIDs)
}

func TestAdminService_UpdateCompositeGroupCopiesAccountsFromConcreteGroups(t *testing.T) {
	var clearedGroupID string
	var copiedFrom []string
	var boundGroupID string
	var boundAccountIDs []string
	groupRepo := &groupRepoStubForAdmin{
		getByIDByID: map[string]*Group{
			"10": {ID: "10", Platform: PlatformOpenAI},
			"20": {ID: "20", Platform: PlatformGrok},
			"99": {ID: "99", Platform: PlatformComposite, RateMultiplier: 1, SubscriptionType: SubscriptionTypeStandard},
		},
		deleteAccountGroupsByGroupIDFn: func(groupID string) (int64, error) {
			clearedGroupID = groupID
			return 2, nil
		},
		getAccountIDsByGroupIDsFn: func(groupIDs []string) ([]string, error) {
			copiedFrom = append([]string{}, groupIDs...)
			return []string{"301", "302"}, nil
		},
		bindAccountsToGroupFn: func(groupID string, accountIDs []string) error {
			boundGroupID = groupID
			boundAccountIDs = append([]string{}, accountIDs...)
			return nil
		},
	}
	svc := &adminServiceImpl{groupRepo: groupRepo}
	maxReasoningEffort := "low"
	reasoningEffortMappings := []ReasoningEffortMapping{{From: "max", To: "high"}}

	group, err := svc.UpdateGroup(context.Background(), "99", &UpdateGroupInput{
		MaxReasoningEffort:       &maxReasoningEffort,
		ReasoningEffortMappings:  &reasoningEffortMappings,
		CopyAccountsFromGroupIDs: []string{"10", "20"},
	})

	require.NoError(t, err)
	require.Equal(t, PlatformComposite, group.Platform)
	require.Equal(t, "low", group.MaxReasoningEffort)
	require.Equal(t, reasoningEffortMappings, group.ReasoningEffortMappings)
	require.Equal(t, "99", clearedGroupID)
	require.ElementsMatch(t, []string{"10", "20"}, copiedFrom)
	require.Equal(t, "99", boundGroupID)
	require.ElementsMatch(t, []string{"301", "302"}, boundAccountIDs)
}

func TestAdminService_CreateAccountAllowsCompositeGroupAssignment(t *testing.T) {
	accountRepo := &accountRepoStubForBulkUpdate{createID: "7"}
	groupRepo := &groupRepoStubForAdmin{
		getByIDByID: map[string]*Group{
			"99": {ID: "99", Platform: PlatformComposite},
		},
	}
	svc := &adminServiceImpl{accountRepo: accountRepo, groupRepo: groupRepo}

	account, err := svc.CreateAccount(context.Background(), &CreateAccountInput{
		Name:                  "OpenAI account",
		Platform:              PlatformOpenAI,
		Type:                  AccountTypeAPIKey,
		Concurrency:           1,
		GroupIDs: []string{"99"},
		SkipDefaultGroupBind:  true,
	})

	require.NoError(t, err)
	require.Equal(t, "7", account.ID)
	require.Equal(t, PlatformOpenAI, accountRepo.createAccount.Platform)
	require.ElementsMatch(t, []string{"99"}, accountRepo.bindGroupsByAccount["7"])
}

func TestAdminService_UpdateAccountAllowsCompositeGroupAssignment(t *testing.T) {
	accountRepo := &accountRepoStubForBulkUpdate{
		getByIDAccounts: map[string]*Account{
			"7": {ID: "7", Platform: PlatformGemini, Type: AccountTypeAPIKey, Status: StatusActive, Extra: map[string]any{}},
		},
	}
	groupRepo := &groupRepoStubForAdmin{
		getByIDByID: map[string]*Group{
			"99": {ID: "99", Platform: PlatformComposite},
		},
	}
	svc := &adminServiceImpl{accountRepo: accountRepo, groupRepo: groupRepo}
	groupIDs := []string{"99"}

	account, err := svc.UpdateAccount(context.Background(), "7", &UpdateAccountInput{
		GroupIDs:              &groupIDs,
	})

	require.NoError(t, err)
	require.Equal(t, "7", account.ID)
	require.Len(t, accountRepo.updatedAccounts, 1)
	require.ElementsMatch(t, []string{"99"}, accountRepo.bindGroupsByAccount["7"])
}

func TestAdminService_CompositeModelsListCandidatesIncludeConcreteAccountMappings(t *testing.T) {
	accountRepo := &accountRepoStubForCompositeModelsList{
		accounts: []Account{
			{
				ID: "1",
				Platform: PlatformOpenAI,
				Credentials: map[string]any{
					"model_mapping": map[string]any{"gpt-custom": "gpt-5"},
				},
			},
			{
				ID: "2",
				Platform: PlatformGemini,
				Credentials: map[string]any{
					"model_mapping": map[string]any{"gemini-custom": "gemini-2.5-flash"},
				},
			},
		},
	}
	groupRepo := &groupRepoStubForAdmin{
		getByIDByID: map[string]*Group{
			"99": {ID: "99", Platform: PlatformComposite},
		},
	}
	svc := &adminServiceImpl{accountRepo: accountRepo, groupRepo: groupRepo}

	candidates, err := svc.GetGroupModelsListCandidates(context.Background(), "99", PlatformComposite)

	require.NoError(t, err)
	require.Contains(t, candidates, "gpt-custom")
	require.Contains(t, candidates, "gemini-custom")
	require.Contains(t, candidates, "gpt-5.5")
	require.Contains(t, candidates, "gemini-2.5-flash")
}
