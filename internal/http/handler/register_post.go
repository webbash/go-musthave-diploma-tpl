package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"go-musthave-diploma-tpl/internal/domain"
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
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	res, err := h.auth.Register(r.Context(), req.Login, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAlreadyExists), errors.Is(err, domain.ErrAlreadyExists):
			http.Error(w, "conflict", http.StatusConflict)
		case errors.Is(err, service.ErrInvalidInput):
			http.Error(w, "invalid request", http.StatusBadRequest)
		default:
			if h.logger != nil {
				h.logger.Error("register failed", zap.Error(err))
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(RegisterResponse{Token: res.Token})
}
