package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-musthave-diploma-tpl/internal/auth"
	"go-musthave-diploma-tpl/internal/http/handler"
	"go-musthave-diploma-tpl/internal/model"
	"go-musthave-diploma-tpl/internal/repository"
	"go-musthave-diploma-tpl/internal/service"
	"go-musthave-diploma-tpl/internal/service/mocks"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestRegisterHandler(t *testing.T) {
	const secret = "secret"

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockUserRepository(ctrl)
		svc := service.NewAuthService(repo, secret, 10)
		h := handler.NewRegister(svc, zap.NewNop())

		repo.EXPECT().
			CreateUser(gomock.Any(), "alice", gomock.Any()).
			DoAndReturn(func(ctx context.Context, login, passwordHash string) (model.User, error) {
				if err := auth.CheckPassword(passwordHash, "password"); err != nil {
					t.Fatalf("hashed password should be valid: %v", err)
				}
				return model.User{ID: 42}, nil
			})

		reqBody, _ := json.Marshal(handler.RegisterRequest{Login: "alice", Password: "password"})
		req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(reqBody))
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if got := rec.Header().Get("Authorization"); got == "" {
			t.Fatal("Authorization header is empty")
		}

		var resp handler.RegisterResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if resp.Token == "" {
			t.Fatal("token is empty")
		}
	})

	t.Run("invalid input", func(t *testing.T) {
		svc := service.NewAuthService(nil, secret, 10)
		h := handler.NewRegister(svc, zap.NewNop())

		reqBody, _ := json.Marshal(handler.RegisterRequest{Login: "", Password: "password"})
		req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(reqBody))
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("duplicate user", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockUserRepository(ctrl)
		svc := service.NewAuthService(repo, secret, 10)
		h := handler.NewRegister(svc, zap.NewNop())

		repo.EXPECT().
			CreateUser(gomock.Any(), "alice", gomock.Any()).
			Return(model.User{}, repository.ErrDuplicateUser)

		reqBody, _ := json.Marshal(handler.RegisterRequest{Login: "alice", Password: "password"})
		req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(reqBody))
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusConflict {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusConflict)
		}
	})
}

func TestLoginHandler(t *testing.T) {
	const secret = "secret"

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockUserRepository(ctrl)
		svc := service.NewAuthService(repo, secret, 10)
		h := handler.NewLogin(svc, zap.NewNop())

		hash, err := auth.HashPassword("password")
		if err != nil {
			t.Fatalf("HashPassword() error = %v", err)
		}

		repo.EXPECT().
			FindUserByLogin(gomock.Any(), "alice").
			Return(model.User{ID: 42, PasswordHash: hash}, nil)

		reqBody, _ := json.Marshal(handler.LoginRequest{Login: "alice", Password: "password"})
		req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(reqBody))
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if got := rec.Header().Get("Authorization"); got == "" {
			t.Fatal("Authorization header is empty")
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockUserRepository(ctrl)
		svc := service.NewAuthService(repo, secret, 10)
		h := handler.NewLogin(svc, zap.NewNop())

		repo.EXPECT().
			FindUserByLogin(gomock.Any(), "alice").
			Return(model.User{}, repository.ErrUserNotFound)

		reqBody, _ := json.Marshal(handler.LoginRequest{Login: "alice", Password: "password"})
		req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(reqBody))
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("invalid input", func(t *testing.T) {
		svc := service.NewAuthService(nil, secret, 10)
		h := handler.NewLogin(svc, zap.NewNop())

		reqBody := []byte("{")
		req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(reqBody))
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}
