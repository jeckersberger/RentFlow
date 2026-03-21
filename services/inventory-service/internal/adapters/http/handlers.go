package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/inventory-service/internal/application"
	"github.com/jeckersberger/rentflow/services/inventory-service/internal/domain"
)

type Handler struct {
	equipmentSvc   *application.EquipmentService
	categorySvc    *application.CategoryService
	flightcaseSvc  *application.FlightcaseService
	logger         logger.Logger
}

func NewHandler(
	equipmentSvc *application.EquipmentService,
	categorySvc *application.CategoryService,
	flightcaseSvc *application.FlightcaseService,
	logger logger.Logger,
) *Handler {
	return &Handler{
		equipmentSvc:  equipmentSvc,
		categorySvc:   categorySvc,
		flightcaseSvc: flightcaseSvc,
		logger:        logger,
	}
}

// Equipment Handlers

func (h *Handler) CreateEquipment(w http.ResponseWriter, r *http.Request) {
	var cmd application.CreateEquipmentCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get tenant from context (middleware should set this)
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}
	cmd.TenantID = tenantID

	dto, err := h.equipmentSvc.CreateEquipment(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) GetEquipment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dto, err := h.equipmentSvc.GetEquipment(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) ListEquipment(w http.ResponseWriter, r *http.Request) {
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

	query := application.ListEquipmentQuery{
		TenantID: tenantID,
		Limit:    limit,
		Offset:   offset,
	}

	if status := r.URL.Query().Get("status"); status != "" {
		s := domain.EquipmentStatus(status)
		query.Status = &s
	}

	if categoryID := r.URL.Query().Get("category_id"); categoryID != "" {
		query.CategoryID = &categoryID
	}

	if locationID := r.URL.Query().Get("location_id"); locationID != "" {
		query.LocationID = &locationID
	}

	result, err := h.equipmentSvc.ListEquipment(r.Context(), query)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

func (h *Handler) SearchEquipment(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	term := r.URL.Query().Get("q")
	if term == "" {
		h.respondError(w, http.StatusBadRequest, "search term required")
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

	result, err := h.equipmentSvc.SearchEquipment(r.Context(), application.SearchEquipmentQuery{
		TenantID:   tenantID,
		SearchTerm: term,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

func (h *Handler) GetEquipmentByBarcode(w http.ResponseWriter, r *http.Request) {
	barcode := r.PathValue("barcode")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dto, err := h.equipmentSvc.GetByBarcode(r.Context(), tenantID, barcode)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) UpdateEquipment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var cmd application.UpdateEquipmentCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd.ID = id
	cmd.TenantID = tenantID

	dto, err := h.equipmentSvc.UpdateEquipment(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) ChangeEquipmentStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		Status domain.EquipmentStatus `json:"status"`
		Reason string                 `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.ChangeStatusCommand{
		ID:       id,
		TenantID: tenantID,
		Status:   payload.Status,
		Reason:   payload.Reason,
	}

	if err := h.equipmentSvc.ChangeStatus(r.Context(), cmd); err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"message": "status updated"})
}

func (h *Handler) UpdateEquipmentCondition(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		Condition domain.EquipmentCondition `json:"condition"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.UpdateConditionCommand{
		ID:        id,
		TenantID:  tenantID,
		Condition: payload.Condition,
	}

	if err := h.equipmentSvc.UpdateCondition(r.Context(), cmd); err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"message": "condition updated"})
}

func (h *Handler) AddEquipmentImage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		h.respondError(w, http.StatusBadRequest, "failed to parse form")
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "image file required")
		return
	}
	defer file.Close()

	dto, err := h.equipmentSvc.AddImage(r.Context(), tenantID, id, file, handler.Filename)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) DeleteEquipment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	if err := h.equipmentSvc.DeleteEquipment(r.Context(), tenantID, id); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Category Handlers

func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var cmd application.CreateCategoryCommand
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

	dto, err := h.categorySvc.CreateCategory(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) ListCategories(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dtos, err := h.categorySvc.ListCategories(r.Context(), tenantID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dtos)
}

func (h *Handler) GetCategory(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dto, err := h.categorySvc.GetCategory(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var cmd application.UpdateCategoryCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd.ID = id
	cmd.TenantID = tenantID

	dto, err := h.categorySvc.UpdateCategory(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	if err := h.categorySvc.DeleteCategory(r.Context(), tenantID, id); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Flightcase Handlers

func (h *Handler) CreateFlightcase(w http.ResponseWriter, r *http.Request) {
	var cmd application.CreateFlightcaseCommand
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

	dto, err := h.flightcaseSvc.CreateFlightcase(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) ListFlightcases(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.flightcaseSvc.ListFlightcases(r.Context(), tenantID, limit, offset)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

func (h *Handler) GetFlightcase(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dto, err := h.flightcaseSvc.GetFlightcase(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) UpdateFlightcase(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var cmd application.UpdateFlightcaseCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd.ID = id
	cmd.TenantID = tenantID

	dto, err := h.flightcaseSvc.UpdateFlightcase(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) AddFlightcaseItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		EquipmentID string `json:"equipment_id"`
		Quantity    int    `json:"quantity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.AddFlightcaseItemCommand{
		FlightcaseID: id,
		TenantID:     tenantID,
		EquipmentID:  payload.EquipmentID,
		Quantity:     payload.Quantity,
	}

	dto, err := h.flightcaseSvc.AddItem(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) RemoveFlightcaseItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		EquipmentID string `json:"equipment_id"`
		Quantity    int    `json:"quantity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.RemoveFlightcaseItemCommand{
		FlightcaseID: id,
		TenantID:     tenantID,
		EquipmentID:  payload.EquipmentID,
		Quantity:     payload.Quantity,
	}

	dto, err := h.flightcaseSvc.RemoveItem(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) DeleteFlightcase(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	if err := h.flightcaseSvc.DeleteFlightcase(r.Context(), tenantID, id); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
		case "VALIDATION_ERROR", "INVALID_STATUS", "INVALID_CONDITION", "INVALID_LOCATION", "INVALID_INPUT":
			h.respondError(w, http.StatusBadRequest, domainErr.Message)
		case "TENANT_REQUIRED", "BARCODE_REQUIRED", "NAME_REQUIRED":
			h.respondError(w, http.StatusBadRequest, domainErr.Message)
		case "UNAUTHORIZED":
			h.respondError(w, http.StatusUnauthorized, domainErr.Message)
		case "BARCODE_EXISTS", "INVALID_PARENT", "EQUIPMENT_NOT_FOUND":
			h.respondError(w, http.StatusConflict, domainErr.Message)
		default:
			h.respondError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	h.respondError(w, http.StatusInternalServerError, "internal server error")
}
