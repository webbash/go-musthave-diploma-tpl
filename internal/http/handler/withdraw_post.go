package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"go-musthave-diploma-tpl/internal/domain"
	httpMiddleware "go-musthave-diploma-tpl/internal/http/middleware"
	"go-musthave-diploma-tpl/internal/service"
	"go.uber.org/zap"
)

type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

type withdrawPostHandler struct {
	balance *service.BalanceService
	logger  *zap.Logger
}

func NewWithdraw(balance *service.BalanceService, logger *zap.Logger) http.Handler {
	return &withdrawPostHandler{balance: balance, logger: logger}
}

func (h *withdrawPostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID := httpMiddleware.UserIDFromContext(r.Context())

	var req WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := h.balance.Withdraw(r.Context(), userID, req.Order, req.Sum); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidOrder), errors.Is(err, domain.ErrInvalidOrder):
			http.Error(w, "invalid order", http.StatusUnprocessableEntity)
		case errors.Is(err, service.ErrInvalidInput):
			http.Error(w, "invalid request", http.StatusBadRequest)
		case errors.Is(err, domain.ErrInsufficientSum), errors.Is(err, service.ErrInsufficientBalance):
			http.Error(w, "insufficient balance", http.StatusPaymentRequired)
		default:
			if h.logger != nil {
				h.logger.Error("withdraw failed", zap.Error(err))
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}
