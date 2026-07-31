package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"go-musthave-diploma-tpl/internal/model"

	"github.com/jackc/pgx/v5/pgconn"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, login, passwordHash string) (model.User, error) {
	const q = `
		INSERT INTO users (login, password_hash, balance)
		VALUES ($1, $2, 0)
		RETURNING id, login, password_hash, COALESCE(balance, 0), created_at
	`

	var user model.User
	err := r.db.QueryRowContext(ctx, q, login, passwordHash).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.Balance,
		&user.CreatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return model.User{}, ErrDuplicateUser
		}
		return model.User{}, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func (r *UserRepository) FindUserByLogin(ctx context.Context, login string) (model.User, error) {
	const q = `
		SELECT id, login, password_hash, COALESCE(balance, 0), created_at
		FROM users
		WHERE login = $1
	`

	var user model.User
	err := r.db.QueryRowContext(ctx, q, login).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.Balance,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, ErrUserNotFound
		}
		return model.User{}, fmt.Errorf("find user by login: %w", err)
	}

	return user, nil
}
