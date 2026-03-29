package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// NewRouter creates and configures the chi router with all transport-service routes.
func NewRouter(
	vehicleHandler *VehicleHandler,
	orderHandler *OrderHandler,
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

		// Vehicle endpoints.
		r.Route("/api/v1/vehicles", func(r chi.Router) {
			r.Post("/", vehicleHandler.Create)
			r.Get("/", vehicleHandler.List)
			r.Get("/{id}", vehicleHandler.GetByID)
			r.Put("/{id}", vehicleHandler.Update)
		})

		// Transport order endpoints.
		r.Route("/api/v1/transport-orders", func(r chi.Router) {
			r.Post("/", orderHandler.Create)
			r.Get("/", orderHandler.List)
			r.Get("/{id}", orderHandler.GetByID)
			r.Put("/{id}", orderHandler.Update)
			r.Patch("/{id}/complete", orderHandler.Complete)
			r.Post("/{id}/items", orderHandler.AddItem)
			r.Get("/{id}/items", orderHandler.ListItems)
			r.Post("/{id}/capacity-check", orderHandler.CapacityCheck)
			r.Post("/{id}/costs", orderHandler.AddCost)
			r.Get("/{id}/costs", orderHandler.ListCosts)
		})
	})

	return r
}
