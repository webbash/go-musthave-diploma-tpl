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

type BalanceRepository struct {
	db *sql.DB
}

func NewBalanceRepository(db *sql.DB) *BalanceRepository {
	return &BalanceRepository{db: db}
}

func (r *BalanceRepository) GetBalance(ctx context.Context, userID int64) (float64, error) {
	const q = `
		SELECT COALESCE(balance, 0)
		FROM users
		WHERE id = $1
	`

	var balance float64
	if err := r.db.QueryRowContext(ctx, q, userID).Scan(&balance); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrUserNotFound
		}
		return 0, fmt.Errorf("get balance: %w", err)
	}

	return balance, nil
}

func (r *BalanceRepository) Withdraw(ctx context.Context, userID int64, order string, sum float64, processedAt time.Time) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin withdraw tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	const lockQuery = `
		SELECT COALESCE(balance, 0)
		FROM users
		WHERE id = $1
		FOR UPDATE
	`
	var current float64
	if err := tx.QueryRowContext(ctx, lockQuery, userID).Scan(&current); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return fmt.Errorf("lock balance: %w", err)
	}
	if current < sum {
		return domain.ErrNotEnoughBalance
	}

	const updateQuery = `
		UPDATE users
		SET balance = COALESCE(balance, 0) - $2
		WHERE id = $1
	`
	if _, err := tx.ExecContext(ctx, updateQuery, userID, sum); err != nil {
		return fmt.Errorf("update balance: %w", err)
	}

	const insertWithdrawalQuery = `
		INSERT INTO withdrawals (user_id, order_number, sum, processed_at)
		VALUES ($1, $2, $3, $4)
	`
	if _, err := tx.ExecContext(ctx, insertWithdrawalQuery, userID, order, sum, processedAt); err != nil {
		return fmt.Errorf("insert withdrawal: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit withdraw tx: %w", err)
	}
	return nil
}

func (r *BalanceRepository) GetWithdrawnTotal(ctx context.Context, userID int64) (float64, error) {
	const q = `
		SELECT COALESCE(SUM(sum), 0)
		FROM withdrawals
		WHERE user_id = $1
	`

	var withdrawn float64
	if err := r.db.QueryRowContext(ctx, q, userID).Scan(&withdrawn); err != nil {
		return 0, fmt.Errorf("get withdrawn total: %w", err)
	}
	return withdrawn, nil
}

func (r *BalanceRepository) ListWithdrawals(ctx context.Context, userID int64) ([]model.Withdrawal, error) {
	const q = `
		SELECT order_number, user_id, sum, processed_at
		FROM withdrawals
		WHERE user_id = $1
		ORDER BY processed_at DESC
	`

	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list withdrawals: %w", err)
	}
	defer rows.Close()

	withdrawals := make([]model.Withdrawal, 0)
	for rows.Next() {
		var item model.Withdrawal
		if err := rows.Scan(&item.Order, &item.UserID, &item.Sum, &item.ProcessedAt); err != nil {
			return nil, fmt.Errorf("scan withdrawal: %w", err)
		}
		withdrawals = append(withdrawals, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate withdrawals: %w", err)
	}

	return withdrawals, nil
}
