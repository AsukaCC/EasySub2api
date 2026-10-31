package service

import (
	"context"
	"log/slog"
	"time"
)

const sessionSchedulingPriorityTTL = 30 * time.Minute

type sessionSchedulingAffinityKey struct{}

type sessionSchedulingAffinity struct {
	cache      GatewayCache
	groupID    string
	cacheKey   string
	priorityID string
}

func sessionSchedulingPriorityKey(cacheKey string) string {
	return "scheduling-priority:" + cacheKey
}

// Keep the short priority window separate from durable protocol/session bindings.
// Cache misses and failures retain the normal scheduling order.
func withSessionSchedulingAffinity(ctx context.Context, cache GatewayCache, groupID *string, cacheKey, stickyID string) context.Context {
	if cache == nil || cacheKey == "" {
		return ctx
	}
	group := derefGroupID(groupID)
	if existing, _ := ctx.Value(sessionSchedulingAffinityKey{}).(*sessionSchedulingAffinity); existing != nil && existing.groupID == group && existing.cacheKey == cacheKey {
		return ctx
	}
	state := &sessionSchedulingAffinity{cache: cache, groupID: group, cacheKey: cacheKey}
	if stickyID != "" {
		id, err := cache.GetSessionAccountID(ctx, group, sessionSchedulingPriorityKey(cacheKey))
		if err == nil && id == stickyID {
			state.priorityID = id
		}
	}
	return context.WithValue(ctx, sessionSchedulingAffinityKey{}, state)
}

func sessionSchedulingPriorityID(ctx context.Context) string {
	if state, _ := ctx.Value(sessionSchedulingAffinityKey{}).(*sessionSchedulingAffinity); state != nil {
		return state.priorityID
	}
	return ""
}

func compareSessionSchedulingPriority(ctx context.Context, a, b *Account) int {
	id := sessionSchedulingPriorityID(ctx)
	if id != "" && a.ID != b.ID {
		if a.ID == id {
			return -1
		}
		if b.ID == id {
			return 1
		}
	}
	return 0
}

func markSessionSchedulingSuccess(ctx context.Context, cache GatewayCache, groupID, cacheKey, accountID string) {
	if cache == nil || cacheKey == "" || accountID == "" {
		return
	}
	if err := cache.SetSessionAccountID(ctx, groupID, sessionSchedulingPriorityKey(cacheKey), accountID, sessionSchedulingPriorityTTL); err != nil {
		slog.Warn("session scheduling priority refresh failed", "account_id", accountID, "error", err)
	}
}

// Called only after acquisition and hydration. Profit-gated and queued requests
// renew at terminal admission in the handler, never at candidate selection.
func renewSessionSchedulingAffinity(ctx context.Context, accountID string) {
	state, _ := ctx.Value(sessionSchedulingAffinityKey{}).(*sessionSchedulingAffinity)
	if state == nil || gatewayProfitControlGateActive(ctx) {
		return
	}
	boundID, err := state.cache.GetSessionAccountID(ctx, state.groupID, state.cacheKey)
	if err == nil && boundID == accountID {
		markSessionSchedulingSuccess(ctx, state.cache, state.groupID, state.cacheKey, accountID)
	}
}

// Promote an eligible account without changing the relative fallback order.
func preferActiveSession[T any](ordered []T, accountID string, accountOf func(T) *Account) {
	if accountID == "" {
		return
	}
	for i, item := range ordered {
		if accountOf(item).ID == accountID {
			copy(ordered[1:i+1], ordered[:i])
			ordered[0] = item
			return
		}
	}
}
