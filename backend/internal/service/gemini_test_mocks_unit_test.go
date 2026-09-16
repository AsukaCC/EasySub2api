//go:build unit

package service

import (
	"context"
	"errors"
)

// mockAccountRepoForGemini is the shared account repository used by unit tests.
// Embedding AccountRepository keeps unrelated methods on the current interface
// available while the explicit methods below provide the behavior exercised by tests.
type mockAccountRepoForGemini struct {
	AccountRepository
	accounts           []Account
	accountsByID       map[string]*Account
	listByGroupFunc    func(context.Context, string, []string) ([]Account, error)
	listByPlatformFunc func(context.Context, []string) ([]Account, error)
}

func (m *mockAccountRepoForGemini) GetByID(_ context.Context, id string) (*Account, error) {
	if account, ok := m.accountsByID[id]; ok {
		return account, nil
	}
	return nil, errors.New("account not found")
}

func (m *mockAccountRepoForGemini) GetByIDs(_ context.Context, ids []string) ([]*Account, error) {
	result := make([]*Account, 0, len(ids))
	for _, id := range ids {
		if account, ok := m.accountsByID[id]; ok {
			result = append(result, account)
		}
	}
	return result, nil
}

func (m *mockAccountRepoForGemini) ExistsByID(_ context.Context, id string) (bool, error) {
	_, ok := m.accountsByID[id]
	return ok, nil
}

func (m *mockAccountRepoForGemini) ListSchedulableByPlatform(_ context.Context, platform string) ([]Account, error) {
	result := make([]Account, 0, len(m.accounts))
	for _, account := range m.accounts {
		if account.Platform == platform && account.IsSchedulable() {
			result = append(result, account)
		}
	}
	return result, nil
}

func (m *mockAccountRepoForGemini) ListSchedulableByGroupIDAndPlatform(ctx context.Context, _ string, platform string) ([]Account, error) {
	return m.ListSchedulableByPlatform(ctx, platform)
}

func (m *mockAccountRepoForGemini) ListSchedulableByPlatforms(ctx context.Context, platforms []string) ([]Account, error) {
	if m.listByPlatformFunc != nil {
		return m.listByPlatformFunc(ctx, platforms)
	}
	allowed := make(map[string]struct{}, len(platforms))
	for _, platform := range platforms {
		allowed[platform] = struct{}{}
	}
	result := make([]Account, 0, len(m.accounts))
	for _, account := range m.accounts {
		if _, ok := allowed[account.Platform]; ok && account.IsSchedulable() {
			result = append(result, account)
		}
	}
	return result, nil
}

func (m *mockAccountRepoForGemini) ListSchedulableByGroupIDAndPlatforms(ctx context.Context, groupID string, platforms []string) ([]Account, error) {
	if m.listByGroupFunc != nil {
		return m.listByGroupFunc(ctx, groupID, platforms)
	}
	return m.ListSchedulableByPlatforms(ctx, platforms)
}

func (m *mockAccountRepoForGemini) ListSchedulableUngroupedByPlatform(ctx context.Context, platform string) ([]Account, error) {
	return m.ListSchedulableByPlatform(ctx, platform)
}

func (m *mockAccountRepoForGemini) ListSchedulableUngroupedByPlatforms(ctx context.Context, platforms []string) ([]Account, error) {
	return m.ListSchedulableByPlatforms(ctx, platforms)
}
