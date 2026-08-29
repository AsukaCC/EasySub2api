package service

import "context"

// Internal500CounterCache tracks consecutive Antigravity INTERNAL 500 rounds.
type Internal500CounterCache interface {
	IncrementInternal500Count(ctx context.Context, accountID string) (int64, error)
	ResetInternal500Count(ctx context.Context, accountID string) error
}
