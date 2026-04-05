package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
)

func NewRouter(
	scheduleHandler *ScheduleHandler,
	taskHandler *TaskHandler,
	echeckHandler *ECheckHandler,
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

		r.Route("/api/v1/maintenance-schedules", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/", scheduleHandler.List)
			r.Get("/{id}", scheduleHandler.Get)

			// Write: manager or admin only
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin", "manager"))
				r.Post("/", scheduleHandler.Create)
				r.Put("/{id}", scheduleHandler.Update)
			})
		})

		r.Route("/api/v1/maintenance-tasks", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/", taskHandler.List)
			r.Get("/{id}", taskHandler.Get)
			r.Get("/{id}/logs", taskHandler.ListLogs)

			// Write: manager or admin only
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin", "manager"))
				r.Post("/", taskHandler.Create)
				r.Put("/{id}", taskHandler.Update)
				r.Patch("/{id}/complete", taskHandler.Complete)
			})

			// Delete: admin only
			r.With(middleware.RequireRole("admin")).Delete("/{id}", taskHandler.Delete)
		})

		r.Route("/api/v1/maintenance/echeck", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/", echeckHandler.List)
			r.Get("/overdue", echeckHandler.ListOverdue)
			r.Get("/{equipmentId}", echeckHandler.ListByEquipment)

			// Write: manager or admin only
			r.With(middleware.RequireRole("admin", "manager")).Post("/", echeckHandler.Create)
		})
	})

	return r
}
