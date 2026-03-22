package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/jeckersberger/rentflow/pkg/common/middleware"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/rentflow/services/insurance-service/internal/adapters/database"
	adaptorshttp "github.com/jeckersberger/rentflow/services/insurance-service/internal/adapters/http"
	"github.com/jeckersberger/rentflow/services/insurance-service/internal/application"
)

func main() {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbHost := getEnvOrDefault("DB_HOST", "localhost")
		dbPort := getEnvOrDefault("DB_PORT", "5432")
		dbName := getEnvOrDefault("DB_NAME", "insurance_service")
		dbUser := getEnvOrDefault("DB_USER", "rentflow")
		dbPassword := getEnvOrDefault("DB_PASSWORD", "rentflow_dev")
		dbURL = fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable", dbHost, dbPort, dbName, dbUser, dbPassword)
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Close()

	// Test database connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		logger.Fatal().Err(err).Msg("database ping failed")
	}

	logger.Info().Msg("connected to database")

	// Initialize repositories
	policyRepo := database.NewPolicyRepository(db)
	claimRepo := database.NewClaimRepository(db)
	claimItemRepo := database.NewClaimItemRepository(db)

	// Initialize services
	policyService := application.NewPolicyService(policyRepo, logger)
	claimService := application.NewClaimService(claimRepo, policyRepo, claimItemRepo, logger)

	// Initialize HTTP handler
	handler := adaptorshttp.NewHandler(policyService, claimService, logger)

	// Setup routes
	mux := http.NewServeMux()
	adaptorshttp.RegisterRoutes(mux, handler)

	// Create HTTP server
	port := getEnvOrDefault("SERVICE_PORT", "8012")
	addr := ":" + port
	server := &http.Server{
		Addr:         addr,
		Handler:      middleware.SimpleCORSMiddleware()(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info().Str("port", addr).Msg("insurance-service starting")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("server error")
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info().Msg("shutdown signal received")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error().Err(err).Msg("server shutdown error")
	}

	logger.Info().Msg("insurance-service stopped")
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
