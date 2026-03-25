package http

import (
	"net/http"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/pkg/common/middleware"
	"github.com/jeckersberger/rentflow/services/scanner-service/internal/application"
)

func NewRouter(scanSvc *application.ScanService, sessionSvc *application.SessionService, logger logger.Logger, jwtSecret string) *http.ServeMux {
	router := http.NewServeMux()
	handler := NewHandler(scanSvc, sessionSvc, logger)

	// Health & readiness (no auth)
	router.HandleFunc("GET /health", healthHandler)
	router.HandleFunc("GET /ready", readyHandler)

	// Auth + Tenant middleware chain for all API routes
	authMw := middleware.JWTAuthMiddleware(jwtSecret)
	tenantMw := middleware.TenantMiddleware()

	// Helper: wrap handler with auth + tenant middleware
	auth := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authMw(tenantMw(http.HandlerFunc(h))).ServeHTTP(w, r)
		}
	}

	// Scan routes (legacy)
	router.HandleFunc("POST /api/v1/scan", auth(handler.ProcessScan))
	router.HandleFunc("POST /api/v1/scan/batch", auth(handler.ProcessBatch))
	router.HandleFunc("POST /api/v1/scan/sync", auth(handler.SyncOfflineScans))
	router.HandleFunc("GET /api/v1/scan/history", auth(handler.GetHistory))
	router.HandleFunc("GET /api/v1/scan/resolve/{barcode}", auth(handler.ResolveBarcode))

	// Device routes
	router.HandleFunc("POST /api/v1/scan/devices", auth(handler.RegisterDevice))
	router.HandleFunc("GET /api/v1/scan/devices", auth(handler.ListDevices))

	// Scanner App API contract endpoints
	router.HandleFunc("POST /api/v1/scanner/scan", auth(handler.ScannerScan))
	router.HandleFunc("POST /api/v1/scanner/checkout", auth(handler.ScannerCheckout))
	router.HandleFunc("POST /api/v1/scanner/checkin", auth(handler.ScannerCheckin))
	router.HandleFunc("POST /api/v1/scanner/bulk", auth(handler.ScannerBulk))
	router.HandleFunc("POST /api/v1/scanner/adhoc-booking", auth(handler.AdhocBooking))

	// Scanner Device Management ("Find My Scanner")
	router.HandleFunc("POST /api/v1/scanner/devices/register", auth(handler.RegisterScannerDevice))
	router.HandleFunc("GET /api/v1/scanner/devices", auth(handler.ListScannerDevices))
	router.HandleFunc("POST /api/v1/scanner/devices/{id}/ring", auth(handler.RingScannerDevice))
	router.HandleFunc("GET /api/v1/scanner/devices/{device_id}/ring", auth(handler.CheckRingRequest))
	router.HandleFunc("POST /api/v1/scanner/devices/{device_id}/ring-ack", auth(handler.AckRing))

	// Session routes (M2.2)
	router.HandleFunc("POST /api/v1/scanner/sessions", auth(handler.StartSession))
	router.HandleFunc("PUT /api/v1/scanner/sessions/{id}/end", auth(handler.EndSession))
	router.HandleFunc("POST /api/v1/scanner/sessions/{id}/scan", auth(handler.ProcessSessionScan))
	router.HandleFunc("GET /api/v1/scanner/sessions/{id}/protocol", auth(handler.GetSessionProtocol))
	router.HandleFunc("POST /api/v1/scanner/sessions/{id}/signature", auth(handler.UploadSignature))
	router.HandleFunc("POST /api/v1/scanner/sync", auth(handler.SyncOfflineQueue))
	router.HandleFunc("POST /api/v1/scanner/offline", auth(handler.QueueOfflineScan))

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
