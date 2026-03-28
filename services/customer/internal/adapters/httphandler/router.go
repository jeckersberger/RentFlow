package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter(
	customerHandler *CustomerHandler,
	contactHandler *ContactHandler,
	healthHandler http.HandlerFunc,
	livenessHandler http.HandlerFunc,
	jwtMiddleware func(http.Handler) http.Handler,
	corsMiddleware func(http.Handler) http.Handler,
	recoveryMiddleware func(http.Handler) http.Handler,
	requestIDMiddleware func(http.Handler) http.Handler,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(recoveryMiddleware)
	r.Use(requestIDMiddleware)
	r.Use(corsMiddleware)

	r.Get("/health", healthHandler)
	r.Get("/livez", livenessHandler)

	r.Group(func(r chi.Router) {
		r.Use(jwtMiddleware)

		// Customer endpoints
		r.Route("/api/v1/customers", func(r chi.Router) {
			r.Get("/", customerHandler.List)
			r.Post("/", customerHandler.Create)
			r.Get("/search", customerHandler.Search)
			r.Get("/{id}", customerHandler.Get)
			r.Put("/{id}", customerHandler.Update)
			r.Delete("/{id}", customerHandler.Delete)

			// Customer contacts (sub-resource)
			r.Get("/{id}/contacts", contactHandler.ListByCustomer)
		})

		// Contact endpoints
		r.Route("/api/v1/contacts", func(r chi.Router) {
			r.Post("/", contactHandler.Create)
			r.Get("/{id}", contactHandler.Get)
			r.Put("/{id}", contactHandler.Update)
			r.Delete("/{id}", contactHandler.Delete)

			// Contact notes (sub-resource)
			r.Post("/{id}/notes", contactHandler.AddNote)
			r.Get("/{id}/notes", contactHandler.ListNotes)
		})
	})

	return r
}
