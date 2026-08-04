package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"go-musthave-diploma-tpl/internal/domain"
	"go-musthave-diploma-tpl/internal/repository"
)

func TestBalanceRepositoryGetBalance(t *testing.T) {
	db, mock := newMockDB(t)
	repo := repository.NewBalanceRepository(db)

	query := regexp.QuoteMeta(`SELECT COALESCE(balance, 0)
		FROM users
		WHERE id = $1`)

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery(query).
			WithArgs(int64(1)).
			WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(100.5))

		balance, err := repo.GetBalance(context.Background(), 1)
		if err != nil {
			t.Fatalf("GetBalance() error = %v", err)
		}
		if balance != 100.5 {
			t.Fatalf("GetBalance() = %v, want 100.5", balance)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(query).
			WithArgs(int64(2)).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.GetBalance(context.Background(), 2)
		if !errors.Is(err, repository.ErrUserNotFound) {
			t.Fatalf("GetBalance() error = %v, want user not found", err)
		}
	})
}

func TestBalanceRepositoryWithdraw(t *testing.T) {
	db, mock := newMockDB(t)
	repo := repository.NewBalanceRepository(db)

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COALESCE(balance, 0)
		FROM users
		WHERE id = $1
		FOR UPDATE`)).
			WithArgs(int64(1)).
			WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(100.0))
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE users
		SET balance = COALESCE(balance, 0) - $2
		WHERE id = $1`)).
			WithArgs(int64(1), 10.0).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO withdrawals (user_id, order_number, sum, processed_at)
		VALUES ($1, $2, $3, $4)`)).
			WithArgs(int64(1), "79927398713", 10.0, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		if err := repo.Withdraw(context.Background(), 1, "79927398713", 10, time.Now()); err != nil {
			t.Fatalf("Withdraw() error = %v", err)
		}
	})

	t.Run("not enough balance", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COALESCE(balance, 0)
		FROM users
		WHERE id = $1
		FOR UPDATE`)).
			WithArgs(int64(2)).
			WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(5.0))
		mock.ExpectRollback()

		err := repo.Withdraw(context.Background(), 2, "79927398713", 10, time.Now())
		if !errors.Is(err, domain.ErrNotEnoughBalance) {
			t.Fatalf("Withdraw() error = %v, want not enough balance", err)
		}
	})
}

func TestBalanceRepositoryGetWithdrawnTotal(t *testing.T) {
	db, mock := newMockDB(t)
	repo := repository.NewBalanceRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COALESCE(SUM(sum), 0)
		FROM withdrawals
		WHERE user_id = $1`)).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(25.25))

	sum, err := repo.GetWithdrawnTotal(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetWithdrawnTotal() error = %v", err)
	}
	if sum != 25.25 {
		t.Fatalf("GetWithdrawnTotal() = %v, want 25.25", sum)
	}
}

func TestBalanceRepositoryListWithdrawals(t *testing.T) {
	db, mock := newMockDB(t)
	repo := repository.NewBalanceRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT order_number, user_id, sum, processed_at
		FROM withdrawals
		WHERE user_id = $1
		ORDER BY processed_at DESC`)).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"order_number", "user_id", "sum", "processed_at"}).
			AddRow("79927398713", int64(1), 10.0, time.Now()))

	withdrawals, err := repo.ListWithdrawals(context.Background(), 1)
	if err != nil {
		t.Fatalf("ListWithdrawals() error = %v", err)
	}
	if len(withdrawals) != 1 || withdrawals[0].Order != "79927398713" {
		t.Fatalf("ListWithdrawals() = %+v", withdrawals)
	}
}
