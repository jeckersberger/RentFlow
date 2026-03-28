package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/federation/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

// CreatePartnerRequest holds the data needed to create a federation partner.
type CreatePartnerRequest struct {
	PartnerName string `json:"partner_name"`
	PartnerURL  string `json:"partner_url,omitempty"`
	APIKeyHash  string `json:"api_key_hash,omitempty"`
	Status      string `json:"status,omitempty"`
	Notes       string `json:"notes,omitempty"`
}

// UpdatePartnerRequest holds the data needed to update a federation partner.
type UpdatePartnerRequest struct {
	PartnerName string `json:"partner_name"`
	PartnerURL  string `json:"partner_url,omitempty"`
	APIKeyHash  string `json:"api_key_hash,omitempty"`
	Status      string `json:"status,omitempty"`
	Notes       string `json:"notes,omitempty"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// PartnerService implements the application-level use cases for federation partners.
type PartnerService struct {
	partnerRepo domain.FederationPartnerRepository
	logger      zerolog.Logger
}

// NewPartnerService constructs a new PartnerService.
func NewPartnerService(
	partnerRepo domain.FederationPartnerRepository,
	logger zerolog.Logger,
) *PartnerService {
	return &PartnerService{
		partnerRepo: partnerRepo,
		logger:      logger.With().Str("service", "partner").Logger(),
	}
}

// Create creates a new federation partner.
func (s *PartnerService) Create(
	ctx context.Context,
	tenantID uuid.UUID,
	req CreatePartnerRequest,
) (*domain.FederationPartner, error) {
	if req.PartnerName == "" {
		return nil, domain.ErrMissingPartnerName
	}

	status := req.Status
	if status == "" {
		status = domain.PartnerStatusPending
	}

	partner := &domain.FederationPartner{
		ID:          uuid.New(),
		TenantID:    tenantID,
		PartnerName: req.PartnerName,
		PartnerURL:  req.PartnerURL,
		APIKeyHash:  req.APIKeyHash,
		Status:      status,
		Notes:       req.Notes,
	}

	if err := s.partnerRepo.Create(ctx, partner); err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to create federation partner")
		return nil, fmt.Errorf("create partner: %w", err)
	}

	s.logger.Info().
		Str("partner_id", partner.ID.String()).
		Str("tenant_id", tenantID.String()).
		Msg("federation partner created")

	return partner, nil
}

// GetByID retrieves a federation partner by ID.
func (s *PartnerService) GetByID(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
) (*domain.FederationPartner, error) {
	partner, err := s.partnerRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("partner_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to get federation partner")
		return nil, fmt.Errorf("get partner: %w", err)
	}
	return partner, nil
}

// List returns a filtered, paginated list of federation partners.
func (s *PartnerService) List(
	ctx context.Context,
	tenantID uuid.UUID,
	filter domain.PartnerFilter,
) ([]*domain.FederationPartner, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 {
		filter.PerPage = 20
	}

	items, total, err := s.partnerRepo.List(ctx, tenantID, filter)
	if err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to list federation partners")
		return nil, 0, fmt.Errorf("list partners: %w", err)
	}
	return items, total, nil
}

// Update updates an existing federation partner.
func (s *PartnerService) Update(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
	req UpdatePartnerRequest,
) (*domain.FederationPartner, error) {
	existing, err := s.partnerRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, fmt.Errorf("update partner: %w", err)
	}

	if req.PartnerName != "" {
		existing.PartnerName = req.PartnerName
	}
	existing.PartnerURL = req.PartnerURL
	existing.APIKeyHash = req.APIKeyHash
	if req.Status != "" {
		existing.Status = req.Status
	}
	existing.Notes = req.Notes

	if err := s.partnerRepo.Update(ctx, existing); err != nil {
		s.logger.Error().Err(err).
			Str("partner_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to update federation partner")
		return nil, fmt.Errorf("update partner: %w", err)
	}

	s.logger.Info().
		Str("partner_id", id.String()).
		Str("tenant_id", tenantID.String()).
		Msg("federation partner updated")

	return existing, nil
}
