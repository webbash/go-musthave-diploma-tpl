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
	"github.com/shopspring/decimal"
	"go-musthave-diploma-tpl/internal/model"
	"go-musthave-diploma-tpl/internal/repository"
)

func TestOrderRepositoryCreateOrder(t *testing.T) {
	db, mock := newMockDB(t)
	repo := repository.NewOrderRepository(db)

	query := regexp.QuoteMeta(`INSERT INTO orders (number, user_id, status, accrual, uploaded_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING number, user_id, status, accrual, uploaded_at, updated_at`)

	rows := sqlmock.NewRows([]string{"number", "user_id", "status", "accrual", "uploaded_at", "updated_at"}).
		AddRow("79927398713", int64(1), model.OrderStatusNew, decimal.Zero, time.Now(), time.Now())
	mock.ExpectQuery(query).
		WithArgs("79927398713", int64(1), model.OrderStatusNew, decimal.Zero, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	order, err := repo.CreateOrder(context.Background(), model.Order{Number: "79927398713", UserID: 1, Status: model.OrderStatusNew})
	if err != nil {
		t.Fatalf("CreateOrder() error = %v", err)
	}
	if order.Number != "79927398713" {
		t.Fatalf("CreateOrder() = %+v", order)
	}
}

func TestOrderRepositoryFindOrderByNumber(t *testing.T) {
	db, mock := newMockDB(t)
	repo := repository.NewOrderRepository(db)

	query := regexp.QuoteMeta(`SELECT number, user_id, status, accrual, uploaded_at, updated_at
		FROM orders
		WHERE number = $1 AND user_id = $2`)

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"number", "user_id", "status", "accrual", "uploaded_at", "updated_at"}).
			AddRow("79927398713", int64(1), model.OrderStatusNew, decimal.Zero, time.Now(), time.Now())
		mock.ExpectQuery(query).
			WithArgs("79927398713", int64(1)).
			WillReturnRows(rows)

		order, err := repo.FindOrderByNumber(context.Background(), "79927398713", 1)
		if err != nil {
			t.Fatalf("FindOrderByNumber() error = %v", err)
		}
		if order.Number != "79927398713" {
			t.Fatalf("FindOrderByNumber() = %+v", order)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(query).
			WithArgs("ghost", int64(1)).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.FindOrderByNumber(context.Background(), "ghost", 1)
		if !errors.Is(err, repository.ErrOrderNotFound) {
			t.Fatalf("FindOrderByNumber() error = %v, want order not found", err)
		}
	})
}

func TestOrderRepositoryListOrdersByUser(t *testing.T) {
	db, mock := newMockDB(t)
	repo := repository.NewOrderRepository(db)

	query := regexp.QuoteMeta(`SELECT number, user_id, status, accrual, uploaded_at, updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY uploaded_at DESC`)
	rows := sqlmock.NewRows([]string{"number", "user_id", "status", "accrual", "uploaded_at", "updated_at"}).
		AddRow("79927398713", int64(1), model.OrderStatusProcessed, decimal.NewFromFloat(10.5), time.Now(), time.Now())
	mock.ExpectQuery(query).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	orders, err := repo.ListOrdersByUser(context.Background(), 1)
	if err != nil {
		t.Fatalf("ListOrdersByUser() error = %v", err)
	}
	if len(orders) != 1 {
		t.Fatalf("ListOrdersByUser() len = %d, want 1", len(orders))
	}
}

func TestOrderRepositoryUpdateOrder(t *testing.T) {
	db, mock := newMockDB(t)
	repo := repository.NewOrderRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT user_id FROM orders WHERE number = $1`)).
		WithArgs("79927398713").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(1)))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE orders
		SET status = $2,
			accrual = $3,
			updated_at = $4
		WHERE number = $1`)).
		WithArgs("79927398713", model.OrderStatusProcessed, decimal.NewFromFloat(15.5), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE users
		SET balance = COALESCE(balance, 0) + $2
		WHERE id = $1`)).
		WithArgs(int64(1), decimal.NewFromFloat(15.5)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := repo.UpdateOrder(context.Background(), "79927398713", model.OrderStatusProcessed, decimal.NewFromFloat(15.5)); err != nil {
		t.Fatalf("UpdateOrder() error = %v", err)
	}
}

func TestOrderRepositoryUpdateOrderNotFound(t *testing.T) {
	db, mock := newMockDB(t)
	repo := repository.NewOrderRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT user_id FROM orders WHERE number = $1`)).
		WithArgs("ghost").
		WillReturnError(sql.ErrNoRows)

	err := repo.UpdateOrder(context.Background(), "ghost", model.OrderStatusProcessed, decimal.NewFromFloat(1))
	if err == nil {
		t.Fatal("UpdateOrder() expected error")
	}
	if !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, repository.ErrOrderNotFound) {
		t.Fatalf("UpdateOrder() error = %v", err)
	}
}

func TestOrderRepositoryCreateOrderDuplicate(t *testing.T) {
	db, mock := newMockDB(t)
	repo := repository.NewOrderRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO orders (number, user_id, status, accrual, uploaded_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING number, user_id, status, accrual, uploaded_at, updated_at`)).
		WithArgs("79927398713", int64(1), model.OrderStatusNew, decimal.Zero, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(&pgconn.PgError{Code: "23505"})

	_, err := repo.CreateOrder(context.Background(), model.Order{Number: "79927398713", UserID: 1, Status: model.OrderStatusNew})
	if !errors.Is(err, repository.ErrDuplicateOrder) {
		t.Fatalf("CreateOrder() error = %v, want duplicate order", err)
	}
}
