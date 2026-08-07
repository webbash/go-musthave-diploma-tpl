package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-musthave-diploma-tpl/internal/auth"
	httpMiddleware "go-musthave-diploma-tpl/internal/http/middleware"
)

func TestAuthMiddlewareAllowsBearerToken(t *testing.T) {
	const secret = "secret"
	token, err := auth.GenerateJWT("42", secret, time.Minute)
	if err != nil {
		t.Fatalf("GenerateJWT() error = %v", err)
	}

	called := false
	handler := httpMiddleware.Auth(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if got := httpMiddleware.UserIDFromContext(r.Context()); got != 42 {
			t.Fatalf("UserIDFromContext() = %d, want 42", got)
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected next handler to be called")
	}
}

func TestAuthMiddlewareRejectsMissingToken(t *testing.T) {
	handler := httpMiddleware.Auth("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Code; got != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", got, http.StatusUnauthorized)
	}
}
