package handler

import (
	"net/http"

	httpMiddleware "go-musthave-diploma-tpl/internal/http/middleware"
	"go-musthave-diploma-tpl/internal/http/response"
	"go-musthave-diploma-tpl/internal/service"

	"go.uber.org/zap"
)

type BalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type balanceGetHandler struct {
	balanceService *service.BalanceService
	logger         *zap.Logger
}

func NewGetBalance(balance *service.BalanceService, logger *zap.Logger) http.Handler {
	return &balanceGetHandler{balanceService: balance, logger: logger}
}

func (h *balanceGetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID := httpMiddleware.UserIDFromContext(r.Context())
	balance, err := h.balanceService.Get(r.Context(), userID)
	if err != nil {
		h.logger.Error("get balance failed", zap.Error(err))
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	_ = response.JSON(w, http.StatusOK, BalanceResponse{
		Current:   balance.Current,
		Withdrawn: balance.Withdrawn,
	})
}
