package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"seeder-service/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx); err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		slog.Error("seeder failed", "error", err)
		os.Exit(1)
	}
}
