package user

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nitin1chandani/ticketmaster/internal/httpx"
)

type UserHandler struct {
	service *UserService
}

func NewUserHandler(service *UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email,max=150"`
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type RegisterResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	RoleID    int       `json:"role_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *UserHandler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := httpx.BindAndValidate(c, &req); err != nil {
		return err
	}

	user, err := h.service.Register(c.Context(), RegisterUserInput{
		Name:     req.Name,
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	})

	if err != nil {
		return err
	}

	resp := RegisterResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Username:  user.Username,
		RoleID:    user.RoleID,
		CreatedAt: user.CreatedAt,
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}
