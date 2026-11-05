package app

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"subscriptions-service/config"
	_ "subscriptions-service/docs"
	"subscriptions-service/internal/handlers"
	"subscriptions-service/internal/repo/subscription"
	"subscriptions-service/internal/services"
	"test_analytics/shared/database"
	"test_analytics/shared/httputil"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
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
	faults, err := httputil.NewFaultsFromEnv()
	if err != nil {
		panic(fmt.Sprintf("faults config: %v", err))
	}
	b.router.Use(faults.Middleware)

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

func (b *ServerBuilder) registerSwagger() {
	b.router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
	b.router.Path("/swagger").HandlerFunc(http.RedirectHandler("/swagger/", http.StatusPermanentRedirect).ServeHTTP)
}

func (b *ServerBuilder) initHandlers() {
	b.registerSwagger()

	planService := services.NewPlanService()
	b.router.HandleFunc("/plans", handlers.GetPlans(planService)).Methods(http.MethodGet)

	subsRepo := subscription.NewRepo(b.store.Db)
	subsService := services.NewSubscriptionService(subsRepo, b.logger)
	b.router.HandleFunc("/subscriptions", handlers.GetSubscriptions(subsService)).Methods(http.MethodGet)
}
