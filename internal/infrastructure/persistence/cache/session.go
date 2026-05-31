package cache

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type SessionCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewSessionCache(client *redis.Client, ttl time.Duration) *SessionCache {
	return &SessionCache{client: client, ttl: ttl}
}

func (c *SessionCache) GetVersion(ctx context.Context, userID string) (int16, error) {
	key := fmt.Sprintf("session_version:%s", userID)
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, nil // no version stored means version 0
		}
		return 0, err
	}
	valInt, err := strconv.ParseInt(val, 10, 16)
	return int16(valInt), err
}

func (c *SessionCache) IncrementVersion(ctx context.Context, userID int64) error {
	key := fmt.Sprintf("session_version:%d", userID)

	pipe := c.client.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, c.ttl)

	_, err := pipe.Exec(ctx)
	return err
}
