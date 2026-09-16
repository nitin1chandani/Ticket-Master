package app

import "github.com/gofiber/fiber/v2"

func (c *Container) registerRoutes() {
	c.App.Get("/health", func(ctx *fiber.Ctx) error {
		return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
		})
	})
}
