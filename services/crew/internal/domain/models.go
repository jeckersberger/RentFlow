package domain

import (
	"time"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Assignment Status constants
// ---------------------------------------------------------------------------

const (
	AssignmentStatusPlanned   = "planned"
	AssignmentStatusConfirmed = "confirmed"
	AssignmentStatusActive    = "active"
	AssignmentStatusCompleted = "completed"
	AssignmentStatusCancelled = "cancelled"
)

// ---------------------------------------------------------------------------
// Domain models
// ---------------------------------------------------------------------------

// CrewMember represents a staff member who can be assigned to projects.
type CrewMember struct {
	ID         uuid.UUID `json:"id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Email      string    `json:"email,omitempty"`
	Phone      string    `json:"phone,omitempty"`
	Role       string    `json:"role"`
	HourlyRate int64     `json:"hourly_rate"`
	IsActive   bool      `json:"is_active"`
	Notes      string    `json:"notes,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CrewAssignment represents a crew member's assignment to a project.
type CrewAssignment struct {
	ID           uuid.UUID `json:"id"`
	CrewMemberID uuid.UUID `json:"crew_member_id"`
	ProjectID    uuid.UUID `json:"project_id"`
	TenantID     uuid.UUID `json:"tenant_id"`
	Role         string    `json:"role,omitempty"`
	StartDate    string    `json:"start_date,omitempty"`
	EndDate      string    `json:"end_date,omitempty"`
	HoursPlanned float64  `json:"hours_planned"`
	HoursActual  float64  `json:"hours_actual"`
	Status       string    `json:"status"`
	Notes        string    `json:"notes,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// CrewQualification represents a qualification or certification held by a crew member.
type CrewQualification struct {
	ID                uuid.UUID `json:"id"`
	CrewMemberID      uuid.UUID `json:"crew_member_id"`
	TenantID          uuid.UUID `json:"tenant_id"`
	Name              string    `json:"name"`
	IssuedAt          string    `json:"issued_at,omitempty"`
	ExpiresAt         string    `json:"expires_at,omitempty"`
	CertificateNumber string   `json:"certificate_number,omitempty"`
	Notes             string    `json:"notes,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}
