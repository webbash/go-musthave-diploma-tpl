package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"go-musthave-diploma-tpl/internal/config"
	httpserver "go-musthave-diploma-tpl/internal/http"
	"go-musthave-diploma-tpl/internal/repository"
	"go-musthave-diploma-tpl/internal/service"
)

func main() {
	cfg := config.Load()

	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	db, err := repository.OpenPostgres(cfg.DatabaseURI)
	if err != nil {
		logger.Fatal("open database", zap.Error(err))
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			logger.Error("close database", zap.Error(closeErr))
		}
	}()

	userRepo := repository.NewUserRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	balanceRepo := repository.NewBalanceRepository(db)

	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret)
	orderSvc := service.NewOrderService(orderRepo)
	balanceSvc := service.NewBalanceService(balanceRepo)
	server := httpserver.NewServer(authSvc, orderSvc, balanceSvc, cfg.JWTSecret, logger)

	srv := &http.Server{
		Addr:    cfg.RunAddress,
		Handler: server.Handler(),
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("gophermart listening", zap.String("addr", cfg.RunAddress))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("listen", zap.Error(err))
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", zap.Error(err))
	}
}
