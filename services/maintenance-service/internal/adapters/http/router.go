package http

import (
	"net/http"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/maintenance-service/internal/application"
)

func NewRouter(
	recordSvc *application.MaintenanceRecordService,
	scheduleSvc *application.MaintenanceScheduleService,
	log logger.Logger,
) *http.ServeMux {
	router := http.NewServeMux()
	handler := NewHandler(recordSvc, scheduleSvc, log)

	router.HandleFunc("GET /health", healthHandler)
	router.HandleFunc("GET /ready", readyHandler)

	// Maintenance Records
	router.HandleFunc("POST /api/v1/maintenance", handler.CreateRecord)
	router.HandleFunc("GET /api/v1/maintenance", handler.ListRecords)
	router.HandleFunc("GET /api/v1/maintenance/{id}", handler.GetRecord)
	router.HandleFunc("PUT /api/v1/maintenance/{id}", handler.UpdateRecord)
	router.HandleFunc("POST /api/v1/maintenance/{id}/complete", handler.CompleteRecord)
	router.HandleFunc("DELETE /api/v1/maintenance/{id}", handler.DeleteRecord)

	// Maintenance Schedules
	router.HandleFunc("GET /api/v1/maintenance/schedule", handler.ListSchedules)
	router.HandleFunc("POST /api/v1/maintenance/schedule", handler.CreateSchedule)
	router.HandleFunc("GET /api/v1/maintenance/equipment/{id}", handler.GetScheduleByEquipment)
	router.HandleFunc("GET /api/v1/maintenance/overdue", handler.ListOverdue)

	// DGUV
	router.HandleFunc("POST /api/v1/dguv-import", handler.ImportDGUV)
	router.HandleFunc("GET /api/v1/dguv/equipment/{id}", handler.GetDGUVStatus)

	return router
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"maintenance-service"}`))
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready","service":"maintenance-service"}`))
}
