package cacher

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// Redis is a Redis-backed cache with string keys and values.
type Redis struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRedis creates new Redis object, ttl=0 means it is never expire.
func NewRedis(db int, ttl time.Duration) *Redis {
	cfg := getConfig()
	return &Redis{
		ttl: ttl,
		client: redis.NewClient(&redis.Options{
			Addr:     cfg.RedisAddr,
			Username: cfg.RedisUsername,
			Password: cfg.RedisPassword,
			DB:       db,
		}),
	}
}

func (c *Redis) Set(ctx context.Context, key, value string) error {
	return c.client.Set(ctx, key, value, c.ttl).Err()
}

func (c *Redis) Get(ctx context.Context, key string) (string, bool, error) {
	value, err := c.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return value, true, nil
}

func (c *Redis) DeletePrefix(ctx context.Context, prefix string) error {
	pattern := prefix + "*"
	iter := c.client.Scan(ctx, 0, pattern, 100).Iterator()

	const batchSize = 100
	keys := make([]string, 0, batchSize)
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
		if len(keys) >= batchSize {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
			keys = keys[:0]
		}
	}
	if len(keys) > 0 {
		if err := c.client.Del(ctx, keys...).Err(); err != nil {
			return err
		}
	}
	if err := iter.Err(); err != nil {
		return err
	}
	return nil
}
