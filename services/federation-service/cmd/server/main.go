package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/config"
	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	httpAdapter "github.com/jeckersberger/rentflow/services/federation-service/internal/adapters/http"
	"github.com/jeckersberger/rentflow/services/federation-service/internal/application"
	"github.com/jeckersberger/rentflow/services/federation-service/internal/infrastructure/repositories"
)

const (
	serviceName = "federation-service"
	servicePort = 8009
)

func main() {
	cfg := config.Load(serviceName)
	if cfg.ServicePort == 8080 {
		cfg.ServicePort = servicePort
	}
	log := logger.New(cfg.LogLevel, serviceName)

	log.Info("Starting service", "name", serviceName, "port", cfg.ServicePort, "env", cfg.Environment)

	dbPool, err := database.NewPostgresPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database", err)
	}
	defer dbPool.Close()

	log.Info("Connected to database")

	peerRepo := repositories.NewPeerPostgres(dbPool)
	equipmentShareRepo := repositories.NewEquipmentSharePostgres(dbPool)
	shareRequestRepo := repositories.NewShareRequestPostgres(dbPool)

	log.Info("Repositories initialized")

	peerSvc := application.NewPeerService(peerRepo, log)
	sharingSvc := application.NewSharingService(equipmentShareRepo, shareRequestRepo, peerRepo, log)

	log.Info("Services initialized")

	router := httpAdapter.NewRouter(peerSvc, sharingSvc, log)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.ServicePort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("Listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server shutdown error", err)
	}

	log.Info("Server stopped")
}
