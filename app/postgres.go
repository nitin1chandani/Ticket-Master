package app

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func newPostgres(databaseURL string, logger *zap.Logger) (*pgxpool.Pool, error) {
	if databaseURL == "" {
		logger.Error("database url is not passed")
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		logger.Error("failed to create pgxpool", zap.Error(err))
		return nil, fmt.Errorf("failed to create pgxpool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		logger.Error("failed to ping postgres", zap.Error(err))
		return nil, fmt.Errorf("failed to ping pgxpool: %w", err)
	}
	return pool, nil
}
