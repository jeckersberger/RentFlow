package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// NewRouter creates and configures the chi router with all scanner-service routes.
func NewRouter(
	scanHandler *ScanHandler,
	deviceHandler *DeviceHandler,
	sessionHandler *SessionHandler,
	healthHandler http.HandlerFunc,
	livenessHandler http.HandlerFunc,
	jwtMiddleware func(http.Handler) http.Handler,
	corsMiddleware func(http.Handler) http.Handler,
	recoveryMiddleware func(http.Handler) http.Handler,
	requestIDMiddleware func(http.Handler) http.Handler,
) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware (order matters: outermost first).
	r.Use(recoveryMiddleware)
	r.Use(requestIDMiddleware)
	r.Use(corsMiddleware)

	// Health endpoints (no auth required).
	r.Get("/health", healthHandler)
	r.Get("/livez", livenessHandler)

	// Protected routes (JWT required).
	r.Group(func(r chi.Router) {
		r.Use(jwtMiddleware)

		// Scanner endpoints.
		r.Route("/api/v1/scanner", func(r chi.Router) {
			// Scan operations.
			r.Post("/scan", scanHandler.Scan)
			r.Post("/checkout", scanHandler.Checkout)
			r.Post("/checkin", scanHandler.Checkin)
			r.Post("/bulk", scanHandler.BulkSync)
			r.Post("/adhoc-booking", scanHandler.AdhocBooking)
			r.Get("/events", scanHandler.ListEvents)

			// Equipment resolve (lookup by barcode/QR/RFID).
			r.Get("/resolve/{identifier}", sessionHandler.Resolve)

			// Scan sessions.
			r.Post("/sessions", sessionHandler.CreateSession)
			r.Put("/sessions/{id}/end", sessionHandler.EndSession)
			r.Get("/sessions/{id}/protocol", sessionHandler.GetProtocol)
			r.Post("/sessions/{id}/signature", sessionHandler.SaveSignature)

			// Device management.
			r.Post("/devices/register", deviceHandler.Register)
			r.Get("/devices", deviceHandler.List)
			r.Get("/devices/ring", deviceHandler.CheckRing)
			r.Post("/devices/ring-ack", deviceHandler.AckRing)
			r.Post("/devices/{id}/ring", deviceHandler.TriggerRing)
		})
	})

	return r
}
