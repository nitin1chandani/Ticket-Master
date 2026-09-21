package auth

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/nitin1chandani/ticketmaster/internal/httpx"
	"github.com/nitin1chandani/ticketmaster/internal/middleware"
	"go.uber.org/zap"
)

type Handler struct {
	logger  *zap.Logger
	service *Service
}

func NewHandler(service *Service, logger *zap.Logger) *Handler {
	return &Handler{
		logger:  logger,
		service: service,
	}
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := httpx.BindAndValidate(c, &req); err != nil {
		return err
	}

	resp, err := h.service.Login(c.Context(), req)

	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			h.logger.Warn("login failed: invalid credentials",
				zap.String("email_or_username", req.EmailOrUsername),
				zap.String("path", c.Path()),
			)
			return &httpx.AppError{
				StatusCode: fiber.StatusUnauthorized,
				Code:       "INVALID_CREDENTIALS",
				Message:    "Invalid credentials",
			}
		}
		h.logger.Error("login failed",
			zap.String("email_or_username", req.EmailOrUsername),
			zap.String("path", c.Path()),
			zap.Error(err),
		)
		return &httpx.AppError{
			StatusCode: fiber.StatusInternalServerError,
			Code:       "LOGIN_FAILED",
			Message:    "Failed to login",
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

func (h *Handler) Me(c *fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return &httpx.AppError{
			StatusCode: fiber.StatusUnauthorized,
			Code:       "UNAUTHORIZED",
			Message:    "Unauthorized access",
		}
	}

	username, _ := middleware.UsernameFromContext(c)
	roleID, _ := middleware.RoleIDFromContext(c)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"user_id":  userID,
			"username": username,
			"role_id":  roleID,
		},
	})
}
