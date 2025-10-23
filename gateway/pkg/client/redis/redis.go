package redis

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Addr string
	DB   int
}

func NewRedisClient(ctx context.Context, cfg Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: cfg.Addr,
		DB:   cfg.DB,
	})

	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Redis connection failed: %v", err)
	}

	return client, nil
}
