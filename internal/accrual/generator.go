package accrual

import (
	"context"
	"go-musthave-diploma-tpl/internal/model"
	"iter"
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

func (g *Generator) Orders(ctx context.Context) iter.Seq[model.Order] {
	return func(yield func(model.Order) bool) {
		ticker := time.NewTicker(g.pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return

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
					if !yield(order) {
						return
					}
				}
			}
		}
	}
}
