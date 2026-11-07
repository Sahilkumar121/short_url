package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisCache(client *redis.Client, ttl time.Duration) *RedisCache {
	return &RedisCache{
		client: client,
		ttl:    ttl,
	}
}

func (r *RedisCache) Key(shortCode string) string {
	return "shortCode:redirect:" + shortCode
}

func (r *RedisCache) Get(ctx context.Context, shortCode string) (string, bool, error) {

	value, err := r.client.Get(ctx, r.Key(shortCode)).Result()

	if errors.Is(err, redis.Nil) {
		return "", false, nil
	}

	if err != nil {
		return "", false, fmt.Errorf("failed to retrieve from redis %w", err)
	}

	return value, true, nil
}

func (r *RedisCache) Set(ctx context.Context, shortCode string, longUrl string) error {

	err := r.client.Set(ctx, r.Key(shortCode), longUrl, r.ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set in redis %w", err)
	}

	return nil
}

func (r *RedisCache) Delete(ctx context.Context, shortCode string) error {

	err := r.client.Del(ctx, r.Key(shortCode)).Err()
	if err != nil {
		return fmt.Errorf("failed to delete in redis %w", err)
	}

	return nil
}
