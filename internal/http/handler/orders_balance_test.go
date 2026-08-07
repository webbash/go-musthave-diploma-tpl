package handler_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"go-musthave-diploma-tpl/internal/domain"
	"go-musthave-diploma-tpl/internal/http/handler"
	"go-musthave-diploma-tpl/internal/model"
	"go-musthave-diploma-tpl/internal/repository"
	"go-musthave-diploma-tpl/internal/service"
	"go-musthave-diploma-tpl/internal/service/mocks"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestPostOrderHandler(t *testing.T) {
	const secret = "secret"
	const number = "79927398713"

	t.Run("accepted", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockOrderRepository(ctrl)
		svc := service.NewOrderService(repo)
		h := handler.NewPostOrder(svc, zap.NewNop())

		repo.EXPECT().
			CreateOrder(gomock.Any(), gomock.Any()).
			Return(model.Order{Number: number, UserID: 42}, nil)

		rec := serveAuthed(t, secret, h, http.MethodPost, "/api/user/orders", []byte(number))

		if rec.Code != http.StatusAccepted {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusAccepted)
		}
	})

	t.Run("current user duplicate", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockOrderRepository(ctrl)
		svc := service.NewOrderService(repo)
		h := handler.NewPostOrder(svc, zap.NewNop())

		repo.EXPECT().
			CreateOrder(gomock.Any(), gomock.Any()).
			Return(model.Order{}, repository.ErrDuplicateOrder)
		repo.EXPECT().
			FindOrderByNumber(gomock.Any(), number, int64(42)).
			Return(model.Order{Number: number, UserID: 42}, nil)

		rec := serveAuthed(t, secret, h, http.MethodPost, "/api/user/orders", []byte(number))

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("another user duplicate", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockOrderRepository(ctrl)
		svc := service.NewOrderService(repo)
		h := handler.NewPostOrder(svc, zap.NewNop())

		repo.EXPECT().
			CreateOrder(gomock.Any(), gomock.Any()).
			Return(model.Order{}, repository.ErrDuplicateOrder)
		repo.EXPECT().
			FindOrderByNumber(gomock.Any(), number, int64(42)).
			Return(model.Order{}, repository.ErrOrderNotFound)

		rec := serveAuthed(t, secret, h, http.MethodPost, "/api/user/orders", []byte(number))

		if rec.Code != http.StatusConflict {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusConflict)
		}
	})

	t.Run("invalid order", func(t *testing.T) {
		svc := service.NewOrderService(nil)
		h := handler.NewPostOrder(svc, zap.NewNop())

		rec := serveAuthed(t, secret, h, http.MethodPost, "/api/user/orders", []byte("123"))
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnprocessableEntity)
		}
	})
}

func TestGetOrdersHandler(t *testing.T) {
	const secret = "secret"

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockOrderRepository(ctrl)
		svc := service.NewOrderService(repo)
		h := handler.NewGetOrders(svc, zap.NewNop())

		repo.EXPECT().
			ListOrdersByUser(gomock.Any(), int64(42)).
			Return([]model.Order{
				{Number: "79927398713", Status: model.OrderStatusProcessed, Accrual: decimal.NewFromFloat(12.5), UploadedAt: time.Unix(100, 0)},
			}, nil)

		rec := serveAuthed(t, secret, h, http.MethodGet, "/api/user/orders", nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		var resp []handler.OrderResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if len(resp) != 1 || resp[0].Number != "79927398713" {
			t.Fatalf("response = %+v", resp)
		}
	})

	t.Run("no content", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockOrderRepository(ctrl)
		svc := service.NewOrderService(repo)
		h := handler.NewGetOrders(svc, zap.NewNop())

		repo.EXPECT().
			ListOrdersByUser(gomock.Any(), int64(42)).
			Return(nil, nil)

		rec := serveAuthed(t, secret, h, http.MethodGet, "/api/user/orders", nil)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
		}
	})
}

func TestBalanceHandler(t *testing.T) {
	const secret = "secret"

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockBalanceRepository(ctrl)
		svc := service.NewBalanceService(repo)
		h := handler.NewGetBalance(svc, zap.NewNop())

		repo.EXPECT().GetBalance(gomock.Any(), int64(42)).Return(100.5, nil)
		repo.EXPECT().GetWithdrawnTotal(gomock.Any(), int64(42)).Return(25.25, nil)

		rec := serveAuthed(t, secret, h, http.MethodGet, "/api/user/balance", nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		var resp handler.BalanceResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if resp.Current != 100.5 || resp.Withdrawn != 25.25 {
			t.Fatalf("response = %+v", resp)
		}
	})
}

func TestWithdrawHandler(t *testing.T) {
	const secret = "secret"
	const number = "79927398713"

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockBalanceRepository(ctrl)
		svc := service.NewBalanceService(repo)
		h := handler.NewWithdraw(svc, zap.NewNop())

		repo.EXPECT().
			Withdraw(gomock.Any(), int64(42), number, 10.0, gomock.AssignableToTypeOf(time.Time{})).
			Return(nil)

		reqBody, _ := json.Marshal(handler.WithdrawRequest{Order: number, Sum: 10})
		rec := serveAuthed(t, secret, h, http.MethodPost, "/api/user/balance/withdraw", reqBody)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("invalid order", func(t *testing.T) {
		svc := service.NewBalanceService(nil)
		h := handler.NewWithdraw(svc, zap.NewNop())

		reqBody, _ := json.Marshal(handler.WithdrawRequest{Order: "123", Sum: 10})
		rec := serveAuthed(t, secret, h, http.MethodPost, "/api/user/balance/withdraw", reqBody)
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnprocessableEntity)
		}
	})

	t.Run("invalid input", func(t *testing.T) {
		svc := service.NewBalanceService(nil)
		h := handler.NewWithdraw(svc, zap.NewNop())

		reqBody, _ := json.Marshal(handler.WithdrawRequest{Order: number, Sum: 0})
		rec := serveAuthed(t, secret, h, http.MethodPost, "/api/user/balance/withdraw", reqBody)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("not enough balance", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockBalanceRepository(ctrl)
		svc := service.NewBalanceService(repo)
		h := handler.NewWithdraw(svc, zap.NewNop())

		repo.EXPECT().
			Withdraw(gomock.Any(), int64(42), number, 10.0, gomock.AssignableToTypeOf(time.Time{})).
			Return(domain.ErrNotEnoughBalance)

		reqBody, _ := json.Marshal(handler.WithdrawRequest{Order: number, Sum: 10})
		rec := serveAuthed(t, secret, h, http.MethodPost, "/api/user/balance/withdraw", reqBody)
		if rec.Code != http.StatusPaymentRequired {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusPaymentRequired)
		}
	})
}

func TestWithdrawalsHandler(t *testing.T) {
	const secret = "secret"

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockBalanceRepository(ctrl)
		svc := service.NewBalanceService(repo)
		h := handler.NewWithdrawals(svc, zap.NewNop())

		repo.EXPECT().
			ListWithdrawals(gomock.Any(), int64(42)).
			Return([]model.Withdrawal{{Order: "1", Sum: 10, ProcessedAt: time.Unix(100, 0)}}, nil)

		rec := serveAuthed(t, secret, h, http.MethodGet, "/api/user/withdrawals", nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		var resp []handler.WithdrawResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if len(resp) != 1 || resp[0].Order != "1" {
			t.Fatalf("response = %+v", resp)
		}
	})

	t.Run("no content", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockBalanceRepository(ctrl)
		svc := service.NewBalanceService(repo)
		h := handler.NewWithdrawals(svc, zap.NewNop())

		repo.EXPECT().
			ListWithdrawals(gomock.Any(), int64(42)).
			Return(nil, nil)

		rec := serveAuthed(t, secret, h, http.MethodGet, "/api/user/withdrawals", nil)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
		}
	})
}
