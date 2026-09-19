package auth

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/nitin1chandani/ticketmaster/internal/httpx"
	"github.com/nitin1chandani/ticketmaster/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
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
			return &httpx.AppError{
				StatusCode: fiber.StatusUnauthorized,
				Code:       "INVALID_CREDENTIALS",
				Message:    "Invalid credentials",
			}
		}
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
