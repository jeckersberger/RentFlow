package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/crew-service/internal/application"
	"github.com/jeckersberger/rentflow/services/crew-service/internal/domain"
)

// Handlers contains all HTTP handlers
type Handlers struct {
	crewService       *application.CrewService
	qualificationSvc  *application.QualificationService
	assignmentSvc     *application.AssignmentService
	timeRecordSvc     *application.TimeRecordService
	logger            logger.Logger
}

// NewHandlers creates a new handlers instance
func NewHandlers(
	crewSvc *application.CrewService,
	qualSvc *application.QualificationService,
	assignmentSvc *application.AssignmentService,
	timeRecordSvc *application.TimeRecordService,
	log logger.Logger,
) *Handlers {
	return &Handlers{
		crewService:      crewSvc,
		qualificationSvc: qualSvc,
		assignmentSvc:    assignmentSvc,
		timeRecordSvc:    timeRecordSvc,
		logger:           log,
	}
}

// ---- Crew Member Handlers ----

// ListCrewMembers lists crew members
func (h *Handlers) ListCrewMembers(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		writeError(w, http.StatusBadRequest, "missing_tenant_id", "X-Tenant-ID header is required")
		return
	}

	page := 1
	perPage := 20
	if p := r.URL.Query().Get("page"); p != "" {
		if pageNum, err := strconv.Atoi(p); err == nil && pageNum > 0 {
			page = pageNum
		}
	}
	if pp := r.URL.Query().Get("per_page"); pp != "" {
		if ppNum, err := strconv.Atoi(pp); err == nil && ppNum > 0 && ppNum <= 100 {
			perPage = ppNum
		}
	}

	result, err := h.crewService.ListCrewMembers(r.Context(), tenantID, page, perPage)
	if err != nil {
		h.logger.Error("failed to list crew members", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to list crew members")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// CreateCrewMember creates a new crew member
func (h *Handlers) CreateCrewMember(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		writeError(w, http.StatusBadRequest, "missing_tenant_id", "X-Tenant-ID header is required")
		return
	}

	var cmd application.CreateCrewMemberCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}

	cmd.TenantID = tenantID

	dto, err := h.crewService.CreateCrewMember(r.Context(), cmd)
	if err != nil {
		switch err {
		case domain.ErrDuplicateEmail:
			writeError(w, http.StatusConflict, "duplicate_email", "email already exists")
		case domain.ErrTenantIDRequired:
			writeError(w, http.StatusBadRequest, "missing_tenant_id", "tenant ID is required")
		default:
			h.logger.Error("failed to create crew member", err)
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to create crew member")
		}
		return
	}

	writeJSON(w, http.StatusCreated, dto)
}

// GetCrewMember retrieves a crew member
func (h *Handlers) GetCrewMember(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing_id", "crew member ID is required")
		return
	}

	dto, err := h.crewService.GetCrewMember(r.Context(), id)
	if err != nil {
		if err == domain.ErrCrewMemberNotFound {
			writeError(w, http.StatusNotFound, "not_found", "crew member not found")
		} else {
			h.logger.Error("failed to get crew member", err)
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to get crew member")
		}
		return
	}

	writeJSON(w, http.StatusOK, dto)
}

// UpdateCrewMember updates a crew member
func (h *Handlers) UpdateCrewMember(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		writeError(w, http.StatusBadRequest, "missing_tenant_id", "X-Tenant-ID header is required")
		return
	}

	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing_id", "crew member ID is required")
		return
	}

	var cmd application.UpdateCrewMemberCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}

	cmd.ID = id
	cmd.TenantID = tenantID

	dto, err := h.crewService.UpdateCrewMember(r.Context(), cmd)
	if err != nil {
		if err == domain.ErrCrewMemberNotFound {
			writeError(w, http.StatusNotFound, "not_found", "crew member not found")
		} else {
			h.logger.Error("failed to update crew member", err)
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to update crew member")
		}
		return
	}

	writeJSON(w, http.StatusOK, dto)
}

// DeleteCrewMember deletes a crew member
func (h *Handlers) DeleteCrewMember(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing_id", "crew member ID is required")
		return
	}

	err := h.crewService.DeleteCrewMember(r.Context(), id)
	if err != nil {
		if err == domain.ErrCrewMemberNotFound {
			writeError(w, http.StatusNotFound, "not_found", "crew member not found")
		} else {
			h.logger.Error("failed to delete crew member", err)
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to delete crew member")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ---- Qualification Handlers ----

// GetQualifications lists qualifications for a crew member
func (h *Handlers) GetQualifications(w http.ResponseWriter, r *http.Request) {
	crewMemberID := r.PathValue("id")
	if crewMemberID == "" {
		writeError(w, http.StatusBadRequest, "missing_id", "crew member ID is required")
		return
	}

	dtos, err := h.qualificationSvc.ListQualificationsForCrewMember(r.Context(), crewMemberID)
	if err != nil {
		h.logger.Error("failed to list qualifications", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to list qualifications")
		return
	}

	writeJSON(w, http.StatusOK, dtos)
}

// CreateQualification creates a new qualification
func (h *Handlers) CreateQualification(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		writeError(w, http.StatusBadRequest, "missing_tenant_id", "X-Tenant-ID header is required")
		return
	}

	crewMemberID := r.PathValue("id")
	if crewMemberID == "" {
		writeError(w, http.StatusBadRequest, "missing_id", "crew member ID is required")
		return
	}

	var cmd application.CreateQualificationCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}

	cmd.TenantID = tenantID
	cmd.CrewMemberID = crewMemberID

	dto, err := h.qualificationSvc.CreateQualification(r.Context(), cmd)
	if err != nil {
		if err == domain.ErrCrewMemberNotFound {
			writeError(w, http.StatusNotFound, "not_found", "crew member not found")
		} else {
			h.logger.Error("failed to create qualification", err)
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to create qualification")
		}
		return
	}

	writeJSON(w, http.StatusCreated, dto)
}

// ---- Assignment Handlers ----

// ListAssignments lists crew assignments
func (h *Handlers) ListAssignments(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		writeError(w, http.StatusBadRequest, "missing_tenant_id", "X-Tenant-ID header is required")
		return
	}

	page := 1
	perPage := 20
	if p := r.URL.Query().Get("page"); p != "" {
		if pageNum, err := strconv.Atoi(p); err == nil && pageNum > 0 {
			page = pageNum
		}
	}
	if pp := r.URL.Query().Get("per_page"); pp != "" {
		if ppNum, err := strconv.Atoi(pp); err == nil && ppNum > 0 && ppNum <= 100 {
			perPage = ppNum
		}
	}

	result, err := h.assignmentSvc.ListAssignments(r.Context(), tenantID, page, perPage)
	if err != nil {
		h.logger.Error("failed to list assignments", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to list assignments")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// CreateAssignment creates a new crew assignment
func (h *Handlers) CreateAssignment(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		writeError(w, http.StatusBadRequest, "missing_tenant_id", "X-Tenant-ID header is required")
		return
	}

	var cmd application.CreateCrewAssignmentCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}

	cmd.TenantID = tenantID

	dto, err := h.assignmentSvc.CreateAssignment(r.Context(), cmd)
	if err != nil {
		switch err {
		case domain.ErrConflictDetected:
			writeError(w, http.StatusConflict, "conflict_detected", "assignment conflicts with existing assignment")
		case domain.ErrCrewMemberNotFound:
			writeError(w, http.StatusNotFound, "not_found", "crew member not found")
		case domain.ErrInvalidDateRange:
			writeError(w, http.StatusBadRequest, "invalid_date", "invalid date range")
		default:
			h.logger.Error("failed to create assignment", err)
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to create assignment")
		}
		return
	}

	writeJSON(w, http.StatusCreated, dto)
}

// GetAssignment retrieves an assignment
func (h *Handlers) GetAssignment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing_id", "assignment ID is required")
		return
	}

	dto, err := h.assignmentSvc.GetAssignment(r.Context(), id)
	if err != nil {
		if err == domain.ErrAssignmentNotFound {
			writeError(w, http.StatusNotFound, "not_found", "assignment not found")
		} else {
			h.logger.Error("failed to get assignment", err)
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to get assignment")
		}
		return
	}

	writeJSON(w, http.StatusOK, dto)
}

// UpdateAssignment updates an assignment
func (h *Handlers) UpdateAssignment(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		writeError(w, http.StatusBadRequest, "missing_tenant_id", "X-Tenant-ID header is required")
		return
	}

	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing_id", "assignment ID is required")
		return
	}

	var cmd application.UpdateCrewAssignmentCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}

	cmd.ID = id
	cmd.TenantID = tenantID

	dto, err := h.assignmentSvc.UpdateAssignment(r.Context(), cmd)
	if err != nil {
		if err == domain.ErrAssignmentNotFound {
			writeError(w, http.StatusNotFound, "not_found", "assignment not found")
		} else {
			h.logger.Error("failed to update assignment", err)
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to update assignment")
		}
		return
	}

	writeJSON(w, http.StatusOK, dto)
}

// DetectConflicts detects assignment conflicts
func (h *Handlers) DetectConflicts(w http.ResponseWriter, r *http.Request) {
	crewMemberID := r.URL.Query().Get("member_id")
	if crewMemberID == "" {
		writeError(w, http.StatusBadRequest, "missing_param", "member_id query parameter is required")
		return
	}

	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	if startStr == "" || endStr == "" {
		writeError(w, http.StatusBadRequest, "missing_param", "start and end query parameters are required")
		return
	}

	startDate, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_date", "invalid start date format")
		return
	}

	endDate, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_date", "invalid end date format")
		return
	}

	conflicts, err := h.assignmentSvc.DetectConflicts(r.Context(), crewMemberID, startDate, endDate)
	if err != nil {
		h.logger.Error("failed to detect conflicts", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to detect conflicts")
		return
	}

	writeJSON(w, http.StatusOK, conflicts)
}

// ---- Availability Handlers ----

// CheckAvailability checks crew member availability
func (h *Handlers) CheckAvailability(w http.ResponseWriter, r *http.Request) {
	crewMemberID := r.PathValue("id")
	if crewMemberID == "" {
		writeError(w, http.StatusBadRequest, "missing_id", "crew member ID is required")
		return
	}

	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	if startStr == "" || endStr == "" {
		writeError(w, http.StatusBadRequest, "missing_param", "start and end query parameters are required")
		return
	}

	startDate, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_date", "invalid start date format")
		return
	}

	endDate, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_date", "invalid end date format")
		return
	}

	availability, err := h.crewService.CheckAvailability(r.Context(), crewMemberID, startDate, endDate)
	if err != nil {
		if err == domain.ErrCrewMemberNotFound {
			writeError(w, http.StatusNotFound, "not_found", "crew member not found")
		} else {
			h.logger.Error("failed to check availability", err)
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to check availability")
		}
		return
	}

	writeJSON(w, http.StatusOK, availability)
}

// ---- Helper functions ----

func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(application.ErrorResponse{
		Code:    code,
		Message: message,
	})
}
