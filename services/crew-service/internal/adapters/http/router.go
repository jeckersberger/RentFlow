package http

import (
	"net/http"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/crew-service/internal/application"
)

func NewRouter(
	crewSvc *application.CrewService,
	timeEntrySvc *application.TimeEntryService,
	assignmentSvc *application.AssignmentService,
	logger logger.Logger,
) *http.ServeMux {
	router := http.NewServeMux()
	handler := NewHandler(crewSvc, timeEntrySvc, assignmentSvc, logger)

	// Health & readiness
	router.HandleFunc("GET /health", healthHandler)
	router.HandleFunc("GET /ready", readyHandler)

	// Crew member routes
	router.HandleFunc("POST /api/v1/crew", handler.CreateCrewMember)
	router.HandleFunc("GET /api/v1/crew", handler.ListCrewMembers)
	router.HandleFunc("GET /api/v1/crew/{id}", handler.GetCrewMember)
	router.HandleFunc("PUT /api/v1/crew/{id}", handler.UpdateCrewMember)
	router.HandleFunc("DELETE /api/v1/crew/{id}", handler.DeleteCrewMember)

	// Time entry routes
	router.HandleFunc("POST /api/v1/time-entries", handler.CreateTimeEntry)
	router.HandleFunc("GET /api/v1/time-entries", handler.ListTimeEntries)
	router.HandleFunc("GET /api/v1/time-entries/{id}", handler.GetTimeEntry)
	router.HandleFunc("PUT /api/v1/time-entries/{id}", handler.UpdateTimeEntry)
	router.HandleFunc("POST /api/v1/time-entries/{id}/approve", handler.ApproveTimeEntry)
	router.HandleFunc("POST /api/v1/time-entries/{id}/reject", handler.RejectTimeEntry)
	router.HandleFunc("GET /api/v1/time-entries/summary", handler.GetTimeEntrySummary)

	// Assignment routes
	router.HandleFunc("POST /api/v1/assignments", handler.CreateAssignment)
	router.HandleFunc("GET /api/v1/assignments", handler.ListAssignments)
	router.HandleFunc("GET /api/v1/assignments/{id}", handler.GetAssignment)
	router.HandleFunc("DELETE /api/v1/assignments/{id}", handler.DeleteAssignment)
	router.HandleFunc("GET /api/v1/assignments/project/{id}", handler.GetAssignmentsByProject)

	return router
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"crew-service"}`))
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready","service":"crew-service"}`))
}
