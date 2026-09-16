package app

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func newRedis(addr, password string, db int) (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		_ = redisClient.Close()
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return redisClient, nil
}
