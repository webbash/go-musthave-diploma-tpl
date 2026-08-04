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
		UpdateOrder(gomock.Any(), "79927398713", model.OrderStatusProcessed, decimal.NewFromFloat(15.5)).
		Return(nil)

	client := mocks.NewMockAccrualClient(ctrl)
	client.EXPECT().
		GetOrder(gomock.Any(), "79927398713").
		Return(model.Order{
			Number:  "79927398713",
			Status:  model.OrderStatusProcessed,
			Accrual: decimal.NewFromFloat(15.5),
		}, nil)
	inputCh := make(chan model.Order, 1)
	inputCh <- model.Order{Number: "79927398713"}
	close(inputCh)

	worker := NewWorker(repo, client, zap.NewNop(), inputCh, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	worker.Run(ctx)
	done := make(chan struct{})
	go func() {
		worker.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Wait() timed out")
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
	worker := NewWorker(repo, client, zap.NewNop(), make(chan model.Order), 3)

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
	worker := NewWorker(repo, client, zap.NewNop(), make(chan model.Order), 3)

	err := worker.processOrder(context.Background(), model.Order{Number: "79927398713"}, NewRateLimiter())
	if err == nil {
		t.Fatal("processOrder() expected error")
	}
}
