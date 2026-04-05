package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/crew/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

// CreateAvailabilityBlockRequest holds the data needed to create a new availability block.
type CreateAvailabilityBlockRequest struct {
	BlockType string `json:"block_type"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Notes     string `json:"notes,omitempty"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// AvailabilityBlockService implements the application-level use cases for
// crew availability blocks (vacation, sick, training, etc.).
type AvailabilityBlockService struct {
	blockRepo domain.AvailabilityBlockRepository
	memberRepo domain.CrewMemberRepository
	logger     zerolog.Logger
}

// NewAvailabilityBlockService constructs a new AvailabilityBlockService.
func NewAvailabilityBlockService(
	blockRepo domain.AvailabilityBlockRepository,
	memberRepo domain.CrewMemberRepository,
	logger zerolog.Logger,
) *AvailabilityBlockService {
	return &AvailabilityBlockService{
		blockRepo:  blockRepo,
		memberRepo: memberRepo,
		logger:     logger.With().Str("service", "availability_block").Logger(),
	}
}

// Create validates the request and persists a new availability block.
func (s *AvailabilityBlockService) Create(
	ctx context.Context,
	tenantID uuid.UUID,
	crewMemberID uuid.UUID,
	req CreateAvailabilityBlockRequest,
) (*domain.AvailabilityBlock, error) {
	// Validate block type.
	switch req.BlockType {
	case domain.BlockTypeVacation,
		domain.BlockTypeSick,
		domain.BlockTypeTraining,
		domain.BlockTypeBlocked,
		domain.BlockTypeOther:
		// valid
	default:
		return nil, domain.ErrInvalidBlockType
	}

	// Validate date formats.
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, domain.ErrInvalidDateFormat
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, domain.ErrInvalidDateFormat
	}

	// end_date must not be before start_date.
	if endDate.Before(startDate) {
		return nil, domain.ErrEndBeforeStart
	}

	// Verify crew member exists.
	if _, err := s.memberRepo.GetByID(ctx, crewMemberID, tenantID); err != nil {
		return nil, fmt.Errorf("create availability block - member lookup: %w", err)
	}

	// Check for overlapping blocks.
	existing, err := s.blockRepo.ListByCrewMember(ctx, crewMemberID, tenantID, req.StartDate, req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("create availability block - conflict check: %w", err)
	}
	if len(existing) > 0 {
		return nil, domain.ErrBlockConflict
	}

	block := &domain.AvailabilityBlock{
		ID:           uuid.New(),
		TenantID:     tenantID,
		CrewMemberID: crewMemberID,
		BlockType:    req.BlockType,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		Notes:        req.Notes,
	}

	if err := s.blockRepo.Create(ctx, block); err != nil {
		s.logger.Error().Err(err).
			Str("crew_member_id", crewMemberID.String()).
			Str("tenant_id", tenantID.String()).
			Str("start_date", req.StartDate).
			Str("end_date", req.EndDate).
			Msg("failed to create availability block")
		return nil, fmt.Errorf("create availability block: %w", err)
	}

	s.logger.Info().
		Str("block_id", block.ID.String()).
		Str("crew_member_id", crewMemberID.String()).
		Str("tenant_id", tenantID.String()).
		Str("block_type", req.BlockType).
		Str("start_date", req.StartDate).
		Str("end_date", req.EndDate).
		Msg("availability block created")

	return block, nil
}

// ListByMember returns availability blocks for a specific crew member within a date range.
func (s *AvailabilityBlockService) ListByMember(
	ctx context.Context,
	tenantID uuid.UUID,
	crewMemberID uuid.UUID,
	from string,
	to string,
) ([]*domain.AvailabilityBlock, error) {
	if _, err := time.Parse("2006-01-02", from); err != nil {
		return nil, domain.ErrInvalidDateFormat
	}
	if _, err := time.Parse("2006-01-02", to); err != nil {
		return nil, domain.ErrInvalidDateFormat
	}

	blocks, err := s.blockRepo.ListByCrewMember(ctx, crewMemberID, tenantID, from, to)
	if err != nil {
		return nil, fmt.Errorf("list member availability blocks: %w", err)
	}
	return blocks, nil
}

// ListAll returns all availability blocks for a tenant within a date range.
func (s *AvailabilityBlockService) ListAll(
	ctx context.Context,
	tenantID uuid.UUID,
	from string,
	to string,
) ([]*domain.AvailabilityBlock, error) {
	if _, err := time.Parse("2006-01-02", from); err != nil {
		return nil, domain.ErrInvalidDateFormat
	}
	if _, err := time.Parse("2006-01-02", to); err != nil {
		return nil, domain.ErrInvalidDateFormat
	}

	blocks, err := s.blockRepo.ListByDateRange(ctx, tenantID, from, to)
	if err != nil {
		return nil, fmt.Errorf("list all availability blocks: %w", err)
	}
	return blocks, nil
}

// Delete removes an availability block by ID within a tenant scope.
func (s *AvailabilityBlockService) Delete(
	ctx context.Context,
	tenantID uuid.UUID,
	blockID uuid.UUID,
) error {
	if err := s.blockRepo.Delete(ctx, blockID, tenantID); err != nil {
		s.logger.Error().Err(err).
			Str("block_id", blockID.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to delete availability block")
		return fmt.Errorf("delete availability block: %w", err)
	}

	s.logger.Info().
		Str("block_id", blockID.String()).
		Str("tenant_id", tenantID.String()).
		Msg("availability block deleted")

	return nil
}
