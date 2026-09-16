package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type groupCapacityAccountRepoStub struct {
	AccountRepository
	rows      []GroupAccountCapacityRow
	requested []string
}

func (s *groupCapacityAccountRepoStub) ListSchedulableCapacityByGroupIDs(_ context.Context, groupIDs []string) ([]GroupAccountCapacityRow, error) {
	s.requested = append([]string(nil), groupIDs...)
	return append([]GroupAccountCapacityRow(nil), s.rows...), nil
}

type groupCapacityGroupRepoStub struct {
	GroupRepository
	groupIDs  []string
	listCalls int
}

func (s *groupCapacityGroupRepoStub) ListActiveIDs(context.Context) ([]string, error) {
	s.listCalls++
	return append([]string(nil), s.groupIDs...), nil
}

type groupCapacityConcurrencyCacheStub struct {
	ConcurrencyCache
	counts    map[string]int
	requested []string
}

func (s *groupCapacityConcurrencyCacheStub) GetAccountConcurrencyBatch(_ context.Context, accountIDs []string) (map[string]int, error) {
	s.requested = append([]string(nil), accountIDs...)
	out := make(map[string]int, len(accountIDs))
	for _, id := range accountIDs {
		out[id] = s.counts[id]
	}
	return out, nil
}

type groupCapacitySessionCacheStub struct {
	SessionLimitCache
	counts       map[string]int
	requested    []string
	idleTimeouts map[string]time.Duration
}

func (s *groupCapacitySessionCacheStub) GetActiveSessionCountBatch(_ context.Context, accountIDs []string, idleTimeouts map[string]time.Duration) (map[string]int, error) {
	s.requested = append([]string(nil), accountIDs...)
	s.idleTimeouts = make(map[string]time.Duration, len(idleTimeouts))
	for id, timeout := range idleTimeouts {
		s.idleTimeouts[id] = timeout
	}
	out := make(map[string]int, len(accountIDs))
	for _, id := range accountIDs {
		out[id] = s.counts[id]
	}
	return out, nil
}

type groupCapacityRPMCacheStub struct {
	RPMCache
	counts    map[string]int
	requested []string
}

func (s *groupCapacityRPMCacheStub) GetRPMBatch(_ context.Context, accountIDs []string) (map[string]int, error) {
	s.requested = append([]string(nil), accountIDs...)
	out := make(map[string]int, len(accountIDs))
	for _, id := range accountIDs {
		out[id] = s.counts[id]
	}
	return out, nil
}

func TestGetAllGroupCapacityBatchAggregatesRuntimeAndLimits(t *testing.T) {
	accountRepo := &groupCapacityAccountRepoStub{
		rows: []GroupAccountCapacityRow{
			{
				GroupID:     "group-10",
				AccountID:   "account-1",
				Concurrency: 2,
				Extra: map[string]any{
					"max_sessions":                 3,
					"session_idle_timeout_minutes": 7,
					"base_rpm":                     11,
				},
			},
			{
				GroupID:     "group-20",
				AccountID:   "account-1",
				Concurrency: 2,
				Extra: map[string]any{
					"max_sessions":                 3,
					"session_idle_timeout_minutes": 7,
					"base_rpm":                     11,
				},
			},
			{
				GroupID:     "group-20",
				AccountID:   "account-2",
				Concurrency: 4,
				Extra: map[string]any{
					"max_sessions":                 1,
					"session_idle_timeout_minutes": 9,
					"base_rpm":                     13,
				},
			},
		},
	}
	groupRepo := &groupCapacityGroupRepoStub{groupIDs: []string{"group-10", "group-20"}}
	concurrencyCache := &groupCapacityConcurrencyCacheStub{counts: map[string]int{"account-1": 1, "account-2": 2}}
	sessionCache := &groupCapacitySessionCacheStub{counts: map[string]int{"account-1": 2, "account-2": 1}}
	rpmCache := &groupCapacityRPMCacheStub{counts: map[string]int{"account-1": 5, "account-2": 7}}
	svc := NewGroupCapacityService(
		accountRepo,
		groupRepo,
		NewConcurrencyService(concurrencyCache),
		sessionCache,
		rpmCache,
	)

	results, err := svc.GetAllGroupCapacity(context.Background())
	require.NoError(t, err)

	require.Equal(t, 1, groupRepo.listCalls)
	require.Equal(t, []string{"group-10", "group-20"}, accountRepo.requested)
	require.Equal(t, []string{"account-1", "account-2"}, concurrencyCache.requested)
	require.ElementsMatch(t, []string{"account-1", "account-2"}, sessionCache.requested)
	require.ElementsMatch(t, []string{"account-1", "account-2"}, rpmCache.requested)
	require.Equal(t, 7*time.Minute, sessionCache.idleTimeouts["account-1"])
	require.Equal(t, 9*time.Minute, sessionCache.idleTimeouts["account-2"])

	require.Equal(t, []GroupCapacitySummary{
		{
			GroupID:         "group-10",
			ConcurrencyUsed: 1,
			ConcurrencyMax:  2,
			SessionsUsed:    2,
			SessionsMax:     3,
			RPMUsed:         5,
			RPMMax:          11,
		},
		{
			GroupID:         "group-20",
			ConcurrencyUsed: 3,
			ConcurrencyMax:  6,
			SessionsUsed:    3,
			SessionsMax:     4,
			RPMUsed:         12,
			RPMMax:          24,
		},
	}, results)
}

func TestGetAllGroupCapacityBatchKeepsEmptyGroupRows(t *testing.T) {
	accountRepo := &groupCapacityAccountRepoStub{
		rows: []GroupAccountCapacityRow{
			{GroupID: "group-20", AccountID: "account-2", Concurrency: 4},
		},
	}
	groupRepo := &groupCapacityGroupRepoStub{groupIDs: []string{"group-10", "group-20"}}
	svc := NewGroupCapacityService(accountRepo, groupRepo, nil, nil, nil)

	results, err := svc.GetAllGroupCapacity(context.Background())
	require.NoError(t, err)

	require.Equal(t, []GroupCapacitySummary{
		{GroupID: "group-10"},
		{GroupID: "group-20", ConcurrencyMax: 4},
	}, results)
}
