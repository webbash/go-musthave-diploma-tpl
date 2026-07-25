package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"go-musthave-diploma-tpl/internal/domain"
	"go-musthave-diploma-tpl/internal/model"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) GetByStatuses(ctx context.Context, statuses ...model.OrderStatus) ([]model.Order, error) {
	selectQuery := `
		SELECT number, user_id, status, accrual, uploaded_at, updated_at
		FROM orders
		WHERE status IN ($1)
	`

	rows, err := r.db.QueryContext(ctx, selectQuery, statuses)
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

func (r *OrderRepository) SaveOrder(ctx context.Context, order model.Order) (model.Order, error) {
	const insertQuery = `
		INSERT INTO orders (number, user_id, status, accrual, uploaded_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (number) DO NOTHING
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
	if !errors.Is(err, sql.ErrNoRows) {
		return model.Order{}, fmt.Errorf("save order: %w", err)
	}

	existing, found, err := r.FindOrderByNumber(ctx, order.Number)
	if err != nil {
		return model.Order{}, err
	}
	if !found {
		return model.Order{}, domain.ErrNotFound
	}
	return existing, nil
}

func (r *OrderRepository) FindOrderByNumber(ctx context.Context, number string) (model.Order, bool, error) {
	const q = `
		SELECT number, user_id, status, accrual, uploaded_at, updated_at
		FROM orders
		WHERE number = $1
	`

	var order model.Order
	err := r.db.QueryRowContext(ctx, q, number).Scan(
		&order.Number,
		&order.UserID,
		&order.Status,
		&order.Accrual,
		&order.UploadedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Order{}, false, nil
		}
		return model.Order{}, false, fmt.Errorf("find order by number: %w", err)
	}

	return order, true, nil
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

func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, number string, status model.OrderStatus, accrual float64) error {
	const q = `
		UPDATE orders
		SET status = $2,
			accrual = $3,
			updated_at = $4
		WHERE number = $1
	`

	res, err := r.db.ExecContext(ctx, q, number, status, accrual, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if affected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
