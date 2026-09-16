package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type opsMetricsProjectionRepo struct {
	AccountRepository
	accounts        []Account
	accountLoads    []AccountWithConcurrency
	listCalls       int
	projectionCalls int
}

func (r *opsMetricsProjectionRepo) ListSchedulable(context.Context) ([]Account, error) {
	r.listCalls++
	return r.accounts, nil
}

func (r *opsMetricsProjectionRepo) ListSchedulableAccountLoads(context.Context) ([]AccountWithConcurrency, error) {
	r.projectionCalls++
	return r.accountLoads, nil
}

type opsMetricsFallbackRepo struct {
	AccountRepository
	accounts  []Account
	listCalls int
}

func (r *opsMetricsFallbackRepo) ListSchedulable(context.Context) ([]Account, error) {
	r.listCalls++
	return r.accounts, nil
}

type opsMetricsLoadCache struct {
	ConcurrencyCache
	loads map[string]*AccountLoadInfo
	got   []AccountWithConcurrency
}

func (c *opsMetricsLoadCache) GetAccountsLoadBatch(_ context.Context, accounts []AccountWithConcurrency) (map[string]*AccountLoadInfo, error) {
	c.got = accounts
	return c.loads, nil
}

func TestCollectConcurrencyQueueDepthUsesProjectionAndPreservesFallbackResult(t *testing.T) {
	loadFactor := 7
	accounts := []Account{
		{ID: "account-11", Concurrency: 2, LoadFactor: &loadFactor},
		{ID: "account-12", Concurrency: 3},
		{ID: "account-13"},
	}
	accountLoads := []AccountWithConcurrency{
		{ID: "account-11", MaxConcurrency: 7},
		{ID: "account-12", MaxConcurrency: 3},
		{ID: "account-13", MaxConcurrency: 1},
	}
	loads := map[string]*AccountLoadInfo{
		"account-11": {AccountID: "account-11", WaitingCount: 2},
		"account-12": {AccountID: "account-12", WaitingCount: 3},
		"account-13": {AccountID: "account-13", WaitingCount: 0},
	}

	projectionRepo := &opsMetricsProjectionRepo{accounts: accounts, accountLoads: accountLoads}
	projectionCache := &opsMetricsLoadCache{loads: loads}
	projectionConcurrency := NewConcurrencyService(projectionCache)
	projectionConcurrency.SetAccountLoadBatchCacheTTL(0)
	projectionCollector := &OpsMetricsCollector{
		accountRepo:        projectionRepo,
		concurrencyService: projectionConcurrency,
	}

	fallbackRepo := &opsMetricsFallbackRepo{accounts: accounts}
	fallbackCache := &opsMetricsLoadCache{loads: loads}
	fallbackConcurrency := NewConcurrencyService(fallbackCache)
	fallbackConcurrency.SetAccountLoadBatchCacheTTL(0)
	fallbackCollector := &OpsMetricsCollector{
		accountRepo:        fallbackRepo,
		concurrencyService: fallbackConcurrency,
	}

	projectionDepth := projectionCollector.collectConcurrencyQueueDepth(context.Background())
	fallbackDepth := fallbackCollector.collectConcurrencyQueueDepth(context.Background())

	require.NotNil(t, projectionDepth)
	require.NotNil(t, fallbackDepth)
	require.Equal(t, 5, *projectionDepth)
	require.Equal(t, *fallbackDepth, *projectionDepth)
	require.Equal(t, 1, projectionRepo.projectionCalls)
	require.Zero(t, projectionRepo.listCalls)
	require.Equal(t, 1, fallbackRepo.listCalls)
	require.Equal(t, accountLoads, projectionCache.got)
	require.Equal(t, fallbackCache.got, projectionCache.got)
}

func BenchmarkOpsMetricsCollectorCollectConcurrencyQueueDepth(b *testing.B) {
	const accountCount = 1000
	loadFactor := 8
	accounts := make([]Account, accountCount)
	accountLoads := make([]AccountWithConcurrency, accountCount)
	for i := range accountCount {
		id := fmt.Sprintf("account-%d", i+1)
		accounts[i] = Account{
			ID:          id,
			Concurrency: 4,
			LoadFactor:  &loadFactor,
		}
		accountLoads[i] = AccountWithConcurrency{ID: id, MaxConcurrency: loadFactor}
	}

	repo := &opsMetricsProjectionRepo{accounts: accounts, accountLoads: accountLoads}
	cache := &opsMetricsLoadCache{loads: map[string]*AccountLoadInfo{}}
	concurrency := NewConcurrencyService(cache)
	concurrency.SetAccountLoadBatchCacheTTL(0)
	collector := &OpsMetricsCollector{accountRepo: repo, concurrencyService: concurrency}

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if depth := collector.collectConcurrencyQueueDepth(context.Background()); depth == nil || *depth != 0 {
			b.Fatalf("unexpected queue depth: %v", depth)
		}
	}
}
