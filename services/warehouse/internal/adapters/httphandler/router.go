package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
)

// NewRouter creates and configures the chi router with all warehouse-service routes.
func NewRouter(
	warehouseHandler *WarehouseHandler,
	zoneHandler *ZoneHandler,
	rackHandler *RackHandler,
	locationHandler *LocationHandler,
	movementHandler *MovementHandler,
	inventoryCheckHandler *InventoryCheckHandler,
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

		// Warehouse endpoints.
		r.Route("/api/v1/warehouses", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/", warehouseHandler.List)
			r.Get("/{id}", warehouseHandler.Get)

			// Write: manager or admin only
			r.With(middleware.RequireRole("admin", "manager")).Post("/", warehouseHandler.Create)
		})

		// Zone endpoints.
		r.Route("/api/v1/zones", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/", zoneHandler.List)

			// Write: manager or admin only
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin", "manager"))
				r.Post("/", zoneHandler.Create)
				r.Put("/{id}", zoneHandler.Update)
			})
		})

		// Rack endpoints.
		r.Route("/api/v1/racks", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/", rackHandler.List)

			// Write: manager or admin only
			r.With(middleware.RequireRole("admin", "manager")).Post("/", rackHandler.Create)
		})

		// Stock location endpoints.
		r.Route("/api/v1/locations", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/", locationHandler.List)
			r.Get("/{id}", locationHandler.Get)

			// Write: manager or admin only
			r.With(middleware.RequireRole("admin", "manager")).Post("/", locationHandler.Create)
		})

		// Movement endpoints.
		r.Route("/api/v1/movements", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/", movementHandler.List)

			// Write: manager or admin only
			r.With(middleware.RequireRole("admin", "manager")).Post("/", movementHandler.Create)
		})

		// Inventory check endpoints.
		r.Route("/api/v1/inventory-checks", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/", inventoryCheckHandler.List)
			r.Get("/{id}/result", inventoryCheckHandler.GetResult)

			// Write: manager or admin only
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin", "manager"))
				r.Post("/", inventoryCheckHandler.Create)
				r.Patch("/{id}/complete", inventoryCheckHandler.Complete)
				r.Post("/{id}/scan", inventoryCheckHandler.ScanItem)
			})
		})
	})

	return r
}
