package middleware

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiterConfig конфиг лимитера
type RateLimiterConfig struct {
	RedisClient *redis.Client
	Limit       int
	Window      time.Duration
	KeyPrefix   string
	Extractor   func(c *gin.Context) string
}

func NewRateLimiter(cfg RateLimiterConfig) gin.HandlerFunc {
	if cfg.KeyPrefix == "" {
		cfg.KeyPrefix = "rl:"
	}
	if cfg.Extractor == nil {
		cfg.Extractor = func(c *gin.Context) string {
			xff := c.Request.Header.Get("X-Forwarded-For")
			if xff != "" {
				parts := strings.Split(xff, ",")
				return strings.TrimSpace(parts[0])
			}
			ra := c.Request.RemoteAddr
			return ra
		}
	}

	return func(c *gin.Context) {
		ctx := c.Request.Context()
		id := cfg.Extractor(c)
		if id == "" {
			id = "anonymous"
		}
		key := fmt.Sprintf("%s%s", cfg.KeyPrefix, id)

		count, err := cfg.RedisClient.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}

		if count == 1 {
			cfg.RedisClient.Expire(ctx, key, cfg.Window)
		}

		if count > int64(cfg.Limit) {
			ttl, _ := cfg.RedisClient.TTL(ctx, key).Result()
			reset := int(ttl.Seconds())
			if reset < 0 {
				reset = 0
			}
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.Limit))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", reset))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":             "rate limit exceeded",
				"rate_limit":        cfg.Limit,
				"rate_limit_window": cfg.Window.String(),
				"retry_after_sec":   reset,
			})
			return
		}

		remaining := cfg.Limit - int(count)
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		ttl, _ := cfg.RedisClient.TTL(ctx, key).Result()
		reset := int(ttl.Seconds())
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", reset))

		c.Next()
	}
}
