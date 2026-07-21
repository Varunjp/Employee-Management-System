package domain

import (
	"context"
	"time"
)

// Cache is the outbound port for a read-through/write-invalidate cache.
// The usecase layer only knows about this interface; the concrete Redis
// implementation lives in internal/cache/redis.
type Cache interface {
	// Get returns the raw cached bytes for key, and false if it is a miss.
	Get(ctx context.Context, key string) ([]byte, bool, error)
	// Set stores value under key with the given expiration (0 = no expiry).
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	// Delete removes one or more keys from the cache. Missing keys are ignored.
	Delete(ctx context.Context, keys ...string) error
}
