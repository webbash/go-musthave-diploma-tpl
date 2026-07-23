package handler

import (
	"encoding/json"
	"net/http"
	"time"

	httpMiddleware "go-musthave-diploma-tpl/internal/http/middleware"
	"go-musthave-diploma-tpl/internal/service"
	"go.uber.org/zap"
)

type GetOrdersResponse struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type getOrdersHandler struct {
	orders *service.OrderService
	logger *zap.Logger
}

func NewGetOrders(orders *service.OrderService, logger *zap.Logger) http.Handler {
	return &getOrdersHandler{orders: orders, logger: logger}
}

func (h *getOrdersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID := httpMiddleware.UserIDFromContext(r.Context())
	orders, err := h.orders.List(r.Context(), userID)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("get orders failed", zap.Error(err))
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	resp := make([]GetOrdersResponse, 0, len(orders))
	for _, order := range orders {
		resp = append(resp, GetOrdersResponse{
			Number:     order.Number,
			Status:     string(order.Status),
			Accrual:    order.Accrual,
			UploadedAt: order.UploadedAt,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
