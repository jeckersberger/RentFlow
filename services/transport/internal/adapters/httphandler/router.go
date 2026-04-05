package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
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
	r.Use(middleware.MaxBodySize(1 << 20)) // 1 MB body limit
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
			// Read: all authenticated users
			r.Get("/", vehicleHandler.List)
			r.Get("/{id}", vehicleHandler.GetByID)

			// Write: manager or admin only
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin", "manager"))
				r.Post("/", vehicleHandler.Create)
				r.Put("/{id}", vehicleHandler.Update)
			})

			// Delete: admin only
			r.With(middleware.RequireRole("admin")).Delete("/{id}", vehicleHandler.Delete)
		})

		// Transport order endpoints.
		r.Route("/api/v1/transport-orders", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/", orderHandler.List)
			r.Get("/{id}", orderHandler.GetByID)
			r.Get("/{id}/items", orderHandler.ListItems)
			r.Get("/{id}/costs", orderHandler.ListCosts)

			// Write: manager or admin only
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin", "manager"))
				r.Post("/", orderHandler.Create)
				r.Put("/{id}", orderHandler.Update)
				r.Patch("/{id}/complete", orderHandler.Complete)
				r.Post("/{id}/items", orderHandler.AddItem)
				r.Post("/{id}/capacity-check", orderHandler.CapacityCheck)
				r.Post("/{id}/costs", orderHandler.AddCost)
			})
		})
	})

	return r
}
