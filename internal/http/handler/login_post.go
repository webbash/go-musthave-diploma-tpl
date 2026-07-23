package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"go-musthave-diploma-tpl/internal/domain"
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
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	res, err := h.auth.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUnauthorized), errors.Is(err, domain.ErrUnauthorized):
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		case errors.Is(err, service.ErrInvalidInput):
			http.Error(w, "invalid request", http.StatusBadRequest)
		default:
			if h.logger != nil {
				h.logger.Error("login failed", zap.Error(err))
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(LoginResponse{Token: res.Token})
}
