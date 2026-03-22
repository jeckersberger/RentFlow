package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jeckersberger/rentflow/services/crew-service/internal/application"
	"github.com/jeckersberger/rentflow/services/crew-service/internal/domain"
)

// StartTimeRecord starts a new time record
func (h *Handlers) StartTimeRecord(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		writeError(w, http.StatusBadRequest, "missing_tenant_id", "X-Tenant-ID header is required")
		return
	}

	var cmd application.StartTimeRecordCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}

	cmd.TenantID = tenantID

	dto, err := h.timeRecordSvc.StartTimeRecord(r.Context(), cmd)
	if err != nil {
		switch err {
		case domain.ErrCrewMemberNotFound:
			writeError(w, http.StatusNotFound, "not_found", "crew member not found")
		case domain.ErrInvalidDateRange:
			writeError(w, http.StatusBadRequest, "invalid_date", "invalid date format")
		default:
			h.logger.Error("failed to start time record", err)
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to start time record")
		}
		return
	}

	writeJSON(w, http.StatusCreated, dto)
}

// StopTimeRecord stops a running time record
func (h *Handlers) StopTimeRecord(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing_id", "time record ID is required")
		return
	}

	var cmd application.StopTimeRecordCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}

	cmd.ID = id

	dto, err := h.timeRecordSvc.StopTimeRecord(r.Context(), cmd)
	if err != nil {
		if err == domain.ErrTimeRecordNotFound {
			writeError(w, http.StatusNotFound, "not_found", "time record not found")
		} else {
			h.logger.Error("failed to stop time record", err)
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to stop time record")
		}
		return
	}

	writeJSON(w, http.StatusOK, dto)
}

// GetTimeRecords lists time records for a crew member
func (h *Handlers) GetTimeRecords(w http.ResponseWriter, r *http.Request) {
	crewMemberID := r.URL.Query().Get("member_id")
	dateStr := r.URL.Query().Get("date")

	if crewMemberID == "" {
		writeError(w, http.StatusBadRequest, "missing_param", "member_id query parameter is required")
		return
	}

	var dtos []*application.TimeRecordDTO
	var err error

	if dateStr != "" {
		// Parse and validate date
		date, parseErr := time.Parse("2006-01-02", dateStr)
		if parseErr != nil {
			writeError(w, http.StatusBadRequest, "invalid_date", "invalid date format (expected YYYY-MM-DD)")
			return
		}
		dtos, err = h.timeRecordSvc.ListTimeRecordsByDate(r.Context(), crewMemberID, date)
	} else {
		dtos, err = h.timeRecordSvc.ListTimeRecordsByCrewMember(r.Context(), crewMemberID)
	}

	if err != nil {
		h.logger.Error("failed to list time records", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to list time records")
		return
	}

	writeJSON(w, http.StatusOK, dtos)
}

// ApproveTimeRecord approves a time record
func (h *Handlers) ApproveTimeRecord(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing_id", "time record ID is required")
		return
	}

	cmd := application.ApproveTimeRecordCommand{
		ID: id,
	}

	dto, err := h.timeRecordSvc.ApproveTimeRecord(r.Context(), cmd)
	if err != nil {
		if err == domain.ErrTimeRecordNotFound {
			writeError(w, http.StatusNotFound, "not_found", "time record not found")
		} else {
			h.logger.Error("failed to approve time record", err)
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to approve time record")
		}
		return
	}

	writeJSON(w, http.StatusOK, dto)
}

// GetDashboard returns a freelancer dashboard summary
func (h *Handlers) GetDashboard(w http.ResponseWriter, r *http.Request) {
	crewMemberID := r.URL.Query().Get("member_id")
	if crewMemberID == "" {
		writeError(w, http.StatusBadRequest, "missing_param", "member_id query parameter is required")
		return
	}

	// Get crew member
	crewDTO, err := h.crewService.GetCrewMember(r.Context(), crewMemberID)
	if err != nil {
		if err == domain.ErrCrewMemberNotFound {
			writeError(w, http.StatusNotFound, "not_found", "crew member not found")
		} else {
			h.logger.Error("failed to get crew member", err)
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to get crew member")
		}
		return
	}

	// Get qualifications
	quals, err := h.qualificationSvc.ListQualificationsForCrewMember(r.Context(), crewMemberID)
	if err != nil {
		h.logger.Error("failed to get qualifications", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to get qualifications")
		return
	}

	// Count valid qualifications
	validQuals := 0
	for _, q := range quals {
		if q.Status == "valid" {
			validQuals++
		}
	}

	// Get time records for this week
	records, err := h.timeRecordSvc.ListTimeRecordsByCrewMember(r.Context(), crewMemberID)
	if err != nil {
		h.logger.Error("failed to get time records", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to get time records")
		return
	}

	// Calculate hours and earnings for this week
	now := time.Now()
	weekStart := now.AddDate(0, 0, -int(now.Weekday()))
	weekEnd := weekStart.AddDate(0, 0, 7)

	totalHours := 0.0
	totalEarnings := 0.0
	for _, record := range records {
		recordTime, _ := time.Parse("2006-01-02", record.Date)
		if recordTime.After(weekStart) && recordTime.Before(weekEnd) {
			totalHours += record.Hours
			if crewDTO.HourlyRate != nil {
				totalEarnings += record.Hours * (*crewDTO.HourlyRate)
			}
		}
	}

	dashboard := application.DashboardDTO{
		CrewMemberID:             crewMemberID,
		Name:                     crewDTO.FirstName + " " + crewDTO.LastName,
		Role:                     crewDTO.Role,
		Status:                   crewDTO.Status,
		TotalHoursThisWeek:       totalHours,
		TotalEarningsThisWeek:    totalEarnings,
		ValidQualifications:      validQuals,
	}

	writeJSON(w, http.StatusOK, dashboard)
}
