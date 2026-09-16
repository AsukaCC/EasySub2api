package handler

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/stretchr/testify/require"
)

type concurrencyCacheMock struct {
	acquireUserSlotFn     func(ctx context.Context, userID string, maxConcurrency int, requestID string) (bool, error)
	acquireAccountSlotFn  func(ctx context.Context, accountID string, maxConcurrency int, requestID string) (bool, error)
	acquireIngressLeaseFn func(ctx context.Context, apiKeyID string, maxConnections int, leaseID string) (bool, error)
	releaseIngressLeaseFn func(ctx context.Context, apiKeyID string, leaseID string) error
	releaseUserCalled     int32
	releaseAccountCalled  int32
	releaseIngressCalled  int32
}

func (m *concurrencyCacheMock) AcquireAccountSlot(ctx context.Context, accountID string, maxConcurrency int, requestID string) (bool, error) {
	if m.acquireAccountSlotFn != nil {
		return m.acquireAccountSlotFn(ctx, accountID, maxConcurrency, requestID)
	}
	return false, nil
}

func (m *concurrencyCacheMock) ReleaseAccountSlot(ctx context.Context, accountID string, requestID string) error {
	atomic.AddInt32(&m.releaseAccountCalled, 1)
	return nil
}

func (m *concurrencyCacheMock) GetAccountConcurrency(ctx context.Context, accountID string) (int, error) {
	return 0, nil
}

func (m *concurrencyCacheMock) GetAccountConcurrencyBatch(ctx context.Context, accountIDs []string) (map[string]int, error) {
	result := make(map[string]int, len(accountIDs))
	for _, accountID := range accountIDs {
		result[accountID] = 0
	}
	return result, nil
}

func (m *concurrencyCacheMock) IncrementAccountWaitCount(ctx context.Context, accountID string, maxWait int) (bool, error) {
	return true, nil
}

func (m *concurrencyCacheMock) DecrementAccountWaitCount(ctx context.Context, accountID string) error {
	return nil
}

func (m *concurrencyCacheMock) GetAccountWaitingCount(ctx context.Context, accountID string) (int, error) {
	return 0, nil
}

func (m *concurrencyCacheMock) AcquireUserSlot(ctx context.Context, userID string, maxConcurrency int, requestID string) (bool, error) {
	if m.acquireUserSlotFn != nil {
		return m.acquireUserSlotFn(ctx, userID, maxConcurrency, requestID)
	}
	return false, nil
}

func (m *concurrencyCacheMock) ReleaseUserSlot(ctx context.Context, userID string, requestID string) error {
	atomic.AddInt32(&m.releaseUserCalled, 1)
	return nil
}

func (m *concurrencyCacheMock) GetUserConcurrency(ctx context.Context, userID string) (int, error) {
	return 0, nil
}

func (m *concurrencyCacheMock) IncrementWaitCount(ctx context.Context, userID string, maxWait int) (bool, error) {
	return true, nil
}

func (m *concurrencyCacheMock) DecrementWaitCount(ctx context.Context, userID string) error {
	return nil
}

func (m *concurrencyCacheMock) GetAccountsLoadBatch(ctx context.Context, accounts []service.AccountWithConcurrency) (map[string]*service.AccountLoadInfo, error) {
	return map[string]*service.AccountLoadInfo{}, nil
}

func (m *concurrencyCacheMock) GetUsersLoadBatch(ctx context.Context, users []service.UserWithConcurrency) (map[string]*service.UserLoadInfo, error) {
	return map[string]*service.UserLoadInfo{}, nil
}

func (m *concurrencyCacheMock) CleanupExpiredAccountSlots(ctx context.Context, accountID string) error {
	return nil
}

func (m *concurrencyCacheMock) CleanupExpiredAccountSlotKeys(ctx context.Context) error {
	return nil
}

func (m *concurrencyCacheMock) CleanupStaleProcessSlots(ctx context.Context, activeRequestPrefix string) error {
	return nil
}

func (m *concurrencyCacheMock) AcquireOpenAIWSIngressLease(ctx context.Context, apiKeyID string, maxConnections int, leaseID string) (bool, error) {
	if m.acquireIngressLeaseFn != nil {
		return m.acquireIngressLeaseFn(ctx, apiKeyID, maxConnections, leaseID)
	}
	return false, nil
}

func (m *concurrencyCacheMock) RefreshOpenAIWSIngressLease(context.Context, string, string) (bool, error) {
	return true, nil
}

func (m *concurrencyCacheMock) ReleaseOpenAIWSIngressLease(ctx context.Context, apiKeyID string, leaseID string) error {
	atomic.AddInt32(&m.releaseIngressCalled, 1)
	if m.releaseIngressLeaseFn != nil {
		return m.releaseIngressLeaseFn(ctx, apiKeyID, leaseID)
	}
	return nil
}

func TestConcurrencyHelper_TryAcquireUserSlot(t *testing.T) {
	cache := &concurrencyCacheMock{
		acquireUserSlotFn: func(ctx context.Context, userID string, maxConcurrency int, requestID string) (bool, error) {
			return true, nil
		},
	}
	helper := NewConcurrencyHelper(service.NewConcurrencyService(cache), SSEPingFormatNone, time.Second)

	release, acquired, err := helper.TryAcquireUserSlot(context.Background(), "user-101", 2)
	require.NoError(t, err)
	require.True(t, acquired)
	require.NotNil(t, release)

	release()
	require.Equal(t, int32(1), atomic.LoadInt32(&cache.releaseUserCalled))
}

func TestConcurrencyHelper_TryAcquireAccountSlot_NotAcquired(t *testing.T) {
	cache := &concurrencyCacheMock{
		acquireAccountSlotFn: func(ctx context.Context, accountID string, maxConcurrency int, requestID string) (bool, error) {
			return false, nil
		},
	}
	helper := NewConcurrencyHelper(service.NewConcurrencyService(cache), SSEPingFormatNone, time.Second)

	release, acquired, err := helper.TryAcquireAccountSlot(context.Background(), "account-201", 1)
	require.NoError(t, err)
	require.False(t, acquired)
	require.Nil(t, release)
	require.Equal(t, int32(0), atomic.LoadInt32(&cache.releaseAccountCalled))
}
