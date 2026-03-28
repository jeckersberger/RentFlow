package domain

import (
	"time"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Status constants
// ---------------------------------------------------------------------------

const (
	PartnerStatusPending  = "pending"
	PartnerStatusActive   = "active"
	PartnerStatusInactive = "inactive"
	PartnerStatusBlocked  = "blocked"
)

const (
	RequestStatusPending  = "pending"
	RequestStatusApproved = "approved"
	RequestStatusRejected = "rejected"
	RequestStatusCanceled = "canceled"
)

// ---------------------------------------------------------------------------
// Domain models
// ---------------------------------------------------------------------------

// FederationPartner represents a cross-company partner for equipment sharing.
type FederationPartner struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	PartnerName string    `json:"partner_name"`
	PartnerURL  string    `json:"partner_url,omitempty"`
	APIKeyHash  string    `json:"-"`
	Status      string    `json:"status"`
	Notes       string    `json:"notes,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SharedListing represents an equipment item shared for cross-company rental.
type SharedListing struct {
	ID             uuid.UUID  `json:"id"`
	TenantID       uuid.UUID  `json:"tenant_id"`
	EquipmentID    uuid.UUID  `json:"equipment_id"`
	DailyRate      int64      `json:"daily_rate"`
	WeeklyRate     int64      `json:"weekly_rate"`
	AvailableFrom  *time.Time `json:"available_from,omitempty"`
	AvailableUntil *time.Time `json:"available_until,omitempty"`
	IsActive       bool       `json:"is_active"`
	Notes          string     `json:"notes,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// FederationRequest represents a request from a partner to rent shared equipment.
type FederationRequest struct {
	ID          uuid.UUID  `json:"id"`
	TenantID    uuid.UUID  `json:"tenant_id"`
	PartnerID   *uuid.UUID `json:"partner_id,omitempty"`
	ListingID   *uuid.UUID `json:"listing_id,omitempty"`
	Status      string     `json:"status"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	TotalCost   int64      `json:"total_cost"`
	Notes       string     `json:"notes,omitempty"`
	RequestedBy *uuid.UUID `json:"requested_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
