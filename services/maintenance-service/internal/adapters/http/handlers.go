package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/maintenance-service/internal/application"
	"github.com/jeckersberger/rentflow/services/maintenance-service/internal/domain"
)

type Handler struct {
	recordSvc   *application.MaintenanceRecordService
	scheduleSvc *application.MaintenanceScheduleService
	logger      logger.Logger
}

func NewHandler(
	recordSvc *application.MaintenanceRecordService,
	scheduleSvc *application.MaintenanceScheduleService,
	log logger.Logger,
) *Handler {
	return &Handler{
		recordSvc:   recordSvc,
		scheduleSvc: scheduleSvc,
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
	if err == domain.ErrRecordNotFound || err == domain.ErrScheduleNotFound {
		h.respondError(w, http.StatusNotFound, err.Error())
		return
	}
	if err == domain.ErrInvalidInput || err == domain.ErrTenantIDRequired {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.logger.Error("Unhandled error", err)
	h.respondError(w, http.StatusInternalServerError, "internal server error")
}
