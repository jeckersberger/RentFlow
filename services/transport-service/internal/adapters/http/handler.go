package http

import (
	"encoding/json"
	"net/http"
	"time"

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

func (h *Handler) ListToursByDate(w http.ResponseWriter, r *http.Request) {
	dateStr := r.PathValue("date")
	date, _ := time.Parse("2006-01-02", dateStr)
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	dtos, err := h.tourSvc.ListToursByDate(r.Context(), tenantID, date)
	if err != nil {
		h.handleError(w, err)
		return
	}
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"data": dtos})
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
	if err == domain.ErrVehicleNotFound || err == domain.ErrTourNotFound {
		h.respondError(w, http.StatusNotFound, err.Error())
		return
	}
	if err == domain.ErrInvalidInput {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.logger.Error("Unhandled error", err)
	h.respondError(w, http.StatusInternalServerError, "internal server error")
}
