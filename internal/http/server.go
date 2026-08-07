package httpserver

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	httpHandler "go-musthave-diploma-tpl/internal/http/handler"
	httpMiddleware "go-musthave-diploma-tpl/internal/http/middleware"
	"go-musthave-diploma-tpl/internal/service"
)

type Server struct {
	router chi.Router
}

func NewServer(auth *service.AuthService, orders *service.OrderService, balance *service.BalanceService, jwtSecret string, logger *zap.Logger) *Server {
	registerHandler := httpHandler.NewRegister(auth, logger)
	loginHandler := httpHandler.NewLogin(auth, logger)
	postOrderHandler := httpHandler.NewPostOrder(orders, logger)
	getOrdersHandler := httpHandler.NewGetOrders(orders, logger)
	getBalanceHandler := httpHandler.NewGetBalance(balance, logger)
	withdrawHandler := httpHandler.NewWithdraw(balance, logger)
	withdrawalsHandler := httpHandler.NewWithdrawals(balance, logger)

	r := chi.NewRouter()
	r.Use(chiMiddleware.Recoverer)
	r.Use(httpMiddleware.RequestLogger(logger))

	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", registerHandler.ServeHTTP)
		r.Post("/login", loginHandler.ServeHTTP)

		r.Group(func(r chi.Router) {
			r.Use(httpMiddleware.Auth(jwtSecret))
			r.Post("/orders", postOrderHandler.ServeHTTP)
			r.Get("/orders", getOrdersHandler.ServeHTTP)
			r.Get("/balance", getBalanceHandler.ServeHTTP)
			r.Post("/balance/withdraw", withdrawHandler.ServeHTTP)
			r.Get("/withdrawals", withdrawalsHandler.ServeHTTP)
		})
	})

	return &Server{router: r}
}

func (s *Server) Handler() http.Handler {
	return s.router
}
