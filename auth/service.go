package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type Service struct {
	userAuthRepo UserAuthRepository
	jwtSecret    []byte
	jwtTTL       time.Duration
}

func NewService(userAuthRepo UserAuthRepository, jwtSecret string, jwtTTL time.Duration) *Service {
	return &Service{
		userAuthRepo: userAuthRepo,
		jwtSecret:    []byte(jwtSecret),
		jwtTTL:       jwtTTL,
	}
}

func (s *Service) Login(ctx context.Context, input LoginRequest) (*LoginResponse, error) {
	user, err := s.userAuthRepo.GetByEmailOrUsername(ctx, input.EmailOrUsername)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	expiresAt := time.Now().Add(s.jwtTTL)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      user.ID,
		"email":    user.Email,
		"username": user.Username,
		"role_id":  user.RoleID,
		"iat":      time.Now().Unix(),
		"exp":      expiresAt.Unix(),
	})

	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		AccessToken: signed,
		TokenType:   "Bearer",
		ExpiresAt:   expiresAt.Unix(),
	}, nil
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}
