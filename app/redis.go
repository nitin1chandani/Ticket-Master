package app

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func newRedis(addr, password string, db int, logger *zap.Logger) (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		_ = redisClient.Close()
		logger.Error("failed to ping redis", zap.Error(err))
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	logger.Info("redis connected", zap.String("addr", addr), zap.Int("db", db))
	return redisClient, nil
}
