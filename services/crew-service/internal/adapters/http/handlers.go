package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
	bookingSvc        *application.BookingService
	logger            logger.Logger
}

// NewHandlers creates a new handlers instance
func NewHandlers(
	crewSvc *application.CrewService,
	qualSvc *application.QualificationService,
	assignmentSvc *application.AssignmentService,
	timeRecordSvc *application.TimeRecordService,
	bookingSvc *application.BookingService,
	log logger.Logger,
) *Handlers {
	return &Handlers{
		crewService:      crewSvc,
		qualificationSvc: qualSvc,
		assignmentSvc:    assignmentSvc,
		timeRecordSvc:    timeRecordSvc,
		bookingSvc:       bookingSvc,
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

// ---- Driver Handlers ----

// GetDriverList retrieves the list of drivers with vehicle assignments
func (h *Handlers) GetDriverList(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		writeError(w, http.StatusBadRequest, "missing_tenant_id", "X-Tenant-ID header is required")
		return
	}

	drivers, err := h.crewService.GetDriverList(r.Context(), tenantID)
	if err != nil {
		h.logger.Error("failed to get driver list", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to get driver list")
		return
	}

	writeJSON(w, http.StatusOK, drivers)
}

// ---- Calendar Handlers ----

// ExportCalendarICS exports a crew member's assignments as an iCalendar feed
func (h *Handlers) ExportCalendarICS(w http.ResponseWriter, r *http.Request) {
	memberID := r.PathValue("id")
	if memberID == "" {
		writeError(w, http.StatusBadRequest, "missing_id", "crew member ID is required")
		return
	}

	// Get crew member details
	crewDTO, err := h.crewService.GetCrewMember(r.Context(), memberID)
	if err != nil {
		if err == domain.ErrCrewMemberNotFound {
			writeError(w, http.StatusNotFound, "not_found", "crew member not found")
		} else {
			h.logger.Error("failed to get crew member", err)
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to get crew member")
		}
		return
	}

	// Get all assignments for the member
	assignments, err := h.assignmentSvc.GetAssignmentsForMember(r.Context(), memberID)
	if err != nil {
		h.logger.Error("failed to get assignments for member", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to get assignments")
		return
	}

	// Generate ICS content
	icsContent := generateICSCalendar(crewDTO, assignments)

	// Set proper headers for calendar feed
	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s-schedule.ics", memberID))
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(icsContent))
}

// generateICSCalendar generates an iCalendar format string from crew member and assignments
func generateICSCalendar(crew *application.CrewMemberDTO, assignments []*application.CrewAssignmentDTO) string {
	sb := strings.Builder{}

	// Write iCalendar header
	sb.WriteString("BEGIN:VCALENDAR\r\n")
	sb.WriteString("VERSION:2.0\r\n")
	sb.WriteString("PRODID:-//EquipFlow//Crew Schedule//EN\r\n")
	sb.WriteString(fmt.Sprintf("CALSCALE:GREGORIAN\r\n"))
	sb.WriteString(fmt.Sprintf("METHOD:PUBLISH\r\n"))
	sb.WriteString(fmt.Sprintf("X-WR-CALNAME:%s Schedule\r\n", crew.FirstName+" "+crew.LastName))
	sb.WriteString(fmt.Sprintf("X-WR-TIMEZONE:UTC\r\n"))

	// Write events for each assignment
	for _, assignment := range assignments {
		// Skip cancelled assignments
		if assignment.Status == "cancelled" {
			continue
		}

		// Parse dates
		startDate, _ := time.Parse(time.RFC3339, assignment.StartDate)
		endDate, _ := time.Parse(time.RFC3339, assignment.EndDate)

		// Create unique identifier for the event
		eventUID := fmt.Sprintf("%s-%s@rentflow.local", assignment.ID, crew.ID)

		sb.WriteString("BEGIN:VEVENT\r\n")
		sb.WriteString(fmt.Sprintf("UID:%s\r\n", eventUID))
		sb.WriteString(fmt.Sprintf("DTSTAMP:%s\r\n", time.Now().UTC().Format("20060102T150405Z")))
		sb.WriteString(fmt.Sprintf("DTSTART:%s\r\n", startDate.UTC().Format("20060102T150405Z")))
		sb.WriteString(fmt.Sprintf("DTEND:%s\r\n", endDate.UTC().Format("20060102T150405Z")))
		sb.WriteString(fmt.Sprintf("SUMMARY:%s - %s\r\n", assignment.Role, assignment.Status))

		// Add optional project or tour reference
		description := "Crew Assignment"
		if assignment.ProjectID != nil && *assignment.ProjectID != "" {
			description += fmt.Sprintf("\nProject ID: %s", *assignment.ProjectID)
		}
		if assignment.TourID != nil && *assignment.TourID != "" {
			description += fmt.Sprintf("\nTour ID: %s", *assignment.TourID)
		}
		if assignment.Notes != "" {
			description += fmt.Sprintf("\nNotes: %s", assignment.Notes)
		}
		sb.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", escapeICSText(description)))
		sb.WriteString(fmt.Sprintf("STATUS:%s\r\n", mapAssignmentStatusToICS(assignment.Status)))
		sb.WriteString(fmt.Sprintf("LOCATION:Assignment\r\n"))
		sb.WriteString("END:VEVENT\r\n")
	}

	// Write iCalendar footer
	sb.WriteString("END:VCALENDAR\r\n")

	return sb.String()
}

// escapeICSText escapes special characters in ICS text fields
func escapeICSText(text string) string {
	text = strings.ReplaceAll(text, "\\", "\\\\")
	text = strings.ReplaceAll(text, ",", "\\,")
	text = strings.ReplaceAll(text, ";", "\\;")
	text = strings.ReplaceAll(text, "\n", "\\n")
	return text
}

// mapAssignmentStatusToICS maps assignment status to ICS status
func mapAssignmentStatusToICS(status string) string {
	switch status {
	case "completed":
		return "COMPLETED"
	case "cancelled":
		return "CANCELLED"
	default:
		return "CONFIRMED"
	}
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

// ---- Booking Handlers ----

// CreateBookingRequest creates a new booking request
func (h *Handlers) CreateBookingRequest(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		writeError(w, http.StatusBadRequest, "missing_tenant_id", "X-Tenant-ID header is required")
		return
	}

	var cmd application.CreateBookingRequestCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}

	cmd.TenantID = tenantID

	dto, err := h.bookingSvc.CreateBookingRequest(r.Context(), cmd)
	if err != nil {
		switch err {
		case domain.ErrAssignmentNotFound:
			writeError(w, http.StatusNotFound, "not_found", "assignment not found")
		case domain.ErrCrewMemberNotFound:
			writeError(w, http.StatusNotFound, "not_found", "crew member not found")
		default:
			h.logger.Error("failed to create booking request", err)
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to create booking request")
		}
		return
	}

	writeJSON(w, http.StatusCreated, dto)
}

// ListBookingRequests lists booking requests for a tenant
func (h *Handlers) ListBookingRequests(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		writeError(w, http.StatusBadRequest, "missing_tenant_id", "X-Tenant-ID header is required")
		return
	}

	bookings, err := h.bookingSvc.ListBookingRequests(r.Context(), tenantID)
	if err != nil {
		h.logger.Error("failed to list booking requests", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to list booking requests")
		return
	}

	writeJSON(w, http.StatusOK, bookings)
}

// GetBookingDetails retrieves booking details by token (public, no auth)
func (h *Handlers) GetBookingDetails(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if token == "" {
		writeError(w, http.StatusBadRequest, "missing_token", "booking token is required")
		return
	}

	dto, err := h.bookingSvc.GetBookingDetails(r.Context(), token)
	if err != nil {
		if err == domain.ErrBookingNotFound {
			writeError(w, http.StatusNotFound, "not_found", "booking request not found")
		} else {
			h.logger.Error("failed to get booking details", err)
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to get booking details")
		}
		return
	}

	writeJSON(w, http.StatusOK, dto)
}

// RespondToBooking processes a freelancer's response (public, no auth)
func (h *Handlers) RespondToBooking(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if token == "" {
		writeError(w, http.StatusBadRequest, "missing_token", "booking token is required")
		return
	}

	var cmd application.BookingResponseCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}

	cmd.Token = token

	if cmd.Status != "accepted" && cmd.Status != "declined" && cmd.Status != "alternative" {
		writeError(w, http.StatusBadRequest, "invalid_status", "status must be accepted, declined, or alternative")
		return
	}

	err := h.bookingSvc.RespondToBooking(r.Context(), cmd)
	if err != nil {
		switch err {
		case domain.ErrBookingNotFound:
			writeError(w, http.StatusNotFound, "not_found", "booking request not found")
		case domain.ErrBookingAlreadyResponded:
			writeError(w, http.StatusConflict, "already_responded", "booking request has already been responded to")
		default:
			h.logger.Error("failed to respond to booking", err)
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to respond to booking")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "response recorded successfully"})
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
