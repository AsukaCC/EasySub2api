package service

import (
	"context"
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/pkg/ctxkey"
)

const keySchedulingIdleInterval = 5 * time.Minute

// AccountKeyActivity separates account ownership from request concurrency.
// OtherKeys counts competing owners; OwnRequests counts every in-flight request.
// Requests are never merged, and each still acquires its own concurrency slot.
type AccountKeyActivity struct {
	OtherKeys       int
	OwnRequests     int
	LastScheduledAt time.Time
}

type AccountKeyActivityReader interface {
	GetAccountKeyActivity(context.Context, []string, string) (map[string]AccountKeyActivity, error)
}

type AccountKeyLeaseStore interface {
	AcquireAccountKeyLease(context.Context, string, string, string, int, []AccountKeyPeer) (bool, error)
	RefreshAccountKeyLease(context.Context, string, string, string) error
	ReleaseAccountKeyLease(context.Context, string, string, string) error
}

type AccountKeyPeer struct {
	ID             string
	MaxConcurrency int
}

type keySchedulingContextKey struct{}
type keySchedulingSnapshot struct {
	activity    map[string]AccountKeyActivity
	now         time.Time
	peers       map[string][]AccountKeyPeer
	mu          sync.Mutex
	unavailable map[string]bool
}

type keySchedulingCaptureKey struct{}
type keySchedulingCapture struct{ snapshot *keySchedulingSnapshot }

func captureKeyScheduling(ctx context.Context) (context.Context, *keySchedulingCapture) {
	capture := &keySchedulingCapture{}
	return context.WithValue(ctx, keySchedulingCaptureKey{}, capture), capture
}

func (c *keySchedulingCapture) context(ctx context.Context) context.Context {
	if c.snapshot == nil {
		return ctx
	}
	return context.WithValue(ctx, keySchedulingContextKey{}, c.snapshot)
}

func deferKeySchedulingAffinity(ctx context.Context) bool {
	capture, _ := ctx.Value(keySchedulingCaptureKey{}).(*keySchedulingCapture)
	return capture != nil && capture.snapshot != nil
}

func withKeyScheduling(ctx context.Context, cache any, accounts []*Account) context.Context {
	keyID, _ := ctx.Value(ctxkey.APIKeyID).(string)
	reader, ok := cache.(AccountKeyActivityReader)
	if keyID == "" || !ok || len(accounts) == 0 {
		return ctx
	}
	ids := make([]string, 0, len(accounts))
	for _, account := range accounts {
		ids = append(ids, account.ID)
	}
	activity, err := reader.GetAccountKeyActivity(ctx, ids, keyID)
	if err != nil {
		slog.Warn("account key scheduling snapshot unavailable", "error", err)
		return ctx
	}
	snapshot := &keySchedulingSnapshot{activity: activity, now: time.Now(), peers: map[string][]AccountKeyPeer{}}
	if capture, _ := ctx.Value(keySchedulingCaptureKey{}).(*keySchedulingCapture); capture != nil {
		capture.snapshot = snapshot
	}
	return context.WithValue(ctx, keySchedulingContextKey{}, snapshot)
}

func withKeySchedulingAccounts(ctx context.Context, cache any, accounts []Account) context.Context {
	pointers := make([]*Account, len(accounts))
	for i := range accounts {
		pointers[i] = &accounts[i]
	}
	return withKeyScheduling(ctx, cache, pointers)
}

func keySchedulingFromContext(ctx context.Context) *keySchedulingSnapshot {
	snapshot, _ := ctx.Value(keySchedulingContextKey{}).(*keySchedulingSnapshot)
	return snapshot
}

func (s *keySchedulingSnapshot) idleSince(account *Account) time.Time {
	activity := s.activity[account.ID]
	if activity.OtherKeys > 0 || activity.OwnRequests > 0 {
		return time.Time{}
	}
	last := activity.LastScheduledAt
	if account.LastUsedAt != nil && account.LastUsedAt.After(last) {
		last = *account.LastUsedAt
	}
	if last.IsZero() {
		last = account.CreatedAt
	}
	if last.IsZero() {
		last = time.Unix(0, 0)
	}
	if s.now.Sub(last) >= keySchedulingIdleInterval {
		return last
	}
	return time.Time{}
}

// compare is used only within an existing business tier and account priority.
func (s *keySchedulingSnapshot) compare(a, b *Account) int {
	if s == nil {
		return 0
	}
	x, xok := s.activity[a.ID]
	y, yok := s.activity[b.ID]
	if !xok || !yok {
		return 0
	}
	if x.OtherKeys < y.OtherKeys {
		return -1
	}
	if x.OtherKeys > y.OtherKeys {
		return 1
	}
	if (x.OwnRequests > 0) != (y.OwnRequests > 0) {
		if x.OwnRequests > 0 {
			return -1
		}
		return 1
	}
	ax, bx := s.idleSince(a), s.idleSince(b)
	if ax.IsZero() != bx.IsZero() {
		if !ax.IsZero() {
			return -1
		}
		return 1
	}
	if !ax.IsZero() {
		if ax.Before(bx) {
			return -1
		}
		if bx.Before(ax) {
			return 1
		}
	}
	return 0
}

// Reorder only interchangeable positions, preserving model/compact/pool tiers.
func applyKeySchedulingOrder[T any](ctx context.Context, ordered []T, accountOf func(T) *Account, samePool func(*Account, *Account) bool) {
	snapshot := keySchedulingFromContext(ctx)
	if snapshot == nil {
		return
	}
	done := make([]bool, len(ordered))
	for i := range ordered {
		if done[i] {
			continue
		}
		indices := []int{}
		pool := []T{}
		for j := i; j < len(ordered); j++ {
			if !done[j] && samePool(accountOf(ordered[i]), accountOf(ordered[j])) {
				indices = append(indices, j)
				pool = append(pool, ordered[j])
				done[j] = true
			}
		}
		sort.SliceStable(pool, func(i, j int) bool { return snapshot.compare(accountOf(pool[i]), accountOf(pool[j])) < 0 })
		for j, index := range indices {
			ordered[index] = pool[j]
		}
	}
}

func filterByKeyScheduling(ctx context.Context, candidates []accountWithLoad) []accountWithLoad {
	snapshot := keySchedulingFromContext(ctx)
	if snapshot == nil || len(candidates) < 2 {
		return candidates
	}
	setKeySchedulingPeers(ctx, candidates, schedulingLoadAccount, sameAccountPriority)
	best := candidates[0].account
	for _, candidate := range candidates {
		if snapshot.compare(candidate.account, best) < 0 {
			best = candidate.account
		}
	}
	result := make([]accountWithLoad, 0, len(candidates))
	for _, candidate := range candidates {
		if snapshot.compare(candidate.account, best) == 0 {
			result = append(result, candidate)
		}
	}
	return result
}

func setKeySchedulingPeers[T any](ctx context.Context, candidates []T, accountOf func(T) *Account, samePool func(*Account, *Account) bool) {
	snapshot := keySchedulingFromContext(ctx)
	if snapshot == nil {
		return
	}
	snapshot.peers = make(map[string][]AccountKeyPeer, len(candidates))
	for _, candidate := range candidates {
		account := accountOf(candidate)
		if _, exists := snapshot.peers[account.ID]; exists {
			continue
		}
		pool := []AccountKeyPeer{}
		for _, other := range candidates {
			peer := accountOf(other)
			if samePool(account, peer) {
				pool = append(pool, AccountKeyPeer{ID: peer.ID, MaxConcurrency: peer.Mode1EffectiveConcurrency()})
			}
		}
		for _, peer := range pool {
			snapshot.peers[peer.ID] = pool
		}
	}
}

func excludeKeySchedulingPeer(ctx context.Context, accountID string) {
	if snapshot := keySchedulingFromContext(ctx); snapshot != nil {
		snapshot.mu.Lock()
		defer snapshot.mu.Unlock()
		if snapshot.unavailable == nil {
			snapshot.unavailable = map[string]bool{}
		}
		snapshot.unavailable[accountID] = true
	}
}

func acquireKeySchedulingLease(ctx context.Context, cache any, accountID, requestID string) (bool, func(), error) {
	keyID, _ := ctx.Value(ctxkey.APIKeyID).(string)
	store, ok := cache.(AccountKeyLeaseStore)
	if keyID == "" || !ok {
		return true, func() {}, nil
	}
	expected := -1
	var peers []AccountKeyPeer
	if snapshot := keySchedulingFromContext(ctx); snapshot != nil {
		if activity, found := snapshot.activity[accountID]; found {
			expected = activity.OtherKeys
		}
		snapshot.mu.Lock()
		for _, peer := range snapshot.peers[accountID] {
			if peer.ID != accountID && !snapshot.unavailable[peer.ID] {
				peers = append(peers, peer)
			}
		}
		snapshot.mu.Unlock()
		if len(peers) == 0 {
			expected = -1
		}
	}
	acquired, err := store.AcquireAccountKeyLease(ctx, accountID, keyID, requestID, expected, peers)
	if err != nil || !acquired {
		return false, nil, err
	}
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(20 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				refreshCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				err := store.RefreshAccountKeyLease(refreshCtx, accountID, keyID, requestID)
				cancel()
				if err != nil {
					slog.Warn("account key lease refresh failed", "account_id", accountID, "error", err)
				}
			}
		}
	}()
	var once sync.Once
	return true, func() {
		once.Do(func() {
			excludeKeySchedulingPeer(ctx, accountID)
			close(stop)
			<-done
			releaseCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if err := store.ReleaseAccountKeyLease(releaseCtx, accountID, keyID, requestID); err != nil {
				slog.Warn("account key lease release failed", "account_id", accountID, "error", err)
			}
		})
	}, nil
}
