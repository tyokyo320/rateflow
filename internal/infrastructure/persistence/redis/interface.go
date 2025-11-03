package redis

import (
	"context"
	"time"
)

// CacheInterface defines the interface for cache operations.
// This allows for easier mocking in tests.
type CacheInterface interface {
	Get(ctx context.Context, key string, dest any) error
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, keys ...string) (int64, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
	Ping(ctx context.Context) error
	Close() error
}

// Ensure Cache implements CacheInterface
var _ CacheInterface = (*Cache)(nil)
