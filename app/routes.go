package app

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nitin1chandani/ticketmaster/auth"
)

func (c *Container) registerRoutes() {
	c.App.Get("/health", func(ctx *fiber.Ctx) error {
		return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
		})
	})

	authRepo := auth.NewRepository(c.DB)
	authService := auth.NewService(authRepo, c.Config.JWTSecret, time.Duration(c.Config.JWTExpiryHours)*time.Hour)
	authHandler := auth.NewHandler(authService)

	api := c.App.Group("/api")
	v1 := api.Group("/v1")

	// auth
	authV1 := v1.Group("/auth")
	authV1.Post("/login", authHandler.Login)

	//next
}
