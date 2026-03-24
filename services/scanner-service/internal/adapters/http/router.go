package http

import (
	"net/http"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/scanner-service/internal/application"
)

func NewRouter(scanSvc *application.ScanService, sessionSvc *application.SessionService, logger logger.Logger) *http.ServeMux {
	router := http.NewServeMux()
	handler := NewHandler(scanSvc, sessionSvc, logger)

	// Health & readiness
	router.HandleFunc("GET /health", healthHandler)
	router.HandleFunc("GET /ready", readyHandler)

	// Scan routes (legacy)
	router.HandleFunc("POST /api/v1/scan", handler.ProcessScan)
	router.HandleFunc("POST /api/v1/scan/batch", handler.ProcessBatch)
	router.HandleFunc("POST /api/v1/scan/sync", handler.SyncOfflineScans)
	router.HandleFunc("GET /api/v1/scan/history", handler.GetHistory)
	router.HandleFunc("GET /api/v1/scan/resolve/{barcode}", handler.ResolveBarcode)

	// Device routes
	router.HandleFunc("POST /api/v1/scan/devices", handler.RegisterDevice)
	router.HandleFunc("GET /api/v1/scan/devices", handler.ListDevices)

	// Scanner App API contract endpoints
	router.HandleFunc("POST /api/v1/scanner/scan", handler.ScannerScan)
	router.HandleFunc("POST /api/v1/scanner/checkout", handler.ScannerCheckout)
	router.HandleFunc("POST /api/v1/scanner/checkin", handler.ScannerCheckin)
	router.HandleFunc("POST /api/v1/scanner/bulk", handler.ScannerBulk)
	router.HandleFunc("POST /api/v1/scanner/adhoc-booking", handler.AdhocBooking)

	// Scanner Device Management ("Find My Scanner")
	router.HandleFunc("POST /api/v1/scanner/devices/register", handler.RegisterScannerDevice)
	router.HandleFunc("GET /api/v1/scanner/devices", handler.ListScannerDevices)
	router.HandleFunc("POST /api/v1/scanner/devices/{id}/ring", handler.RingScannerDevice)
	router.HandleFunc("GET /api/v1/scanner/devices/{device_id}/ring", handler.CheckRingRequest)
	router.HandleFunc("POST /api/v1/scanner/devices/{device_id}/ring-ack", handler.AckRing)

	// Session routes (M2.2)
	router.HandleFunc("POST /api/v1/scanner/sessions", handler.StartSession)
	router.HandleFunc("PUT /api/v1/scanner/sessions/{id}/end", handler.EndSession)
	router.HandleFunc("POST /api/v1/scanner/sessions/{id}/scan", handler.ProcessSessionScan)
	router.HandleFunc("GET /api/v1/scanner/sessions/{id}/protocol", handler.GetSessionProtocol)
	router.HandleFunc("POST /api/v1/scanner/sync", handler.SyncOfflineQueue)
	router.HandleFunc("POST /api/v1/scanner/offline", handler.QueueOfflineScan)

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
