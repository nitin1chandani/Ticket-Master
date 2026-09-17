package auth

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserAuthRepository interface {
	GetByEmailOrUsername(ctx context.Context, value string) (*UserAuth, error)
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetByEmailOrUsername(ctx context.Context, value string) (*UserAuth, error) {
	const query = `
		SELECT id, email, username, role_id, password_hash
		FROM users
		WHERE email = $1 OR username = $1
		LIMIT 1
	`

	var user UserAuth

	err := r.db.QueryRow(ctx, query, value).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.RoleID,
		&user.PasswordHash,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}
