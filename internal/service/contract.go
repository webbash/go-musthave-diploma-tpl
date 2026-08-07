package service

import (
	"context"
	"go-musthave-diploma-tpl/internal/model"
	"time"

	"github.com/shopspring/decimal"
)

type UserRepository interface {
	CreateUser(ctx context.Context, login, passwordHash string) (model.User, error)
	FindUserByLogin(ctx context.Context, login string) (model.User, error)
}

type BalanceRepository interface {
	GetBalance(ctx context.Context, userID int64) (float64, error)
	GetWithdrawnTotal(ctx context.Context, userID int64) (float64, error)
	Withdraw(ctx context.Context, userID int64, order string, sum float64, processedAt time.Time) error
	ListWithdrawals(ctx context.Context, userID int64) ([]model.Withdrawal, error)
}

type OrderRepository interface {
	CreateOrder(ctx context.Context, order model.Order) (model.Order, error)
	FindOrderByNumber(ctx context.Context, number string, userId int64) (model.Order, error)
	ListOrdersByUser(ctx context.Context, userID int64) ([]model.Order, error)
	UpdateOrder(ctx context.Context, number string, status model.OrderStatus, accrual decimal.Decimal) error
}
