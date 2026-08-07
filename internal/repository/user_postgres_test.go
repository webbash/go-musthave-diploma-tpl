package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
	"go-musthave-diploma-tpl/internal/repository"
)

func TestUserRepositoryCreateUser(t *testing.T) {
	db, mock := newMockDB(t)
	repo := repository.NewUserRepository(db)

	t.Run("success", func(t *testing.T) {
		rows := sqlmockNewRows([]string{"id", "login", "password_hash", "balance", "created_at"}).
			AddRow(int64(1), "alice", "hash", 0, time.Now())
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO users (login, password_hash, balance)
		VALUES ($1, $2, 0)
		RETURNING id, login, password_hash, COALESCE(balance, 0), created_at`)).
			WithArgs("alice", "hash").
			WillReturnRows(rows)

		user, err := repo.CreateUser(context.Background(), "alice", "hash")
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		if user.ID != 1 || user.Login != "alice" {
			t.Fatalf("CreateUser() = %+v", user)
		}
	})

	t.Run("duplicate", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO users (login, password_hash, balance)
		VALUES ($1, $2, 0)
		RETURNING id, login, password_hash, COALESCE(balance, 0), created_at`)).
			WithArgs("alice", "hash").
			WillReturnError(&pgconn.PgError{Code: "23505"})

		_, err := repo.CreateUser(context.Background(), "alice", "hash")
		if !errors.Is(err, repository.ErrDuplicateUser) {
			t.Fatalf("CreateUser() error = %v, want duplicate user", err)
		}
	})
}

func TestUserRepositoryFindUserByLogin(t *testing.T) {
	db, mock := newMockDB(t)
	repo := repository.NewUserRepository(db)

	t.Run("success", func(t *testing.T) {
		rows := sqlmockNewRows([]string{"id", "login", "password_hash", "balance", "created_at"}).
			AddRow(int64(1), "alice", "hash", 0, time.Now())
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, login, password_hash, COALESCE(balance, 0), created_at
		FROM users
		WHERE login = $1`)).
			WithArgs("alice").
			WillReturnRows(rows)

		user, err := repo.FindUserByLogin(context.Background(), "alice")
		if err != nil {
			t.Fatalf("FindUserByLogin() error = %v", err)
		}
		if user.Login != "alice" {
			t.Fatalf("FindUserByLogin() = %+v", user)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, login, password_hash, COALESCE(balance, 0), created_at
		FROM users
		WHERE login = $1`)).
			WithArgs("ghost").
			WillReturnError(sql.ErrNoRows)

		_, err := repo.FindUserByLogin(context.Background(), "ghost")
		if !errors.Is(err, repository.ErrUserNotFound) {
			t.Fatalf("FindUserByLogin() error = %v, want user not found", err)
		}
	})
}

func sqlmockNewRows(columns []string) *sqlmock.Rows {
	return sqlmock.NewRows(columns)
}
