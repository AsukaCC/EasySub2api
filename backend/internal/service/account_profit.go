package service

import (
	"context"
	"fmt"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/pkg/usagestats"
)

// Optional interface: alternate UsageLogRepository implementations do not
// need to implement profit statistics until they support migration 261.
type AccountProfitRepository interface {
	GetAccountProfit(ctx context.Context, accountID string, from, to time.Time, page, pageSize int) (*usagestats.AccountProfitResponse, error)
}

type AccountProfitListRepository interface {
	ListAccountProfit(ctx context.Context, params usagestats.AccountProfitListParams) (*usagestats.AccountProfitListResponse, error)
}

func (s *AccountUsageService) GetAccountProfit(ctx context.Context, accountID string, from, to time.Time, page, pageSize int) (*usagestats.AccountProfitResponse, error) {
	repo, ok := s.usageLogRepo.(AccountProfitRepository)
	if !ok {
		return nil, fmt.Errorf("account profit repository is unavailable")
	}
	result, err := repo.GetAccountProfit(ctx, accountID, from, to, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("get account profit failed: %w", err)
	}
	return result, nil
}

func (s *AccountUsageService) ListAccountProfit(ctx context.Context, params usagestats.AccountProfitListParams) (*usagestats.AccountProfitListResponse, error) {
	repo, ok := s.usageLogRepo.(AccountProfitListRepository)
	if !ok {
		return nil, fmt.Errorf("account profit list repository is unavailable")
	}
	result, err := repo.ListAccountProfit(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list account profit failed: %w", err)
	}
	return result, nil
}
