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

// CreateCrewMemberRequest holds the data needed to create a new crew member.
type CreateCrewMemberRequest struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Email      string `json:"email,omitempty"`
	Phone      string `json:"phone,omitempty"`
	Role       string `json:"role,omitempty"`
	HourlyRate int64  `json:"hourly_rate,omitempty"`
	Notes      string `json:"notes,omitempty"`
}

// UpdateCrewMemberRequest holds optional fields for patching a crew member.
type UpdateCrewMemberRequest struct {
	FirstName  *string `json:"first_name,omitempty"`
	LastName   *string `json:"last_name,omitempty"`
	Email      *string `json:"email,omitempty"`
	Phone      *string `json:"phone,omitempty"`
	Role       *string `json:"role,omitempty"`
	HourlyRate *int64  `json:"hourly_rate,omitempty"`
	IsActive   *bool   `json:"is_active,omitempty"`
	Notes      *string `json:"notes,omitempty"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// CrewService implements the application-level use cases for crew members.
type CrewService struct {
	memberRepo domain.CrewMemberRepository
	logger     zerolog.Logger
}

// NewCrewService constructs a new CrewService.
func NewCrewService(
	memberRepo domain.CrewMemberRepository,
	logger zerolog.Logger,
) *CrewService {
	return &CrewService{
		memberRepo: memberRepo,
		logger:     logger.With().Str("service", "crew").Logger(),
	}
}

// Create validates the request and persists a new crew member.
func (s *CrewService) Create(
	ctx context.Context,
	tenantID uuid.UUID,
	req CreateCrewMemberRequest,
) (*domain.CrewMember, error) {
	if req.FirstName == "" || req.LastName == "" {
		return nil, domain.ErrFirstNameRequired
	}

	role := req.Role
	if role == "" {
		role = "technician"
	}

	now := time.Now()
	member := &domain.CrewMember{
		ID:         uuid.New(),
		TenantID:   tenantID,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Email:      req.Email,
		Phone:      req.Phone,
		Role:       role,
		HourlyRate: req.HourlyRate,
		IsActive:   true,
		Notes:      req.Notes,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.memberRepo.Create(ctx, member); err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to create crew member")
		return nil, fmt.Errorf("create crew member: %w", err)
	}

	s.logger.Info().
		Str("crew_member_id", member.ID.String()).
		Str("tenant_id", tenantID.String()).
		Msg("crew member created")

	return member, nil
}

// GetByID retrieves a single crew member by ID within a tenant scope.
func (s *CrewService) GetByID(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
) (*domain.CrewMember, error) {
	member, err := s.memberRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("crew_member_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to get crew member by id")
		return nil, fmt.Errorf("get crew member by id: %w", err)
	}
	return member, nil
}

// List returns all crew members for a tenant.
func (s *CrewService) List(
	ctx context.Context,
	tenantID uuid.UUID,
) ([]*domain.CrewMember, error) {
	members, err := s.memberRepo.List(ctx, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to list crew members")
		return nil, fmt.Errorf("list crew members: %w", err)
	}
	return members, nil
}

// Update patches an existing crew member with non-nil fields from the request.
func (s *CrewService) Update(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
	req UpdateCrewMemberRequest,
) (*domain.CrewMember, error) {
	existing, err := s.memberRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("crew_member_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to fetch crew member for update")
		return nil, fmt.Errorf("update crew member – fetch: %w", err)
	}

	if req.FirstName != nil {
		existing.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		existing.LastName = *req.LastName
	}
	if req.Email != nil {
		existing.Email = *req.Email
	}
	if req.Phone != nil {
		existing.Phone = *req.Phone
	}
	if req.Role != nil {
		existing.Role = *req.Role
	}
	if req.HourlyRate != nil {
		existing.HourlyRate = *req.HourlyRate
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}
	if req.Notes != nil {
		existing.Notes = *req.Notes
	}

	existing.UpdatedAt = time.Now()

	if err := s.memberRepo.Update(ctx, existing); err != nil {
		s.logger.Error().Err(err).
			Str("crew_member_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to update crew member")
		return nil, fmt.Errorf("update crew member: %w", err)
	}

	s.logger.Info().
		Str("crew_member_id", id.String()).
		Str("tenant_id", tenantID.String()).
		Msg("crew member updated")

	return existing, nil
}

// Delete removes a crew member by ID within a tenant scope.
func (s *CrewService) Delete(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
) error {
	if err := s.memberRepo.Delete(ctx, id, tenantID); err != nil {
		s.logger.Error().Err(err).
			Str("crew_member_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to delete crew member")
		return fmt.Errorf("delete crew member: %w", err)
	}

	s.logger.Info().
		Str("crew_member_id", id.String()).
		Str("tenant_id", tenantID.String()).
		Msg("crew member deleted")

	return nil
}
