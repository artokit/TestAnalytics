package app

import (
	"fmt"
	"log/slog"
	"net/http"
	"shared/database"
	"shared/httputil"
	middleware "shared/httputil/middleware"
	middleware_config "shared/httputil/middleware/config"
	"time"
	"users-service/config"
	"users-service/internal/handlers"
	"users-service/internal/repo/user"
	"users-service/internal/services"

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
	cfg, err := middleware_config.FaultsConfigFromEnv()
	if err != nil {
		panic(fmt.Sprintf("faults config: %v", err))
	}

	faults := middleware.NewFaults(cfg.ErrorRate500, cfg.ErrorRate429, cfg.RetryAfter)
	b.router.Use(middleware.NewMetrics("users-service").Middleware)
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

func (b *ServerBuilder) initHandlers() {
	usersRepo := user.NewRepo(b.store.Db)
	usersService := services.New(usersRepo, b.logger)
	b.router.HandleFunc("/users", handlers.GetUsers(usersService)).Methods(http.MethodGet)

	b.router.HandleFunc("/healthz", httputil.Healthz()).Methods(http.MethodGet)
	b.router.Handle("/metrics", httputil.PrometheusHandler()).Methods(http.MethodGet)
}
