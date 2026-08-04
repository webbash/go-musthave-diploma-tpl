package accrual

import (
	"context"
	"testing"
	"time"

	"go-musthave-diploma-tpl/internal/accrual/mocks"
	"go-musthave-diploma-tpl/internal/model"

	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestGeneratorOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockOrderRepository(ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo.EXPECT().
		GetByStatuses(gomock.Any(), model.OrderStatusNew, model.OrderStatusProcessing).
		Return([]model.Order{
			{Number: "79927398713"},
		}, nil)

	generator := NewGenerator(repo, 5*time.Millisecond, zap.NewNop())

	var got model.Order

	for order := range generator.Orders(ctx) {
		got = order
		cancel()
		break
	}

	if got.Number != "79927398713" {
		t.Fatalf("Orders() order = %+v", got)
	}
}
