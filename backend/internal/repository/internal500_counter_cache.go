package repository

import (
	"context"
	"fmt"

	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/redis/go-redis/v9"
)

const (
	internal500CounterPrefix     = "internal500_count:account:"
	internal500CounterTTLSeconds = 86400
)

var internal500CounterIncrScript = redis.NewScript(`
	local key = KEYS[1]
	local ttl = tonumber(ARGV[1])
	local count = redis.call('INCR', key)
	if count == 1 then
		redis.call('EXPIRE', key, ttl)
	end
	return count
`)

type internal500CounterCache struct {
	rdb *redis.Client
}

func NewInternal500CounterCache(rdb *redis.Client) service.Internal500CounterCache {
	return &internal500CounterCache{rdb: rdb}
}

func (c *internal500CounterCache) IncrementInternal500Count(ctx context.Context, accountID string) (int64, error) {
	key := internal500CounterPrefix + accountID
	result, err := internal500CounterIncrScript.Run(ctx, c.rdb, []string{key}, internal500CounterTTLSeconds).Int64()
	if err != nil {
		return 0, fmt.Errorf("increment internal500 count: %w", err)
	}
	return result, nil
}

func (c *internal500CounterCache) ResetInternal500Count(ctx context.Context, accountID string) error {
	return c.rdb.Del(ctx, internal500CounterPrefix+accountID).Err()
}
