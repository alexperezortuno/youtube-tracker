package cache

import (
	"context"
	"time"

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
	err := r.Client.SAdd(ctx, "streams:active", videoID).Err()
	if err == nil {
		r.Client.Expire(ctx, "streams:active", 24*time.Hour)
	}
	return err
}

func (r *RedisClient) GetStreams(ctx context.Context) ([]string, error) {
	return r.Client.SMembers(ctx, "streams:active").Result()
}

func (r *RedisClient) DeleteStream(ctx context.Context, videoID string) error {
	return r.Client.SRem(ctx, "streams:active", videoID).Err()
}

func (r *RedisClient) IncrementDeadCounter(ctx context.Context, videoID string) (int64, error) {
	key := "stream:dead:" + videoID
	return r.Client.Incr(ctx, key).Result()
}

func (r *RedisClient) ResetDeadCounter(ctx context.Context, videoID string) error {
	key := "stream:dead:" + videoID
	return r.Client.Del(ctx, key).Err()
}

func (r *RedisClient) RemoveStream(ctx context.Context, videoID string) error {
	return r.Client.SRem(ctx, "streams:active", videoID).Err()
}
