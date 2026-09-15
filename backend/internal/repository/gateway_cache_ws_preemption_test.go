package repository

import (
	"context"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestWSPreemptionCacheOwnership(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	defer client.Close()
	cache := &gatewayCache{rdb: client}
	ctx := context.Background()
	old, err := cache.ClaimOpenAIResponsesSessionWindow(ctx, "group", "scope", []byte("first"), time.Minute)
	require.NoError(t, err)
	require.Empty(t, old)
	old, err = cache.ClaimOpenAIResponsesSessionWindow(ctx, "group", "scope", []byte("second"), time.Minute)
	require.NoError(t, err)
	require.Equal(t, []byte("first"), old)
	deleted, err := cache.CompareAndDeleteOpenAIResponsesSessionWindow(ctx, "group", "scope", []byte("first"))
	require.NoError(t, err)
	require.False(t, deleted)
	refreshed, err := cache.CompareAndRefreshOpenAIResponsesSessionWindow(ctx, "group", "scope", []byte("first"), time.Hour)
	require.NoError(t, err)
	require.False(t, refreshed)
	refreshed, err = cache.CompareAndRefreshOpenAIResponsesSessionWindow(ctx, "group", "scope", []byte("second"), time.Hour)
	require.NoError(t, err)
	require.True(t, refreshed)
	old, err = cache.ClaimOpenAIResponsesSessionWindow(ctx, "other-group", "scope", []byte("third"), time.Minute)
	require.NoError(t, err)
	require.Empty(t, old)
	server.FastForward(2 * time.Minute)
	refreshed, err = cache.CompareAndRefreshOpenAIResponsesSessionWindow(ctx, "other-group", "scope", []byte("third"), time.Minute)
	require.NoError(t, err)
	require.False(t, refreshed)
	deleted, err = cache.CompareAndDeleteOpenAIResponsesSessionWindow(ctx, "group", "scope", []byte("second"))
	require.NoError(t, err)
	require.True(t, deleted)
}
