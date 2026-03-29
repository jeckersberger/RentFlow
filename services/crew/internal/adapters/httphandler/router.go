package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// NewRouter creates and configures the chi router with all crew-service routes.
func NewRouter(
	crewHandler *CrewHandler,
	assignmentHandler *AssignmentHandler,
	qualificationHandler *QualificationHandler,
	timeTrackingHandler *TimeTrackingHandler,
	availabilityHandler *AvailabilityHandler,
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

		// Crew member endpoints.
		r.Route("/api/v1/crew", func(r chi.Router) {
			r.Post("/", crewHandler.Create)
			r.Get("/", crewHandler.List)

			// Time-tracking endpoints (must be before /{id} to avoid route conflict).
			r.Route("/time-tracking", func(r chi.Router) {
				r.Post("/check-in", timeTrackingHandler.CheckIn)
				r.Post("/check-out", timeTrackingHandler.CheckOut)
				r.Get("/active/{memberId}", timeTrackingHandler.GetActive)
				r.Get("/entries", timeTrackingHandler.ListEntries)
				r.Get("/export", timeTrackingHandler.Export)
			})

			// Availability endpoints (global).
			r.Post("/availability", availabilityHandler.Set)
			r.Get("/availability", availabilityHandler.ListAll)

			// Crew member by ID endpoints.
			r.Get("/{id}", crewHandler.Get)
			r.Put("/{id}", crewHandler.Update)
			r.Delete("/{id}", crewHandler.Delete)

			// Nested assignment endpoints for a crew member.
			r.Post("/{id}/assignments", assignmentHandler.CreateForMember)
			r.Get("/{id}/assignments", assignmentHandler.ListForMember)

			// Nested qualification endpoints for a crew member.
			r.Post("/{id}/qualifications", qualificationHandler.CreateForMember)
			r.Get("/{id}/qualifications", qualificationHandler.ListForMember)

			// Nested availability for a specific member.
			r.Get("/{id}/availability", availabilityHandler.ListForMember)
		})

		// Global assignment listing (with optional project_id filter).
		r.Get("/api/v1/crew-assignments", assignmentHandler.ListAll)
	})

	return r
}
