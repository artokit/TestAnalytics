package app

import (
	"fmt"
	"log/slog"
	"net/http"
	"payments-service/config"
	"payments-service/internal/handlers"
	"payments-service/internal/repo/payment"
	"payments-service/internal/services"
	"test_analytics/shared/database"
	"time"

	"github.com/gorilla/mux"
)

type ServerBuilder struct {
	router *mux.Router
	logger *slog.Logger
	store  *database.Storage
	config *config.Config
}

func NewServerBuilder(logger *slog.Logger, store *database.Storage, config *config.Config) *ServerBuilder {
	return &ServerBuilder{
		router: mux.NewRouter(),
		logger: logger,
		store:  store,
		config: config,
	}
}

func (b *ServerBuilder) Build() *http.Server {
	b.initHandlers()
	addr := fmt.Sprintf("%s:%d", b.config.Address, b.config.Port)
	b.logger.Info("listening", "addr", addr)

	return &http.Server{
		Addr:              addr,
		Handler:           b.router,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func (b *ServerBuilder) initHandlers() {
	paymentsRepo := payment.NewRepo(b.store.Db)
	paymentsService := services.NewPaymentService(paymentsRepo, b.logger)
	b.router.HandleFunc("/payments", handlers.GetPayments(paymentsService)).Methods(http.MethodGet)
}
