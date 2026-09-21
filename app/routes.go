package app

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nitin1chandani/ticketmaster/auth"
	"github.com/nitin1chandani/ticketmaster/bookings"
	"github.com/nitin1chandani/ticketmaster/event"
	"github.com/nitin1chandani/ticketmaster/internal/middleware"
	"github.com/nitin1chandani/ticketmaster/user"
	"go.uber.org/zap"
)

func (c *Container) registerRoutes(logger *zap.Logger) {
	c.App.Get("/health", func(ctx *fiber.Ctx) error {
		return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
		})
	})

	authMW := middleware.NewAuthMiddleware(c.Config.JWTSecret)

	//auth
	authRepo := auth.NewRepository(c.DB)
	authService := auth.NewService(authRepo, c.Config.JWTSecret, time.Duration(c.Config.JWTExpiryHours)*time.Hour)
	authHandler := auth.NewHandler(authService, logger)

	//user
	userRepo := user.NewRepository(c.DB)
	userService := user.NewService(userRepo)
	userHandler := user.NewUserHandler(userService)

	//booking
	bookingRepo := bookings.NewBookingRepo(c.DB)
	bookingService := bookings.NewBookingService(bookingRepo)
	bookingHandler := bookings.NewBookingHandler(bookingService)

	// event
	eventRepo := event.NewEventRepo(c.DB)
	eventService := event.NewEventService(eventRepo)
	eventHandler := event.NewEventHandler(eventService)

	api := c.App.Group("/api")
	v1 := api.Group("/v1")

	// auth
	authV1 := v1.Group("/auth")
	authV1.Post("/login", authHandler.Login)

	// user
	userV1 := v1.Group("/user")
	userV1.Post("/register", userHandler.Register)

	//protected
	protectedV1 := v1.Group("/", authMW.RequireJWT)
	protectedV1.Get("/me", authHandler.Me)

	bookingV1 := protectedV1.Group("/booking")
	bookingV1.Post("/:event_id/reserve", bookingHandler.ReserveTickets)
	bookingV1.Post("/:event_id/confirm", bookingHandler.ConfirmBooking)

	eventV1 := protectedV1.Group("/event")
	eventV1.Post("/", eventHandler.CreateNewEvent) // only admin can create we need a middleware
	eventV1.Get("/:event_id", eventHandler.GetEventDetailsWithAvailableTickets)

}
