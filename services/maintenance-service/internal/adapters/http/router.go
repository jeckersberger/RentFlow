package http

import (
	"net/http"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/maintenance-service/internal/application"
)

func NewRouter(
	recordSvc *application.MaintenanceRecordService,
	scheduleSvc *application.MaintenanceScheduleService,
	planSvc *application.MaintenancePlanService,
	taskSvc *application.MaintenanceTaskService,
	checklistSvc *application.ChecklistService,
	testSvc *application.ElectricalTestService,
	log logger.Logger,
) *http.ServeMux {
	router := http.NewServeMux()
	handler := NewHandler(recordSvc, scheduleSvc, planSvc, taskSvc, checklistSvc, testSvc, log)

	router.HandleFunc("GET /health", healthHandler)
	router.HandleFunc("GET /ready", readyHandler)

	// Maintenance Records (legacy)
	router.HandleFunc("POST /api/v1/maintenance", handler.CreateRecord)
	router.HandleFunc("GET /api/v1/maintenance", handler.ListRecords)
	router.HandleFunc("GET /api/v1/maintenance/{id}", handler.GetRecord)
	router.HandleFunc("PUT /api/v1/maintenance/{id}", handler.UpdateRecord)
	router.HandleFunc("POST /api/v1/maintenance/{id}/complete", handler.CompleteRecord)
	router.HandleFunc("DELETE /api/v1/maintenance/{id}", handler.DeleteRecord)

	// Maintenance Schedules (legacy)
	router.HandleFunc("GET /api/v1/maintenance/schedule", handler.ListSchedules)
	router.HandleFunc("POST /api/v1/maintenance/schedule", handler.CreateSchedule)
	router.HandleFunc("GET /api/v1/maintenance/equipment/{id}", handler.GetScheduleByEquipment)
	router.HandleFunc("GET /api/v1/maintenance/overdue", handler.ListOverdue)

	// Maintenance Plans
	router.HandleFunc("POST /api/v1/maintenance/plans", handler.CreatePlan)
	router.HandleFunc("GET /api/v1/maintenance/plans", handler.ListPlans)
	router.HandleFunc("GET /api/v1/maintenance/plans/{id}", handler.GetPlan)
	router.HandleFunc("PUT /api/v1/maintenance/plans/{id}", handler.UpdatePlan)
	router.HandleFunc("DELETE /api/v1/maintenance/plans/{id}", handler.DeletePlan)

	// Maintenance Tasks
	router.HandleFunc("POST /api/v1/maintenance/tasks", handler.CreateTask)
	router.HandleFunc("GET /api/v1/maintenance/tasks", handler.ListTasks)
	router.HandleFunc("GET /api/v1/maintenance/tasks/{id}", handler.GetTask)
	router.HandleFunc("PUT /api/v1/maintenance/tasks/{id}/start", handler.StartTask)
	router.HandleFunc("PUT /api/v1/maintenance/tasks/{id}/complete", handler.CompleteTask)
	router.HandleFunc("GET /api/v1/maintenance/tasks/due", handler.GetDueTasks)
	router.HandleFunc("GET /api/v1/maintenance/tasks/overdue", handler.GetOverdueTasks)

	// Checklists
	router.HandleFunc("POST /api/v1/maintenance/checklists", handler.CreateChecklist)
	router.HandleFunc("GET /api/v1/maintenance/checklists", handler.ListChecklists)
	router.HandleFunc("GET /api/v1/maintenance/checklists/{id}", handler.GetChecklist)

	// Electrical Tests
	router.HandleFunc("POST /api/v1/maintenance/electrical-tests", handler.RecordTest)
	router.HandleFunc("GET /api/v1/maintenance/electrical-tests", handler.GetTestsByEquipment)

	// IZYTRON Import
	router.HandleFunc("POST /api/v1/maintenance/import/izytron", handler.ImportIzytron)

	// Dashboard
	router.HandleFunc("GET /api/v1/maintenance/dashboard", handler.GetDashboard)

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
