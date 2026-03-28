package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
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
	r.Use(requestIDMiddleware)
	r.Use(corsMiddleware)

	r.Get("/health", healthHandler)
	r.Get("/livez", livenessHandler)

	r.Group(func(r chi.Router) {
		r.Use(jwtMiddleware)

		// Project endpoints
		r.Route("/api/v1/projects", func(r chi.Router) {
			r.Get("/", projectHandler.List)
			r.Post("/", projectHandler.Create)
			r.Get("/search", projectHandler.Search)
			r.Get("/{id}", projectHandler.Get)
			r.Put("/{id}", projectHandler.Update)
			r.Delete("/{id}", projectHandler.Delete)
			r.Patch("/{id}/status", projectHandler.UpdateStatus)

			// Project equipment
			r.Post("/{id}/equipment", projectHandler.AddEquipment)
			r.Get("/{id}/equipment", projectHandler.ListEquipment)
			r.Delete("/{id}/equipment/{equipmentId}", projectHandler.RemoveEquipment)

			// Project packlists
			r.Get("/{id}/packlists", packlistHandler.ListByProject)
		})

		// Packlist endpoints
		r.Route("/api/v1/packlists", func(r chi.Router) {
			r.Post("/", packlistHandler.Create)
			r.Get("/{id}", packlistHandler.Get)
			r.Patch("/{id}/status", packlistHandler.UpdateStatus)
			r.Post("/{id}/items", packlistHandler.AddItem)
			r.Get("/{id}/items", packlistHandler.GetItems)
			r.Patch("/{id}/items/{itemId}/packed", packlistHandler.UpdateItemPacked)
			r.Patch("/{id}/items/{itemId}/returned", packlistHandler.UpdateItemReturned)
		})

		// Reservation endpoints
		r.Route("/api/v1/reservations", func(r chi.Router) {
			r.Get("/", reservationHandler.List)
			r.Post("/", reservationHandler.Create)
			r.Get("/{id}", reservationHandler.Get)
			r.Patch("/{id}/status", reservationHandler.UpdateStatus)
		})
	})

	return r
}
