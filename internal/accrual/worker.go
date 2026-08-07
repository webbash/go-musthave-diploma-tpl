package accrual

import (
	"context"
	"errors"
	"fmt"
	"go-musthave-diploma-tpl/internal/model"
	"iter"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Worker struct {
	repository OrderRepository
	workers    int
	client     AccrualClient
	logger     *zap.Logger
}

func NewWorker(repository OrderRepository, client AccrualClient, logger *zap.Logger, workers int) *Worker {
	return &Worker{
		repository: repository,
		client:     client,
		logger:     logger,
		workers:    workers,
	}
}

func (w *Worker) Run(ctx context.Context, orders iter.Seq[model.Order]) error {
	limiter := NewRateLimiter()

	group, ctx := errgroup.WithContext(ctx)
	group.SetLimit(w.workers)

	for order := range orders {
		order := order

		group.Go(func() error {
			return w.processOrder(ctx, order, limiter)
		})
	}

	return group.Wait()
}

func (w *Worker) processOrder(
	ctx context.Context,
	order model.Order,
	limiter *RateLimiter,
) error {
	for {
		limiter.Wait(ctx)

		updatedOrder, err := w.client.GetOrder(ctx, order.Number)

		if rateLimitErr, ok := errors.AsType[*RateLimitError](err); ok {
			limiter.Block(rateLimitErr.RetryAfter)
			continue
		}

		if err != nil {
			return fmt.Errorf("get order %s: %w", order.Number, err)
		}

		err = w.repository.UpdateOrder(ctx, updatedOrder.Number, updatedOrder.Status, updatedOrder.Accrual)
		if err != nil {
			return fmt.Errorf("update order %s: %w", updatedOrder.Number, err)
		}

		return nil
	}
}
