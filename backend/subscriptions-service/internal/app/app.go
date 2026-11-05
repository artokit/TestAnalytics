package app

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"shared/database"
	"shared/httputil"
	middleware "shared/httputil/middleware"
	middleware_config "shared/httputil/middleware/config"
	"subscriptions-service/config"
	"subscriptions-service/internal/handlers"
	"subscriptions-service/internal/repo/subscription"
	"subscriptions-service/internal/services"

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
	cfg, err := middleware_config.FaultsConfigFromEnv()
	if err != nil {
		panic(fmt.Sprintf("faults config: %v", err))
	}

	faults := middleware.NewFaults(cfg.ErrorRate500, cfg.ErrorRate429, cfg.RetryAfter)
	b.router.Use(middleware.HealthSafe(faults.Middleware))

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

	b.router.HandleFunc("/healthz", httputil.Healthz()).Methods(http.MethodGet)
}
