package user

import (
	"context"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nitin1chandani/ticketmaster/internal/httpx"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo UserRepository
}

func NewService(repo *Repository) *UserService {
	return &UserService{
		repo: repo,
	}
}

type RegisterUserInput struct {
	Name     string
	Email    string
	Username string
	Password string
}

func (s *UserService) Register(ctx context.Context, input RegisterUserInput) (*User, error) {
	roleID, err := s.repo.GetRoleIDByName(ctx, "user")
	if err != nil {
		if errors.Is(err, ErrUserRoleNotFound) {
			return nil, &httpx.AppError{
				StatusCode: fiber.StatusInternalServerError,
				Code:       "INTERNAL_SERVER_ERROR",
				Message:    "Internal Server Error",
			}
		}
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	created, err := s.repo.CreateUser(ctx, roleID, CreateUserParams{
		Name:         strings.TrimSpace(input.Name),
		Email:        strings.TrimSpace(strings.ToLower(input.Email)),
		Username:     strings.TrimSpace(input.Username),
		PasswordHash: string(hash),
	})

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, &httpx.AppError{
				StatusCode: fiber.StatusConflict,
				Code:       "USER_ALREADY_EXISTS",
				Message:    "Email or username already exists",
			}
		}
		return nil, err
	}

	return created, nil
}
