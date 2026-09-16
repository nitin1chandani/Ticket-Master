package app

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	// server port
	Port string
	// DB Config
	DatabaseURL string
	// Redis Config
	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

type Container struct {
	Config Config

	DB    *pgxpool.Pool
	Redis *redis.Client
	App   *fiber.App
}

func NewContainer() (*Container, error) {
	config := loadConfig()
	db, err := newPostgres(config.DatabaseURL)
	if err != nil {
		return nil, err
	}

	redisClient, err := newRedis(config.RedisAddr, config.RedisPassword, config.RedisDB)
	if err != nil {
		return nil, err
	}

	app := fiber.New()
	c := &Container{
		Config: config,
		DB:     db,
		Redis:  redisClient,
		App:    app,
	}
	c.registerRoutes()
	return c, nil
}

func (c *Container) Close() error {
	if c.DB != nil {
		c.DB.Close()
	}
	if c.Redis != nil {
		_ = c.Redis.Close()
	}
	return nil
}
