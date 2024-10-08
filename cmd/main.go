package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sync-algo/internal/config"
	algorithmController "sync-algo/internal/controller/algorithm"
	clientController "sync-algo/internal/controller/client"
	"sync-algo/internal/deployer"
	"sync-algo/internal/lib/logger"
	"sync-algo/internal/lib/logger/sl"
	"sync-algo/internal/scheduler"
	algorithmService "sync-algo/internal/service/algorithm"
	clientService "sync-algo/internal/service/client"
	"sync-algo/internal/storage/postgres"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Algo Sync Service
// @version 1.0
// @description Microservice from syncing user's algorythms in kubernates.
func main() {
	cfg := config.MustLoad()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logger.New(cfg.Env)
	log.Info("initializing server...", slog.String("port", cfg.Server.Port))

	// Data layer
	storage, err := postgres.New(cfg.Storage)
	if err != nil {
		log.Error("storage initial error", sl.Error(err))
		os.Exit(1)
	}

	// Service layer
	clientService := clientService.New(storage, log)
	algorithmService := algorithmService.New(storage, log)

	// Controller layer
	clientController := clientController.New(clientService, log)
	algorithmController := algorithmController.New(algorithmService, log)

	// Deployer initialization
	deployer, err := deployer.New(cfg.Kubernates)
	if err != nil {
		log.Error(`failed to init 'deployer'`, sl.Error(err))
		os.Exit(1)
	}

	// Scheduler initialization
	sch := scheduler.New(log, storage)
	go sch.Start(ctx, deployer)

	// Init router
	r := chi.NewRouter()
	chi.NewMux()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Route("/clients", clientController.Register())
	r.Route("/algorithms", algorithmController.Register())

	// Swagger documentation
	r.Get("/docs/*", httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http://%s:%s/docs/swagger.json", cfg.Server.Host, cfg.Server.Port)), //The url pointing to API definition
	))

	r.Get("/docs/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/swagger.json")
	})

	// Init server
	srv := http.Server{
		Handler:      r,
		Addr:         cfg.Server.Host + ":" + cfg.Server.Port,
		ReadTimeout:  cfg.Server.Timeout * time.Second,
		WriteTimeout: cfg.Server.Timeout * time.Second,
		IdleTimeout:  cfg.Server.Timeout * time.Second,
	}

	log.Info("server initialized")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start server
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", sl.Error(err))
		}
	}()

	log.Info("server is running...")

	<-stop

	log.Info("shutting down server...")

	c, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(c); err != nil {
		log.Error("failed to shutdown server", sl.Error(err))
	}

	storage.Close()
	log.Info("server stopped")
}
