package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// NewRouter creates and configures the chi router with all inventory-service routes.
func NewRouter(
	equipmentHandler *EquipmentHandler,
	categoryHandler *CategoryHandler,
	typeHandler *EquipmentTypeHandler,
	flightcaseHandler *FlightcaseHandler,
	pricingHandler *PricingHandler,
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

		// Equipment endpoints.
		r.Route("/api/v1/equipment", func(r chi.Router) {
			r.Get("/", equipmentHandler.List)
			r.Post("/", equipmentHandler.Create)
			r.Get("/search", equipmentHandler.Search)
			r.Get("/lookup/barcode/{barcode}", equipmentHandler.LookupBarcode)
			r.Get("/lookup/rfid/{tag}", equipmentHandler.LookupRFID)
			r.Get("/{id}", equipmentHandler.Get)
			r.Put("/{id}", equipmentHandler.Update)
			r.Delete("/{id}", equipmentHandler.Delete)
			r.Patch("/{id}/status", equipmentHandler.UpdateStatus)
			r.Patch("/{id}/condition", equipmentHandler.UpdateCondition)
			r.Patch("/{id}/rfid", equipmentHandler.AssignRFID)
			r.Get("/{id}/history", equipmentHandler.GetHistory)
			r.Post("/{id}/calculate-price", pricingHandler.CalculatePrice)
		})

		// Category endpoints.
		r.Route("/api/v1/categories", func(r chi.Router) {
			r.Get("/", categoryHandler.List)
			r.Post("/", categoryHandler.Create)
			r.Get("/{id}", categoryHandler.Get)
			r.Put("/{id}", categoryHandler.Update)
			r.Delete("/{id}", categoryHandler.Delete)
		})

		// Equipment type endpoints.
		r.Route("/api/v1/equipment-types", func(r chi.Router) {
			r.Get("/", typeHandler.List)
			r.Post("/", typeHandler.Create)
			r.Get("/{id}", typeHandler.Get)
			r.Put("/{id}", typeHandler.Update)
			r.Delete("/{id}", typeHandler.Delete)
			r.Post("/{id}/create-items", typeHandler.BulkCreate)
		})

		// Price rule endpoints.
		r.Route("/api/v1/price-rules", func(r chi.Router) {
			r.Get("/", pricingHandler.ListPriceRules)
			r.Post("/", pricingHandler.CreatePriceRule)
			r.Put("/{id}", pricingHandler.UpdatePriceRule)
		})

		// Flightcase endpoints.
		r.Route("/api/v1/flightcases", func(r chi.Router) {
			r.Get("/", flightcaseHandler.List)
			r.Post("/", flightcaseHandler.Create)
			r.Get("/{id}", flightcaseHandler.Get)
			r.Put("/{id}", flightcaseHandler.Update)
			r.Delete("/{id}", flightcaseHandler.Delete)
			r.Post("/{id}/items", flightcaseHandler.AddItem)
			r.Delete("/{id}/items/{itemId}", flightcaseHandler.RemoveItem)
			r.Get("/{id}/items", flightcaseHandler.GetItems)
		})
	})

	return r
}
