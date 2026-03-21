package http

import (
	"net/http"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/scanner-service/internal/application"
)

func NewRouter(scanSvc *application.ScanService, logger logger.Logger) *http.ServeMux {
	router := http.NewServeMux()
	handler := NewHandler(scanSvc, logger)

	// Health & readiness
	router.HandleFunc("GET /health", healthHandler)
	router.HandleFunc("GET /ready", readyHandler)

	// Scan routes
	router.HandleFunc("POST /api/v1/scan", handler.ProcessScan)
	router.HandleFunc("POST /api/v1/scan/batch", handler.ProcessBatch)
	router.HandleFunc("POST /api/v1/scan/sync", handler.SyncOfflineScans)
	router.HandleFunc("GET /api/v1/scan/history", handler.GetHistory)
	router.HandleFunc("GET /api/v1/scan/resolve/{barcode}", handler.ResolveBarcode)

	// Device routes
	router.HandleFunc("POST /api/v1/scan/devices", handler.RegisterDevice)
	router.HandleFunc("GET /api/v1/scan/devices", handler.ListDevices)

	return router
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"scanner-service"}`))
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready","service":"scanner-service"}`))
}
