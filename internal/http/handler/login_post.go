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

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type loginPostHandler struct {
	auth   *service.AuthService
	logger *zap.Logger
}

func NewLogin(auth *service.AuthService, logger *zap.Logger) http.Handler {
	return &loginPostHandler{auth: auth, logger: logger}
}

func (h *loginPostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request")
		return
	}

	res, err := h.auth.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUnauthorized):
			response.Error(w, http.StatusUnauthorized, "unauthorized")
		case errors.Is(err, domain.ErrInvalidInput):
			response.Error(w, http.StatusBadRequest, "invalid request")
		default:
			h.logger.Error("login failed", zap.Error(err))
			response.Error(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	w.Header().Set("Authorization", "Bearer "+res.Token)

	_ = response.JSON(w, http.StatusOK, LoginResponse{Token: res.Token})
}
