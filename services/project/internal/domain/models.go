package domain

import (
	"time"

	"github.com/google/uuid"
)

// Project status constants
const (
	ProjectStatusDraft     = "draft"
	ProjectStatusConfirmed = "confirmed"
	ProjectStatusActive    = "active"
	ProjectStatusCompleted = "completed"
	ProjectStatusCancelled = "cancelled"
	ProjectStatusArchived  = "archived"
)

var validProjectStatuses = map[string]bool{
	ProjectStatusDraft:     true,
	ProjectStatusConfirmed: true,
	ProjectStatusActive:    true,
	ProjectStatusCompleted: true,
	ProjectStatusCancelled: true,
	ProjectStatusArchived:  true,
}

// Project Equipment status constants
const (
	PEStatusPlanned    = "planned"
	PEStatusConfirmed  = "confirmed"
	PEStatusCheckedOut = "checked_out"
	PEStatusReturned   = "returned"
)

var validPEStatuses = map[string]bool{
	PEStatusPlanned:    true,
	PEStatusConfirmed:  true,
	PEStatusCheckedOut: true,
	PEStatusReturned:   true,
}

// Packlist status constants
const (
	PacklistStatusDraft    = "draft"
	PacklistStatusPacking  = "packing"
	PacklistStatusComplete = "complete"
)

// Reservation status constants
const (
	ReservationStatusPending   = "pending"
	ReservationStatusConfirmed = "confirmed"
	ReservationStatusCancelled = "cancelled"
)

// Project represents a rental project with all its metadata.
type Project struct {
	ID            uuid.UUID  `json:"id"`
	TenantID      uuid.UUID  `json:"tenant_id"`
	Name          string     `json:"name"`
	ProjectNumber string     `json:"project_number"`
	Description   string     `json:"description"`
	Status        string     `json:"status"`
	CustomerID    *uuid.UUID `json:"customer_id,omitempty"`
	ContactName   string     `json:"contact_name"`
	ContactEmail  string     `json:"contact_email"`
	ContactPhone  string     `json:"contact_phone"`
	VenueName     string     `json:"venue_name"`
	VenueAddress  string     `json:"venue_address"`
	VenueLat      *float64   `json:"venue_lat,omitempty"`
	VenueLng      *float64   `json:"venue_lng,omitempty"`
	StartDate     *time.Time `json:"start_date,omitempty"`
	EndDate       *time.Time `json:"end_date,omitempty"`
	SetupDate     *time.Time `json:"setup_date,omitempty"`
	TeardownDate  *time.Time `json:"teardown_date,omitempty"`
	Color         string     `json:"color"`
	Budget        int64      `json:"budget"`
	Currency      string     `json:"currency"`
	ManagerID     *uuid.UUID `json:"manager_id,omitempty"`
	Notes         string     `json:"notes"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// ValidateStatus checks whether the given status string is a valid project status.
func (p *Project) ValidateStatus(s string) bool {
	return validProjectStatuses[s]
}

// ProjectEquipment represents an equipment assignment to a project.
type ProjectEquipment struct {
	ID             uuid.UUID  `json:"id"`
	ProjectID      uuid.UUID  `json:"project_id"`
	EquipmentID    uuid.UUID  `json:"equipment_id"`
	Quantity       int        `json:"quantity"`
	AllocatedFrom  *time.Time `json:"allocated_from,omitempty"`
	AllocatedUntil *time.Time `json:"allocated_until,omitempty"`
	Status         string     `json:"status"`
	Notes          string     `json:"notes"`
	CreatedAt      time.Time  `json:"created_at"`
}

// Packlist represents a packing list for a project.
type Packlist struct {
	ID        uuid.UUID  `json:"id"`
	ProjectID uuid.UUID  `json:"project_id"`
	TenantID  uuid.UUID  `json:"tenant_id"`
	Name      string     `json:"name"`
	Status    string     `json:"status"`
	CreatedBy *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// PacklistItem represents a single item within a packlist.
type PacklistItem struct {
	ID               uuid.UUID  `json:"id"`
	PacklistID       uuid.UUID  `json:"packlist_id"`
	EquipmentID      uuid.UUID  `json:"equipment_id"`
	QuantityPlanned  int        `json:"quantity_planned"`
	QuantityPacked   int        `json:"quantity_packed"`
	QuantityReturned int        `json:"quantity_returned"`
	PackedBy         *uuid.UUID `json:"packed_by,omitempty"`
	PackedAt         *time.Time `json:"packed_at,omitempty"`
	Notes            string     `json:"notes"`
}

// Reservation represents an equipment reservation for a project.
type Reservation struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	ProjectID   uuid.UUID `json:"project_id"`
	EquipmentID uuid.UUID `json:"equipment_id"`
	Quantity    int       `json:"quantity"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// ProjectFilter holds filter/pagination parameters for listing projects.
type ProjectFilter struct {
	Page       int        `json:"page"`
	PerPage    int        `json:"per_page"`
	Status     *string    `json:"status,omitempty"`
	Search     *string    `json:"search,omitempty"`
	CustomerID *uuid.UUID `json:"customer_id,omitempty"`
	StartAfter *time.Time `json:"start_after,omitempty"`
	EndBefore  *time.Time `json:"end_before,omitempty"`
}
