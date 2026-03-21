package http

import (
	"net/http"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/project-service/internal/application"
)

func NewRouter(
	projectSvc *application.ProjectService,
	packlistSvc *application.PacklistService,
	reservationSvc *application.ReservationService,
	logger *logger.Logger,
) *http.ServeMux {
	router := http.NewServeMux()
	handler := NewHandler(projectSvc, packlistSvc, reservationSvc, logger)

	// Health & readiness
	router.HandleFunc("GET /health", healthHandler)
	router.HandleFunc("GET /ready", readyHandler)

	// Project routes
	router.HandleFunc("POST /api/v1/projects", handler.CreateProject)
	router.HandleFunc("GET /api/v1/projects", handler.ListProjects)
	router.HandleFunc("GET /api/v1/projects/search", handler.SearchProjects)
	router.HandleFunc("GET /api/v1/projects/{id}", handler.GetProject)
	router.HandleFunc("PUT /api/v1/projects/{id}", handler.UpdateProject)
	router.HandleFunc("PATCH /api/v1/projects/{id}/status", handler.ChangeProjectStatus)
	router.HandleFunc("PATCH /api/v1/projects/{id}/manager", handler.SetProjectManager)
	router.HandleFunc("DELETE /api/v1/projects/{id}", handler.DeleteProject)

	// Packlist routes
	router.HandleFunc("POST /api/v1/packlists", handler.CreatePacklist)
	router.HandleFunc("GET /api/v1/packlists", handler.ListPacklists)
	router.HandleFunc("GET /api/v1/packlists/{id}", handler.GetPacklist)
	router.HandleFunc("POST /api/v1/packlists/{id}/items", handler.AddPacklistItem)
	router.HandleFunc("DELETE /api/v1/packlists/{id}/items", handler.RemovePacklistItem)
	router.HandleFunc("PATCH /api/v1/packlists/{id}/items/pack", handler.MarkItemPacked)
	router.HandleFunc("PATCH /api/v1/packlists/{id}/items/return", handler.MarkItemReturned)
	router.HandleFunc("PATCH /api/v1/packlists/{id}/status", handler.ChangePacklistStatus)
	router.HandleFunc("DELETE /api/v1/packlists/{id}", handler.DeletePacklist)

	// Reservation routes
	router.HandleFunc("POST /api/v1/reservations", handler.CreateReservation)
	router.HandleFunc("GET /api/v1/reservations", handler.ListReservations)
	router.HandleFunc("GET /api/v1/reservations/{id}", handler.GetReservation)
	router.HandleFunc("GET /api/v1/reservations/conflicts", handler.CheckConflicts)
	router.HandleFunc("PATCH /api/v1/reservations/{id}/confirm", handler.ConfirmReservation)
	router.HandleFunc("PATCH /api/v1/reservations/{id}/cancel", handler.CancelReservation)
	router.HandleFunc("DELETE /api/v1/reservations/{id}", handler.DeleteReservation)

	return router
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"project-service"}`))
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready","service":"project-service"}`))
}
