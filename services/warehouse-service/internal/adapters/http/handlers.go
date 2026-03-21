package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/warehouse-service/internal/application"
	"github.com/jeckersberger/rentflow/services/warehouse-service/internal/domain"
)

type Handler struct {
	locationSvc       *application.LocationService
	movementSvc       *application.MovementService
	inventoryCheckSvc *application.InventoryCheckService
	logger            logger.Logger
}

func NewHandler(
	locationSvc *application.LocationService,
	movementSvc *application.MovementService,
	inventoryCheckSvc *application.InventoryCheckService,
	logger logger.Logger,
) *Handler {
	return &Handler{
		locationSvc:       locationSvc,
		movementSvc:       movementSvc,
		inventoryCheckSvc: inventoryCheckSvc,
		logger:            logger,
	}
}

// Location Handlers

func (h *Handler) CreateLocation(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		Name      string  `json:"name"`
		Type      string  `json:"type"`
		ParentID  *string `json:"parent_id"`
		Capacity  int     `json:"capacity"`
		Barcode   string  `json:"barcode"`
		SortOrder int     `json:"sort_order"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.CreateLocationCommand{
		TenantID:  tenantID,
		Name:      payload.Name,
		Type:      domain.LocationType(payload.Type),
		ParentID:  payload.ParentID,
		Capacity:  payload.Capacity,
		Barcode:   payload.Barcode,
		SortOrder: payload.SortOrder,
	}

	dto, err := h.locationSvc.CreateLocation(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) GetLocation(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	locationID := r.PathValue("id")
	dto, err := h.locationSvc.GetLocation(r.Context(), tenantID, locationID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) ListLocations(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.locationSvc.ListLocations(r.Context(), tenantID, limit, offset)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

func (h *Handler) GetLocationTree(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	tree, err := h.locationSvc.GetLocationTree(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, tree)
}

func (h *Handler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	locationID := r.PathValue("id")

	var payload struct {
		Name      string `json:"name"`
		Capacity  int    `json:"capacity"`
		Barcode   string `json:"barcode"`
		SortOrder int    `json:"sort_order"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.UpdateLocationCommand{
		ID:        locationID,
		TenantID:  tenantID,
		Name:      payload.Name,
		Capacity:  payload.Capacity,
		Barcode:   payload.Barcode,
		SortOrder: payload.SortOrder,
	}

	dto, err := h.locationSvc.UpdateLocation(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) DeleteLocation(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	locationID := r.PathValue("id")
	if err := h.locationSvc.DeleteLocation(r.Context(), tenantID, locationID); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetLocationOccupancy(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	locationID := r.PathValue("id")
	occupancy, err := h.locationSvc.GetLocationOccupancy(r.Context(), tenantID, locationID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, occupancy)
}

// Movement Handlers

func (h *Handler) RecordMovement(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		EquipmentID    string  `json:"equipment_id"`
		FromLocationID *string `json:"from_location_id"`
		ToLocationID   string  `json:"to_location_id"`
		MovementType   string  `json:"movement_type"`
		Quantity       int     `json:"quantity"`
		Reason         string  `json:"reason"`
		UserID         string  `json:"user_id"`
		ProjectID      *string `json:"project_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.RecordMovementCommand{
		TenantID:       tenantID,
		EquipmentID:    payload.EquipmentID,
		FromLocationID: payload.FromLocationID,
		ToLocationID:   payload.ToLocationID,
		MovementType:   domain.MovementType(payload.MovementType),
		Quantity:       payload.Quantity,
		Reason:         payload.Reason,
		UserID:         payload.UserID,
		ProjectID:      payload.ProjectID,
	}

	dto, err := h.movementSvc.RecordMovement(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) GetMovementHistory(w http.ResponseWriter, r *http.Request) {
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

	query := application.HistoryQuery{
		Limit:  limit,
		Offset: offset,
	}

	if eq := r.URL.Query().Get("equipment_id"); eq != "" {
		query.EquipmentID = &eq
	}

	result, err := h.movementSvc.GetMovementHistory(r.Context(), tenantID, query)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

func (h *Handler) GetEquipmentHistory(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	equipmentID := r.PathValue("id")

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

	result, err := h.movementSvc.GetEquipmentHistory(r.Context(), tenantID, equipmentID, limit, offset)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

// Inventory Check Handlers

func (h *Handler) StartInventoryCheck(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		Name       string  `json:"name"`
		LocationID *string `json:"location_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.StartInventoryCheckCommand{
		TenantID:   tenantID,
		Name:       payload.Name,
		LocationID: payload.LocationID,
	}

	dto, err := h.inventoryCheckSvc.StartCheck(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) GetInventoryCheck(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	checkID := r.PathValue("id")
	dto, err := h.inventoryCheckSvc.GetCheck(r.Context(), tenantID, checkID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) ListInventoryChecks(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.inventoryCheckSvc.ListChecks(r.Context(), tenantID, limit, offset)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

func (h *Handler) ScanInventoryItem(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	checkID := r.PathValue("id")

	var payload struct {
		EquipmentID string `json:"equipment_id"`
		Notes       string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.ScanInventoryItemCommand{
		TenantID:    tenantID,
		CheckID:     checkID,
		EquipmentID: payload.EquipmentID,
		Notes:       payload.Notes,
	}

	dto, err := h.inventoryCheckSvc.ScanItem(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) CompleteInventoryCheck(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	checkID := r.PathValue("id")

	cmd := application.CompleteInventoryCheckCommand{
		TenantID: tenantID,
		CheckID:  checkID,
	}

	dto, err := h.inventoryCheckSvc.CompleteCheck(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) GetDiscrepancies(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	checkID := r.PathValue("id")
	result, err := h.inventoryCheckSvc.GetDiscrepancies(r.Context(), tenantID, checkID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, result)
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
		case "VALIDATION_ERROR", "INVALID_INPUT", "INVALID_STATE":
			h.respondError(w, http.StatusBadRequest, domainErr.Message)
		case "TENANT_REQUIRED", "NAME_REQUIRED", "EQUIPMENT_REQUIRED":
			h.respondError(w, http.StatusBadRequest, domainErr.Message)
		case "UNAUTHORIZED":
			h.respondError(w, http.StatusUnauthorized, domainErr.Message)
		case "LOCATION_NOT_FOUND", "FROM_LOCATION_NOT_FOUND":
			h.respondError(w, http.StatusNotFound, domainErr.Message)
		case "INSUFFICIENT_CAPACITY", "CAPACITY_ERROR":
			h.respondError(w, http.StatusConflict, domainErr.Message)
		case "INVALID_PARENT":
			h.respondError(w, http.StatusConflict, domainErr.Message)
		case "SCAN_ERROR", "COMPLETION_ERROR":
			h.respondError(w, http.StatusBadRequest, domainErr.Message)
		default:
			h.respondError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	h.respondError(w, http.StatusInternalServerError, "internal server error")
}
