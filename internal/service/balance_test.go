package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"go-musthave-diploma-tpl/internal/domain"
	"go-musthave-diploma-tpl/internal/model"
	"go-musthave-diploma-tpl/internal/service"
	"go-musthave-diploma-tpl/internal/service/mocks"
	"go.uber.org/mock/gomock"
)

func TestBalanceServiceGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockBalanceRepository(ctrl)
	svc := service.NewBalanceService(repo)

	repo.EXPECT().GetBalance(gomock.Any(), int64(1)).Return(100.5, nil)
	repo.EXPECT().GetWithdrawnTotal(gomock.Any(), int64(1)).Return(25.25, nil)

	balance, err := svc.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if balance.Current != 100.5 || balance.Withdrawn != 25.25 {
		t.Fatalf("Get() = %+v, want current=100.5 withdrawn=25.25", balance)
	}
}

func TestBalanceServiceWithdraw(t *testing.T) {
	t.Run("invalid order", func(t *testing.T) {
		svc := service.NewBalanceService(nil)
		err := svc.Withdraw(context.Background(), 1, "123", 10)
		if !errors.Is(err, domain.ErrInvalidOrderNumber) {
			t.Fatalf("Withdraw() error = %v, want invalid order", err)
		}
	})

	t.Run("invalid amount", func(t *testing.T) {
		svc := service.NewBalanceService(nil)
		err := svc.Withdraw(context.Background(), 1, "79927398713", 0)
		if !errors.Is(err, domain.ErrInvalidInput) {
			t.Fatalf("Withdraw() error = %v, want invalid input", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockBalanceRepository(ctrl)
		svc := service.NewBalanceService(repo)

		repo.EXPECT().
			Withdraw(gomock.Any(), int64(1), "79927398713", 10.0, gomock.AssignableToTypeOf(time.Time{})).
			Return(nil)

		if err := svc.Withdraw(context.Background(), 1, "79927398713", 10); err != nil {
			t.Fatalf("Withdraw() error = %v", err)
		}
	})
}

func TestBalanceServiceWithdrawals(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockBalanceRepository(ctrl)
	svc := service.NewBalanceService(repo)

	repo.EXPECT().
		ListWithdrawals(gomock.Any(), int64(1)).
		Return([]model.Withdrawal{{Order: "1", Sum: 10}}, nil)

	withdrawals, err := svc.Withdrawals(context.Background(), 1)
	if err != nil {
		t.Fatalf("Withdrawals() error = %v", err)
	}
	if len(withdrawals) != 1 {
		t.Fatalf("Withdrawals() len = %d, want 1", len(withdrawals))
	}
}
