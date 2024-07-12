package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"sync-algo/internal/config"
	algorithmController "sync-algo/internal/controller/algorithm"
	clientController "sync-algo/internal/controller/client"
	"sync-algo/internal/lib/logger"
	"sync-algo/internal/lib/logger/sl"
	algorithmService "sync-algo/internal/service/algorithm"
	clientService "sync-algo/internal/service/client"
	"sync-algo/internal/storage/postgres"
	"syscall"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

// @title Algo Sync Service API
// @version 1.0
// @description Test task for Effective-mobile.
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
		return
	}

	// Service layer
	clientService := clientService.New(storage, log)
	algorithmService := algorithmService.New(storage, log)

	// Controller layer
	clientController := clientController.New(clientService, log)
	algorithmController := algorithmController.New(algorithmService, log)

	// Deployer and Scheduler initialization
	// deployer := NewK8sDeployer() // Assuming this function is implemented elsewhere
	// sch := scheduler.NewScheduler(deployer, storage)

	// go func() {
	// 	if err := sch.Start(ctx); err != nil {
	// 		log.Error("scheduler error", sl.Error(err))
	// 	}
	// }()

	// Init router
	r := chi.NewRouter()
	chi.NewMux()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Route("/clients", clientController.Register())
	r.Route("/algorithms", algorithmController.Register())

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
