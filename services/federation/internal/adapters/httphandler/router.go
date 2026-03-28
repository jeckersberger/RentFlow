package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
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
			r.Post("/", partnerHandler.Create)
			r.Get("/", partnerHandler.List)
			r.Get("/{id}", partnerHandler.Get)
			r.Put("/{id}", partnerHandler.Update)
		})

		// Shared listings.
		r.Route("/api/v1/shared-listings", func(r chi.Router) {
			r.Post("/", listingHandler.Create)
			r.Get("/", listingHandler.List)
			r.Get("/{id}", listingHandler.Get)
			r.Put("/{id}", listingHandler.Update)
		})

		// Federation requests.
		r.Route("/api/v1/federation-requests", func(r chi.Router) {
			r.Post("/", requestHandler.Create)
			r.Get("/", requestHandler.List)
			r.Get("/{id}", requestHandler.Get)
			r.Put("/{id}", requestHandler.Update)
		})
	})

	return r
}
