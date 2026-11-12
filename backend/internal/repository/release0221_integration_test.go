//go:build integration

package repository

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRelease0221RefreshTokenAtomicConsumption(t *testing.T) {
	ctx := context.Background()
	cache := NewRefreshTokenCache(integrationRedis)
	token := uuid.NewString()
	data := &service.RefreshTokenData{UserID: uuid.NewString(), FamilyID: uuid.NewString(), ExpiresAt: time.Now().Add(time.Minute)}
	require.NoError(t, cache.StoreRefreshToken(ctx, token, data, time.Minute))
	t.Cleanup(func() { _ = cache.DeleteRefreshToken(ctx, token) })
	const workers = 32
	results := make(chan error, workers)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			got, err := cache.ConsumeRefreshToken(ctx, token)
			if err == nil && (got.UserID != data.UserID || got.FamilyID != data.FamilyID) {
				err = errors.New("token owner changed")
			}
			results <- err
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	winners := 0
	for err := range results {
		if err == nil {
			winners++
		} else {
			require.ErrorIs(t, err, service.ErrRefreshTokenNotFound)
		}
	}
	require.Equal(t, 1, winners)
	_, err := cache.GetRefreshToken(ctx, token)
	require.ErrorIs(t, err, service.ErrRefreshTokenNotFound)
}
