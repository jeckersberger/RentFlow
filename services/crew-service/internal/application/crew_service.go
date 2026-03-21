package application

import (
	"context"
	"fmt"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/crew-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/crew-service/internal/ports"
)

type CrewService struct {
	crewRepo ports.CrewRepository
	logger   *logger.Logger
}

func NewCrewService(crewRepo ports.CrewRepository, logger *logger.Logger) *CrewService {
	return &CrewService{
		crewRepo: crewRepo,
		logger:   logger,
	}
}

func (s *CrewService) CreateCrewMember(ctx context.Context, cmd CreateCrewMemberCommand) (*CrewMemberDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if cmd.Email == "" {
		return nil, domain.NewDomainError("EMAIL_REQUIRED", "email is required", nil)
	}

	cmID := fmt.Sprintf("crew_%d", hashString(cmd.TenantID+cmd.Email))
	crewType := domain.CrewMemberType(cmd.Type)

	cm := domain.NewCrewMember(cmID, cmd.TenantID, cmd.UserID, cmd.FirstName, cmd.LastName, cmd.Email, cmd.Phone, crewType)

	if err := cm.Validate(); err != nil {
		return nil, domain.NewDomainError("VALIDATION_ERROR", err.Error(), nil)
	}

	if err := s.crewRepo.CreateCrewMember(ctx, cm); err != nil {
		return nil, domain.NewDomainError("CREATE_FAILED", "failed to create crew member", err)
	}

	return CrewMemberToDTO(cm), nil
}

func (s *CrewService) GetCrewMember(ctx context.Context, tenantID, crewMemberID string) (*CrewMemberDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	cm, err := s.crewRepo.GetCrewMember(ctx, tenantID, crewMemberID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "crew member not found", err)
	}

	return CrewMemberToDTO(cm), nil
}

func (s *CrewService) ListCrewMembers(ctx context.Context, tenantID string, limit, offset int) ([]*CrewMemberDTO, int64, error) {
	if tenantID == "" {
		return nil, 0, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	members, total, err := s.crewRepo.ListCrewMembers(ctx, tenantID, limit, offset)
	if err != nil {
		return nil, 0, domain.NewDomainError("LIST_FAILED", "failed to list crew members", err)
	}

	dtos := make([]*CrewMemberDTO, len(members))
	for i, m := range members {
		dtos[i] = CrewMemberToDTO(m)
	}

	return dtos, total, nil
}

func (s *CrewService) UpdateCrewMember(ctx context.Context, cmd UpdateCrewMemberCommand) (*CrewMemberDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	cm, err := s.crewRepo.GetCrewMember(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "crew member not found", err)
	}

	if cmd.FirstName != "" {
		cm.FirstName = cmd.FirstName
	}
	if cmd.LastName != "" {
		cm.LastName = cmd.LastName
	}
	if cmd.Email != "" {
		cm.Email = cmd.Email
	}
	if cmd.Phone != "" {
		cm.Phone = cmd.Phone
	}
	if cmd.HourlyRate > 0 {
		cm.HourlyRate = cmd.HourlyRate
	}
	if cmd.DailyRate > 0 {
		cm.DailyRate = cmd.DailyRate
	}

	if err := s.crewRepo.UpdateCrewMember(ctx, cm); err != nil {
		return nil, domain.NewDomainError("UPDATE_FAILED", "failed to update crew member", err)
	}

	return CrewMemberToDTO(cm), nil
}

func (s *CrewService) DeleteCrewMember(ctx context.Context, tenantID, crewMemberID string) error {
	if tenantID == "" {
		return domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	if err := s.crewRepo.DeleteCrewMember(ctx, tenantID, crewMemberID); err != nil {
		return domain.NewDomainError("DELETE_FAILED", "failed to delete crew member", err)
	}

	return nil
}
