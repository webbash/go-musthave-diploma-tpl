package handler_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-musthave-diploma-tpl/internal/auth"
	httpMiddleware "go-musthave-diploma-tpl/internal/http/middleware"
)

func authRequest(t *testing.T, secret string, method, target string, body []byte) (*http.Request, *httptest.ResponseRecorder) {
	t.Helper()

	token, err := auth.GenerateJWT("42", secret, time.Minute)
	if err != nil {
		t.Fatalf("GenerateJWT() error = %v", err)
	}

	req := httptest.NewRequest(method, target, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)

	return req, httptest.NewRecorder()
}

func serveAuthed(t *testing.T, secret string, h http.Handler, method, target string, body []byte) *httptest.ResponseRecorder {
	t.Helper()

	req, rec := authRequest(t, secret, method, target, body)
	httpMiddleware.Auth(secret)(h).ServeHTTP(rec, req)
	return rec
}
