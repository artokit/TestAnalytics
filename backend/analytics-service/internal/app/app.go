package app

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"analytics-service/config"
	"analytics-service/internal/handlers"
	"analytics-service/internal/services"
	worker "analytics-service/internal/worker"

	"shared/httputil"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

type ServerBuilder struct {
	router      *mux.Router
	logger      *slog.Logger
	worker      *worker.Service
	metricsRead *services.MetricsReadService
	cfg         *config.Config
}

func NewServerBuilder(logger *slog.Logger, worker *worker.Service, read *services.MetricsReadService, cfg *config.Config) *ServerBuilder {
	return &ServerBuilder{
		router:      mux.NewRouter(),
		logger:      logger,
		worker:      worker,
		metricsRead: read,
		cfg:         cfg,
	}
}

func (b *ServerBuilder) Build() *http.Server {
	b.registerSwagger()
	b.initHandlers()
	addr := fmt.Sprintf("%s:%d", b.cfg.Address, b.cfg.Port)
	b.logger.Info("listening", "addr", addr)

	return &http.Server{
		Addr:         addr,
		Handler:      b.router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func (b *ServerBuilder) registerSwagger() {
	b.router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
	b.router.Path("/swagger").HandlerFunc(http.RedirectHandler("/swagger/", http.StatusPermanentRedirect).ServeHTTP)
}

func (b *ServerBuilder) initHandlers() {
	b.router.HandleFunc("/metrics/ltv/summary", handlers.GetLtvSummary(b.metricsRead, b.worker)).Methods(http.MethodGet)

	b.router.HandleFunc("/metrics/ltv/by-plan", handlers.GetLtvByPlan()).Methods(http.MethodGet)
	b.router.HandleFunc("/metrics/retention", handlers.GetRetention(b.metricsRead, b.worker)).Methods(http.MethodGet)

	b.router.HandleFunc("/status", handlers.GetStatus(b.worker)).Methods(http.MethodGet)
	b.router.HandleFunc("/healthz", httputil.Healthz()).Methods(http.MethodGet)
}
