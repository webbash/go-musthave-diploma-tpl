package handler

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"go-musthave-diploma-tpl/internal/domain"
	httpMiddleware "go-musthave-diploma-tpl/internal/http/middleware"
	"go-musthave-diploma-tpl/internal/http/response"
	"go-musthave-diploma-tpl/internal/service"

	"go.uber.org/zap"
)

type postOrderHandler struct {
	service *service.OrderService
	logger  *zap.Logger
}

func NewPostOrder(orders *service.OrderService, logger *zap.Logger) http.Handler {
	return &postOrderHandler{service: orders, logger: logger}
}

func (h *postOrderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID := httpMiddleware.UserIDFromContext(r.Context())
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request")
		return
	}

	number := strings.TrimSpace(string(body))
	_, err = h.service.Create(r.Context(), userID, number)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidOrderNumber):
			response.Error(w, http.StatusUnprocessableEntity, "invalid order")
		case errors.Is(err, domain.ErrInvalidInput):
			response.Error(w, http.StatusBadRequest, "invalid request")
		case errors.Is(err, domain.ErrOrderAlreadyCreatedByAnotherUser):
			response.Error(w, http.StatusConflict, "another user have order")
		case errors.Is(err, domain.ErrOrderAlreadyCreatedByCurrentUser):
			w.WriteHeader(http.StatusOK)
			return
		default:
			h.logger.Error("post order failed", zap.Error(err))
			response.Error(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
