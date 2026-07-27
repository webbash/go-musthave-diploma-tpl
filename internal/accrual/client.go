package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"go-musthave-diploma-tpl/internal/model"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"go.uber.org/zap"
)

type Client struct {
	baseUrl    string
	httpClient *http.Client
	logger     *zap.Logger
}

func NewAccrualClient(baseUrl string, httpClient *http.Client, logger *zap.Logger) *Client {
	return &Client{
		baseUrl:    baseUrl,
		httpClient: httpClient,
		logger:     logger,
	}
}

func (ac *Client) GetOrder(ctx context.Context, orderId string) (model.Order, error) {
	getOrderUrl, err := url.JoinPath(ac.baseUrl, "/api/orders/", orderId)

	if err != nil {
		return model.Order{}, fmt.Errorf("create url: %w", err)
	}

	ac.logger.Info("get order by url", zap.String("url", getOrderUrl))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, getOrderUrl, nil)
	if err != nil {
		return model.Order{}, fmt.Errorf("create request: %w", err)
	}

	resp, err := ac.httpClient.Do(req)
	if err != nil {
		return model.Order{}, fmt.Errorf("send get: %w", err)
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var order model.Order
		if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
			return model.Order{}, fmt.Errorf("failed to deserialize json: %w", err)
		}

		return order, nil
	case http.StatusNoContent:
		return model.Order{}, ErrOrderNotFound
	case http.StatusTooManyRequests:
		retryAfter := resp.Header.Get("Retry-After")
		seconds, err := strconv.Atoi(retryAfter)
		if err != nil {
			return model.Order{}, fmt.Errorf("failed to parse Retry-After header: %w", err)
		}

		return model.Order{}, &RateLimitError{RetryAfter: time.Duration(seconds) * time.Second}
	default:
		return model.Order{}, fmt.Errorf("send get: %d", resp.StatusCode)
	}
}
