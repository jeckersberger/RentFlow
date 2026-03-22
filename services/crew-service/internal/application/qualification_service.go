package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/crew-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/crew-service/internal/ports"
)

// QualificationService handles qualification business logic
type QualificationService struct {
	qualificationRepo ports.QualificationRepository
	crewRepo          ports.CrewMemberRepository
	logger            logger.Logger
}

// NewQualificationService creates a new qualification service
func NewQualificationService(
	qualificationRepo ports.QualificationRepository,
	crewRepo ports.CrewMemberRepository,
	log logger.Logger,
) *QualificationService {
	return &QualificationService{
		qualificationRepo: qualificationRepo,
		crewRepo:          crewRepo,
		logger:            log,
	}
}

// CreateQualification creates a new qualification
func (s *QualificationService) CreateQualification(ctx context.Context, cmd CreateQualificationCommand) (*QualificationDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}
	if cmd.CrewMemberID == "" {
		return nil, domain.ErrCrewMemberIDRequired
	}

	// Verify crew member exists
	member, err := s.crewRepo.FindByID(ctx, cmd.CrewMemberID)
	if err != nil {
		s.logger.Error("failed to find crew member", err)
		return nil, err
	}
	if member == nil {
		return nil, domain.ErrCrewMemberNotFound
	}

	// Parse issued date
	issuedAt, err := time.Parse(time.RFC3339, cmd.IssuedAt)
	if err != nil {
		s.logger.Error("invalid issued date", err)
		return nil, domain.ErrInvalidDateRange
	}

	qual := &domain.Qualification{
		ID:                uuid.New().String(),
		TenantID:          cmd.TenantID,
		CrewMemberID:      cmd.CrewMemberID,
		QualificationType: domain.QualificationType(cmd.QualificationType),
		IssuedAt:          issuedAt,
		CertificateNumber: cmd.CertificateNumber,
		IssuingAuthority:  cmd.IssuingAuthority,
		Status:            domain.QualStatusValid,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}

	// Parse expiry date if provided
	if cmd.ExpiresAt != nil && *cmd.ExpiresAt != "" {
		expiresAt, err := time.Parse(time.RFC3339, *cmd.ExpiresAt)
		if err != nil {
			s.logger.Error("invalid expiry date", err)
			return nil, domain.ErrInvalidDateRange
		}
		qual.ExpiresAt = &expiresAt

		// Check if already expired
		if expiresAt.Before(time.Now()) {
			qual.Status = domain.QualStatusExpired
		}
	}

	if err := s.qualificationRepo.Save(ctx, qual); err != nil {
		s.logger.Error("failed to save qualification", err)
		return nil, err
	}

	return ToQualificationDTO(qual), nil
}

// GetQualification retrieves a qualification by ID
func (s *QualificationService) GetQualification(ctx context.Context, id string) (*QualificationDTO, error) {
	qual, err := s.qualificationRepo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to find qualification", err)
		return nil, err
	}
	if qual == nil {
		return nil, domain.ErrQualificationNotFound
	}
	return ToQualificationDTO(qual), nil
}

// ListQualificationsForCrewMember lists qualifications for a crew member
func (s *QualificationService) ListQualificationsForCrewMember(ctx context.Context, crewMemberID string) ([]*QualificationDTO, error) {
	quals, err := s.qualificationRepo.ListByCrewMember(ctx, crewMemberID)
	if err != nil {
		s.logger.Error("failed to list qualifications", err)
		return nil, err
	}

	dtos := make([]*QualificationDTO, len(quals))
	for i, qual := range quals {
		dtos[i] = ToQualificationDTO(qual)
	}

	return dtos, nil
}

// DeleteQualification deletes a qualification
func (s *QualificationService) DeleteQualification(ctx context.Context, id string) error {
	qual, err := s.qualificationRepo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to find qualification", err)
		return err
	}
	if qual == nil {
		return domain.ErrQualificationNotFound
	}

	if err := s.qualificationRepo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete qualification", err)
		return err
	}

	return nil
}
