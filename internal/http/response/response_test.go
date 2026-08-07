package response_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-musthave-diploma-tpl/internal/http/response"
)

func TestJSON(t *testing.T) {
	rec := httptest.NewRecorder()

	if err := response.JSON(rec, http.StatusOK, map[string]string{"token": "abc"}); err != nil {
		t.Fatalf("JSON() error = %v", err)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}

	if body["token"] != "abc" {
		t.Fatalf("body token = %q, want abc", body["token"])
	}
}

func TestError(t *testing.T) {
	rec := httptest.NewRecorder()

	response.Error(rec, http.StatusBadRequest, "invalid request")

	if got := rec.Code; got != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", got, http.StatusBadRequest)
	}
}
