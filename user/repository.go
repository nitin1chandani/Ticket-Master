package user

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUserRoleNotFound = errors.New("user role not found")

type CreateUserParams struct {
	Name         string
	Email        string
	RoleID       int64
	Username     string
	PasswordHash string
}

type UserRepository interface {
	GetRoleIDByName(ctx context.Context, roleName string) (int, error)
	CreateUser(ctx context.Context, roleID int, p CreateUserParams) (*User, error)
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) GetRoleIDByName(ctx context.Context, roleName string) (int, error) {
	q := "SELECT id FROM roles WHERE name = $1 LIMIT 1"
	var id int
	err := r.db.QueryRow(ctx, q, roleName).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrUserRoleNotFound
		}
		return 0, err
	}
	return id, nil
}

func (r *Repository) CreateUser(ctx context.Context, roleID int, p CreateUserParams) (*User, error) {
	q := "" +
		"INSERT INTO users (name, email, username, role_id, password_hash, created_by, created_at) " +
		"VALUES ($1, $2, $3, $4, $5, $6, $7) " +
		"RETURNING id, name, email, username, role_id, created_at"

	u := &User{}
	err := r.db.QueryRow(
		ctx, q,
		p.Name, p.Email, p.Username, roleID, p.PasswordHash, "self_registration", time.Now(),
	).Scan(&u.ID, &u.Name, &u.Email, &u.Username, &u.RoleID, &u.CreatedAt)

	if err != nil {
		return nil, err
	}
	return u, nil
}
