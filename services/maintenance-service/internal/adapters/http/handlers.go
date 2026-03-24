package http

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/maintenance-service/internal/application"
	"github.com/jeckersberger/rentflow/services/maintenance-service/internal/domain"
)

type Handler struct {
	recordSvc   *application.MaintenanceRecordService
	scheduleSvc *application.MaintenanceScheduleService
	planSvc     *application.MaintenancePlanService
	taskSvc     *application.MaintenanceTaskService
	checklistSvc *application.ChecklistService
	testSvc     *application.ElectricalTestService
	logger      logger.Logger
}

func NewHandler(
	recordSvc *application.MaintenanceRecordService,
	scheduleSvc *application.MaintenanceScheduleService,
	planSvc *application.MaintenancePlanService,
	taskSvc *application.MaintenanceTaskService,
	checklistSvc *application.ChecklistService,
	testSvc *application.ElectricalTestService,
	log logger.Logger,
) *Handler {
	return &Handler{
		recordSvc:   recordSvc,
		scheduleSvc: scheduleSvc,
		planSvc:     planSvc,
		taskSvc:     taskSvc,
		checklistSvc: checklistSvc,
		testSvc:     testSvc,
		logger:      log,
	}
}

func (h *Handler) CreateRecord(w http.ResponseWriter, r *http.Request) {
	var cmd application.CreateMaintenanceRecordCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	cmd.TenantID = tenantID

	dto, err := h.recordSvc.CreateRecord(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) GetRecord(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dto, err := h.recordSvc.GetRecord(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) ListRecords(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dtos, err := h.recordSvc.ListRecords(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": dtos})
}

func (h *Handler) UpdateRecord(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dto, err := h.recordSvc.GetRecord(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) CompleteRecord(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var cmd application.CompleteMaintenanceCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	cmd.TenantID = tenantID
	cmd.RecordID = id
	if cmd.CompletedDate.IsZero() {
		cmd.CompletedDate = time.Now()
	}

	dto, err := h.recordSvc.CompleteRecord(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) DeleteRecord(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	if err := h.recordSvc.DeleteRecord(r.Context(), tenantID, id); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	var cmd application.CreateScheduleCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	cmd.TenantID = tenantID

	dto, err := h.scheduleSvc.CreateSchedule(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) ListSchedules(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dtos, err := h.scheduleSvc.ListSchedules(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": dtos})
}

func (h *Handler) GetScheduleByEquipment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dto, err := h.scheduleSvc.GetScheduleByEquipment(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) ListOverdue(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dtos, err := h.recordSvc.ListOverdue(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": dtos})
}

func (h *Handler) ImportDGUV(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, http.StatusOK, map[string]string{"status": "import_queued"})
}

func (h *Handler) GetDGUVStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dto, err := h.recordSvc.ListByEquipment(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"records": dto})
}

func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	if err == domain.ErrRecordNotFound || err == domain.ErrScheduleNotFound ||
		err == domain.ErrPlanNotFound || err == domain.ErrTaskNotFound ||
		err == domain.ErrChecklistNotFound || err == domain.ErrTestNotFound {
		h.respondError(w, http.StatusNotFound, err.Error())
		return
	}
	if err == domain.ErrInvalidInput || err == domain.ErrTenantIDRequired ||
		err == domain.ErrInvalidTaskStatus {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.logger.Error("Unhandled error", err)
	h.respondError(w, http.StatusInternalServerError, "internal server error")
}

// Maintenance Plans
func (h *Handler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	var cmd application.CreateMaintenancePlanCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	cmd.TenantID = tenantID

	dto, err := h.planSvc.CreatePlan(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) GetPlan(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dto, err := h.planSvc.GetPlan(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) ListPlans(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dtos, err := h.planSvc.ListPlans(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": dtos})
}

func (h *Handler) UpdatePlan(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var cmd application.CreateMaintenancePlanCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dto, err := h.planSvc.UpdatePlan(r.Context(), tenantID, id, cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) DeletePlan(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	if err := h.planSvc.DeletePlan(r.Context(), tenantID, id); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Maintenance Tasks
func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var cmd application.CreateMaintenanceTaskCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	cmd.TenantID = tenantID

	dto, err := h.taskSvc.CreateTask(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dto, err := h.taskSvc.GetTask(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dtos, err := h.taskSvc.ListTasks(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": dtos})
}

func (h *Handler) StartTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	cmd := application.StartMaintenanceTaskCommand{
		TenantID: tenantID,
		TaskID:   id,
	}

	dto, err := h.taskSvc.StartTask(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) CompleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var cmd application.CompleteMaintenanceTaskCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	cmd.TenantID = tenantID
	cmd.TaskID = id

	dto, err := h.taskSvc.CompleteTask(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) GetDueTasks(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dtos, err := h.taskSvc.GetDueTasks(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": dtos})
}

func (h *Handler) GetOverdueTasks(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dtos, err := h.taskSvc.GetOverdueTasks(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": dtos})
}

// Checklists
func (h *Handler) CreateChecklist(w http.ResponseWriter, r *http.Request) {
	var cmd application.CreateChecklistCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	cmd.TenantID = tenantID

	dto, err := h.checklistSvc.CreateChecklist(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) GetChecklist(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dto, err := h.checklistSvc.GetChecklist(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) ListChecklists(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dtos, err := h.checklistSvc.ListChecklists(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": dtos})
}

// Electrical Tests
func (h *Handler) RecordTest(w http.ResponseWriter, r *http.Request) {
	var cmd application.RecordElectricalTestCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	cmd.TenantID = tenantID

	dto, err := h.testSvc.RecordTest(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) GetTestsByEquipment(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	// Support both path parameter and query parameter for equipment_id
	equipmentID := r.PathValue("id")
	if equipmentID == "" {
		equipmentID = r.URL.Query().Get("equipment_id")
	}

	// If no equipment ID provided, list all tests for the tenant
	if equipmentID == "" {
		dtos, err := h.testSvc.ListTestsByTenant(r.Context(), tenantID)
		if err != nil {
			h.handleError(w, err)
			return
		}
		h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": dtos})
		return
	}

	dtos, err := h.testSvc.ListTestsByEquipment(r.Context(), tenantID, equipmentID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": dtos})
}

func (h *Handler) ImportIzytron(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dtos, err := h.testSvc.ImportIzytronXML(r.Context(), tenantID, string(body))
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, map[string]interface{}{"data": dtos})
}

func (h *Handler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dueTasks, err := h.taskSvc.GetDueTasks(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	overdueTasks, err := h.taskSvc.GetOverdueTasks(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	duePlans, err := h.planSvc.GetDuePlans(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	recentTests, err := h.testSvc.ListTestsByTenant(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	dashboard := application.MaintenanceDashboardDTO{
		DueTasks:     dueTasks,
		OverdueTasks: overdueTasks,
		DuePlans:     duePlans,
		RecentTests:  recentTests,
	}

	h.respondJSON(w, http.StatusOK, dashboard)
}
