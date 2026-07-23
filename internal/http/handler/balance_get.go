package handler

import (
	"encoding/json"
	"net/http"

	httpMiddleware "go-musthave-diploma-tpl/internal/http/middleware"
	"go-musthave-diploma-tpl/internal/service"
	"go.uber.org/zap"
)

type BalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type balanceGetHandler struct {
	balance *service.BalanceService
	logger  *zap.Logger
}

func NewGetBalance(balance *service.BalanceService, logger *zap.Logger) http.Handler {
	return &balanceGetHandler{balance: balance, logger: logger}
}

func (h *balanceGetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID := httpMiddleware.UserIDFromContext(r.Context())
	balance, err := h.balance.Get(r.Context(), userID)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("get balance failed", zap.Error(err))
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(BalanceResponse{
		Current:   balance.Current,
		Withdrawn: balance.Withdrawn,
	})
}
