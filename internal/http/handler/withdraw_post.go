package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"go-musthave-diploma-tpl/internal/domain"
	httpMiddleware "go-musthave-diploma-tpl/internal/http/middleware"
	"go-musthave-diploma-tpl/internal/http/response"
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
		response.Error(w, http.StatusBadRequest, "invalid request")
		return
	}

	if err := h.balance.Withdraw(r.Context(), userID, req.Order, req.Sum); err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidOrder):
			response.Error(w, http.StatusUnprocessableEntity, "invalid order")
		case errors.Is(err, domain.ErrInvalidInput):
			response.Error(w, http.StatusBadRequest, "invalid request")
		case errors.Is(err, domain.ErrInsufficientSum):
			response.Error(w, http.StatusPaymentRequired, "insufficient balance")
		default:
			h.logger.Error("withdraw failed", zap.Error(err))
			response.Error(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}
