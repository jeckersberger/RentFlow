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
	warehouseSvc      *application.WarehouseService
	zplSvc            *application.ZPLService
	logger            logger.Logger
}

func NewHandler(
	locationSvc *application.LocationService,
	movementSvc *application.MovementService,
	inventoryCheckSvc *application.InventoryCheckService,
	warehouseSvc *application.WarehouseService,
	zplSvc *application.ZPLService,
	logger logger.Logger,
) *Handler {
	return &Handler{
		locationSvc:       locationSvc,
		movementSvc:       movementSvc,
		inventoryCheckSvc: inventoryCheckSvc,
		warehouseSvc:      warehouseSvc,
		zplSvc:            zplSvc,
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
		ZoneID     *string `json:"zone_id"`
		CheckType  string  `json:"check_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.StartInventoryCheckCommand{
		TenantID:  tenantID,
		Name:      payload.Name,
		ZoneID:    payload.ZoneID,
		CheckType: payload.CheckType,
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
		EquipmentID string  `json:"equipment_id"`
		LocationID  *string `json:"location_id"`
		Notes       string  `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.ScanInventoryItemCommand{
		TenantID:    tenantID,
		CheckID:     checkID,
		EquipmentID: payload.EquipmentID,
		LocationID:  payload.LocationID,
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

	var payload struct {
		CompletedBy string `json:"completed_by"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.CompleteInventoryCheckCommand{
		TenantID:    tenantID,
		CheckID:     checkID,
		CompletedBy: payload.CompletedBy,
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

// Warehouse Handlers

func (h *Handler) CreateWarehouse(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		Name          string   `json:"name"`
		Code          string   `json:"code"`
		Address       string   `json:"address"`
		City          string   `json:"city"`
		PostalCode    string   `json:"postal_code"`
		Country       string   `json:"country"`
		Latitude      *float64 `json:"latitude"`
		Longitude     *float64 `json:"longitude"`
		TotalCapacity int      `json:"total_capacity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.CreateWarehouseCommand{
		TenantID:      tenantID,
		Name:          payload.Name,
		Code:          payload.Code,
		Address:       payload.Address,
		City:          payload.City,
		PostalCode:    payload.PostalCode,
		Country:       payload.Country,
		Latitude:      payload.Latitude,
		Longitude:     payload.Longitude,
		TotalCapacity: payload.TotalCapacity,
	}

	dto, err := h.warehouseSvc.CreateWarehouse(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) GetWarehouse(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	warehouseID := r.PathValue("id")
	dto, err := h.warehouseSvc.GetWarehouse(r.Context(), tenantID, warehouseID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) ListWarehouses(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.warehouseSvc.ListWarehouses(r.Context(), tenantID, limit, offset)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

func (h *Handler) CreateZone(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		WarehouseID    string   `json:"warehouse_id"`
		Name           string   `json:"name"`
		Code           string   `json:"code"`
		ZoneType       string   `json:"zone_type"`
		Description    string   `json:"description"`
		TemperatureMin *float64 `json:"temperature_min"`
		TemperatureMax *float64 `json:"temperature_max"`
		HumidityMin    *float64 `json:"humidity_min"`
		HumidityMax    *float64 `json:"humidity_max"`
		Capacity       int      `json:"capacity"`
		SortOrder      int      `json:"sort_order"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.CreateZoneCommand{
		TenantID:       tenantID,
		WarehouseID:    payload.WarehouseID,
		Name:           payload.Name,
		Code:           payload.Code,
		ZoneType:       payload.ZoneType,
		Description:    payload.Description,
		TemperatureMin: payload.TemperatureMin,
		TemperatureMax: payload.TemperatureMax,
		HumidityMin:    payload.HumidityMin,
		HumidityMax:    payload.HumidityMax,
		Capacity:       payload.Capacity,
		SortOrder:      payload.SortOrder,
	}

	dto, err := h.warehouseSvc.CreateZone(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) CreateRack(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		ZoneID         string   `json:"zone_id"`
		Name           string   `json:"name"`
		Code           string   `json:"code"`
		RackType       string   `json:"rack_type"`
		Aisle          string   `json:"aisle"`
		RowNumber      int      `json:"row_number"`
		ColumnNumber   int      `json:"column_number"`
		Capacity       int      `json:"capacity"`
		Height         *float64 `json:"height"`
		Width          *float64 `json:"width"`
		Depth          *float64 `json:"depth"`
		WeightCapacity *float64 `json:"weight_capacity"`
		SortOrder      int      `json:"sort_order"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.CreateRackCommand{
		TenantID:       tenantID,
		ZoneID:         payload.ZoneID,
		Name:           payload.Name,
		Code:           payload.Code,
		RackType:       payload.RackType,
		Aisle:          payload.Aisle,
		RowNumber:      payload.RowNumber,
		ColumnNumber:   payload.ColumnNumber,
		Capacity:       payload.Capacity,
		Height:         payload.Height,
		Width:          payload.Width,
		Depth:          payload.Depth,
		WeightCapacity: payload.WeightCapacity,
		SortOrder:      payload.SortOrder,
	}

	dto, err := h.warehouseSvc.CreateRack(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) CreateBay(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		RackID         string   `json:"rack_id"`
		Name           string   `json:"name"`
		Code           string   `json:"code"`
		BayNumber      int      `json:"bay_number"`
		BayLevel       int      `json:"bay_level"`
		Capacity       int      `json:"capacity"`
		WeightCapacity *float64 `json:"weight_capacity"`
		SortOrder      int      `json:"sort_order"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.CreateBayCommand{
		TenantID:       tenantID,
		RackID:         payload.RackID,
		Name:           payload.Name,
		Code:           payload.Code,
		BayNumber:      payload.BayNumber,
		BayLevel:       payload.BayLevel,
		Capacity:       payload.Capacity,
		WeightCapacity: payload.WeightCapacity,
		SortOrder:      payload.SortOrder,
	}

	dto, err := h.warehouseSvc.CreateBay(r.Context(), cmd)
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
		case "NOT_FOUND", "WAREHOUSE_NOT_FOUND", "ZONE_NOT_FOUND", "RACK_NOT_FOUND", "BAY_NOT_FOUND":
			h.respondError(w, http.StatusNotFound, domainErr.Message)
		case "VALIDATION_ERROR", "INVALID_INPUT", "INVALID_STATE":
			h.respondError(w, http.StatusBadRequest, domainErr.Message)
		case "TENANT_REQUIRED", "NAME_REQUIRED", "EQUIPMENT_REQUIRED", "CODE_REQUIRED":
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
		case "CREATE_ERROR", "UPDATE_ERROR", "DELETE_ERROR", "QUERY_ERROR":
			h.respondError(w, http.StatusInternalServerError, domainErr.Message)
		default:
			h.respondError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	h.respondError(w, http.StatusInternalServerError, "internal server error")
}
