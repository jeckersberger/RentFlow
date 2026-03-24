package application

import (
	"time"

	"github.com/jeckersberger/rentflow/services/crew-service/internal/domain"
)

// CrewMemberDTO represents a crew member in responses
type CrewMemberDTO struct {
	ID                 string   `json:"id"`
	TenantID           string   `json:"tenant_id"`
	FirstName          string   `json:"first_name"`
	LastName           string   `json:"last_name"`
	Email              string   `json:"email"`
	Phone              string   `json:"phone"`
	Role               string   `json:"role"`
	Status             string   `json:"status"`
	HourlyRate         *float64 `json:"hourly_rate,omitempty"`
	DailyRate          *float64 `json:"daily_rate,omitempty"`
	PreferredVehicleID *string  `json:"preferred_vehicle_id,omitempty"`
	EmergencyContact   string   `json:"emergency_contact"`
	Notes              string   `json:"notes"`
	CreatedAt          string   `json:"created_at"`
	UpdatedAt          string   `json:"updated_at"`
}

// QualificationDTO represents a qualification in responses
type QualificationDTO struct {
	ID                string `json:"id"`
	CrewMemberID      string `json:"crew_member_id"`
	QualificationType string `json:"qualification_type"`
	IssuedAt          string `json:"issued_at"`
	ExpiresAt         *string `json:"expires_at,omitempty"`
	CertificateNumber string `json:"certificate_number"`
	IssuingAuthority  string `json:"issuing_authority"`
	Status            string `json:"status"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

// CrewAssignmentDTO represents a crew assignment in responses
type CrewAssignmentDTO struct {
	ID           string `json:"id"`
	TenantID     string `json:"tenant_id"`
	CrewMemberID string `json:"crew_member_id"`
	ProjectID    *string `json:"project_id,omitempty"`
	TourID       *string `json:"tour_id,omitempty"`
	Role         string `json:"role"`
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date"`
	Status       string `json:"status"`
	Notes        string `json:"notes"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// TimeRecordDTO represents a time record in responses
type TimeRecordDTO struct {
	ID              string `json:"id"`
	TenantID        string `json:"tenant_id"`
	CrewMemberID    string `json:"crew_member_id"`
	AssignmentID    *string `json:"assignment_id,omitempty"`
	Date            string `json:"date"`
	StartTime       string `json:"start_time"`
	EndTime         *string `json:"end_time,omitempty"`
	BreakMinutes    int `json:"break_minutes"`
	OvertimeMinutes int `json:"overtime_minutes"`
	Status          string `json:"status"`
	Notes           string `json:"notes"`
	Hours           float64 `json:"hours,omitempty"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

// AvailabilityDTO represents crew member availability
type AvailabilityDTO struct {
	CrewMemberID string `json:"crew_member_id"`
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date"`
	IsAvailable  bool `json:"is_available"`
}

// ConflictDTO represents an assignment conflict
type ConflictDTO struct {
	CrewMemberID     string `json:"crew_member_id"`
	CrewMemberName   string `json:"crew_member_name"`
	ConflictingDates string `json:"conflicting_dates"`
	ExistingAssignment *CrewAssignmentDTO `json:"existing_assignment,omitempty"`
}

// DashboardDTO represents freelancer dashboard summary
type DashboardDTO struct {
	CrewMemberID      string `json:"crew_member_id"`
	Name              string `json:"name"`
	Role              string `json:"role"`
	Status            string `json:"status"`
	TotalHoursThisWeek float64 `json:"total_hours_this_week"`
	TotalEarningsThisWeek float64 `json:"total_earnings_this_week"`
	UpcomingAssignments int `json:"upcoming_assignments"`
	ValidQualifications int `json:"valid_qualifications"`
}

// DriverDTO represents a driver with vehicle assignment info
type DriverDTO struct {
	MemberID           string     `json:"member_id"`
	Name               string     `json:"name"`
	PreferredVehicleID *string    `json:"preferred_vehicle_id,omitempty"`
	AssignmentStartDate *time.Time `json:"assignment_start_date,omitempty"`
	AssignmentEndDate   *time.Time `json:"assignment_end_date,omitempty"`
}

// BookingRequestDTO represents a booking request in responses
type BookingRequestDTO struct {
	ID              string  `json:"id"`
	AssignmentID    string  `json:"assignment_id"`
	CrewMemberID    string  `json:"crew_member_id"`
	CrewMemberName  string  `json:"crew_member_name"`
	ProjectID       string  `json:"project_id"`
	Token           string  `json:"token"`
	Status          string  `json:"status"`
	ResponseMessage *string `json:"response_message,omitempty"`
	RespondedAt     *string `json:"responded_at,omitempty"`
	CreatedAt       string  `json:"created_at"`
}

// BookingDetailsDTO represents booking details for the public response page
type BookingDetailsDTO struct {
	ProjectName    string `json:"project_name"`
	ProjectDates   string `json:"project_dates"`
	Role           string `json:"role"`
	Location       string `json:"location"`
	Message        string `json:"message"`
	FreelancerName string `json:"freelancer_name"`
	Status         string `json:"status"`
}

// PaginatedResult represents a paginated response
type PaginatedResult struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	Total      int         `json:"total"`
	TotalPages int         `json:"total_pages"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Convert functions

// ToCrewMemberDTO converts a domain CrewMember to DTO
func ToCrewMemberDTO(m *domain.CrewMember) *CrewMemberDTO {
	if m == nil {
		return nil
	}
	return &CrewMemberDTO{
		ID:                 m.ID,
		TenantID:           m.TenantID,
		FirstName:          m.FirstName,
		LastName:           m.LastName,
		Email:              m.Email,
		Phone:              m.Phone,
		Role:               string(m.Role),
		Status:             string(m.Status),
		HourlyRate:         m.HourlyRate,
		DailyRate:          m.DailyRate,
		PreferredVehicleID: m.PreferredVehicleID,
		EmergencyContact:   m.EmergencyContact,
		Notes:              m.Notes,
		CreatedAt:          m.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          m.UpdatedAt.Format(time.RFC3339),
	}
}

// ToQualificationDTO converts a domain Qualification to DTO
func ToQualificationDTO(q *domain.Qualification) *QualificationDTO {
	if q == nil {
		return nil
	}
	expiresAt := (*string)(nil)
	if q.ExpiresAt != nil {
		expiresAtStr := q.ExpiresAt.Format(time.RFC3339)
		expiresAt = &expiresAtStr
	}
	return &QualificationDTO{
		ID:                q.ID,
		CrewMemberID:      q.CrewMemberID,
		QualificationType: string(q.QualificationType),
		IssuedAt:          q.IssuedAt.Format(time.RFC3339),
		ExpiresAt:         expiresAt,
		CertificateNumber: q.CertificateNumber,
		IssuingAuthority:  q.IssuingAuthority,
		Status:            string(q.Status),
		CreatedAt:         q.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         q.UpdatedAt.Format(time.RFC3339),
	}
}

// ToCrewAssignmentDTO converts a domain CrewAssignment to DTO
func ToCrewAssignmentDTO(a *domain.CrewAssignment) *CrewAssignmentDTO {
	if a == nil {
		return nil
	}
	return &CrewAssignmentDTO{
		ID:           a.ID,
		TenantID:     a.TenantID,
		CrewMemberID: a.CrewMemberID,
		ProjectID:    a.ProjectID,
		TourID:       a.TourID,
		Role:         a.Role,
		StartDate:    a.StartDate.Format(time.RFC3339),
		EndDate:      a.EndDate.Format(time.RFC3339),
		Status:       string(a.Status),
		Notes:        a.Notes,
		CreatedAt:    a.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    a.UpdatedAt.Format(time.RFC3339),
	}
}

// ToTimeRecordDTO converts a domain TimeRecord to DTO
func ToTimeRecordDTO(t *domain.TimeRecord) *TimeRecordDTO {
	if t == nil {
		return nil
	}
	endTime := (*string)(nil)
	if t.EndTime != nil {
		endTimeStr := t.EndTime.Format(time.RFC3339)
		endTime = &endTimeStr
	}
	return &TimeRecordDTO{
		ID:              t.ID,
		TenantID:        t.TenantID,
		CrewMemberID:    t.CrewMemberID,
		AssignmentID:    t.AssignmentID,
		Date:            t.Date.Format("2006-01-02"),
		StartTime:       t.StartTime.Format(time.RFC3339),
		EndTime:         endTime,
		BreakMinutes:    t.BreakMinutes,
		OvertimeMinutes: t.OvertimeMinutes,
		Status:          string(t.Status),
		Notes:           t.Notes,
		Hours:           t.CalculateHours(),
		CreatedAt:       t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       t.UpdatedAt.Format(time.RFC3339),
	}
}
