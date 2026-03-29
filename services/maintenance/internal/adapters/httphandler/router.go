package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
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
	r.Use(requestIDMiddleware)
	r.Use(corsMiddleware)

	r.Get("/health", healthHandler)
	r.Get("/livez", livenessHandler)

	r.Group(func(r chi.Router) {
		r.Use(jwtMiddleware)

		r.Route("/api/v1/maintenance-schedules", func(r chi.Router) {
			r.Get("/", scheduleHandler.List)
			r.Post("/", scheduleHandler.Create)
			r.Get("/{id}", scheduleHandler.Get)
			r.Put("/{id}", scheduleHandler.Update)
		})

		r.Route("/api/v1/maintenance-tasks", func(r chi.Router) {
			r.Get("/", taskHandler.List)
			r.Post("/", taskHandler.Create)
			r.Get("/{id}", taskHandler.Get)
			r.Put("/{id}", taskHandler.Update)
			r.Patch("/{id}/complete", taskHandler.Complete)
			r.Get("/{id}/logs", taskHandler.ListLogs)
		})

		r.Route("/api/v1/maintenance/echeck", func(r chi.Router) {
			r.Post("/", echeckHandler.Create)
			r.Get("/", echeckHandler.List)
			r.Get("/overdue", echeckHandler.ListOverdue)
			r.Get("/{equipmentId}", echeckHandler.ListByEquipment)
		})
	})

	return r
}
