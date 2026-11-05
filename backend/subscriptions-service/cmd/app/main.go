package main

import (
	"log/slog"
	"os"
	"subscriptions-service/internal/app"
)

func main() {
	handler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.NewLogLogger(handler, slog.LevelInfo)
	serverBuilder := app.New(logger)
	server, err := serverBuilder.Build()
	if err != nil {
		logger.Println(err)
		os.Exit(1)
	}

	err = server.ListenAndServe()
	if err != nil {
		logger.Fatal(err)
	}
}
