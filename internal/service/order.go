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

func NewOrderService(repository OrderRepository) *OrderService {
	return &OrderService{repository: repository}
}

func (s *OrderService) Create(ctx context.Context, userID int64, number string) (model.Order, error) {
	if !domain.IsValidLuhn(number) {
		return model.Order{}, domain.ErrInvalidOrder
	}

	existing, found, err := s.repository.FindOrderByNumber(ctx, number)
	if err != nil {
		return model.Order{}, err
	}
	if found {
		if existing.UserID != userID {
			return model.Order{}, domain.ErrConflict
		}
		return model.Order{}, domain.ErrAlreadyExists
	}

	order, err := s.repository.SaveOrder(ctx, model.Order{
		Number:     number,
		UserID:     userID,
		Status:     model.OrderStatusNew,
		UploadedAt: time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	})
	if err != nil {
		return model.Order{}, fmt.Errorf("save order: %w", err)
	}

	return order, nil
}

func (s *OrderService) List(ctx context.Context, userID int64) ([]model.Order, error) {
	return s.repository.ListOrdersByUser(ctx, userID)
}
