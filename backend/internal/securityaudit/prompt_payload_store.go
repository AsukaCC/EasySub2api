package securityaudit

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type PayloadStore interface {
	Set(ctx context.Context, jobID string, scanText string, ttl time.Duration) error
	Get(ctx context.Context, jobID string) (string, error)
	Delete(ctx context.Context, jobID string) error
	Ping(ctx context.Context) error
}

type RedisPayloadStore struct {
	client *redis.Client
}

func NewRedisPayloadStore(client *redis.Client) *RedisPayloadStore {
	return &RedisPayloadStore{client: client}
}

func (s *RedisPayloadStore) Set(ctx context.Context, jobID string, scanText string, ttl time.Duration) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("prompt audit payload store unavailable")
	}
	if strings.TrimSpace(jobID) == "" || scanText == "" {
		return fmt.Errorf("prompt audit payload input invalid")
	}
	if ttl <= 0 || ttl > DefaultPayloadTTL {
		ttl = DefaultPayloadTTL
	}
	return s.client.Set(ctx, payloadKey(jobID), scanText, ttl).Err()
}

func (s *RedisPayloadStore) Get(ctx context.Context, jobID string) (string, error) {
	if s == nil || s.client == nil {
		return "", fmt.Errorf("prompt audit payload store unavailable")
	}
	value, err := s.client.Get(ctx, payloadKey(jobID)).Result()
	if err == redis.Nil {
		return s.client.Get(ctx, LegacyPayloadKeyPrefix+jobID).Result()
	}
	return value, err
}

func (s *RedisPayloadStore) Delete(ctx context.Context, jobID string) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("prompt audit payload store unavailable")
	}
	return s.client.Del(ctx, payloadKey(jobID), LegacyPayloadKeyPrefix+jobID).Err()
}

func (s *RedisPayloadStore) Ping(ctx context.Context) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("prompt audit payload store unavailable")
	}
	return s.client.Ping(ctx).Err()
}

func payloadKey(jobID string) string {
	return PayloadKeyPrefix + jobID
}
