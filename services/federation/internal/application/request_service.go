package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/federation/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

// CreateFederationRequestRequest holds the data needed to create a federation request.
type CreateFederationRequestRequest struct {
	PartnerID *uuid.UUID `json:"partner_id,omitempty"`
	ListingID *uuid.UUID `json:"listing_id,omitempty"`
	Status    string     `json:"status,omitempty"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	TotalCost int64      `json:"total_cost"`
	Notes     string     `json:"notes,omitempty"`
}

// UpdateFederationRequestRequest holds the data needed to update a federation request.
type UpdateFederationRequestRequest struct {
	PartnerID *uuid.UUID `json:"partner_id,omitempty"`
	ListingID *uuid.UUID `json:"listing_id,omitempty"`
	Status    string     `json:"status,omitempty"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	TotalCost int64      `json:"total_cost"`
	Notes     string     `json:"notes,omitempty"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// RequestService implements the application-level use cases for federation requests.
type RequestService struct {
	requestRepo domain.FederationRequestRepository
	logger      zerolog.Logger
}

// NewRequestService constructs a new RequestService.
func NewRequestService(
	requestRepo domain.FederationRequestRepository,
	logger zerolog.Logger,
) *RequestService {
	return &RequestService{
		requestRepo: requestRepo,
		logger:      logger.With().Str("service", "request").Logger(),
	}
}

// Create creates a new federation request.
func (s *RequestService) Create(
	ctx context.Context,
	tenantID uuid.UUID,
	userID uuid.UUID,
	req CreateFederationRequestRequest,
) (*domain.FederationRequest, error) {
	status := req.Status
	if status == "" {
		status = domain.RequestStatusPending
	}

	fedReq := &domain.FederationRequest{
		ID:          uuid.New(),
		TenantID:    tenantID,
		PartnerID:   req.PartnerID,
		ListingID:   req.ListingID,
		Status:      status,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		TotalCost:   req.TotalCost,
		Notes:       req.Notes,
		RequestedBy: &userID,
	}

	if err := s.requestRepo.Create(ctx, fedReq); err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to create federation request")
		return nil, fmt.Errorf("create request: %w", err)
	}

	s.logger.Info().
		Str("request_id", fedReq.ID.String()).
		Str("tenant_id", tenantID.String()).
		Msg("federation request created")

	return fedReq, nil
}

// GetByID retrieves a federation request by ID.
func (s *RequestService) GetByID(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
) (*domain.FederationRequest, error) {
	fedReq, err := s.requestRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("request_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to get federation request")
		return nil, fmt.Errorf("get request: %w", err)
	}
	return fedReq, nil
}

// List returns a filtered, paginated list of federation requests.
func (s *RequestService) List(
	ctx context.Context,
	tenantID uuid.UUID,
	filter domain.RequestFilter,
) ([]*domain.FederationRequest, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 {
		filter.PerPage = 20
	}

	items, total, err := s.requestRepo.List(ctx, tenantID, filter)
	if err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to list federation requests")
		return nil, 0, fmt.Errorf("list requests: %w", err)
	}
	return items, total, nil
}

// Update updates an existing federation request.
func (s *RequestService) Update(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
	req UpdateFederationRequestRequest,
) (*domain.FederationRequest, error) {
	existing, err := s.requestRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, fmt.Errorf("update request: %w", err)
	}

	existing.PartnerID = req.PartnerID
	existing.ListingID = req.ListingID
	if req.Status != "" {
		existing.Status = req.Status
	}
	existing.StartDate = req.StartDate
	existing.EndDate = req.EndDate
	existing.TotalCost = req.TotalCost
	existing.Notes = req.Notes

	if err := s.requestRepo.Update(ctx, existing); err != nil {
		s.logger.Error().Err(err).
			Str("request_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to update federation request")
		return nil, fmt.Errorf("update request: %w", err)
	}

	s.logger.Info().
		Str("request_id", id.String()).
		Str("tenant_id", tenantID.String()).
		Msg("federation request updated")

	return existing, nil
}
