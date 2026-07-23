package handler

import (
	"encoding/json"
	"net/http"
	"time"

	httpMiddleware "go-musthave-diploma-tpl/internal/http/middleware"
	"go-musthave-diploma-tpl/internal/service"
	"go.uber.org/zap"
)

type WithdrawResponse struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

type withdrawalsGetHandler struct {
	balance *service.BalanceService
	logger  *zap.Logger
}

func NewWithdrawals(balance *service.BalanceService, logger *zap.Logger) http.Handler {
	return &withdrawalsGetHandler{balance: balance, logger: logger}
}

func (h *withdrawalsGetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID := httpMiddleware.UserIDFromContext(r.Context())
	withdrawals, err := h.balance.Withdrawals(r.Context(), userID)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("get withdrawals failed", zap.Error(err))
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	resp := make([]WithdrawResponse, 0, len(withdrawals))
	for _, item := range withdrawals {
		resp = append(resp, WithdrawResponse{
			Order:       item.Order,
			Sum:         item.Sum,
			ProcessedAt: item.ProcessedAt,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
