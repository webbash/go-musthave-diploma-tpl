package accrual

import (
	"context"
	"go-musthave-diploma-tpl/internal/model"
)

type OrderRepository interface {
	GetByStatuses(ctx context.Context, statuses ...model.OrderStatus) ([]model.Order, error)
	UpdateOrderStatus(ctx context.Context, number string, status model.OrderStatus, accrual float64) error
}
