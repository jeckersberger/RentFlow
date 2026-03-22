package http

import (
	"net/http"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/crew-service/internal/application"
)

// SetupRoutes sets up all HTTP routes
func SetupRoutes(
	mux *http.ServeMux,
	crewSvc *application.CrewService,
	qualSvc *application.QualificationService,
	assignmentSvc *application.AssignmentService,
	timeRecordSvc *application.TimeRecordService,
	log logger.Logger,
) {
	handlers := NewHandlers(crewSvc, qualSvc, assignmentSvc, timeRecordSvc, log)

	// Crew member routes
	mux.HandleFunc("GET /api/v1/crew/members", handlers.ListCrewMembers)
	mux.HandleFunc("POST /api/v1/crew/members", handlers.CreateCrewMember)
	mux.HandleFunc("GET /api/v1/crew/members/{id}", handlers.GetCrewMember)
	mux.HandleFunc("PUT /api/v1/crew/members/{id}", handlers.UpdateCrewMember)
	mux.HandleFunc("DELETE /api/v1/crew/members/{id}", handlers.DeleteCrewMember)

	// Qualification routes
	mux.HandleFunc("GET /api/v1/crew/members/{id}/qualifications", handlers.GetQualifications)
	mux.HandleFunc("POST /api/v1/crew/members/{id}/qualifications", handlers.CreateQualification)

	// Availability routes
	mux.HandleFunc("GET /api/v1/crew/members/{id}/availability", handlers.CheckAvailability)

	// Crew assignment routes
	mux.HandleFunc("GET /api/v1/crew/assignments", handlers.ListAssignments)
	mux.HandleFunc("POST /api/v1/crew/assignments", handlers.CreateAssignment)
	mux.HandleFunc("GET /api/v1/crew/assignments/{id}", handlers.GetAssignment)
	mux.HandleFunc("PUT /api/v1/crew/assignments/{id}", handlers.UpdateAssignment)

	// Conflict detection routes
	mux.HandleFunc("GET /api/v1/crew/assignments/conflicts", handlers.DetectConflicts)

	// Time record routes
	mux.HandleFunc("POST /api/v1/crew/time-records/start", handlers.StartTimeRecord)
	mux.HandleFunc("POST /api/v1/crew/time-records/{id}/stop", handlers.StopTimeRecord)
	mux.HandleFunc("GET /api/v1/crew/time-records", handlers.GetTimeRecords)

	// Dashboard route
	mux.HandleFunc("GET /api/v1/crew/dashboard", handlers.GetDashboard)
}
