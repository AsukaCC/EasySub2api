//go:build unit

package service

import (
	"context"
	"sync"
	"testing"

	"github.com/AsukaCC/EasySub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type bulkEventAccountRepo struct {
	*batchAccountQueryRepo
	accounts []*Account
}

func newBulkEventAccountRepo(accounts ...*Account) *bulkEventAccountRepo {
	return &bulkEventAccountRepo{
		batchAccountQueryRepo: newBatchAccountQueryRepo(),
		accounts:              accounts,
	}
}

func (r *bulkEventAccountRepo) GetByIDs(context.Context, []string) ([]*Account, error) {
	return append([]*Account(nil), r.accounts...), nil
}

type bulkEventSnapshotCache struct {
	*batchSnapshotCache

	accountMu        sync.Mutex
	setAccountIDs    []string
	deleteAccountIDs []string
}

func newBulkEventSnapshotCache() *bulkEventSnapshotCache {
	return &bulkEventSnapshotCache{batchSnapshotCache: newBatchSnapshotCache()}
}

func (c *bulkEventSnapshotCache) SetAccount(_ context.Context, account *Account) error {
	c.accountMu.Lock()
	defer c.accountMu.Unlock()
	c.setAccountIDs = append(c.setAccountIDs, account.ID)
	return nil
}

func (c *bulkEventSnapshotCache) DeleteAccount(_ context.Context, accountID string) error {
	c.accountMu.Lock()
	defer c.accountMu.Unlock()
	c.deleteAccountIDs = append(c.deleteAccountIDs, accountID)
	return nil
}

func (c *bulkEventSnapshotCache) accountWrites() (set []string, deleted []string) {
	c.accountMu.Lock()
	defer c.accountMu.Unlock()
	return append([]string(nil), c.setAccountIDs...), append([]string(nil), c.deleteAccountIDs...)
}

func (c *bulkEventSnapshotCache) capturedBuckets() []SchedulerBucket {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]SchedulerBucket(nil), c.captures...)
}

func newBulkEventTestService(cache SchedulerCache, accounts AccountRepository) *SchedulerSnapshotService {
	return NewSchedulerSnapshotService(cache, nil, accounts, nil, &config.Config{RunMode: config.RunModeStandard})
}

func bulkEventPayload(accountIDs []string, groupIDs []string) map[string]any {
	accountValues := make([]any, 0, len(accountIDs))
	for _, id := range accountIDs {
		accountValues = append(accountValues, id)
	}
	groupValues := make([]any, 0, len(groupIDs))
	for _, id := range groupIDs {
		groupValues = append(groupValues, id)
	}
	return map[string]any{
		"account_ids": accountValues,
		"group_ids":   groupValues,
	}
}

func schedulerBucketsForTest(groupIDs []string, platforms ...string) []SchedulerBucket {
	buckets := make([]SchedulerBucket, 0, len(groupIDs)*len(platforms)*3)
	for _, platform := range platforms {
		for _, groupID := range groupIDs {
			buckets = append(buckets,
				SchedulerBucket{GroupID: groupID, Platform: platform, Mode: SchedulerModeSingle},
				SchedulerBucket{GroupID: groupID, Platform: platform, Mode: SchedulerModeForced},
			)
			if platform == PlatformAnthropic || platform == PlatformGemini {
				buckets = append(buckets, SchedulerBucket{GroupID: groupID, Platform: platform, Mode: SchedulerModeMixed})
			}
		}
	}
	return buckets
}

func TestSchedulerBulkAccountEventScopesOpenAIRebuildToFreshPlatform(t *testing.T) {
	cache := newBulkEventSnapshotCache()
	repo := newBulkEventAccountRepo(&Account{ID: "1", Platform: PlatformOpenAI, GroupIDs: []string{"12"}})
	svc := newBulkEventTestService(cache, repo)

	err := svc.handleBulkAccountEvent(context.Background(), bulkEventPayload([]string{"1"}, []string{"11"}), make(map[batchSeenKey]struct{}))

	require.NoError(t, err)
	require.ElementsMatch(t, schedulerBucketsForTest([]string{"11", "12"}, PlatformOpenAI), cache.capturedBuckets())
	set, deleted := cache.accountWrites()
	require.Equal(t, []string{"1"}, set)
	require.Empty(t, deleted)
}

func TestSchedulerBulkAccountEventRebuildsOpenAIUngroupedBucket(t *testing.T) {
	cache := newBulkEventSnapshotCache()
	repo := newBulkEventAccountRepo(&Account{ID: "6", Platform: PlatformOpenAI})
	svc := newBulkEventTestService(cache, repo)

	err := svc.handleBulkAccountEvent(context.Background(), bulkEventPayload([]string{"6"}, nil), make(map[batchSeenKey]struct{}))

	require.NoError(t, err)
	require.ElementsMatch(t, schedulerBucketsForTest([]string{"0"}, PlatformOpenAI), cache.capturedBuckets())
}

func TestSchedulerBulkAccountEventKeepsGroupedAndUngroupedBuckets(t *testing.T) {
	cache := newBulkEventSnapshotCache()
	repo := newBulkEventAccountRepo(
		&Account{ID: "7", Platform: PlatformOpenAI, GroupIDs: []string{"51"}},
		&Account{ID: "8", Platform: PlatformOpenAI},
	)
	svc := newBulkEventTestService(cache, repo)

	err := svc.handleBulkAccountEvent(context.Background(), bulkEventPayload([]string{"7", "8"}, nil), make(map[batchSeenKey]struct{}))

	require.NoError(t, err)
	require.ElementsMatch(t, schedulerBucketsForTest([]string{"0", "51"}, PlatformOpenAI), cache.capturedBuckets())
}

func TestSchedulerBulkAccountEventDoesNotCrossCurrentGroupsBetweenPlatforms(t *testing.T) {
	cache := newBulkEventSnapshotCache()
	repo := newBulkEventAccountRepo(
		&Account{ID: "9", Platform: PlatformOpenAI, GroupIDs: []string{"61"}},
		&Account{ID: "10", Platform: PlatformGrok, GroupIDs: []string{"62"}},
	)
	svc := newBulkEventTestService(cache, repo)

	err := svc.handleBulkAccountEvent(context.Background(), bulkEventPayload([]string{"9", "10"}, []string{"63"}), make(map[batchSeenKey]struct{}))

	require.NoError(t, err)
	want := append(
		schedulerBucketsForTest([]string{"61", "63"}, PlatformOpenAI),
		schedulerBucketsForTest([]string{"62", "63"}, PlatformGrok)...,
	)
	require.ElementsMatch(t, want, cache.capturedBuckets())
}

func TestSchedulerBulkAccountEventUsesGroupZeroInSimpleMode(t *testing.T) {
	cache := newBulkEventSnapshotCache()
	repo := newBulkEventAccountRepo(&Account{ID: "11", Platform: PlatformOpenAI, GroupIDs: []string{"71"}})
	svc := NewSchedulerSnapshotService(cache, nil, repo, nil, &config.Config{RunMode: config.RunModeSimple})

	err := svc.handleBulkAccountEvent(context.Background(), bulkEventPayload([]string{"11"}, []string{"72"}), make(map[batchSeenKey]struct{}))

	require.NoError(t, err)
	require.ElementsMatch(t, schedulerBucketsForTest([]string{"0"}, PlatformOpenAI), cache.capturedBuckets())
}

func TestSchedulerBulkAccountEventConservativelyExpandsAntigravityPlatforms(t *testing.T) {
	cache := newBulkEventSnapshotCache()
	// fresh 值可能已经关闭 mixed_scheduling，兼容平台仍要重建以清理旧快照。
	repo := newBulkEventAccountRepo(&Account{ID: "2", Platform: PlatformAntigravity, GroupIDs: []string{"22"}})
	svc := newBulkEventTestService(cache, repo)

	err := svc.handleBulkAccountEvent(context.Background(), bulkEventPayload([]string{"2"}, []string{"21"}), make(map[batchSeenKey]struct{}))

	require.NoError(t, err)
	require.ElementsMatch(t,
		schedulerBucketsForTest([]string{"21", "22"}, PlatformAnthropic, PlatformGemini, PlatformAntigravity),
		cache.capturedBuckets(),
	)
}

func TestSchedulerBulkAccountEventMissingAccountFallsBackToAllPlatforms(t *testing.T) {
	cache := newBulkEventSnapshotCache()
	repo := newBulkEventAccountRepo(&Account{ID: "3", Platform: PlatformOpenAI, GroupIDs: []string{"32"}})
	svc := newBulkEventTestService(cache, repo)

	err := svc.handleBulkAccountEvent(context.Background(), bulkEventPayload([]string{"3", "4"}, []string{"31"}), make(map[batchSeenKey]struct{}))

	require.NoError(t, err)
	platforms := schedulerSnapshotPlatforms()
	require.ElementsMatch(t, schedulerBucketsForTest([]string{"31", "32"}, platforms[:]...), cache.capturedBuckets())
	set, deleted := cache.accountWrites()
	require.Equal(t, []string{"3"}, set)
	require.Equal(t, []string{"4"}, deleted)
}

func TestSchedulerBulkAccountEventUnknownPlatformFallsBackToAllPlatforms(t *testing.T) {
	cache := newBulkEventSnapshotCache()
	repo := newBulkEventAccountRepo(&Account{ID: "5", GroupIDs: []string{"42"}})
	svc := newBulkEventTestService(cache, repo)

	err := svc.handleBulkAccountEvent(context.Background(), bulkEventPayload([]string{"5"}, []string{"41"}), make(map[batchSeenKey]struct{}))

	require.NoError(t, err)
	platforms := schedulerSnapshotPlatforms()
	require.ElementsMatch(t, schedulerBucketsForTest([]string{"41", "42"}, platforms[:]...), cache.capturedBuckets())
}
