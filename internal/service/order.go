package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go-musthave-diploma-tpl/internal/domain"
	"go-musthave-diploma-tpl/internal/model"
	"go-musthave-diploma-tpl/internal/repository"
)

type OrderService struct {
	repository OrderRepository
}

func NewOrderService(repository OrderRepository) *OrderService {
	return &OrderService{repository: repository}
}

func (s *OrderService) Create(ctx context.Context, userID int64, number string) (model.Order, error) {
	if !domain.IsValidLuhn(number) {
		return model.Order{}, domain.ErrInvalidOrderNumber
	}

	order, err := s.repository.CreateOrder(ctx, model.Order{
		Number:     number,
		UserID:     userID,
		Status:     model.OrderStatusNew,
		UploadedAt: time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	})

	if errors.Is(err, repository.ErrDuplicateOrder) {
		existingOrder, err := s.repository.FindOrderByNumber(ctx, number, userID)
		if errors.Is(err, repository.ErrOrderNotFound) {
			return model.Order{}, domain.ErrOrderAlreadyCreatedByAnotherUser
		}
		if err != nil {
			return model.Order{}, fmt.Errorf("orderService.Create: failed to get order: %w", err)
		}
		return existingOrder, domain.ErrOrderAlreadyCreatedByCurrentUser
	}
	if err != nil {
		return model.Order{}, fmt.Errorf("orderService.Create: failed to create order: %w", err)
	}

	return order, nil
}

func (s *OrderService) List(ctx context.Context, userID int64) ([]model.Order, error) {
	return s.repository.ListOrdersByUser(ctx, userID)
}
