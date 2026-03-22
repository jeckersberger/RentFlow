package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/jeckersberger/rentflow/services/reporting-service/internal/application"
	httphandlers "github.com/jeckersberger/rentflow/services/reporting-service/internal/infrastructure/http"
	"github.com/jeckersberger/rentflow/services/reporting-service/internal/infrastructure/persistence"
)

func main() {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://localhost/rentflow?sslmode=disable"
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

	port := 8012
	addr := fmt.Sprintf(":%d", port)

	logger.Info().Int("port", port).Msg("Starting reporting-service")

	if err := http.ListenAndServe(addr, mux); err != nil && err != http.ErrServerClosed {
		logger.Fatal().Err(err).Msg("Server error")
	}
}
