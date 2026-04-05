package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
)

// NewRouter creates and configures the chi router with all crew-service routes.
func NewRouter(
	crewHandler *CrewHandler,
	assignmentHandler *AssignmentHandler,
	qualificationHandler *QualificationHandler,
	timeTrackingHandler *TimeTrackingHandler,
	availabilityHandler *AvailabilityHandler,
	availabilityBlockHandler *AvailabilityBlockHandler,
	skillMatchHandler *SkillMatchHandler,
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

		// Crew member endpoints.
		r.Route("/api/v1/crew", func(r chi.Router) {
			// Read: all authenticated users
			r.Get("/", crewHandler.List)
			r.Get("/members", crewHandler.List)
			r.Get("/availability", availabilityHandler.ListAll)
			r.Get("/availability-blocks", availabilityBlockHandler.ListAll)
			r.Post("/match-skills", skillMatchHandler.MatchSkills)
			r.Get("/{id}", crewHandler.Get)
			r.Get("/{id}/assignments", assignmentHandler.ListForMember)
			r.Get("/{id}/qualifications", qualificationHandler.ListForMember)
			r.Get("/{id}/availability", availabilityHandler.ListForMember)
			r.Get("/{id}/availability-blocks", availabilityBlockHandler.ListForMember)

			// Time-tracking read endpoints (must be before /{id} to avoid route conflict).
			r.Route("/time-tracking", func(r chi.Router) {
				r.Get("/active/{memberId}", timeTrackingHandler.GetActive)
				r.Get("/entries", timeTrackingHandler.ListEntries)
				r.Get("/export", timeTrackingHandler.Export)

				// Time-tracking write: manager or admin only
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireRole("admin", "manager"))
					r.Post("/check-in", timeTrackingHandler.CheckIn)
					r.Post("/check-out", timeTrackingHandler.CheckOut)
				})
			})

			// Write: manager or admin only
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin", "manager"))
				r.Post("/", crewHandler.Create)
				r.Post("/availability", availabilityHandler.Set)
				r.Put("/{id}", crewHandler.Update)
				r.Post("/{id}/assignments", assignmentHandler.CreateForMember)
				r.Post("/{id}/qualifications", qualificationHandler.CreateForMember)
				r.Post("/{id}/availability-blocks", availabilityBlockHandler.Create)
				r.Delete("/{id}/availability-blocks/{blockId}", availabilityBlockHandler.Delete)
			})

			// Delete: admin only
			r.With(middleware.RequireRole("admin")).Delete("/{id}", crewHandler.Delete)
		})

		// Global assignment listing (with optional project_id filter).
		r.Get("/api/v1/crew-assignments", assignmentHandler.ListAll)
	})

	return r
}
