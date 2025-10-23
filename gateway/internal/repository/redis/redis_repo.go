package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
	Client *redis.Client
}

func NewRedisRepo(client *redis.Client) *RedisRepo {
	return &RedisRepo{Client: client}
}

func (r *RedisRepo) SetStatus(ctx context.Context, jobID, status string) error {
	return r.Client.Set(ctx, "job_status:"+jobID, status, time.Hour).Err()
}

func (r *RedisRepo) GetStatus(ctx context.Context, jobID string) (string, error) {
	return r.Client.Get(ctx, "job_status:"+jobID).Result()
}
