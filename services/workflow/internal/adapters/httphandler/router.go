package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
)

// NewRouter creates and configures the chi router with all workflow-service routes.
func NewRouter(
	definitionHandler *DefinitionHandler,
	instanceHandler *InstanceHandler,
	templateHandler *TemplateHandler,
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

		// Workflow template endpoints.
		r.Get("/api/v1/workflow-templates", templateHandler.ListTemplates)

		// Workflow definition endpoints.
		r.Route("/api/v1/workflow-definitions", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/", definitionHandler.List)
			r.Get("/{id}", definitionHandler.GetByID)

			// Write: manager or admin only
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin", "manager"))
				r.Post("/", definitionHandler.Create)
				r.Post("/from-template", templateHandler.CreateFromTemplate)
				r.Put("/{id}", definitionHandler.Update)
			})
		})

		// Workflow instance endpoints.
		r.Route("/api/v1/workflow-instances", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/", instanceHandler.List)
			r.Get("/{id}", instanceHandler.GetByID)

			// Write: manager or admin only
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin", "manager"))
				r.Post("/", instanceHandler.Create)
				r.Post("/{id}/action", instanceHandler.PerformAction)
			})
		})
	})

	return r
}
