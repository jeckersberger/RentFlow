package domain

import (
	"context"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Filters
// ---------------------------------------------------------------------------

// PartnerFilter holds optional criteria for listing federation partners.
type PartnerFilter struct {
	Status  *string `json:"status,omitempty"`
	Page    int     `json:"page"`
	PerPage int     `json:"per_page"`
}

// ListingFilter holds optional criteria for listing shared listings.
type ListingFilter struct {
	EquipmentID *uuid.UUID `json:"equipment_id,omitempty"`
	IsActive    *bool      `json:"is_active,omitempty"`
	Page        int        `json:"page"`
	PerPage     int        `json:"per_page"`
}

// RequestFilter holds optional criteria for listing federation requests.
type RequestFilter struct {
	PartnerID *uuid.UUID `json:"partner_id,omitempty"`
	ListingID *uuid.UUID `json:"listing_id,omitempty"`
	Status    *string    `json:"status,omitempty"`
	Page      int        `json:"page"`
	PerPage   int        `json:"per_page"`
}

// ---------------------------------------------------------------------------
// Repository ports (driven / secondary adapters)
// ---------------------------------------------------------------------------

// FederationPartnerRepository defines persistence operations for FederationPartner aggregates.
type FederationPartnerRepository interface {
	Create(ctx context.Context, partner *FederationPartner) error
	GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*FederationPartner, error)
	List(ctx context.Context, tenantID uuid.UUID, filter PartnerFilter) ([]*FederationPartner, int64, error)
	Update(ctx context.Context, partner *FederationPartner) error
}

// SharedListingRepository defines persistence operations for SharedListing aggregates.
type SharedListingRepository interface {
	Create(ctx context.Context, listing *SharedListing) error
	GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*SharedListing, error)
	List(ctx context.Context, tenantID uuid.UUID, filter ListingFilter) ([]*SharedListing, int64, error)
	Update(ctx context.Context, listing *SharedListing) error
}

// FederationRequestRepository defines persistence operations for FederationRequest aggregates.
type FederationRequestRepository interface {
	Create(ctx context.Context, req *FederationRequest) error
	GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*FederationRequest, error)
	List(ctx context.Context, tenantID uuid.UUID, filter RequestFilter) ([]*FederationRequest, int64, error)
	Update(ctx context.Context, req *FederationRequest) error
}
