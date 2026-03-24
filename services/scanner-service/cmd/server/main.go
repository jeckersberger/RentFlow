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
	"github.com/jeckersberger/rentflow/pkg/common/middleware"
	httpAdapter "github.com/jeckersberger/rentflow/services/scanner-service/internal/adapters/http"
	"github.com/jeckersberger/rentflow/services/scanner-service/internal/application"
	"github.com/jeckersberger/rentflow/services/scanner-service/internal/infrastructure/clients"
	"github.com/jeckersberger/rentflow/services/scanner-service/internal/infrastructure/repositories"
)

const (
	serviceName = "scanner-service"
	servicePort = 8004
)

func main() {
	cfg := config.Load(serviceName)
	if cfg.ServicePort == 8080 {
		cfg.ServicePort = servicePort
	}
	log := logger.New(cfg.LogLevel, serviceName)

	log.Info("Starting service", "name", serviceName, "port", cfg.ServicePort, "env", cfg.Environment)

	// Initialize database
	dbPool, err := database.NewPostgresPool(cfg.ConnectionString())
	if err != nil {
		log.Fatal("Failed to connect to database", err)
	}
	defer dbPool.Close()

	log.Info("Connected to database")

	// Initialize repositories
	scanEventRepo := repositories.NewScanEventPostgres(dbPool)
	deviceRepo := repositories.NewDevicePostgres(dbPool)
	sessionRepo := repositories.NewScanSessionPostgres(dbPool)
	queueRepo := repositories.NewOfflineQueuePostgres(dbPool)

	log.Info("Repositories initialized")

	// Initialize inventory service HTTP client
	inventoryURL := os.Getenv("INVENTORY_SERVICE_URL")
	if inventoryURL == "" {
		inventoryURL = "http://localhost:8002"
	}
	inventoryClient := clients.NewInventoryHTTPClient(inventoryURL)
	log.Info("Inventory service client initialized", "url", inventoryURL)

	// Initialize services
	scanSvc := application.NewScanService(scanEventRepo, deviceRepo, inventoryClient, log)
	sessionSvc := application.NewSessionService(sessionRepo, scanEventRepo, queueRepo, deviceRepo, inventoryClient, log)

	log.Info("Services initialized")

	// Setup router
	router := httpAdapter.NewRouter(scanSvc, sessionSvc, log)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.ServicePort),
		Handler:      middleware.SimpleCORSMiddleware()(router),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info("Listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server error", err)
		}
	}()

	// Wait for interrupt signal
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
