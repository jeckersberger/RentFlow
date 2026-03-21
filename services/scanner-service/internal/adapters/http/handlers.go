package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/scanner-service/internal/application"
	"github.com/jeckersberger/rentflow/services/scanner-service/internal/domain"
)

type Handler struct {
	scanSvc *application.ScanService
	logger  *logger.Logger
}

func NewHandler(scanSvc *application.ScanService, logger *logger.Logger) *Handler {
	return &Handler{
		scanSvc: scanSvc,
		logger:  logger,
	}
}

func (h *Handler) ProcessScan(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		Barcode    string `json:"barcode"`
		ScanType   string `json:"scan_type"`
		UserID     string `json:"user_id"`
		DeviceID   string `json:"device_id"`
		DeviceType string `json:"device_type"`
		ProjectID  *string `json:"project_id"`
		LocationID *string `json:"location_id"`
		Latitude   *float64 `json:"latitude"`
		Longitude  *float64 `json:"longitude"`
		Notes      string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.ProcessScanCommand{
		TenantID:   tenantID,
		Barcode:    payload.Barcode,
		ScanType:   domain.ScanType(payload.ScanType),
		UserID:     payload.UserID,
		DeviceID:   payload.DeviceID,
		DeviceType: domain.DeviceType(payload.DeviceType),
		ProjectID:  payload.ProjectID,
		LocationID: payload.LocationID,
		Latitude:   payload.Latitude,
		Longitude:  payload.Longitude,
		Notes:      payload.Notes,
	}

	dto, err := h.scanSvc.ProcessScan(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) ProcessBatch(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		Scans []struct {
			Barcode    string `json:"barcode"`
			ScanType   string `json:"scan_type"`
			UserID     string `json:"user_id"`
			DeviceID   string `json:"device_id"`
			DeviceType string `json:"device_type"`
			ProjectID  *string `json:"project_id"`
			LocationID *string `json:"location_id"`
			Latitude   *float64 `json:"latitude"`
			Longitude  *float64 `json:"longitude"`
			Notes      string `json:"notes"`
		} `json:"scans"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	scans := make([]application.ProcessScanCommand, len(payload.Scans))
	for i, s := range payload.Scans {
		scans[i] = application.ProcessScanCommand{
			Barcode:    s.Barcode,
			ScanType:   domain.ScanType(s.ScanType),
			UserID:     s.UserID,
			DeviceID:   s.DeviceID,
			DeviceType: domain.DeviceType(s.DeviceType),
			ProjectID:  s.ProjectID,
			LocationID: s.LocationID,
			Latitude:   s.Latitude,
			Longitude:  s.Longitude,
			Notes:      s.Notes,
		}
	}

	cmd := application.BatchScanCommand{
		TenantID: tenantID,
		Scans:    scans,
	}

	dtos, err := h.scanSvc.ProcessBatch(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, map[string]interface{}{
		"count": len(dtos),
		"scans": dtos,
	})
}

func (h *Handler) SyncOfflineScans(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		Scans []struct {
			Barcode    string `json:"barcode"`
			ScanType   string `json:"scan_type"`
			UserID     string `json:"user_id"`
			DeviceID   string `json:"device_id"`
			DeviceType string `json:"device_type"`
			ProjectID  *string `json:"project_id"`
			LocationID *string `json:"location_id"`
			Latitude   *float64 `json:"latitude"`
			Longitude  *float64 `json:"longitude"`
			Notes      string `json:"notes"`
		} `json:"scans"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	scans := make([]application.ProcessScanCommand, len(payload.Scans))
	for i, s := range payload.Scans {
		scans[i] = application.ProcessScanCommand{
			Barcode:    s.Barcode,
			ScanType:   domain.ScanType(s.ScanType),
			UserID:     s.UserID,
			DeviceID:   s.DeviceID,
			DeviceType: domain.DeviceType(s.DeviceType),
			ProjectID:  s.ProjectID,
			LocationID: s.LocationID,
			Latitude:   s.Latitude,
			Longitude:  s.Longitude,
			Notes:      s.Notes,
		}
	}

	cmd := application.SyncOfflineCommand{
		TenantID: tenantID,
		Scans:    scans,
	}

	dtos, err := h.scanSvc.SyncOfflineScans(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"count": len(dtos),
		"scans": dtos,
	})
}

func (h *Handler) GetHistory(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	limit := 20
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	query := application.ScanHistoryQuery{
		TenantID: tenantID,
		Limit:    limit,
		Offset:   offset,
	}

	if eq := r.URL.Query().Get("equipment_id"); eq != "" {
		query.EquipmentID = &eq
	}
	if proj := r.URL.Query().Get("project_id"); proj != "" {
		query.ProjectID = &proj
	}
	if usr := r.URL.Query().Get("user_id"); usr != "" {
		query.UserID = &usr
	}
	if dev := r.URL.Query().Get("device_id"); dev != "" {
		query.DeviceID = &dev
	}

	result, err := h.scanSvc.GetHistory(r.Context(), query)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

func (h *Handler) ResolveBarcode(w http.ResponseWriter, r *http.Request) {
	barcode := r.PathValue("barcode")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dto, err := h.scanSvc.ResolveBarcode(r.Context(), tenantID, barcode)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) ListDevices(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	limit := 20
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	result, err := h.scanSvc.ListDevices(r.Context(), tenantID, limit, offset)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

func (h *Handler) RegisterDevice(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		Name     string `json:"name"`
		Type     string `json:"type"`
		Serial   string `json:"serial"`
		Location string `json:"location"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.RegisterDeviceCommand{
		TenantID: tenantID,
		Name:     payload.Name,
		Type:     domain.DeviceType(payload.Type),
		Serial:   payload.Serial,
		Location: payload.Location,
	}

	dto, err := h.scanSvc.RegisterDevice(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

// Helper methods

func (h *Handler) respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) respondError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	if domainErr, ok := err.(*domain.DomainError); ok {
		switch domainErr.Code {
		case "NOT_FOUND":
			h.respondError(w, http.StatusNotFound, domainErr.Message)
		case "VALIDATION_ERROR", "INVALID_INPUT":
			h.respondError(w, http.StatusBadRequest, domainErr.Message)
		case "TENANT_REQUIRED", "BARCODE_REQUIRED", "NAME_REQUIRED":
			h.respondError(w, http.StatusBadRequest, domainErr.Message)
		case "UNAUTHORIZED":
			h.respondError(w, http.StatusUnauthorized, domainErr.Message)
		case "DEVICE_EXISTS":
			h.respondError(w, http.StatusConflict, domainErr.Message)
		case "SERVICE_UNAVAILABLE":
			h.respondError(w, http.StatusServiceUnavailable, domainErr.Message)
		default:
			h.respondError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	h.respondError(w, http.StatusInternalServerError, "internal server error")
}
