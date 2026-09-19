package auth

type UserAuth struct {
	ID           int
	Email        string
	Username     string
	RoleID       int
	PasswordHash string
}

type LoginRequest struct {
	EmailOrUsername string `json:"email_or_username" validate:"required,min=8,max=200"`
	Password        string `json:"password" validate:"required,min=8,max=80"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresAt   int64  `json:"expires_at"`
}
