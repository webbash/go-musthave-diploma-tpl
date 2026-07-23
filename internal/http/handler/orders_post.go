package handler

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"go-musthave-diploma-tpl/internal/domain"
	httpMiddleware "go-musthave-diploma-tpl/internal/http/middleware"
	"go-musthave-diploma-tpl/internal/service"
	"go.uber.org/zap"
)

type postOrderHandler struct {
	orders *service.OrderService
	logger *zap.Logger
}

func NewPostOrder(orders *service.OrderService, logger *zap.Logger) http.Handler {
	return &postOrderHandler{orders: orders, logger: logger}
}

func (h *postOrderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID := httpMiddleware.UserIDFromContext(r.Context())
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	number := strings.TrimSpace(string(body))
	already, err := h.orders.Submit(r.Context(), userID, number)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrConflict), errors.Is(err, domain.ErrConflict):
			http.Error(w, "conflict", http.StatusConflict)
		case errors.Is(err, service.ErrInvalidOrder), errors.Is(err, domain.ErrInvalidOrder):
			http.Error(w, "invalid order", http.StatusUnprocessableEntity)
		case errors.Is(err, service.ErrInvalidInput):
			http.Error(w, "invalid request", http.StatusBadRequest)
		default:
			if h.logger != nil {
				h.logger.Error("post order failed", zap.Error(err))
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	if already {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
