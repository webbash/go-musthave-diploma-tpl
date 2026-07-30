package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"go-musthave-diploma-tpl/internal/model"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) GetByStatuses(ctx context.Context, statuses ...model.OrderStatus) ([]model.Order, error) {
	if len(statuses) == 0 {
		return []model.Order{}, nil
	}

	placeholders := make([]string, 0, len(statuses))
	args := make([]any, 0, len(statuses))
	for i, status := range statuses {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
		args = append(args, string(status))
	}

	selectQuery := fmt.Sprintf(`
		SELECT number, user_id, status, accrual, uploaded_at, updated_at
		FROM orders
		WHERE status IN (%s)
	`, strings.Join(placeholders, ", "))

	rows, err := r.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("list orders by statuses: %w", err)
	}
	defer rows.Close()

	orders := make([]model.Order, 0)
	for rows.Next() {
		var order model.Order
		if err := rows.Scan(
			&order.Number,
			&order.UserID,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
			&order.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate orders: %w", err)
	}

	return orders, nil

}

func (r *OrderRepository) CreateOrder(ctx context.Context, order model.Order) (model.Order, error) {
	const insertQuery = `
		INSERT INTO orders (number, user_id, status, accrual, uploaded_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING number, user_id, status, accrual, uploaded_at, updated_at
	`

	var created model.Order
	err := r.db.QueryRowContext(
		ctx,
		insertQuery,
		order.Number,
		order.UserID,
		order.Status,
		order.Accrual,
		order.UploadedAt,
		order.UpdatedAt,
	).Scan(
		&created.Number,
		&created.UserID,
		&created.Status,
		&created.Accrual,
		&created.UploadedAt,
		&created.UpdatedAt,
	)
	if err == nil {
		return created, nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return model.Order{}, ErrDuplicateOrder
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return model.Order{}, fmt.Errorf("save order: %w", err)
	}
	return model.Order{}, ErrDuplicateOrder
}

func (r *OrderRepository) FindOrderByNumber(ctx context.Context, number string, userId int64) (model.Order, error) {
	const q = `
		SELECT number, user_id, status, accrual, uploaded_at, updated_at
		FROM orders
		WHERE number = $1 AND user_id = $2
	`

	var order model.Order
	err := r.db.QueryRowContext(ctx, q, number, userId).Scan(
		&order.Number,
		&order.UserID,
		&order.Status,
		&order.Accrual,
		&order.UploadedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Order{}, ErrOrderNotFound
		}

		return model.Order{}, fmt.Errorf("orderRepository.FindOrderByNumber: failed to get order by number and user id: %w", err)
	}

	return order, nil
}

func (r *OrderRepository) ListOrdersByUser(ctx context.Context, userID int64) ([]model.Order, error) {
	const q = `
		SELECT number, user_id, status, accrual, uploaded_at, updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY uploaded_at DESC
	`

	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list orders by user: %w", err)
	}
	defer rows.Close()

	orders := make([]model.Order, 0)
	for rows.Next() {
		var order model.Order
		if err := rows.Scan(
			&order.Number,
			&order.UserID,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
			&order.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate orders: %w", err)
	}

	return orders, nil
}

func (r *OrderRepository) UpdateOrder(ctx context.Context, number string, status model.OrderStatus, accrual decimal.Decimal) error {
	var userID int64
	userIdRes := r.db.QueryRowContext(ctx, `SELECT user_id FROM orders WHERE number = $1`, number)
	if err := userIdRes.Scan(&userID); err != nil {
		return fmt.Errorf("orderRepository.UpdateOrder: get user id: %w", err)
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("orderRepository.UpdateOrder: begin update order tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	const q = `
		UPDATE orders
		SET status = $2,
			accrual = $3,
			updated_at = $4
		WHERE number = $1
	`

	res, err := tx.ExecContext(ctx, q, number, status, accrual, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("orderRepository.UpdateOrder: update order status: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("orderRepository.UpdateOrder: rows affected: %w", err)
	}
	if affected == 0 {
		return ErrOrderNotFound
	}

	const qUpdateBalance = `
		UPDATE users
		SET balance = COALESCE(balance, 0) + $2
		WHERE id = $1
	`
	res, err = tx.ExecContext(ctx, qUpdateBalance, userID, accrual)
	if err != nil {
		return fmt.Errorf("orderRepository.UpdateOrder: update user balance: %w", err)
	}

	affected, err = res.RowsAffected()
	if err != nil {
		return fmt.Errorf("orderRepository.UpdateOrder: rows affected: %w", err)
	}
	if affected == 0 {
		return ErrOrderNotFound
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("orderRepository.UpdateOrder: commit tx: %w", err)
	}

	return nil
}
