package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"go-musthave-diploma-tpl/internal/domain"
	"go-musthave-diploma-tpl/internal/http/response"
	"go-musthave-diploma-tpl/internal/service"

	"go.uber.org/zap"
)

type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	Token string `json:"token"`
}

type registerPostHandler struct {
	auth   *service.AuthService
	logger *zap.Logger
}

func NewRegister(auth *service.AuthService, logger *zap.Logger) http.Handler {
	return &registerPostHandler{auth: auth, logger: logger}
}

func (h *registerPostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request")
		return
	}

	res, err := h.auth.Register(r.Context(), req.Login, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrAlreadyExists):
			response.Error(w, http.StatusConflict, "conflict")
		case errors.Is(err, domain.ErrInvalidInput):
			response.Error(w, http.StatusBadRequest, "invalid request")
		default:
			h.logger.Error("register failed", zap.Error(err))
			response.Error(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	_ = response.JSON(w, http.StatusOK, RegisterResponse{Token: res.Token})
}
