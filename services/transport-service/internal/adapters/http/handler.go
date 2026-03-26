package http

import (
	"encoding/json"
	"net/http"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/transport-service/internal/application"
	"github.com/jeckersberger/rentflow/services/transport-service/internal/domain"
)

type Handler struct {
	vehicleSvc *application.VehicleService
	tourSvc    *application.TourService
	logger     logger.Logger
}

func NewHandler(vehicleSvc *application.VehicleService, tourSvc *application.TourService, log logger.Logger) *Handler {
	return &Handler{vehicleSvc: vehicleSvc, tourSvc: tourSvc, logger: log}
}

// Vehicles
func (h *Handler) CreateVehicle(w http.ResponseWriter, r *http.Request) {
	var cmd application.CreateVehicleCommand
	json.NewDecoder(r.Body).Decode(&cmd)
	cmd.TenantID = r.Header.Get("X-Tenant-ID")
	if cmd.TenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	dto, err := h.vehicleSvc.CreateVehicle(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}
	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) ListVehicles(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	dtos, err := h.vehicleSvc.ListVehicles(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": dtos})
}

func (h *Handler) GetVehicle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	dto, err := h.vehicleSvc.GetVehicle(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}
	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) UpdateVehicle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var cmd application.UpdateVehicleCommand
	json.NewDecoder(r.Body).Decode(&cmd)
	cmd.TenantID = r.Header.Get("X-Tenant-ID")
	cmd.VehicleID = id
	if cmd.TenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	dto, err := h.vehicleSvc.UpdateVehicle(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}
	h.respondJSON(w, http.StatusOK, dto)
}

// Tours
func (h *Handler) CreateTour(w http.ResponseWriter, r *http.Request) {
	var cmd application.CreateTourCommand
	json.NewDecoder(r.Body).Decode(&cmd)
	cmd.TenantID = r.Header.Get("X-Tenant-ID")
	if cmd.TenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	dto, err := h.tourSvc.CreateTour(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}
	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) ListTours(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	dtos, err := h.tourSvc.ListTours(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": dtos})
}

func (h *Handler) GetTour(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	dto, err := h.tourSvc.GetTour(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}
	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) UpdateTourStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var cmd application.UpdateTourStatusCommand
	json.NewDecoder(r.Body).Decode(&cmd)
	cmd.TenantID = r.Header.Get("X-Tenant-ID")
	cmd.TourID = id
	if cmd.TenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	dto, err := h.tourSvc.UpdateTourStatus(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}
	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) StartTour(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var cmd application.StartTourCommand
	json.NewDecoder(r.Body).Decode(&cmd)
	cmd.TenantID = r.Header.Get("X-Tenant-ID")
	cmd.TourID = id
	if cmd.TenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	dto, err := h.tourSvc.StartTour(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}
	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) CompleteTour(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var cmd application.CompleteTourCommand
	json.NewDecoder(r.Body).Decode(&cmd)
	cmd.TenantID = r.Header.Get("X-Tenant-ID")
	cmd.TourID = id
	if cmd.TenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	dto, err := h.tourSvc.CompleteTour(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}
	h.respondJSON(w, http.StatusOK, dto)
}

// Tour Equipment
func (h *Handler) AddEquipmentToTour(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var cmd application.AddEquipmentToTourCommand
	json.NewDecoder(r.Body).Decode(&cmd)
	cmd.TenantID = r.Header.Get("X-Tenant-ID")
	cmd.TourID = id
	if cmd.TenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	dto, err := h.tourSvc.AddEquipmentToTour(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}
	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) RemoveEquipmentFromTour(w http.ResponseWriter, r *http.Request) {
	tourID := r.PathValue("id")
	equipmentID := r.PathValue("equipmentId")
	cmd := application.RemoveEquipmentFromTourCommand{
		TenantID:    r.Header.Get("X-Tenant-ID"),
		TourID:      tourID,
		EquipmentID: equipmentID,
	}
	if cmd.TenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	err := h.tourSvc.RemoveEquipmentFromTour(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}
	h.respondJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

// Tour Capacity
func (h *Handler) GetTourCapacity(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	dto, err := h.tourSvc.GetTourCapacity(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}
	h.respondJSON(w, http.StatusOK, dto)
}

// Driver Logs
func (h *Handler) LogDriverActivity(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var cmd application.LogDriverActivityCommand
	json.NewDecoder(r.Body).Decode(&cmd)
	cmd.TenantID = r.Header.Get("X-Tenant-ID")
	cmd.TourID = id
	if cmd.TenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	dto, err := h.tourSvc.LogDriverActivity(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}
	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) GetDriverLogs(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	dtos, err := h.tourSvc.GetDriverLogs(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": dtos})
}

// Response helpers
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
	if err == domain.ErrVehicleNotFound || err == domain.ErrTourNotFound || err == domain.ErrTourEquipmentNotFound || err == domain.ErrDriverLogNotFound {
		h.respondError(w, http.StatusNotFound, err.Error())
		return
	}
	if err == domain.ErrInvalidInput {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err == domain.ErrCapacityExceeded {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.logger.Error("Unhandled error", err)
	h.respondError(w, http.StatusInternalServerError, "internal server error")
}
