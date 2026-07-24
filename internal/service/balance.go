package service

import (
	"context"
	"time"

	"go-musthave-diploma-tpl/internal/domain"
	"go-musthave-diploma-tpl/internal/model"
)

type BalanceRepository interface {
	GetBalance(ctx context.Context, userID int64) (float64, error)
	GetWithdrawnTotal(ctx context.Context, userID int64) (float64, error)
	AddAccrual(ctx context.Context, userID int64, sum float64) error
	Withdraw(ctx context.Context, userID int64, order string, sum float64, processedAt time.Time) error
	ListWithdrawals(ctx context.Context, userID int64) ([]model.Withdrawal, error)
}

type Balance struct {
	Current   float64
	Withdrawn float64
}

type BalanceService struct {
	repository BalanceRepository
}

func NewBalanceService(repository BalanceRepository) *BalanceService {
	return &BalanceService{repository: repository}
}

func (s *BalanceService) Get(ctx context.Context, userID int64) (Balance, error) {
	current, err := s.repository.GetBalance(ctx, userID)
	if err != nil {
		return Balance{}, err
	}

	withdrawn, err := s.repository.GetWithdrawnTotal(ctx, userID)
	if err != nil {
		return Balance{}, err
	}

	return Balance{Current: current, Withdrawn: withdrawn}, nil
}

func (s *BalanceService) Withdraw(ctx context.Context, userID int64, order string, sum float64) error {
	if !domain.IsDigits(order) || !domain.IsValidLuhn(order) {
		return domain.ErrInvalidOrder
	}
	if sum <= 0 {
		return domain.ErrInvalidInput
	}

	return s.repository.Withdraw(ctx, userID, order, sum, time.Now().UTC())
}

func (s *BalanceService) Withdrawals(ctx context.Context, userID int64) ([]model.Withdrawal, error) {
	return s.repository.ListWithdrawals(ctx, userID)
}
