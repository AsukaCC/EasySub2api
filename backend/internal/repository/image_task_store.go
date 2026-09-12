package repository

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/redis/go-redis/v9"
)

const imageTaskKeyPrefix = "image_task:"

type imageTaskStore struct {
	rdb *redis.Client
}

const imageTaskQueueStream = "image_tasks:queue"
const imageTaskQueueGroup = "easysub2api-workers"

func NewImageTaskStore(rdb *redis.Client) service.ImageTaskStore {
	return &imageTaskStore{rdb: rdb}
}

func (s *imageTaskStore) Save(ctx context.Context, task *service.ImageTaskRecord, ttl time.Duration) error {
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, imageTaskKey(task.ID), data, ttl).Err()
}

func (s *imageTaskStore) Get(ctx context.Context, id string) (*service.ImageTaskRecord, error) {
	data, err := s.rdb.Get(ctx, imageTaskKey(id)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, service.ErrImageTaskNotFound
		}
		return nil, err
	}
	var task service.ImageTaskRecord
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *imageTaskStore) Delete(ctx context.Context, id string) error {
	return s.rdb.Del(ctx, imageTaskKey(id)).Err()
}

func (s *imageTaskStore) CompareAndSetStatus(ctx context.Context, id, expected string, task *service.ImageTaskRecord, ttl time.Duration) error {
	key := imageTaskKeyPrefix + id
	payload, err := json.Marshal(task)
	if err != nil {
		return err
	}
	secs := int64(ttl / time.Second)
	if secs < 1 {
		secs = 1
	}
	script := redis.NewScript(`local current = redis.call('GET', KEYS[1]); if not current then return 0 end; local ok = string.find(current, '"status":"' .. ARGV[1] .. '"', 1, true); if not ok then return 0 end; redis.call('SET', KEYS[1], ARGV[2], 'EX', ARGV[3]); return 1`)
	result, err := script.Run(ctx, s.rdb, []string{key}, expected, string(payload), secs).Int()
	if err != nil {
		return err
	}
	if result == 0 {
		return service.ErrImageTaskCanceled
	}
	return nil
}

type ImageTaskQueue struct{ rdb *redis.Client }

func NewImageTaskQueue(rdb *redis.Client) *ImageTaskQueue { return &ImageTaskQueue{rdb: rdb} }

func (q *ImageTaskQueue) Ensure(ctx context.Context) error {
	_, err := q.rdb.XGroupCreateMkStream(ctx, imageTaskQueueStream, imageTaskQueueGroup, "0").Result()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return err
	}
	return nil
}

func (q *ImageTaskQueue) Enqueue(ctx context.Context, taskID string) error {
	_, err := q.rdb.XAdd(ctx, &redis.XAddArgs{Stream: imageTaskQueueStream, Values: map[string]any{"task_id": taskID}}).Result()
	return err
}

func (q *ImageTaskQueue) Read(ctx context.Context, consumer string, block time.Duration) ([]redis.XMessage, error) {
	entries, err := q.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{Group: imageTaskQueueGroup, Consumer: consumer, Streams: []string{imageTaskQueueStream, ">"}, Count: 1, Block: block, NoAck: false}).Result()
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, nil
	}
	return entries[0].Messages, nil
}

func (q *ImageTaskQueue) Claim(ctx context.Context, consumer string, minIdle time.Duration) ([]redis.XMessage, error) {
	result, _, err := q.rdb.XAutoClaim(ctx, &redis.XAutoClaimArgs{Stream: imageTaskQueueStream, Group: imageTaskQueueGroup, Consumer: consumer, MinIdle: minIdle, Start: "0-0", Count: 10}).Result()
	return result, err
}

func (q *ImageTaskQueue) Ack(ctx context.Context, id string) error {
	_, err := q.rdb.XAck(ctx, imageTaskQueueStream, imageTaskQueueGroup, id).Result()
	return err
}

func imageTaskKey(id string) string {
	return imageTaskKeyPrefix + strings.TrimSpace(id)
}
