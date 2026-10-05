package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"game_backlog_api/app/model"
)

// 1. Interface & Structs
type UserRepository interface {
	FindByID(ctx context.Context, id int) (model.User, error)
	FindByUsernameOrEmail(ctx context.Context, identifier string) (model.User, error)
	Create(ctx context.Context, u model.User) (model.User, error)
	UpdateRole(ctx context.Context, id int, role string) (model.User, error)
}

type userPostgresRepository struct {
	pool *pgxpool.Pool
}

// 2. Constructor
func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userPostgresRepository{pool: pool}
}

// 3. Helper
const userColumns = "id, username, email, password, role, is_active, created_at"

func scanUser(row pgx.Row) (model.User, error) {
	var u model.User
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Role, &u.IsActive, &u.CreatedAt)
	return u, err
}

// 4. Find By ID (Get One)
func (r *userPostgresRepository) FindByID(ctx context.Context, id int) (model.User, error) {
	query := fmt.Sprintf(`SELECT %s FROM users WHERE id = $1`, userColumns)
	u, err := scanUser(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("fetch user: %w", err)
	}
	return u, nil
}

// 5. Find by Username or Email
func (r *userPostgresRepository) FindByUsernameOrEmail(ctx context.Context, identifier string) (model.User, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM users
         WHERE LOWER(username) = LOWER($1) OR LOWER(email) = LOWER($1)`,
		userColumns,
	)
	u, err := scanUser(r.pool.QueryRow(ctx, query, identifier))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("fetch user by identifier: %w", err)
	}
	return u, nil
}

// 6. Create (Insert)
func (r *userPostgresRepository) Create(ctx context.Context, u model.User) (model.User, error) {
	if u.Role == "" {
		u.Role = "user"
	}
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (username, email, password, role, is_active)
         VALUES ($1, $2, $3, $4, $5)
         RETURNING id, created_at`,
		u.Username, u.Email, u.Password, u.Role, u.IsActive,
	).Scan(&u.ID, &u.CreatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return model.User{}, ErrDuplicate
		}
		return model.User{}, fmt.Errorf("save user: %w", err)
	}

	return u, nil
}

// 7. Update Role
func (r *userPostgresRepository) UpdateRole(ctx context.Context, id int, role string) (model.User, error) {
	query := fmt.Sprintf(
		`UPDATE users SET role = $1
         WHERE id = $2
         RETURNING %s`,
		userColumns,
	)
	u, err := scanUser(r.pool.QueryRow(ctx, query, role, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("update user role: %w", err)
	}
	return u, nil
}
