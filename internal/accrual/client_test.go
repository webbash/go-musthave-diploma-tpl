package accrual

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-musthave-diploma-tpl/internal/model"

	"go.uber.org/zap"
)

func TestClientGetOrder(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Fatalf("method = %s", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"order":"79927398713","status":"PROCESSED","accrual":15.5}`))
		}))
		defer server.Close()

		client := NewAccrualClient(server.URL, server.Client(), zap.NewNop())
		order, err := client.GetOrder(context.Background(), "79927398713")
		if err != nil {
			t.Fatalf("GetOrder() error = %v", err)
		}
		if order.Number != "79927398713" || order.Status != model.OrderStatusProcessed {
			t.Fatalf("GetOrder() = %+v", order)
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client := NewAccrualClient(server.URL, server.Client(), zap.NewNop())
		_, err := client.GetOrder(context.Background(), "79927398713")
		if err != ErrOrderNotFound {
			t.Fatalf("GetOrder() error = %v, want ErrOrderNotFound", err)
		}
	})

	t.Run("rate limited", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Retry-After", "2")
			w.WriteHeader(http.StatusTooManyRequests)
		}))
		defer server.Close()

		client := NewAccrualClient(server.URL, server.Client(), zap.NewNop())
		_, err := client.GetOrder(context.Background(), "79927398713")
		rateLimitErr, ok := err.(*RateLimitError)
		if !ok {
			t.Fatalf("GetOrder() error = %T %v, want *RateLimitError", err, err)
		}
		if rateLimitErr.RetryAfter != 2*time.Second {
			t.Fatalf("RetryAfter = %v, want 2s", rateLimitErr.RetryAfter)
		}
	})
}
