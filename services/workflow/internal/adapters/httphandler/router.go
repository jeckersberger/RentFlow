package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// NewRouter creates and configures the chi router with all workflow-service routes.
func NewRouter(
	definitionHandler *DefinitionHandler,
	instanceHandler *InstanceHandler,
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

		// Workflow definition endpoints.
		r.Route("/api/v1/workflow-definitions", func(r chi.Router) {
			r.Post("/", definitionHandler.Create)
			r.Get("/", definitionHandler.List)
			r.Get("/{id}", definitionHandler.GetByID)
			r.Put("/{id}", definitionHandler.Update)
		})

		// Workflow instance endpoints.
		r.Route("/api/v1/workflow-instances", func(r chi.Router) {
			r.Post("/", instanceHandler.Create)
			r.Get("/", instanceHandler.List)
			r.Get("/{id}", instanceHandler.GetByID)
			r.Post("/{id}/action", instanceHandler.PerformAction)
		})
	})

	return r
}
