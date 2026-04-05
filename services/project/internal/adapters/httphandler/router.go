package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
)

func NewRouter(
	projectHandler *ProjectHandler,
	packlistHandler *PacklistHandler,
	reservationHandler *ReservationHandler,
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

		// Project endpoints
		r.Route("/api/v1/projects", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/", projectHandler.List)
			r.Get("/search", projectHandler.Search)
			r.Get("/calendar", projectHandler.GetCalendar)
			r.Get("/calendar.ics", projectHandler.ExportICS)
			r.Get("/{id}", projectHandler.Get)
			r.Get("/{id}/equipment", projectHandler.ListEquipment)
			r.Get("/{id}/packlists", packlistHandler.ListByProject)

			// Write: manager or admin only
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin", "manager"))
				r.Post("/", projectHandler.Create)
				r.Put("/{id}", projectHandler.Update)
				r.Patch("/{id}/status", projectHandler.UpdateStatus)
				r.Post("/{id}/equipment", projectHandler.AddEquipment)
			})

			// Delete: admin only
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin"))
				r.Delete("/{id}", projectHandler.Delete)
				r.Delete("/{id}/equipment/{equipmentId}", projectHandler.RemoveEquipment)
			})
		})

		// Packlist endpoints
		r.Route("/api/v1/packlists", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/", packlistHandler.List)
			r.Get("/{id}", packlistHandler.Get)
			r.Get("/{id}/summary", packlistHandler.GetSummary)
			r.Get("/{id}/items", packlistHandler.GetItems)

			// Write: manager or admin only
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin", "manager"))
				r.Post("/", packlistHandler.Create)
				r.Patch("/{id}/status", packlistHandler.UpdateStatus)
				r.Post("/{id}/items", packlistHandler.AddItem)
				r.Post("/{id}/bulk-status", packlistHandler.BulkUpdateItemStatus)
				r.Patch("/{id}/items/{itemId}/packed", packlistHandler.UpdateItemPacked)
				r.Patch("/{id}/items/{itemId}/returned", packlistHandler.UpdateItemReturned)
				r.Patch("/{id}/items/{itemId}/status", packlistHandler.UpdateItemStatus)
			})
		})

		// Reservation endpoints
		r.Route("/api/v1/reservations", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/", reservationHandler.List)
			r.Get("/{id}", reservationHandler.Get)

			// Write: manager or admin only
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin", "manager"))
				r.Post("/", reservationHandler.Create)
				r.Patch("/{id}/status", reservationHandler.UpdateStatus)
			})
		})
	})

	return r
}
