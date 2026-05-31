package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	infraErrors "github.com/victorotene80/medilog-api/internal/infrastructure"
	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
)

type RedisCache[K comparable, V any] struct {
	client     *redis.Client
	prefix     string
	defaultTTL time.Duration
}

var _ appContracts.Cache[string, struct{}] = (*RedisCache[string, struct{}])(nil)

func NewRedisCache[K comparable, V any](
	client *redis.Client,
	prefix string,
	defaultTTL time.Duration,
) *RedisCache[K, V] {
	if prefix == "" {
		prefix = "cache:"
	}
	return &RedisCache[K, V]{
		client:     client,
		prefix:     prefix,
		defaultTTL: defaultTTL,
	}
}

func (c *RedisCache[K, V]) key(k K) string {
	return c.prefix + fmt.Sprint(k)
}


func (c *RedisCache[K, V]) Set(
	ctx context.Context,
	key K,
	value *V,
) error {
	redisKey := c.key(key)

	if value == nil {
		return c.client.Del(ctx, redisKey).Err()
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, redisKey, data, c.defaultTTL).Err()
}

func (c *RedisCache[K, V]) Get(
	ctx context.Context,
	key K,
) (*V, error) {
	redisKey := c.key(key)

	data, err := c.client.Get(ctx, redisKey).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, infraErrors.ErrCacheMiss
		}
		return nil, err
	}

	var v V
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}

	return &v, nil
}

func (c *RedisCache[K, V]) Delete(
	ctx context.Context,
	key K,
) error {
	redisKey := c.key(key)
	return c.client.Del(ctx, redisKey).Err()
}

func (c *RedisCache[K, V]) RefreshTTL(
	ctx context.Context,
	key K,
) error {
	redisKey := c.key(key)

	ok, err := c.client.Expire(ctx, redisKey, c.defaultTTL).Result()
	if err != nil {
		return err
	}
	if !ok {
		return infraErrors.ErrCacheMiss
	}
	return nil
}