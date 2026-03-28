package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
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
			r.Get("/", warehouseHandler.List)
			r.Post("/", warehouseHandler.Create)
			r.Get("/{id}", warehouseHandler.Get)
		})

		// Zone endpoints.
		r.Route("/api/v1/zones", func(r chi.Router) {
			r.Get("/", zoneHandler.List)
			r.Post("/", zoneHandler.Create)
			r.Put("/{id}", zoneHandler.Update)
		})

		// Rack endpoints.
		r.Route("/api/v1/racks", func(r chi.Router) {
			r.Get("/", rackHandler.List)
			r.Post("/", rackHandler.Create)
		})

		// Stock location endpoints.
		r.Route("/api/v1/locations", func(r chi.Router) {
			r.Get("/", locationHandler.List)
			r.Post("/", locationHandler.Create)
			r.Get("/{id}", locationHandler.Get)
		})

		// Movement endpoints.
		r.Route("/api/v1/movements", func(r chi.Router) {
			r.Get("/", movementHandler.List)
			r.Post("/", movementHandler.Create)
		})

		// Inventory check endpoints.
		r.Route("/api/v1/inventory-checks", func(r chi.Router) {
			r.Get("/", inventoryCheckHandler.List)
			r.Post("/", inventoryCheckHandler.Create)
			r.Patch("/{id}/complete", inventoryCheckHandler.Complete)
			r.Post("/{id}/scan", inventoryCheckHandler.ScanItem)
			r.Get("/{id}/result", inventoryCheckHandler.GetResult)
		})
	})

	return r
}
