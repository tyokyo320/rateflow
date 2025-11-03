package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/tyokyo320/rateflow/internal/infrastructure/config"
)

// Cache provides Redis caching functionality.
type Cache struct {
	client *redis.Client
	logger *slog.Logger
}

// NewCache creates a new Redis cache instance.
func NewCache(cfg config.RedisConfig, logger *slog.Logger) *Cache {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	logger.Info("redis cache initialized", "addr", cfg.Addr())

	return &Cache{
		client: client,
		logger: logger,
	}
}

// Get retrieves a value from the cache.
func (c *Cache) Get(ctx context.Context, key string, dest any) error {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key not found: %s", key)
		}
		c.logger.Error("cache get error", "key", key, "error", err)
		return err
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		c.logger.Error("cache unmarshal error", "key", key, "error", err)
		return err
	}

	c.logger.Debug("cache hit", "key", key)
	return nil
}

// Set stores a value in the cache with TTL.
func (c *Cache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		c.logger.Error("cache marshal error", "key", key, "error", err)
		return err
	}

	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		c.logger.Error("cache set error", "key", key, "error", err)
		return err
	}

	c.logger.Debug("cache set", "key", key, "ttl", ttl)
	return nil
}

// Delete removes one or more keys from the cache.
func (c *Cache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	if err := c.client.Del(ctx, keys...).Err(); err != nil {
		c.logger.Error("cache delete error", "keys", keys, "error", err)
		return err
	}

	c.logger.Debug("cache delete", "keys", keys)
	return nil
}

// Exists checks if a key exists in the cache.
func (c *Cache) Exists(ctx context.Context, keys ...string) (int64, error) {
	if len(keys) == 0 {
		return 0, nil
	}

	count, err := c.client.Exists(ctx, keys...).Result()
	if err != nil {
		c.logger.Error("cache exists error", "keys", keys, "error", err)
		return 0, err
	}

	return count, nil
}

// Expire sets a timeout on a key.
func (c *Cache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if err := c.client.Expire(ctx, key, ttl).Err(); err != nil {
		c.logger.Error("cache expire error", "key", key, "error", err)
		return err
	}

	c.logger.Debug("cache expire set", "key", key, "ttl", ttl)
	return nil
}

// Ping checks if the Redis server is reachable.
func (c *Cache) Ping(ctx context.Context) error {
	if err := c.client.Ping(ctx).Err(); err != nil {
		c.logger.Error("redis ping failed", "error", err)
		return err
	}

	return nil
}

// Close closes the Redis connection.
func (c *Cache) Close() error {
	return c.client.Close()
}
