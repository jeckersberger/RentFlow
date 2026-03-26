package http

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/pkg/common/middleware"
	"github.com/jeckersberger/rentflow/services/scanner-service/internal/application"
	"github.com/jeckersberger/rentflow/services/scanner-service/internal/domain"
)

// getTenantID extracts tenant ID from JWT claims (set by middleware), falls back to X-Tenant-ID header
func (h *Handler) getTenantID(r *http.Request) string {
	if tid := middleware.GetTenantID(r.Context()); tid != "" {
		return tid
	}
	return r.Header.Get("X-Tenant-ID")
}

type Handler struct {
	scanSvc    *application.ScanService
	sessionSvc *application.SessionService
	logger     logger.Logger
}

func NewHandler(scanSvc *application.ScanService, sessionSvc *application.SessionService, logger logger.Logger) *Handler {
	return &Handler{
		scanSvc:    scanSvc,
		sessionSvc: sessionSvc,
		logger:     logger,
	}
}

func (h *Handler) ProcessScan(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		Barcode    string   `json:"barcode"`
		ScanType   string   `json:"scan_type"`
		UserID     string   `json:"user_id"`
		DeviceID   string   `json:"device_id"`
		DeviceType string   `json:"device_type"`
		ProjectID  *string  `json:"project_id"`
		LocationID *string  `json:"location_id"`
		Latitude   *float64 `json:"latitude"`
		Longitude  *float64 `json:"longitude"`
		Notes      string   `json:"notes"`
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
	tenantID := h.getTenantID(r)
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		Scans []struct {
			Barcode    string   `json:"barcode"`
			ScanType   string   `json:"scan_type"`
			UserID     string   `json:"user_id"`
			DeviceID   string   `json:"device_id"`
			DeviceType string   `json:"device_type"`
			ProjectID  *string  `json:"project_id"`
			LocationID *string  `json:"location_id"`
			Latitude   *float64 `json:"latitude"`
			Longitude  *float64 `json:"longitude"`
			Notes      string   `json:"notes"`
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
	tenantID := h.getTenantID(r)
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		Scans []struct {
			Barcode    string   `json:"barcode"`
			ScanType   string   `json:"scan_type"`
			UserID     string   `json:"user_id"`
			DeviceID   string   `json:"device_id"`
			DeviceType string   `json:"device_type"`
			ProjectID  *string  `json:"project_id"`
			LocationID *string  `json:"location_id"`
			Latitude   *float64 `json:"latitude"`
			Longitude  *float64 `json:"longitude"`
			Notes      string   `json:"notes"`
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
	tenantID := h.getTenantID(r)
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
	tenantID := h.getTenantID(r)
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
	tenantID := h.getTenantID(r)
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
	tenantID := h.getTenantID(r)
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

// =====================================================
// Scanner Device Management handlers ("Find My Scanner")
// =====================================================

// RegisterScannerDevice handles POST /api/v1/scanner/devices/register
func (h *Handler) RegisterScannerDevice(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		DeviceID   string `json:"device_id"`
		DeviceName string `json:"device_name"`
		FCMToken   string `json:"fcm_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if payload.DeviceID == "" {
		h.respondError(w, http.StatusBadRequest, "device_id is required")
		return
	}
	if payload.DeviceName == "" {
		h.respondError(w, http.StatusBadRequest, "device_name is required")
		return
	}

	device, err := h.scanSvc.RegisterScannerDevice(r.Context(), tenantID, payload.DeviceID, payload.DeviceName, payload.FCMToken)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, device)
}

// ListScannerDevices handles GET /api/v1/scanner/devices
func (h *Handler) ListScannerDevices(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	devices, err := h.scanSvc.ListScannerDevices(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"devices": devices,
	})
}

// RingScannerDevice handles POST /api/v1/scanner/devices/{id}/ring
func (h *Handler) RingScannerDevice(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.respondError(w, http.StatusBadRequest, "device id required")
		return
	}

	if err := h.scanSvc.RingScannerDevice(r.Context(), id); err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "ring requested",
	})
}

// CheckRingRequest handles GET /api/v1/scanner/devices/{device_id}/ring
func (h *Handler) CheckRingRequest(w http.ResponseWriter, r *http.Request) {
	deviceID := r.PathValue("device_id")
	tenantID := h.getTenantID(r)
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	if deviceID == "" {
		h.respondError(w, http.StatusBadRequest, "device_id required")
		return
	}

	ringRequested, err := h.scanSvc.CheckRingRequest(r.Context(), tenantID, deviceID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"ring_requested": ringRequested,
	})
}

// AckRing handles POST /api/v1/scanner/devices/{device_id}/ring-ack
func (h *Handler) AckRing(w http.ResponseWriter, r *http.Request) {
	deviceID := r.PathValue("device_id")
	tenantID := h.getTenantID(r)
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	if deviceID == "" {
		h.respondError(w, http.StatusBadRequest, "device_id required")
		return
	}

	if err := h.scanSvc.AckRing(r.Context(), tenantID, deviceID); err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "ring acknowledged",
	})
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
		case "CHECKOUT_ERROR", "CHECKIN_ERROR":
			h.respondError(w, http.StatusUnprocessableEntity, domainErr.Message)
		case "QUEUE_FULL":
			h.respondError(w, http.StatusTooManyRequests, domainErr.Message)
		default:
			h.respondError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	h.respondError(w, http.StatusInternalServerError, "internal server error")
}

// =====================================================
// Scanner App API contract handlers
// =====================================================

// ScannerScan handles POST /api/v1/scanner/scan
// Resolves a barcode/RFID to full equipment detail.
func (h *Handler) ScannerScan(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		Code      string `json:"code"`
		ScanType  string `json:"scan_type"`
		Timestamp string `json:"timestamp"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if payload.Code == "" {
		h.respondError(w, http.StatusBadRequest, "code is required")
		return
	}
	if payload.ScanType == "" {
		payload.ScanType = "barcode"
	}

	equipment, err := h.scanSvc.ResolveEquipment(r.Context(), tenantID, payload.Code, payload.ScanType)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"equipment": equipment,
	})
}

// ScannerCheckout handles POST /api/v1/scanner/checkout
// Checks out equipment to a project.
func (h *Handler) ScannerCheckout(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		EquipmentIDs []string `json:"equipment_ids"`
		ProjectID    string   `json:"project_id"`
		Notes        string   `json:"notes"`
		Timestamp    string   `json:"timestamp"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(payload.EquipmentIDs) == 0 {
		h.respondError(w, http.StatusBadRequest, "equipment_ids is required")
		return
	}
	if payload.ProjectID == "" {
		h.respondError(w, http.StatusBadRequest, "project_id is required")
		return
	}

	result, err := h.scanSvc.CheckOutEquipment(r.Context(), tenantID, payload.EquipmentIDs, payload.ProjectID, payload.Notes)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":     true,
		"checked_out": result.CheckedOut,
		"project":     result.Project,
	})
}

// ScannerCheckin handles POST /api/v1/scanner/checkin
// Checks in equipment with condition ratings.
func (h *Handler) ScannerCheckin(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		EquipmentIDs     []string                            `json:"equipment_ids"`
		ConditionRatings map[string]application.ConditionRatingPayload `json:"condition_ratings"`
		Timestamp        string                              `json:"timestamp"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(payload.EquipmentIDs) == 0 {
		h.respondError(w, http.StatusBadRequest, "equipment_ids is required")
		return
	}

	result, err := h.scanSvc.CheckInEquipment(r.Context(), tenantID, payload.EquipmentIDs, payload.ConditionRatings)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":        true,
		"checked_in":     result.CheckedIn,
		"damage_reports": result.DamageReports,
	})
}

// ScannerBulk handles POST /api/v1/scanner/bulk
// Processes batch offline actions idempotently.
func (h *Handler) ScannerBulk(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		Actions []application.BulkAction `json:"actions"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(payload.Actions) == 0 {
		h.respondError(w, http.StatusBadRequest, "actions array is required")
		return
	}

	result, err := h.scanSvc.ProcessBulkActions(r.Context(), tenantID, payload.Actions)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

// AdhocBooking handles POST /api/v1/scanner/adhoc-booking
// Creates a reservation on-the-fly from the scanner app.
func (h *Handler) AdhocBooking(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		EquipmentID string `json:"equipment_id"`
		ProjectID   string `json:"project_id"`
		Notes       string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if payload.EquipmentID == "" {
		h.respondError(w, http.StatusBadRequest, "equipment_id is required")
		return
	}
	if payload.ProjectID == "" {
		h.respondError(w, http.StatusBadRequest, "project_id is required")
		return
	}

	cmd := application.AdhocBookingCommand{
		TenantID:    tenantID,
		EquipmentID: payload.EquipmentID,
		ProjectID:   payload.ProjectID,
		Notes:       payload.Notes,
	}

	result, err := h.scanSvc.AdhocBooking(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, result)
}

// Session handlers

func (h *Handler) StartSession(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		UserID   string `json:"user_id"`
		Context  string `json:"context"`
		ProjectID *string `json:"project_id"`
		DeviceType string `json:"device_type"`
		DeviceID string `json:"device_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.StartScanSessionCommand{
		TenantID:   tenantID,
		UserID:     payload.UserID,
		Context:    payload.Context,
		ProjectID:  payload.ProjectID,
		DeviceType: domain.DeviceType(payload.DeviceType),
		DeviceID:   payload.DeviceID,
	}

	dto, err := h.sessionSvc.StartSession(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) EndSession(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	sessionID := r.PathValue("id")
	if sessionID == "" {
		h.respondError(w, http.StatusBadRequest, "session ID required")
		return
	}

	cmd := application.EndScanSessionCommand{
		TenantID:  tenantID,
		SessionID: sessionID,
	}

	dto, err := h.sessionSvc.EndSession(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) ProcessSessionScan(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	sessionID := r.PathValue("id")
	if sessionID == "" {
		h.respondError(w, http.StatusBadRequest, "session ID required")
		return
	}

	var payload struct {
		Barcode    string   `json:"barcode"`
		DeviceID   string   `json:"device_id"`
		DeviceType string   `json:"device_type"`
		Latitude   *float64 `json:"latitude"`
		Longitude  *float64 `json:"longitude"`
		Notes      string   `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.ProcessSessionScanCommand{
		SessionID:   sessionID,
		TenantID:    tenantID,
		Barcode:     payload.Barcode,
		DeviceID:    payload.DeviceID,
		DeviceType:  domain.DeviceType(payload.DeviceType),
		Latitude:    payload.Latitude,
		Longitude:   payload.Longitude,
		Notes:       payload.Notes,
	}

	dto, err := h.sessionSvc.ProcessSessionScan(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) GetSessionProtocol(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	sessionID := r.PathValue("id")
	if sessionID == "" {
		h.respondError(w, http.StatusBadRequest, "session ID required")
		return
	}

	dtos, err := h.sessionSvc.GetSessionProtocol(r.Context(), tenantID, sessionID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"count": len(dtos),
		"scans": dtos,
	})
}

func (h *Handler) SyncOfflineQueue(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 500 {
			limit = parsed
		}
	}

	result, err := h.sessionSvc.SyncOfflineQueue(r.Context(), tenantID, limit)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

func (h *Handler) QueueOfflineScan(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		Barcode    string   `json:"barcode"`
		ScanType   string   `json:"scan_type"`
		UserID     string   `json:"user_id"`
		DeviceID   string   `json:"device_id"`
		DeviceType string   `json:"device_type"`
		ProjectID  *string  `json:"project_id"`
		LocationID *string  `json:"location_id"`
		Latitude   *float64 `json:"latitude"`
		Longitude  *float64 `json:"longitude"`
		Notes      string   `json:"notes"`
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

	dto, err := h.sessionSvc.QueueOfflineScan(r.Context(), tenantID, payload.DeviceID, cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) UploadSignature(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	sessionID := r.PathValue("id")
	if sessionID == "" {
		h.respondError(w, http.StatusBadRequest, "session ID required")
		return
	}

	var payload struct {
		Signature string `json:"signature"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 5*1024*1024)).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if payload.Signature == "" {
		h.respondError(w, http.StatusBadRequest, "signature data required")
		return
	}

	if err := h.sessionSvc.UploadSignature(r.Context(), tenantID, sessionID, payload.Signature); err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data":    nil,
		"message": "Signature uploaded",
	})
}
