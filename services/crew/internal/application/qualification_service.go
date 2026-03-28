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

// CreateQualificationRequest holds the data needed to create a new qualification.
type CreateQualificationRequest struct {
	Name              string `json:"name"`
	IssuedAt          string `json:"issued_at,omitempty"`
	ExpiresAt         string `json:"expires_at,omitempty"`
	CertificateNumber string `json:"certificate_number,omitempty"`
	Notes             string `json:"notes,omitempty"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// QualificationService implements the application-level use cases for crew qualifications.
type QualificationService struct {
	qualificationRepo domain.QualificationRepository
	memberRepo        domain.CrewMemberRepository
	logger            zerolog.Logger
}

// NewQualificationService constructs a new QualificationService.
func NewQualificationService(
	qualificationRepo domain.QualificationRepository,
	memberRepo domain.CrewMemberRepository,
	logger zerolog.Logger,
) *QualificationService {
	return &QualificationService{
		qualificationRepo: qualificationRepo,
		memberRepo:        memberRepo,
		logger:            logger.With().Str("service", "qualification").Logger(),
	}
}

// Create validates the request and persists a new qualification for a crew member.
func (s *QualificationService) Create(
	ctx context.Context,
	crewMemberID uuid.UUID,
	tenantID uuid.UUID,
	req CreateQualificationRequest,
) (*domain.CrewQualification, error) {
	if req.Name == "" {
		return nil, domain.ErrQualificationName
	}

	// Verify that the crew member exists.
	if _, err := s.memberRepo.GetByID(ctx, crewMemberID, tenantID); err != nil {
		return nil, fmt.Errorf("create qualification – member: %w", err)
	}

	qualification := &domain.CrewQualification{
		ID:                uuid.New(),
		CrewMemberID:      crewMemberID,
		TenantID:          tenantID,
		Name:              req.Name,
		IssuedAt:          req.IssuedAt,
		ExpiresAt:         req.ExpiresAt,
		CertificateNumber: req.CertificateNumber,
		Notes:             req.Notes,
		CreatedAt:         time.Now(),
	}

	if err := s.qualificationRepo.Create(ctx, qualification); err != nil {
		s.logger.Error().Err(err).
			Str("crew_member_id", crewMemberID.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to create qualification")
		return nil, fmt.Errorf("create qualification: %w", err)
	}

	s.logger.Info().
		Str("qualification_id", qualification.ID.String()).
		Str("crew_member_id", crewMemberID.String()).
		Str("tenant_id", tenantID.String()).
		Msg("qualification created")

	return qualification, nil
}

// ListByMember returns all qualifications for a specific crew member.
func (s *QualificationService) ListByMember(
	ctx context.Context,
	crewMemberID uuid.UUID,
	tenantID uuid.UUID,
) ([]*domain.CrewQualification, error) {
	qualifications, err := s.qualificationRepo.ListByMember(ctx, crewMemberID, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("crew_member_id", crewMemberID.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to list qualifications for member")
		return nil, fmt.Errorf("list qualifications by member: %w", err)
	}
	return qualifications, nil
}
