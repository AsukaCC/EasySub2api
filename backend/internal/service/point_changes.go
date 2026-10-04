package service

import (
	"context"
	"time"

	infraerrors "github.com/AsukaCC/EasySub2api/internal/pkg/errors"
)

type PointChangeFilter struct {
	Keyword, UserID, PointType, Direction string
	StartTime, EndTime                    *time.Time
	Page, PageSize                        int
}

type PointChange struct {
	ID            string    `json:"id"`
	TransactionID string    `json:"transaction_id"`
	UserID        string    `json:"user_id"`
	Email         string    `json:"email"`
	Username      string    `json:"username"`
	PointType     string    `json:"point_type"`
	Amount        float64   `json:"amount"`
	FrozenAmount  float64   `json:"frozen_amount"`
	BalanceBefore *float64  `json:"balance_before"`
	BalanceAfter  *float64  `json:"balance_after"`
	Action        string    `json:"action"`
	SourceType    string    `json:"source_type"`
	SourceID      string    `json:"source_id"`
	Notes         string    `json:"notes"`
	CreatedAt     time.Time `json:"created_at"`
}

type PointChangePage struct {
	Items    []PointChange `json:"items"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

type PointChangeReader interface {
	ListPointChanges(context.Context, PointChangeFilter) (PointChangePage, error)
}

func (s *adminServiceImpl) ListPointChanges(ctx context.Context, filter PointChangeFilter) (PointChangePage, error) {
	repo, ok := s.userRepo.(PointChangeReader)
	if !ok {
		return PointChangePage{}, infraerrors.InternalServer("POINT_CHANGES_UNAVAILABLE", "point change query is unavailable")
	}
	return repo.ListPointChanges(ctx, filter)
}
