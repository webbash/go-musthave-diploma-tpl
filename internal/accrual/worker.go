package accrual

import (
	"context"
	"errors"
	"fmt"
	"go-musthave-diploma-tpl/internal/model"
	"sync"

	"go.uber.org/zap"
)

type Worker struct {
	repository OrderRepository
	workers    int
	inputCh    chan model.Order
	client     *Client
	logger     *zap.Logger
	wg         *sync.WaitGroup
}

func NewWorker(repository OrderRepository, client *Client, logger *zap.Logger, inputCh chan model.Order) *Worker {
	return &Worker{
		repository: repository,
		client:     client,
		logger:     logger,
		wg:         &sync.WaitGroup{},
		workers:    3, // TODO
		inputCh:    inputCh,
	}
}

func (w *Worker) Run(ctx context.Context) {
	rateLimiter := NewRateLimiter()
	w.wg.Add(w.workers)

	for i := 0; i < w.workers; i++ {
		go func() {

			defer w.wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case order, ok := <-w.inputCh:
					if !ok {
						return
					}
					err := w.processOrder(ctx, order, rateLimiter)
					if err != nil {
						w.logger.Error("Error processing order", zap.String("order_number", order.Number), zap.Error(err))
					}
				}
			}
		}()
	}
}

func (w *Worker) Wait() {
	w.wg.Wait()
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

		err = w.repository.UpdateOrderStatus(ctx, updatedOrder.Number, updatedOrder.Status, updatedOrder.Accrual)
		if err != nil {
			return fmt.Errorf("update order %s: %w", updatedOrder.Number, err)
		}

		// TODO Обновлять баланс пользователя в таблице users

		return nil
	}
}
