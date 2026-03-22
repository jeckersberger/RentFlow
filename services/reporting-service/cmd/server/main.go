package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/jeckersberger/rentflow/services/reporting-service/internal/application"
	httphandlers "github.com/jeckersberger/rentflow/services/reporting-service/internal/infrastructure/http"
	"github.com/jeckersberger/rentflow/services/reporting-service/internal/infrastructure/persistence"
)

func main() {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dbHost := getEnvOrDefault("DB_HOST", "localhost")
		dbPort := getEnvOrDefault("DB_PORT", "5432")
		dbName := getEnvOrDefault("DB_NAME", "reporting_service")
		dbUser := getEnvOrDefault("DB_USER", "rentflow")
		dbPassword := getEnvOrDefault("DB_PASSWORD", "rentflow_dev")
		dsn = fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable", dbHost, dbPort, dbName, dbUser, dbPassword)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Fatal().Err(err).Msg("Failed to ping database")
	}

	logger.Info().Msg("Connected to database")

	reportRepo := persistence.NewPostgresReportRepository(db)
	kpiRepo := persistence.NewPostgresKPIRepository(db)

	reportService := application.NewReportService(reportRepo)
	kpiService := application.NewKPIService(kpiRepo)

	mux := http.NewServeMux()
	httphandlers.SetupRoutes(mux, reportService, kpiService, logger)

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the scheduled reports processor in the background
	go scheduleReportsProcessor(ctx, reportService, logger)

	port := 8012
	addr := fmt.Sprintf(":%d", port)

	logger.Info().Int("port", port).Msg("Starting reporting-service")

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Run server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Server error")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Shutting down server and scheduler")
	cancel()

	// Shutdown server
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error().Err(err).Msg("Server shutdown error")
	}

	logger.Info().Msg("Server stopped")
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// scheduleReportsProcessor runs every hour and processes scheduled reports for all tenants
func scheduleReportsProcessor(ctx context.Context, reportService *application.ReportService, logger zerolog.Logger) {
	logger.Info().Msg("Starting scheduled reports processor")
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("Scheduled reports processor stopping")
			return
		case <-ticker.C:
			logger.Info().Msg("Processing scheduled reports")
			// For now, process for a default tenant ID (in production, iterate over all tenants)
			tenantID := uuid.Nil
			if _, err := reportService.ProcessScheduledReports(ctx, tenantID); err != nil {
				logger.Error().Err(err).Msg("Failed to process scheduled reports")
			}
		}
	}
}
