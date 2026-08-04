package accrual

import (
	"context"
	"errors"
	"testing"
	"time"

	"go-musthave-diploma-tpl/internal/accrual/mocks"
	"go-musthave-diploma-tpl/internal/model"

	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestGeneratorRun(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockOrderRepository(ctrl)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	repo.EXPECT().
		GetByStatuses(gomock.Any(), model.OrderStatusNew, model.OrderStatusProcessing).
		DoAndReturn(func(context.Context, ...model.OrderStatus) ([]model.Order, error) {
			return []model.Order{{Number: "79927398713"}}, nil
		})

	generator := NewGenerator(repo, 5*time.Millisecond, zap.NewNop())
	ch := generator.Run(ctx)

	select {
	case order := <-ch:
		if order.Number != "79927398713" {
			t.Fatalf("Run() order = %+v", order)
		}
		cancel()
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Run() timed out waiting for order")
	}
}

func TestGeneratorRunRepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockOrderRepository(ctrl)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	repo.EXPECT().
		GetByStatuses(gomock.Any(), model.OrderStatusNew, model.OrderStatusProcessing).
		DoAndReturn(func(context.Context, ...model.OrderStatus) ([]model.Order, error) {
			cancel()
			return nil, errors.New("boom")
		})

	generator := NewGenerator(repo, 5*time.Millisecond, zap.NewNop())
	ch := generator.Run(ctx)

	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("Run() unexpectedly yielded order")
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Run() timed out waiting for channel close")
	}
}
