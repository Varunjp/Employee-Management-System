package redis

import (
	"context"
	"employee_management/internal/domain"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type cache struct {
	client *redis.Client
}

func NewCache(client *redis.Client) domain.Cache {
	return &cache{client: client}
}

func (c *cache) Get(ctx context.Context, key string) ([]byte, bool, error) {
	value, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("redis: get %q: %w", key, err)
	}
	return value, true, nil
}

func (c *cache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if err := c.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("redis: set %q: %w", key, err)
	}
	return nil
}

func (c *cache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	if err := c.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("redis: delete %v: %w", keys, err)
	}
	return nil
}
