package app

import (
	"fmt"
	"log"
	"net/http"
	"subscriptions-service/config"
	_ "subscriptions-service/docs"
	"subscriptions-service/internal/handlers"
	"subscriptions-service/internal/repo/subscription"
	"subscriptions-service/internal/services"
	"test_analytics/shared/database"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

type ServerBuilder struct {
	router *mux.Router
	logger *log.Logger
	store  *database.Storage
}

func New(logger *log.Logger) *ServerBuilder {
	return &ServerBuilder{
		router: mux.NewRouter(),
		logger: logger,
	}
}

func (b *ServerBuilder) Build() (*http.Server, error) {
	cfg, err := config.NewConfig()
	if err != nil {
		return nil, err
	}

	b.store, err = database.New(cfg.Database.ConnectionString)
	if err != nil {
		return nil, err
	}
	//defer func() { _ = b.store.Close() }()

	if err := b.store.RunMigrations("./migrations"); err != nil {
		return nil, err
	}

	b.initHandlers()
	addr := fmt.Sprintf("%s:%d", cfg.Address, cfg.Port)
	b.logger.Printf("subscriptions-service listening on %s", addr)
	return &http.Server{
		Addr:    addr,
		Handler: b.router,
	}, nil
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
	subsService := services.NewSubscriptionService(subsRepo)
	b.router.HandleFunc("/subscriptions", handlers.GetSubscriptions(subsService)).Methods(http.MethodGet)
}
