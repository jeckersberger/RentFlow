package http

import (
	"net/http"

	"github.com/rs/zerolog"
	"github.com/jeckersberger/rentflow/services/reporting-service/internal/application"
)

func SetupRoutes(mux *http.ServeMux, reportService *application.ReportService, kpiService *application.KPIService, logger zerolog.Logger) {
	handler := NewReportHandler(reportService, kpiService, logger)

	mux.HandleFunc("GET /health", handler.Health)
	mux.HandleFunc("GET /ready", handler.Ready)

	mux.HandleFunc("GET /api/v1/reports/definitions", handler.ListReportDefinitions)
	mux.HandleFunc("POST /api/v1/reports/definitions", handler.CreateReportDefinition)
	mux.HandleFunc("GET /api/v1/reports/definitions/{id}", handler.GetReportDefinition)
	mux.HandleFunc("PUT /api/v1/reports/definitions/{id}", handler.UpdateReportDefinition)
	mux.HandleFunc("POST /api/v1/reports/definitions/{id}/run", handler.GenerateReport)

	mux.HandleFunc("GET /api/v1/reports/runs", handler.ListReportRuns)
	mux.HandleFunc("GET /api/v1/reports/runs/{id}", handler.GetReportRun)

	mux.HandleFunc("GET /api/v1/reports/kpi/dashboard", handler.GetKPIDashboard)
	mux.HandleFunc("GET /api/v1/reports/kpi/trends", handler.GetKPITrends)
	mux.HandleFunc("GET /api/v1/reports/kpi/snapshot", handler.GetKPISnapshot)
}
