package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
)

// NewRouter creates and configures the chi router with all inventory-service routes.
func NewRouter(
	equipmentHandler *EquipmentHandler,
	categoryHandler *CategoryHandler,
	typeHandler *EquipmentTypeHandler,
	flightcaseHandler *FlightcaseHandler,
	pricingHandler *PricingHandler,
	availabilityHandler *AvailabilityHandler,
	scanResolveHandler *ScanResolveHandler,
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

		// Equipment endpoints.
		r.Route("/api/v1/equipment", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/", equipmentHandler.List)
			r.Get("/search", equipmentHandler.Search)
			r.Get("/availability", availabilityHandler.ListAvailability)
			r.Get("/lookup/barcode/{barcode}", equipmentHandler.LookupBarcode)
			r.Get("/lookup/rfid/{tag}", equipmentHandler.LookupRFID)
			r.Get("/{id}", equipmentHandler.Get)
			r.Get("/{id}/history", equipmentHandler.GetHistory)
			r.Get("/{id}/availability", availabilityHandler.GetItemAvailability)

			// Write: manager or admin only
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin", "manager"))
				r.Post("/", equipmentHandler.Create)
				r.Post("/availability-check", availabilityHandler.CheckSingle)
				r.Post("/batch-availability", availabilityHandler.CheckBatch)
				r.Put("/{id}", equipmentHandler.Update)
				r.Patch("/{id}/status", equipmentHandler.UpdateStatus)
				r.Patch("/{id}/condition", equipmentHandler.UpdateCondition)
				r.Patch("/{id}/rfid", equipmentHandler.AssignRFID)
				r.Post("/{id}/calculate-price", pricingHandler.CalculatePrice)
			})

			// Delete: admin only
			r.With(middleware.RequireRole("admin")).Delete("/{id}", equipmentHandler.Delete)
		})

		// Category endpoints.
		r.Route("/api/v1/categories", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/", categoryHandler.List)
			r.Get("/{id}", categoryHandler.Get)

			// Write: manager or admin only
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin", "manager"))
				r.Post("/", categoryHandler.Create)
				r.Put("/{id}", categoryHandler.Update)
			})

			// Delete: admin only
			r.With(middleware.RequireRole("admin")).Delete("/{id}", categoryHandler.Delete)
		})

		// Equipment type endpoints.
		r.Route("/api/v1/equipment-types", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/", typeHandler.List)
			r.Get("/{id}", typeHandler.Get)

			// Write: manager or admin only
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin", "manager"))
				r.Post("/", typeHandler.Create)
				r.Put("/{id}", typeHandler.Update)
				r.Post("/{id}/create-items", typeHandler.BulkCreate)
			})

			// Delete: admin only
			r.With(middleware.RequireRole("admin")).Delete("/{id}", typeHandler.Delete)
		})

		// Price rule endpoints.
		r.Route("/api/v1/price-rules", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/", pricingHandler.ListPriceRules)

			// Write: manager or admin only
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin", "manager"))
				r.Post("/", pricingHandler.CreatePriceRule)
				r.Put("/{id}", pricingHandler.UpdatePriceRule)
			})
		})

		// Scan resolve endpoint (used by scanner-service).
		r.Route("/api/v1/scan", func(r chi.Router) {
			r.Get("/resolve/{identifier}", scanResolveHandler.Resolve)
		})

		// Flightcase endpoints.
		r.Route("/api/v1/flightcases", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/", flightcaseHandler.List)
			r.Get("/{id}", flightcaseHandler.Get)
			r.Get("/{id}/items", flightcaseHandler.GetItems)

			// Write: manager or admin only
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin", "manager"))
				r.Post("/", flightcaseHandler.Create)
				r.Put("/{id}", flightcaseHandler.Update)
				r.Post("/{id}/items", flightcaseHandler.AddItem)
			})

			// Delete: admin only
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin"))
				r.Delete("/{id}", flightcaseHandler.Delete)
				r.Delete("/{id}/items/{itemId}", flightcaseHandler.RemoveItem)
			})
		})
	})

	return r
}
