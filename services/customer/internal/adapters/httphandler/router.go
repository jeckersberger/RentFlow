package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
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
	r.Use(middleware.MaxBodySize(1 << 20)) // 1 MB body limit
	r.Use(requestIDMiddleware)
	r.Use(corsMiddleware)

	r.Get("/health", healthHandler)
	r.Get("/livez", livenessHandler)

	r.Group(func(r chi.Router) {
		r.Use(jwtMiddleware)

		// Customer endpoints
		r.Route("/api/v1/customers", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/", customerHandler.List)
			r.Get("/search", customerHandler.Search)
			r.Get("/{id}", customerHandler.Get)
			r.Get("/{id}/contacts", contactHandler.ListByCustomer)

			// Write: admin, manager, or user
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin", "manager", "user"))
				r.Post("/", customerHandler.Create)
				r.Put("/{id}", customerHandler.Update)
			})

			// Delete: admin or manager only
			r.With(middleware.RequireRole("admin", "manager")).Delete("/{id}", customerHandler.Delete)
		})

		// Contact endpoints
		r.Route("/api/v1/contacts", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/{id}", contactHandler.Get)
			r.Get("/{id}/notes", contactHandler.ListNotes)

			// Write: admin, manager, or user
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin", "manager", "user"))
				r.Post("/", contactHandler.Create)
				r.Put("/{id}", contactHandler.Update)
				r.Post("/{id}/notes", contactHandler.AddNote)
			})

			// Delete: admin or manager only
			r.With(middleware.RequireRole("admin", "manager")).Delete("/{id}", contactHandler.Delete)
		})
	})

	return r
}
