package repository

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/pkg/ctxkey"
	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestKeySchedulingLeaseLifecycle(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	reader := NewGatewayCache(client).(service.AccountKeyActivityReader)
	first := service.NewConcurrencyService(NewConcurrencyCache(client, 15, 30))
	second := service.NewConcurrencyService(NewConcurrencyCache(client, 15, 30))
	ctx := context.WithValue(context.Background(), ctxkey.APIKeyID, "key-a")
	a, err := first.AcquireAccountSlot(ctx, "account", 10)
	require.NoError(t, err)
	require.True(t, a.Acquired)
	t.Cleanup(a.ReleaseFunc)
	b, err := second.AcquireAccountSlot(ctx, "account", 10)
	require.NoError(t, err)
	require.True(t, b.Acquired)
	t.Cleanup(b.ReleaseFunc)
	read := func(key string) service.AccountKeyActivity {
		values, readErr := reader.GetAccountKeyActivity(context.Background(), []string{"account"}, key)
		require.NoError(t, readErr)
		return values["account"]
	}
	require.Equal(t, 2, read("key-a").OwnRequests)
	require.Equal(t, 1, read("key-b").OtherKeys, "one account owner with two independent requests")
	a.ReleaseFunc()
	a.ReleaseFunc()
	require.Equal(t, 1, read("key-b").OtherKeys)
	b.ReleaseFunc()
	require.Zero(t, read("key-b").OtherKeys)
	require.False(t, read("key-b").LastScheduledAt.IsZero())

	// Unlimited concurrency still participates in key isolation.
	c, err := first.AcquireAccountSlot(ctx, "account", 0)
	require.NoError(t, err)
	require.True(t, c.Acquired)
	t.Cleanup(c.ReleaseFunc)
	require.Equal(t, 1, read("key-b").OtherKeys)
	c.ReleaseFunc()
	require.Zero(t, read("key-b").OtherKeys)
}

func TestKeySchedulingAtomicFirstAndSecondRound(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	store := NewConcurrencyCache(client, 15, 30).(service.AccountKeyLeaseStore)
	reader := NewGatewayCache(client).(service.AccountKeyActivityReader)
	ctx := context.Background()
	var wg sync.WaitGroup
	results := make(chan string, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", i)
			for _, account := range []string{"a", "b"} {
				ok, err := store.AcquireAccountKeyLease(ctx, account, key, "req", 0, nil)
				if err != nil {
					results <- "error"
					return
				}
				if ok {
					results <- account
					return
				}
			}
			results <- ""
		}(i)
	}
	wg.Wait()
	close(results)
	counts := map[string]int{}
	for result := range results {
		counts[result]++
	}
	require.Zero(t, counts["error"])
	require.Equal(t, 1, counts["a"])
	require.Equal(t, 1, counts["b"])
	values, err := reader.GetAccountKeyActivity(ctx, []string{"a", "b"}, "new-key")
	require.NoError(t, err)
	for _, account := range []string{"a", "b"} {
		require.Equal(t, 1, values[account].OtherKeys)
		ok, err := store.AcquireAccountKeyLease(ctx, account, "new-key-"+account, "req", 1, nil)
		require.NoError(t, err)
		require.True(t, ok)
	}
	values, err = reader.GetAccountKeyActivity(ctx, []string{"a", "b"}, "another-key")
	require.NoError(t, err)
	require.Equal(t, 2, values["a"].OtherKeys)
	require.Equal(t, 2, values["b"].OtherKeys)
}

func TestKeySchedulingSameKeyRequestsRemainIndependent(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	cache := NewConcurrencyCache(client, 15, 30)
	instances := []*service.ConcurrencyService{
		service.NewConcurrencyService(cache),
		service.NewConcurrencyService(NewConcurrencyCache(client, 15, 30)),
	}
	reader := NewGatewayCache(client).(service.AccountKeyActivityReader)
	ctx := context.WithValue(context.Background(), ctxkey.APIKeyID, "same-key")
	const requests = 4
	type attempt struct {
		result *service.AcquireResult
		err    error
	}
	start := make(chan struct{})
	results := make(chan attempt, requests)
	for i := 0; i < requests; i++ {
		go func(instance *service.ConcurrencyService) {
			<-start
			result, err := instance.AcquireAccountSlot(ctx, "account", requests)
			results <- attempt{result: result, err: err}
		}(instances[i%len(instances)])
	}
	close(start)
	var releases []func()
	for i := 0; i < requests; i++ {
		result := <-results
		require.NoError(t, result.err)
		require.True(t, result.result.Acquired)
		releases = append(releases, result.result.ReleaseFunc)
		t.Cleanup(result.result.ReleaseFunc)
	}
	assertRequests := func(expected int) {
		count, err := cache.GetAccountConcurrency(ctx, "account")
		require.NoError(t, err)
		require.Equal(t, expected, count, "each request consumes a separate concurrency slot")
		activity, err := reader.GetAccountKeyActivity(ctx, []string{"account"}, "same-key")
		require.NoError(t, err)
		require.Equal(t, expected, activity["account"].OwnRequests)
	}
	assertRequests(requests)
	blocked, err := instances[0].AcquireAccountSlot(ctx, "account", requests)
	require.NoError(t, err)
	require.False(t, blocked.Acquired, "same-key affinity cannot bypass the concurrency limit")
	for i, release := range releases {
		release()
		assertRequests(requests - i - 1)
	}
}

func TestKeySchedulingLeaseRefreshAndCrashExpiry(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	store := NewConcurrencyCache(client, 15, 30).(service.AccountKeyLeaseStore)
	reader := NewGatewayCache(client).(service.AccountKeyActivityReader)
	ctx := context.Background()
	now := time.Now().Truncate(time.Second)
	server.SetTime(now)
	ok, err := store.AcquireAccountKeyLease(ctx, "a", "key", "request", 0, nil)
	require.NoError(t, err)
	require.True(t, ok)
	server.SetTime(now.Add(60 * time.Second))
	require.NoError(t, store.RefreshAccountKeyLease(ctx, "a", "key", "request"))
	server.SetTime(now.Add(100 * time.Second))
	values, err := reader.GetAccountKeyActivity(ctx, []string{"a"}, "other")
	require.NoError(t, err)
	require.Equal(t, 1, values["a"].OtherKeys)
	server.SetTime(now.Add(151 * time.Second))
	values, err = reader.GetAccountKeyActivity(ctx, []string{"a"}, "other")
	require.NoError(t, err)
	require.Zero(t, values["a"].OtherKeys)
	// Late refresh cannot resurrect a released/expired request.
	require.NoError(t, store.RefreshAccountKeyLease(ctx, "a", "key", "request"))
	values, err = reader.GetAccountKeyActivity(ctx, []string{"a"}, "other")
	require.NoError(t, err)
	require.Zero(t, values["a"].OtherKeys)
}

func TestKeySchedulingBalancesStaleSnapshotsAcrossRounds(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	store := NewConcurrencyCache(client, 15, 30).(service.AccountKeyLeaseStore)
	reader := NewGatewayCache(client).(service.AccountKeyActivityReader)
	ctx := context.Background()
	// Every request deliberately has the same stale zero-occupancy snapshot.
	// Atomic peer comparison must still fill each round evenly.
	for i := 0; i < 40; i++ {
		claimed := false
		for _, candidate := range []struct{ account, peer string }{{"a", "b"}, {"b", "a"}} {
			ok, err := store.AcquireAccountKeyLease(ctx, candidate.account, fmt.Sprintf("key-%d", i), "req", 0, []service.AccountKeyPeer{{ID: candidate.peer, MaxConcurrency: 100}})
			require.NoError(t, err)
			if ok {
				claimed = true
				break
			}
		}
		require.True(t, claimed)
		values, err := reader.GetAccountKeyActivity(ctx, []string{"a", "b"}, "observer")
		require.NoError(t, err)
		require.LessOrEqual(t, values["a"].OtherKeys-values["b"].OtherKeys, 1)
		require.LessOrEqual(t, values["b"].OtherKeys-values["a"].OtherKeys, 1)
	}
	// An empty peer at its physical concurrency limit cannot block a usable account.
	require.NoError(t, client.ZAdd(ctx, accountSlotKeyPrefix+"full", redis.Z{Score: float64(time.Now().Unix()), Member: "running"}).Err())
	ok, err := store.AcquireAccountKeyLease(ctx, "a", "overflow-key", "req", 0, []service.AccountKeyPeer{{ID: "full", MaxConcurrency: 1}})
	require.NoError(t, err)
	require.True(t, ok)
}

func TestKeySchedulingLiveLeaseHandoff(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	cache := NewConcurrencyCache(client, 15, 30)
	live := cache.(service.LiveConcurrencyCache)
	reader := NewGatewayCache(client).(service.AccountKeyActivityReader)
	ctx := context.WithValue(context.Background(), ctxkey.APIKeyID, "key")
	selection, err := service.NewConcurrencyService(cache).AcquireAccountSlot(ctx, "a", 10)
	require.NoError(t, err)
	require.True(t, selection.Acquired)
	t.Cleanup(selection.ReleaseFunc)
	ok, err := live.AcquireLiveLease(ctx, "a", 10, "user", 10, "key", "live-lease", true)
	require.NoError(t, err)
	require.True(t, ok)
	selection.ReleaseFunc()
	values, err := reader.GetAccountKeyActivity(ctx, []string{"a"}, "observer")
	require.NoError(t, err)
	require.Equal(t, 1, values["a"].OtherKeys)
	ok, err = live.RefreshLiveLease(ctx, "a", "user", "key", "live-lease")
	require.NoError(t, err)
	require.True(t, ok)
	require.NoError(t, live.ReleaseLiveLease(ctx, "a", "user", "key", "live-lease"))
	values, err = reader.GetAccountKeyActivity(ctx, []string{"a"}, "observer")
	require.NoError(t, err)
	require.Zero(t, values["a"].OtherKeys)
}
