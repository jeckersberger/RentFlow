package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter(
	definitionHandler *DefinitionHandler,
	snapshotHandler *SnapshotHandler,
	widgetHandler *WidgetHandler,
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

		r.Route("/api/v1/report-definitions", func(r chi.Router) {
			r.Get("/", definitionHandler.List)
			r.Post("/", definitionHandler.Create)
			r.Get("/{id}", definitionHandler.Get)
			r.Put("/{id}", definitionHandler.Update)
			r.Post("/{id}/generate", snapshotHandler.Generate)
		})

		r.Route("/api/v1/report-snapshots", func(r chi.Router) {
			r.Get("/", snapshotHandler.List)
			r.Get("/{id}", snapshotHandler.Get)
		})

		r.Route("/api/v1/dashboard-widgets", func(r chi.Router) {
			r.Get("/", widgetHandler.List)
			r.Post("/", widgetHandler.Create)
			r.Put("/{id}", widgetHandler.Update)
			r.Delete("/{id}", widgetHandler.Delete)
		})
	})

	return r
}
