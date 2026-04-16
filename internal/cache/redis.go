package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
}

func NewRedis(addr string) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisClient{Client: rdb}
}

func (r *RedisClient) AddStream(ctx context.Context, videoID string) error {
	return r.Client.SAdd(ctx, "streams:active", videoID).Err()
}

func (r *RedisClient) GetStreams(ctx context.Context) ([]string, error) {
	return r.Client.SMembers(ctx, "streams:active").Result()
}
