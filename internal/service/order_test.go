package service_test

import (
	"context"
	"errors"
	"testing"

	"go-musthave-diploma-tpl/internal/domain"
	"go-musthave-diploma-tpl/internal/model"
	"go-musthave-diploma-tpl/internal/repository"
	"go-musthave-diploma-tpl/internal/service"
	"go-musthave-diploma-tpl/internal/service/mocks"
	"go.uber.org/mock/gomock"
)

func TestOrderServiceCreate(t *testing.T) {
	const number = "79927398713"

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockOrderRepository(ctrl)
		svc := service.NewOrderService(repo)

		repo.EXPECT().
			CreateOrder(gomock.Any(), gomock.Any()).
			Return(model.Order{Number: number, UserID: 1}, nil)

		order, err := svc.Create(context.Background(), 1, number)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if order.Number != number {
			t.Fatalf("Create() number = %q, want %q", order.Number, number)
		}
	})

	t.Run("invalid number", func(t *testing.T) {
		svc := service.NewOrderService(nil)
		_, err := svc.Create(context.Background(), 1, "123")
		if !errors.Is(err, domain.ErrInvalidOrderNumber) {
			t.Fatalf("Create() error = %v, want invalid order number", err)
		}
	})

	t.Run("duplicate by current user", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockOrderRepository(ctrl)
		svc := service.NewOrderService(repo)

		repo.EXPECT().
			CreateOrder(gomock.Any(), gomock.Any()).
			Return(model.Order{}, repository.ErrDuplicateOrder)
		repo.EXPECT().
			FindOrderByNumber(gomock.Any(), number, int64(1)).
			Return(model.Order{Number: number, UserID: 1}, nil)

		_, err := svc.Create(context.Background(), 1, number)
		if !errors.Is(err, domain.ErrOrderAlreadyCreatedByCurrentUser) {
			t.Fatalf("Create() error = %v, want current user duplicate", err)
		}
	})

	t.Run("duplicate by another user", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockOrderRepository(ctrl)
		svc := service.NewOrderService(repo)

		repo.EXPECT().
			CreateOrder(gomock.Any(), gomock.Any()).
			Return(model.Order{}, repository.ErrDuplicateOrder)
		repo.EXPECT().
			FindOrderByNumber(gomock.Any(), number, int64(1)).
			Return(model.Order{}, repository.ErrOrderNotFound)

		_, err := svc.Create(context.Background(), 1, number)
		if !errors.Is(err, domain.ErrOrderAlreadyCreatedByAnotherUser) {
			t.Fatalf("Create() error = %v, want another user duplicate", err)
		}
	})
}

func TestOrderServiceList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockOrderRepository(ctrl)
	svc := service.NewOrderService(repo)

	repo.EXPECT().
		ListOrdersByUser(gomock.Any(), int64(1)).
		Return([]model.Order{{Number: "79927398713"}}, nil)

	orders, err := svc.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(orders) != 1 {
		t.Fatalf("List() len = %d, want 1", len(orders))
	}
}
