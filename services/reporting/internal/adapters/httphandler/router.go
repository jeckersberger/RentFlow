package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
)

func NewRouter(
	definitionHandler *DefinitionHandler,
	snapshotHandler *SnapshotHandler,
	widgetHandler *WidgetHandler,
	kpiHandler *KPIHandler,
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

		// Reporting: read-only for all authenticated users (no creates/deletes from API)
		r.Route("/api/v1/report-definitions", func(r chi.Router) {
			r.Get("/", definitionHandler.List)
			r.Get("/{id}", definitionHandler.Get)
		})

		r.Route("/api/v1/report-snapshots", func(r chi.Router) {
			r.Get("/", snapshotHandler.List)
			r.Get("/{id}", snapshotHandler.Get)
		})

		r.Route("/api/v1/dashboard-widgets", func(r chi.Router) {
			r.Get("/", widgetHandler.List)
		})

		r.Route("/api/v1/reporting/kpis", func(r chi.Router) {
			r.Get("/", kpiHandler.GetLatest)
			r.Get("/history", kpiHandler.GetHistory)
		})
	})

	return r
}
