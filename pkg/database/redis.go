package database

import (
	"context"
	"employee_management/config"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient creates and validates a Redis client, retrying briefly to
// tolerate Redis starting up concurrently with the API in Docker Compose.
func NewRedisClient(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	var lastErr error
	for attempt := 1; attempt <= 5; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		lastErr = client.Ping(ctx).Err()
		cancel()
		if lastErr == nil {
			return client, nil
		}
		time.Sleep(time.Duration(attempt) * time.Second)
	}

	return nil, fmt.Errorf("database: ping redis after retries: %w", lastErr)
}
