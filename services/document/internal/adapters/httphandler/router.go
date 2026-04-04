package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// NewRouter creates and configures the chi router with all document-service routes.
func NewRouter(
	templateHandler *TemplateHandler,
	documentHandler *DocumentHandler,
	attachmentHandler *AttachmentHandler,
	renderHandler *RenderHandler,
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

		// Document template endpoints (CRUD).
		r.Route("/api/v1/document-templates", func(r chi.Router) {
			r.Post("/", templateHandler.Create)
			r.Get("/", templateHandler.List)
			r.Get("/{id}", templateHandler.GetByID)
			r.Put("/{id}", templateHandler.Update)
		})

		// Document endpoints.
		r.Route("/api/v1/documents", func(r chi.Router) {
			r.Post("/", documentHandler.Create)
			r.Get("/", documentHandler.List)
			r.Get("/templates", templateHandler.List)

			r.Get("/{id}", documentHandler.GetByID)

			// Template rendering endpoints.
			r.Post("/render", renderHandler.Render)
			r.Route("/templates/defaults", func(r chi.Router) {
				r.Get("/", renderHandler.ListDefaults)
				r.Get("/{type}/preview", renderHandler.PreviewDefault)
			})
		})

		// Attachment endpoints.
		r.Route("/api/v1/attachments", func(r chi.Router) {
			r.Post("/", attachmentHandler.Create)
			r.Get("/", attachmentHandler.List)
			r.Get("/{id}", attachmentHandler.GetByID)
			r.Delete("/{id}", attachmentHandler.Delete)
		})
	})

	return r
}
