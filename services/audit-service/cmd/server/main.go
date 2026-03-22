package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	nethttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"github.com/jeckersberger/rentflow/pkg/common/config"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	audithttp "github.com/jeckersberger/rentflow/services/audit-service/internal/adapters/http"
	"github.com/jeckersberger/rentflow/services/audit-service/internal/application"
	"github.com/jeckersberger/rentflow/services/audit-service/internal/infrastructure/repositories"
)

const (
	serviceName = "audit-service"
	servicePort = 8017
)

func main() {
	// Load configuration
	cfg := config.Load(serviceName)
	if cfg.ServicePort == 8080 {
		cfg.ServicePort = servicePort
	}
	log := logger.New(cfg.LogLevel, serviceName)

	log.Info("Starting service", "name", serviceName, "port", cfg.ServicePort, "env", cfg.Environment)

	// Connect to PostgreSQL
	db, err := connectPostgres(cfg.ConnectionString(), log)
	if err != nil {
		log.Fatal("Failed to connect to database", err)
	}
	defer db.Close()

	log.Info("Connected to database")

	// Create repositories
	auditRepo := repositories.NewPostgresAuditRepository(db, log)
	exportRepo := repositories.NewPostgresExportRepository(db, log)

	// Create services
	auditService := application.NewAuditService(auditRepo, log)
	verificationService := application.NewVerificationService(auditRepo, log)
	pseudonymizationService := application.NewPseudonymizationService(auditRepo, log)
	exportService := application.NewExportService(auditRepo, exportRepo, log)

	// Setup router
	router := nethttp.NewServeMux()

	// Health & readiness
	router.HandleFunc("GET /health", healthHandler(serviceName))
	router.HandleFunc("GET /ready", readyHandler(serviceName, db, log))

	// Setup API routes
	audithttp.SetupRoutes(router, auditService, verificationService, pseudonymizationService, exportService, log)

	// Create HTTP server
	srv := &nethttp.Server{
		Addr:         fmt.Sprintf(":%d", cfg.ServicePort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Info("Listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != nethttp.ErrServerClosed {
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
	srv.Shutdown(ctx)
	log.Info("Server stopped")
}

func healthHandler(serviceName string) nethttp.HandlerFunc {
	return func(w nethttp.ResponseWriter, r *nethttp.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(nethttp.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"service":   serviceName,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}
}

func readyHandler(serviceName string, db *sql.DB, log logger.Logger) nethttp.HandlerFunc {
	return func(w nethttp.ResponseWriter, r *nethttp.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			log.Error("Database not ready", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(nethttp.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "not ready",
				"service": serviceName,
				"reason":  "database connection failed",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(nethttp.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ready",
			"service": serviceName,
		})
	}
}

func connectPostgres(connectionString string, log logger.Logger) (*sql.DB, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}
