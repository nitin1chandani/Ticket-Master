package app

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nitin1chandani/ticketmaster/internal/httpx"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
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

	// development or production
	Environment string
}

type Container struct {
	Config Config

	DB     *pgxpool.Pool
	Redis  *redis.Client
	App    *fiber.App
	Logger *zap.Logger
}

func NewContainer() (*Container, error) {
	config := loadConfig()
	logger, err := newLogger(config.Environment)
	if err != nil {
		return nil, err
	}
	db, err := newPostgres(config.DatabaseURL, logger)
	if err != nil {
		logger.Error("failed to initialize postgres", zap.Error(err))
		return nil, err
	}

	redisClient, err := newRedis(config.RedisAddr, config.RedisPassword, config.RedisDB, logger)
	if err != nil {
		logger.Error("failed to initialize redis", zap.Error(err))
		return nil, err
	}

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			var appErr *httpx.AppError
			if errors.As(err, &appErr) {
				logger.Warn("request failed with app error",
					zap.String("path", c.Path()),
					zap.String("method", c.Method()),
					zap.String("code", appErr.Code),
					zap.Error(err),
				)
				return c.Status(appErr.StatusCode).JSON(fiber.Map{
					"success": false,
					"code":    appErr.Code,
					"message": appErr.Message,
					"fields":  appErr.Fields,
				})
			}

			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				logger.Warn("request failed with fiber error",
					zap.String("path", c.Path()),
					zap.String("method", c.Method()),
					zap.Int("status", fiberErr.Code),
					zap.Error(err),
				)
				return c.Status(fiberErr.Code).JSON(fiber.Map{
					"success": false,
					"code":    "REQUEST_FAILED",
					"message": fiberErr.Message,
				})
			}

			logger.Error("unhandled request error",
				zap.String("path", c.Path()),
				zap.String("method", c.Method()),
				zap.Error(err),
			)

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
		Logger: logger,
	}
	c.registerRoutes(logger)
	return c, nil
}

func newLogger(env string) (*zap.Logger, error) {
	if env == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}

func (c *Container) Close() error {
	if c.DB != nil {
		c.DB.Close()
	}
	if c.Redis != nil {
		_ = c.Redis.Close()
	}
	if c.Logger != nil {
		c.Logger.Sync()
	}
	return nil
}
