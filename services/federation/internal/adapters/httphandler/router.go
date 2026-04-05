package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
)

// NewRouter creates and configures the chi router with all federation-service routes.
func NewRouter(
	partnerHandler *PartnerHandler,
	listingHandler *ListingHandler,
	requestHandler *RequestHandler,
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

		// Federation partners.
		r.Route("/api/v1/federation-partners", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/", partnerHandler.List)
			r.Get("/{id}", partnerHandler.Get)

			// Write: admin only
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin"))
				r.Post("/", partnerHandler.Create)
				r.Put("/{id}", partnerHandler.Update)
			})
		})

		// Shared listings.
		r.Route("/api/v1/shared-listings", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/", listingHandler.List)
			r.Get("/{id}", listingHandler.Get)

			// Write: admin only
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin"))
				r.Post("/", listingHandler.Create)
				r.Put("/{id}", listingHandler.Update)
			})
		})

		// Federation requests.
		r.Route("/api/v1/federation-requests", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/", requestHandler.List)
			r.Get("/{id}", requestHandler.Get)

			// Write: admin only
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin"))
				r.Post("/", requestHandler.Create)
				r.Put("/{id}", requestHandler.Update)
			})
		})
	})

	return r
}
