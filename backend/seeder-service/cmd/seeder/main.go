package main

import (
	"context"
	"log/slog"
	"os"

	"seeder-service/internal/app"
)

func main() {
	if err := app.Run(context.Background()); err != nil {
		slog.Error("seeder failed", "error", err)
		os.Exit(1)
	}
}
