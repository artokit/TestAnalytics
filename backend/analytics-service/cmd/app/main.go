package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"analytics-service/config"
	_ "analytics-service/docs"
	"analytics-service/internal/app"
	"analytics-service/internal/cache"
	"analytics-service/internal/services"
	"analytics-service/internal/storage"
	worker "analytics-service/internal/worker"
	"analytics-service/pkg"

	"github.com/redis/go-redis/v9"
)

// @title        Analytics Service API
// @version      1.0
// @description  REST API сервиса аналитики.
// @BasePath     /
// @Schemes      http
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "analytics-service")

	cfg, err := config.NewConfig()
	if err != nil {
		logger.Error("error on load config", "error", err)
		os.Exit(1)
	}

	rdb := redis.NewClient(&redis.Options{Addr: cfg.Redis.Address})
	defer func() {
		if err := rdb.Close(); err != nil {
			logger.Error("error on close redis", "error", err)
		}
	}()

	cacheRepo := cache.NewCache(rdb, 24*time.Hour)
	metricsService := services.NewMetricsService(cacheRepo)
	metricsRead := services.NewMetricsReadService(cacheRepo, cfg.Sync.Interval)
	userStorage := storage.NewUserStorage()
	paymentStorage := storage.NewPaymentStorage()
	subStorage := storage.NewSubscriptionStorage()
	planStorage := storage.NewPlanStorage()

	hc := pkg.NewHttpClient(&http.Client{Timeout: 10 * time.Second})
	usersService := services.NewUserService(hc, logger, cfg)
	paymentsService := services.NewPaymentService(hc, logger, cfg)
	subsService := services.NewSubscriptionService(hc, logger, cfg)
	wrk := worker.New(logger, cfg.Sync, userStorage, paymentStorage, subStorage, planStorage, usersService, paymentsService, subsService, metricsService)
	go wrk.Run(ctx)

	currencyService := services.NewCurrencyService(hc, logger, cfg)
	server := app.NewServerBuilder(logger, wrk, metricsRead, currencyService, cfg).Build()

	go func() {
		<-ctx.Done()
		logger.Info("shutting down...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown error", "error", err)
		}
	}()

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("listen and serve error", "error", err)
		os.Exit(1)
	}
}
