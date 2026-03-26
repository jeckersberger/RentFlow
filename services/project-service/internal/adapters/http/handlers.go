package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/project-service/internal/application"
	"github.com/jeckersberger/rentflow/services/project-service/internal/domain"
)

type Handler struct {
	projectSvc     *application.ProjectService
	packlistSvc    *application.PacklistService
	reservationSvc *application.ReservationService
	customerSvc    *application.CustomerService
	contactSvc     *application.ContactService
	logger         logger.Logger
}

func NewHandler(
	projectSvc *application.ProjectService,
	packlistSvc *application.PacklistService,
	reservationSvc *application.ReservationService,
	customerSvc *application.CustomerService,
	contactSvc *application.ContactService,
	logger logger.Logger,
) *Handler {
	return &Handler{
		projectSvc:     projectSvc,
		packlistSvc:    packlistSvc,
		reservationSvc: reservationSvc,
		customerSvc:    customerSvc,
		contactSvc:     contactSvc,
		logger:         logger,
	}
}

// Project Handlers

func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var cmd application.CreateProjectCommand
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

	// Set created_by_user_id from JWT middleware header
	userID := r.Header.Get("X-User-ID")
	if userID != "" {
		cmd.CreatedByUserID = userID
	}

	// Basic input validation
	if strings.TrimSpace(cmd.Name) == "" {
		h.respondError(w, http.StatusBadRequest, "project name is required")
		return
	}

	dto, err := h.projectSvc.CreateProject(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) GetProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dto, err := h.projectSvc.GetProject(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) ListProjects(w http.ResponseWriter, r *http.Request) {
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

	filter := r.URL.Query().Get("filter")

	query := application.ListProjectsQuery{
		TenantID: tenantID,
		Filter:   filter,
		Limit:    limit,
		Offset:   offset,
	}

	result, err := h.projectSvc.ListProjects(r.Context(), query)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

func (h *Handler) SearchProjects(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.projectSvc.SearchProjects(r.Context(), application.SearchProjectsQuery{
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

func (h *Handler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var cmd application.UpdateProjectCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd.ID = id
	cmd.TenantID = tenantID

	dto, err := h.projectSvc.UpdateProject(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) ChangeProjectStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		Status domain.ProjectStatus `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.ChangeProjectStatusCommand{
		ID:       id,
		TenantID: tenantID,
		Status:   payload.Status,
	}

	if err := h.projectSvc.ChangeStatus(r.Context(), cmd); err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"message": "status updated"})
}

func (h *Handler) SetProjectManager(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		ManagerID string `json:"manager_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.SetProjectManagerCommand{
		ID:        id,
		TenantID:  tenantID,
		ManagerID: payload.ManagerID,
	}

	if err := h.projectSvc.SetProjectManager(r.Context(), cmd); err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"message": "project manager updated"})
}

func (h *Handler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	if err := h.projectSvc.DeleteProject(r.Context(), tenantID, id); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Packlist Handlers

func (h *Handler) CreatePacklist(w http.ResponseWriter, r *http.Request) {
	var cmd application.CreatePacklistCommand
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

	dto, err := h.packlistSvc.CreatePacklist(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) GetPacklist(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dto, err := h.packlistSvc.GetPacklist(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) ListPacklists(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	projectID := r.URL.Query().Get("project_id")
	if projectID == "" {
		h.respondError(w, http.StatusBadRequest, "project_id required")
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

	query := application.ListPacklistsQuery{
		TenantID:  tenantID,
		ProjectID: projectID,
		Limit:     limit,
		Offset:    offset,
	}

	result, err := h.packlistSvc.ListByProject(r.Context(), query)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

func (h *Handler) AddPacklistItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		EquipmentID   string `json:"equipment_id"`
		EquipmentName string `json:"equipment_name"`
		Quantity      int    `json:"quantity"`
		Notes         string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.AddPacklistItemCommand{
		TenantID:      tenantID,
		PacklistID:    id,
		EquipmentID:   payload.EquipmentID,
		EquipmentName: payload.EquipmentName,
		Quantity:      payload.Quantity,
		Notes:         payload.Notes,
	}

	dto, err := h.packlistSvc.AddItem(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) RemovePacklistItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		EquipmentID string `json:"equipment_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.RemovePacklistItemCommand{
		TenantID:    tenantID,
		PacklistID:  id,
		EquipmentID: payload.EquipmentID,
	}

	dto, err := h.packlistSvc.RemoveItem(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) MarkItemPacked(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		EquipmentID    string `json:"equipment_id"`
		QuantityPacked int    `json:"quantity_packed"`
		Notes          string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.MarkItemPackedCommand{
		TenantID:       tenantID,
		PacklistID:     id,
		EquipmentID:    payload.EquipmentID,
		QuantityPacked: payload.QuantityPacked,
		Notes:          payload.Notes,
	}

	dto, err := h.packlistSvc.MarkItemPacked(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) MarkItemReturned(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		EquipmentID      string `json:"equipment_id"`
		QuantityReturned int    `json:"quantity_returned"`
		Notes            string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.MarkItemReturnedCommand{
		TenantID:         tenantID,
		PacklistID:       id,
		EquipmentID:      payload.EquipmentID,
		QuantityReturned: payload.QuantityReturned,
		Notes:            payload.Notes,
	}

	dto, err := h.packlistSvc.MarkItemReturned(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) ChangePacklistStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		Status domain.PacklistStatus `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.ChangePacklistStatusCommand{
		ID:       id,
		TenantID: tenantID,
		Status:   payload.Status,
	}

	if err := h.packlistSvc.ChangeStatus(r.Context(), cmd); err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"message": "status updated"})
}

func (h *Handler) DeletePacklist(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	if err := h.packlistSvc.DeletePacklist(r.Context(), tenantID, id); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Reservation Handlers

func (h *Handler) CreateReservation(w http.ResponseWriter, r *http.Request) {
	var cmd application.CreateReservationCommand
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

	dto, err := h.reservationSvc.CreateReservation(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) GetReservation(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dto, err := h.reservationSvc.GetReservation(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) ListReservations(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	projectID := r.URL.Query().Get("project_id")
	if projectID == "" {
		h.respondError(w, http.StatusBadRequest, "project_id required")
		return
	}

	query := application.ListReservationsQuery{
		TenantID:  tenantID,
		ProjectID: projectID,
	}

	dtos, err := h.reservationSvc.ListByProject(r.Context(), query)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dtos)
}

func (h *Handler) CheckConflicts(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	equipmentID := r.URL.Query().Get("equipment_id")
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	excludeProjectID := r.URL.Query().Get("exclude_project_id")

	if equipmentID == "" || startStr == "" || endStr == "" {
		h.respondError(w, http.StatusBadRequest, "equipment_id, start, and end required")
		return
	}

	// Parse dates - try RFC3339 then date-only format
	startDate, err := parseHandlerDate(startStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid start date format")
		return
	}
	endDate, err := parseHandlerDate(endStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid end date format")
		return
	}

	query := application.CheckReservationConflictQuery{
		TenantID:         tenantID,
		EquipmentID:      equipmentID,
		StartDate:        startDate,
		EndDate:          endDate,
		ExcludeProjectID: excludeProjectID,
	}

	dtos, err := h.reservationSvc.CheckConflicts(r.Context(), query)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dtos)
}

func parseHandlerDate(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse date: %s", s)
}

func (h *Handler) ConfirmReservation(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	cmd := application.ConfirmReservationCommand{
		ID:       id,
		TenantID: tenantID,
	}

	if err := h.reservationSvc.ConfirmReservation(r.Context(), cmd); err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"message": "reservation confirmed"})
}

func (h *Handler) CancelReservation(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		Reason string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := application.CancelReservationCommand{
		ID:       id,
		TenantID: tenantID,
		Reason:   payload.Reason,
	}

	if err := h.reservationSvc.CancelReservation(r.Context(), cmd); err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"message": "reservation cancelled"})
}

func (h *Handler) DeleteReservation(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	if err := h.reservationSvc.DeleteReservation(r.Context(), tenantID, id); err != nil {
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
		case "VALIDATION_ERROR", "INVALID_STATUS", "INVALID_INPUT", "INVALID_DATE":
			h.respondError(w, http.StatusBadRequest, domainErr.Message)
		case "TENANT_REQUIRED", "NAME_REQUIRED", "CLIENT_REQUIRED", "PROJECT_REQUIRED", "EQUIPMENT_REQUIRED":
			h.respondError(w, http.StatusBadRequest, domainErr.Message)
		case "CREATE_ERROR", "UPDATE_ERROR", "DELETE_ERROR", "QUERY_ERROR", "SEARCH_ERROR":
			h.logger.Error("Database operation failed", domainErr.Err)
			h.respondError(w, http.StatusInternalServerError, domainErr.Message)
		case "UNAUTHORIZED":
			h.respondError(w, http.StatusUnauthorized, domainErr.Message)
		case "CONFLICT":
			h.respondError(w, http.StatusConflict, domainErr.Message)
		default:
			h.respondError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	h.respondError(w, http.StatusInternalServerError, "internal server error")
}

// Customer Handlers

func (h *Handler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	var cmd application.CreateCustomerCommand
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

	dto, err := h.customerSvc.CreateCustomer(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) GetCustomer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dto, err := h.customerSvc.GetCustomer(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) ListCustomers(w http.ResponseWriter, r *http.Request) {
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

	query := application.ListCustomersQuery{
		TenantID: tenantID,
		Limit:    limit,
		Offset:   offset,
	}

	result, err := h.customerSvc.ListCustomers(r.Context(), query)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

func (h *Handler) UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var cmd application.UpdateCustomerCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd.ID = id
	cmd.TenantID = tenantID

	dto, err := h.customerSvc.UpdateCustomer(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) DeleteCustomer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	if err := h.customerSvc.DeleteCustomer(r.Context(), tenantID, id); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetProjectEquipment returns the equipment list (Soll-Liste) for a project
func (h *Handler) GetProjectEquipment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dto, err := h.projectSvc.GetProjectEquipment(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

// Calendar and Project Extended Handlers

func (h *Handler) GetCalendar(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	if startStr == "" || endStr == "" {
		h.respondError(w, http.StatusBadRequest, "start and end dates required")
		return
	}

	query := application.GetCalendarQuery{
		TenantID:  tenantID,
		StartDate: startStr,
		EndDate:   endStr,
	}

	projects, err := h.projectSvc.GetCalendar(r.Context(), query)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, projects)
}

func (h *Handler) CopyProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var payload struct {
		NewStartDate string `json:"new_start_date"`
		NewEndDate   string `json:"new_end_date"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	dto, err := h.projectSvc.CopyProject(r.Context(), id, tenantID, payload.NewStartDate, payload.NewEndDate)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) GetPackingListJSON(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dto, err := h.projectSvc.GeneratePackingListJSON(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) GetPackingListHTML(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	html, err := h.projectSvc.GeneratePackingListHTML(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// Contact Handlers

func (h *Handler) CreateContact(w http.ResponseWriter, r *http.Request) {
	var cmd application.CreateContactCommand
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
	cmd.CreatedBy = r.Header.Get("X-User-ID")

	dto, err := h.contactSvc.CreateContact(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto)
}

func (h *Handler) GetContact(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	dto, err := h.contactSvc.GetContact(r.Context(), tenantID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) ListContacts(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	limit := 50
	offset := 0
	searchTerm := ""

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

	if s := r.URL.Query().Get("search"); s != "" {
		searchTerm = s
	}

	query := application.ListContactsQuery{
		TenantID:   tenantID,
		SearchTerm: searchTerm,
		Limit:      limit,
		Offset:     offset,
	}

	result, err := h.contactSvc.ListContacts(r.Context(), query)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

func (h *Handler) UpdateContact(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	var cmd application.UpdateContactCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd.ID = id
	cmd.TenantID = tenantID

	dto, err := h.contactSvc.UpdateContact(r.Context(), cmd)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto)
}

func (h *Handler) DeleteContact(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant ID required")
		return
	}

	if err := h.contactSvc.DeleteContact(r.Context(), tenantID, id); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
