package accrual

import (
	"context"
	"go-musthave-diploma-tpl/internal/model"
	"time"

	"go.uber.org/zap"
)

type Generator struct {
	repository   OrderRepository
	pollInterval time.Duration
	logger       *zap.Logger
}

func NewGenerator(repository OrderRepository, pollInterval time.Duration, logger *zap.Logger) *Generator {
	return &Generator{repository: repository, pollInterval: pollInterval, logger: logger}
}

func (g *Generator) Run(ctx context.Context) chan model.Order {
	ch := make(chan model.Order)

	go func() {
		ticker := time.NewTicker(g.pollInterval)
		defer close(ch)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				orders, err := g.repository.GetByStatuses(
					ctx,
					model.OrderStatusNew,
					model.OrderStatusProcessing,
				)
				if err != nil {
					g.logger.Error("failed to get orders", zap.Error(err))
					continue
				}
				for _, order := range orders {
					select {
					case <-ctx.Done():
						return
					case ch <- order:
					}
				}
			case <-ctx.Done():
				return
			}

		}
	}()

	return ch
}
