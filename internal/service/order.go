package service

import (
	"context"
	"fmt"
	"time"

	"go-musthave-diploma-tpl/internal/domain"
	"go-musthave-diploma-tpl/internal/model"
)

type OrderService struct {
	repository OrderRepository
}

type OrderRepository interface {
	SaveOrder(ctx context.Context, order model.Order) (model.Order, error)
	FindOrderByNumber(ctx context.Context, number string) (model.Order, bool, error)
	ListOrdersByUser(ctx context.Context, userID int64) ([]model.Order, error)
	UpdateOrderStatus(ctx context.Context, number string, status model.OrderStatus, accrual float64) error
}

func NewOrderService(repository OrderRepository) *OrderService {
	return &OrderService{repository: repository}
}

func (s *OrderService) Submit(ctx context.Context, userID int64, number string) (bool, error) {
	if !domain.IsDigits(number) {
		return false, domain.ErrInvalidInput
	}
	if !domain.IsValidLuhn(number) {
		return false, domain.ErrInvalidOrder
	}

	existing, found, err := s.repository.FindOrderByNumber(ctx, number)
	if err != nil {
		return false, err
	}
	if found {
		if existing.UserID != userID {
			return false, domain.ErrConflict
		}
		return true, nil
	}

	_, err = s.repository.SaveOrder(ctx, model.Order{
		Number:     number,
		UserID:     userID,
		Status:     model.OrderStatusNew,
		UploadedAt: time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	})
	if err != nil {
		return false, fmt.Errorf("save order: %w", err)
	}
	return false, nil
}

func (s *OrderService) List(ctx context.Context, userID int64) ([]model.Order, error) {
	return s.repository.ListOrdersByUser(ctx, userID)
}
