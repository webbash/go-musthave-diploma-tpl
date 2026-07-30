package accrual

import (
	"context"
	"go-musthave-diploma-tpl/internal/model"

	"github.com/shopspring/decimal"
)

type OrderRepository interface {
	GetByStatuses(ctx context.Context, statuses ...model.OrderStatus) ([]model.Order, error)
	UpdateOrder(ctx context.Context, number string, status model.OrderStatus, accrual decimal.Decimal) error
}
