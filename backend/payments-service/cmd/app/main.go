package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"payments-service/config"
	"payments-service/internal/app"
	"syscall"
	"shared/database"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "payments-service")

	cfg, err := config.NewConfig()
	if err != nil {
		logger.Error("error on load config", "error", err)
		os.Exit(1)
	}

	store, err := database.New(cfg.Database.ConnectionString)
	if err != nil {
		logger.Error("error connecting to database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := store.Close(); err != nil {
			logger.Error("error on close db", "error", err)
		}
	}()

	if err := store.RunMigrations("./migrations"); err != nil {
		logger.Error("error running migrations", "error", err)
		os.Exit(1)
	}

	server := app.NewServerBuilder(logger, store, cfg).Build()

	go func() {
		<-ctx.Done()
		logger.Info("shutting down...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown error", "error", err)
		}
	}()

	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("listen and serve error", "error", err)
		os.Exit(1)
	}
}
