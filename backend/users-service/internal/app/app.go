package app

import (
	"fmt"
	"log/slog"
	"net/http"
	"shared/database"
	"shared/httputil"
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

func (b *ServerBuilder) initHandlers() {
	usersRepo := user.NewRepo(b.store.Db)
	usersService := services.New(usersRepo, b.logger)
	b.router.HandleFunc("/users", handlers.GetUsers(usersService)).Methods(http.MethodGet)
}
