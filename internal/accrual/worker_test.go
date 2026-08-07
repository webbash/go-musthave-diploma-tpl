package accrual

import (
	"context"
	"errors"
	"testing"
	"time"

	"go-musthave-diploma-tpl/internal/accrual/mocks"
	"go-musthave-diploma-tpl/internal/model"

	"github.com/shopspring/decimal"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestWorkerRun(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockOrderRepository(ctrl)

	repo.EXPECT().
		GetByStatuses(gomock.Any(), model.OrderStatusNew).
		Return([]model.Order{
			{
				Number:  "79927398713",
				Status:  model.OrderStatusNew,
				Accrual: decimal.Zero,
			},
		}, nil).
		Times(1)

	processed := make(chan struct{})

	repo.EXPECT().
		UpdateOrder(
			gomock.Any(),
			"79927398713",
			model.OrderStatusProcessed,
			decimal.NewFromFloat(15.5),
		).
		DoAndReturn(func(
			context.Context,
			string,
			model.OrderStatus,
			decimal.Decimal,
		) error {
			close(processed)
			return nil
		}).
		Times(1)

	client := mocks.NewMockAccrualClient(ctrl)

	client.EXPECT().
		GetOrder(gomock.Any(), "79927398713").
		Return(model.Order{
			Number:  "79927398713",
			Status:  model.OrderStatusProcessed,
			Accrual: decimal.NewFromFloat(15.5),
		}, nil).
		Times(1)

	generator := NewGenerator(
		repo,
		5*time.Millisecond,
		zap.NewNop(),
	)

	worker := NewWorker(
		repo,
		client,
		zap.NewNop(),
		1,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)

	go func() {
		done <- worker.Run(ctx, generator.Orders(ctx))
	}()

	// Ждём именно фактической обработки заказа.
	select {
	case <-processed:
		cancel()

	case <-time.After(200 * time.Millisecond):
		t.Fatal("order was not processed")
	}

	// После cancel Run должен завершиться.
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

	case <-time.After(200 * time.Millisecond):
		t.Fatal("Run() did not stop after context cancellation")
	}
}

func TestWorkerProcessOrderRateLimitRetry(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockOrderRepository(ctrl)
	repo.EXPECT().
		UpdateOrder(gomock.Any(), "79927398713", model.OrderStatusProcessed, decimal.NewFromFloat(15.5)).
		Return(nil)

	client := mocks.NewMockAccrualClient(ctrl)
	client.EXPECT().
		GetOrder(gomock.Any(), "79927398713").
		Return(model.Order{}, &RateLimitError{RetryAfter: 0})
	client.EXPECT().
		GetOrder(gomock.Any(), "79927398713").
		Return(model.Order{
			Number:  "79927398713",
			Status:  model.OrderStatusProcessed,
			Accrual: decimal.NewFromFloat(15.5),
		}, nil)
	worker := NewWorker(repo, client, zap.NewNop(), 3)

	err := worker.processOrder(context.Background(), model.Order{Number: "79927398713"}, NewRateLimiter())
	if err != nil {
		t.Fatalf("processOrder() error = %v", err)
	}
}

func TestWorkerProcessOrderUpdateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockOrderRepository(ctrl)
	repo.EXPECT().
		UpdateOrder(gomock.Any(), "79927398713", model.OrderStatusProcessed, decimal.NewFromFloat(15.5)).
		Return(errors.New("update failed"))

	client := mocks.NewMockAccrualClient(ctrl)
	client.EXPECT().
		GetOrder(gomock.Any(), "79927398713").
		Return(model.Order{
			Number:  "79927398713",
			Status:  model.OrderStatusProcessed,
			Accrual: decimal.NewFromFloat(15.5),
		}, nil)
	worker := NewWorker(repo, client, zap.NewNop(), 3)

	err := worker.processOrder(context.Background(), model.Order{Number: "79927398713"}, NewRateLimiter())
	if err == nil {
		t.Fatal("processOrder() expected error")
	}
}
