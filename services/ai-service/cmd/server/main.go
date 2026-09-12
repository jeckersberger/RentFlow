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
	aihttp "github.com/jeckersberger/rentflow/services/ai-service/internal/adapters/http"
	"github.com/jeckersberger/rentflow/services/ai-service/internal/application"
	"github.com/jeckersberger/rentflow/services/ai-service/internal/infrastructure/repositories"
)

const (
	serviceName = "ai-service"
	servicePort = 8013
)

func main() {
	cfg := config.Load(serviceName)
	if cfg.ServicePort == 8080 {
		cfg.ServicePort = servicePort
	}
	log := logger.New(cfg.LogLevel, serviceName)

	log.Info("Starting service", "name", serviceName, "port", cfg.ServicePort, "env", cfg.Environment)

	db, err := connectPostgres(cfg.ConnectionString(), log)
	if err != nil {
		log.Fatal("Failed to connect to database", err)
	}
	defer db.Close()

	log.Info("Connected to database")

	requestRepo := repositories.NewPostgresAIRequestRepository(db, log)
	feedbackRepo := repositories.NewPostgresAIFeedbackRepository(db, log)
	exampleRepo := repositories.NewPostgresFewShotExampleRepository(db, log)
	providerRepo := repositories.NewPostgresAIProviderRepository(db, log)

	providerSvc := application.NewProviderService(log)
	anonymizationSvc := application.NewAnonymizationService(log)
	predictionSvc := application.NewPredictionService(providerSvc, log)
	aiService := application.NewAIService(
		requestRepo,
		feedbackRepo,
		exampleRepo,
		providerRepo,
		providerSvc,
		anonymizationSvc,
		predictionSvc,
		log,
	)
	visionSvc := application.NewReceiptVisionService(providerSvc, log)

	router := nethttp.NewServeMux()
	router.HandleFunc("GET /health", healthHandler(serviceName))
	router.HandleFunc("GET /ready", readyHandler(serviceName, db, log))

	if config.FeatureEnabled("FEATURE_AI_ENABLED") {
		aihttp.SetupRoutes(router, aiService, visionSvc, log)
		log.Warn("AI business API enabled by feature gate", "env", "FEATURE_AI_ENABLED")
	} else {
		log.Info("AI business API disabled", "env", "FEATURE_AI_ENABLED")
	}

	srv := &nethttp.Server{
		Addr:         fmt.Sprintf(":%d", cfg.ServicePort),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 90 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("Listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != nethttp.ErrServerClosed {
			log.Fatal("Server error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Info("Server stopped")
}

func healthHandler(serviceName string) nethttp.HandlerFunc {
	return func(w nethttp.ResponseWriter, r *nethttp.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(nethttp.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
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
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "not ready",
				"service": serviceName,
				"reason":  "database connection failed",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(nethttp.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
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
