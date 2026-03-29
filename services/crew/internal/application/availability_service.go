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

// SetAvailabilityRequest holds the data needed to set availability for a date.
type SetAvailabilityRequest struct {
	CrewMemberID uuid.UUID `json:"crew_member_id"`
	Date         string    `json:"date"`
	Status       string    `json:"status"`
	Note         string    `json:"note,omitempty"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// AvailabilityService implements the application-level use cases for crew availability.
type AvailabilityService struct {
	availabilityRepo domain.AvailabilityRepository
	memberRepo       domain.CrewMemberRepository
	logger           zerolog.Logger
}

// NewAvailabilityService constructs a new AvailabilityService.
func NewAvailabilityService(
	availabilityRepo domain.AvailabilityRepository,
	memberRepo domain.CrewMemberRepository,
	logger zerolog.Logger,
) *AvailabilityService {
	return &AvailabilityService{
		availabilityRepo: availabilityRepo,
		memberRepo:       memberRepo,
		logger:           logger.With().Str("service", "availability").Logger(),
	}
}

// SetAvailability creates or updates an availability entry for a member on a date.
func (s *AvailabilityService) SetAvailability(
	ctx context.Context,
	tenantID uuid.UUID,
	req SetAvailabilityRequest,
) (*domain.CrewAvailability, error) {
	if req.CrewMemberID == uuid.Nil {
		return nil, domain.ErrCrewMemberIDRequired
	}

	// Validate date format.
	if _, err := time.Parse("2006-01-02", req.Date); err != nil {
		return nil, domain.ErrInvalidDateFormat
	}

	// Validate status.
	switch req.Status {
	case domain.AvailabilityStatusAvailable,
		domain.AvailabilityStatusUnavailable,
		domain.AvailabilityStatusOnRequest:
		// valid
	default:
		return nil, domain.ErrInvalidAvailStatus
	}

	// Verify crew member exists.
	if _, err := s.memberRepo.GetByID(ctx, req.CrewMemberID, tenantID); err != nil {
		return nil, fmt.Errorf("set availability – member lookup: %w", err)
	}

	availability := &domain.CrewAvailability{
		ID:           uuid.New(),
		TenantID:     tenantID,
		CrewMemberID: req.CrewMemberID,
		Date:         req.Date,
		Status:       req.Status,
		Note:         req.Note,
	}

	if err := s.availabilityRepo.Upsert(ctx, availability); err != nil {
		s.logger.Error().Err(err).
			Str("crew_member_id", req.CrewMemberID.String()).
			Str("tenant_id", tenantID.String()).
			Str("date", req.Date).
			Msg("failed to set availability")
		return nil, fmt.Errorf("set availability: %w", err)
	}

	s.logger.Info().
		Str("availability_id", availability.ID.String()).
		Str("crew_member_id", req.CrewMemberID.String()).
		Str("tenant_id", tenantID.String()).
		Str("date", req.Date).
		Str("status", req.Status).
		Msg("availability set")

	return availability, nil
}

// ListAll returns all availability entries for a tenant within a date range.
func (s *AvailabilityService) ListAll(
	ctx context.Context,
	tenantID uuid.UUID,
	from string,
	to string,
) ([]*domain.CrewAvailability, error) {
	// Validate date formats.
	if _, err := time.Parse("2006-01-02", from); err != nil {
		return nil, domain.ErrInvalidDateFormat
	}
	if _, err := time.Parse("2006-01-02", to); err != nil {
		return nil, domain.ErrInvalidDateFormat
	}

	entries, err := s.availabilityRepo.ListAll(ctx, tenantID, from, to)
	if err != nil {
		return nil, fmt.Errorf("list all availability: %w", err)
	}
	return entries, nil
}

// ListByMember returns availability entries for a specific crew member.
func (s *AvailabilityService) ListByMember(
	ctx context.Context,
	tenantID uuid.UUID,
	memberID uuid.UUID,
	from string,
	to string,
) ([]*domain.CrewAvailability, error) {
	// Validate date formats.
	if _, err := time.Parse("2006-01-02", from); err != nil {
		return nil, domain.ErrInvalidDateFormat
	}
	if _, err := time.Parse("2006-01-02", to); err != nil {
		return nil, domain.ErrInvalidDateFormat
	}

	entries, err := s.availabilityRepo.ListByMember(ctx, memberID, tenantID, from, to)
	if err != nil {
		return nil, fmt.Errorf("list member availability: %w", err)
	}
	return entries, nil
}
