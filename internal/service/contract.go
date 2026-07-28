package service

import (
	"context"
	"go-musthave-diploma-tpl/internal/model"
	"time"
)

type UserRepository interface {
	CreateUser(ctx context.Context, login, passwordHash string) (model.User, error)
	FindUserByLogin(ctx context.Context, login string) (model.User, error)
}

type BalanceRepository interface {
	GetBalance(ctx context.Context, userID int64) (float64, error)
	GetWithdrawnTotal(ctx context.Context, userID int64) (float64, error)
	AddAccrual(ctx context.Context, userID int64, sum float64) error
	Withdraw(ctx context.Context, userID int64, order string, sum float64, processedAt time.Time) error
	ListWithdrawals(ctx context.Context, userID int64) ([]model.Withdrawal, error)
}

type OrderRepository interface {
	SaveOrder(ctx context.Context, order model.Order) (model.Order, error)
	FindOrderByNumber(ctx context.Context, number string) (model.Order, bool, error)
	ListOrdersByUser(ctx context.Context, userID int64) ([]model.Order, error)
	UpdateOrderStatus(ctx context.Context, number string, status model.OrderStatus, accrual float64) error
}
