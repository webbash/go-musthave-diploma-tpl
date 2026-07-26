package accrual

import (
	"encoding/json"
	"fmt"
	"go-musthave-diploma-tpl/internal/model"
	"net/http"
	"net/url"
)

type AccrualClient struct {
	baseUrl    string
	httpClient *http.Client
}

func NewAccrualClient(baseUrl string, httpClient *http.Client) *AccrualClient {
	return &AccrualClient{
		baseUrl:    baseUrl,
		httpClient: httpClient,
	}
}

func (ac *AccrualClient) GetOrder(orderId int64) (model.Order, error) {
	getOrderUrl, err := url.JoinPath(ac.baseUrl, "/orders/", fmt.Sprint(orderId))
	if err != nil {
		return model.Order{}, fmt.Errorf("create url: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, getOrderUrl, nil)
	if err != nil {
		return model.Order{}, fmt.Errorf("create request: %w", err)
	}

	resp, err := ac.httpClient.Do(req)
	if err != nil {
		return model.Order{}, fmt.Errorf("send get: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return model.Order{}, fmt.Errorf("send update: %d", resp.StatusCode)
	}

	var order model.Order
	if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
		return model.Order{}, fmt.Errorf("failed to deserialize json: %w", err)
	}

	return order, nil
}
