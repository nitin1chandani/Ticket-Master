package app

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nitin1chandani/ticketmaster/internal/httpx"
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

	// JWT CONFIG
	JWTSecret      string
	JWTExpiryHours int
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

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			var appErr *httpx.AppError
			if errors.As(err, &appErr) {
				return c.Status(appErr.StatusCode).JSON(fiber.Map{
					"success": false,
					"code":    appErr.Code,
					"message": appErr.Message,
					"fields":  appErr.Fields,
				})
			}

			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				return c.Status(fiberErr.Code).JSON(fiber.Map{
					"success": false,
					"code":    "REQUEST_FAILED",
					"message": fiberErr.Message,
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Internal server error",
			})
		},
	})

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
