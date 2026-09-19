package middleware

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/nitin1chandani/ticketmaster/internal/httpx"
)

const (
	LocalUserID   = "auth.user_id"
	LocalUsername = "auth.username"
	LocalRoleID   = "auth.role_id"
)

type AuthMiddleware struct {
	jwtSecret []byte
}

func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret: []byte(jwtSecret),
	}
}

func (m *AuthMiddleware) RequireJWT(c *fiber.Ctx) error {
	authHeader := strings.TrimSpace(c.Get("Authorization"))

	if authHeader == "" {
		return unauthorizedError("Missing authorization header")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return unauthorizedError("Invalid authorization header format")
	}

	tokenString := strings.TrimSpace(parts[1])
	if tokenString == "" {
		return unauthorizedError("Token is required")
	}

	token, err := jwt.Parse(
		tokenString,
		func(token *jwt.Token) (any, error) {
			return m.jwtSecret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)

	if err != nil || !token.Valid {
		return unauthorizedError("Invalid or expired token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return unauthorizedError("Invalid token claims")
	}

	userID, ok := intFromClaim(claims["sub"])
	if !ok {
		return unauthorizedError("Invalid token subject")
	}

	roleID, ok := intFromClaim(claims["role_id"])
	if !ok {
		return unauthorizedError("Invalid token role")
	}

	username, ok := claims["username"].(string)
	if !ok || strings.TrimSpace(username) == "" {
		return unauthorizedError("Invalid token username")
	}

	c.Locals(LocalUserID, userID)
	c.Locals(LocalUsername, username)
	c.Locals(LocalRoleID, roleID)

	return c.Next()
}

func UserIDFromContext(c *fiber.Ctx) (int, bool) {
	v := c.Locals(LocalUserID)
	id, ok := v.(int)
	return id, ok
}

func RoleIDFromContext(c *fiber.Ctx) (int, bool) {
	v := c.Locals(LocalRoleID)
	id, ok := v.(int)
	return id, ok
}

func UsernameFromContext(c *fiber.Ctx) (string, bool) {
	v := c.Locals(LocalUsername)
	username, ok := v.(string)
	return username, ok
}

func unauthorizedError(msg string) *httpx.AppError {
	return &httpx.AppError{
		StatusCode: fiber.StatusUnauthorized,
		Code:       "UNAUTHORIZED",
		Message:    msg,
	}
}

func intFromClaim(v any) (int, bool) {
	switch t := v.(type) {
	case float64:
		return int(t), true
	case int:
		return t, true
	case int64:
		return int(t), true
	case string:
		n, err := strconv.Atoi(t)
		if err != nil {
			return 0, false
		}
		return n, true
	default:
		return 0, false
	}
}
