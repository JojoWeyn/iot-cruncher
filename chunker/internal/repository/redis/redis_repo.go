package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
	client *redis.Client
}

func NewRedisRepo(client *redis.Client) *RedisRepo {
	return &RedisRepo{client: client}
}

func (r *RedisRepo) SetChunkStatus(ctx context.Context, jobID string, chunkID int, status string) error {
	key := fmt.Sprintf("job:%s:chunk:%d", jobID, chunkID)
	return r.client.Set(ctx, key, status, 0).Err()
}

func (r *RedisRepo) GetJobProgress(ctx context.Context, jobID string) (completed, total int, err error) {
	pattern := fmt.Sprintf("job:%s:chunk:*", jobID)
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return 0, 0, err
	}
	total = len(keys)
	for _, k := range keys {
		val, err := r.client.Get(ctx, k).Result()
		if err != nil {
			continue
		}
		if val == "PUBLISHED" {
			completed++
		}
	}
	return completed, total, nil
}
