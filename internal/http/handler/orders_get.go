package handler

import (
	"net/http"
	"time"

	"go-musthave-diploma-tpl/internal/http/middleware"
	"go-musthave-diploma-tpl/internal/http/response"
	"go-musthave-diploma-tpl/internal/service"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type OrderResponse struct {
	Number     string          `json:"number"`
	Status     string          `json:"status"`
	Accrual    decimal.Decimal `json:"accrual,omitempty"`
	UploadedAt time.Time       `json:"uploaded_at"`
}

type getOrdersHandler struct {
	orders *service.OrderService
	logger *zap.Logger
}

func NewGetOrders(orders *service.OrderService, logger *zap.Logger) http.Handler {
	return &getOrdersHandler{orders: orders, logger: logger}
}

func (h *getOrdersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	orders, err := h.orders.List(r.Context(), userID)
	if err != nil {
		h.logger.Error("get orders failed", zap.Error(err))
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	resp := make([]OrderResponse, 0, len(orders))
	for _, order := range orders {
		resp = append(resp, OrderResponse{
			Number:     order.Number,
			Status:     string(order.Status),
			Accrual:    order.Accrual,
			UploadedAt: order.UploadedAt,
		})
	}
	_ = response.JSON(w, http.StatusOK, resp)
}
