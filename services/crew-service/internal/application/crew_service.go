package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/crew-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/crew-service/internal/ports"
)

// CrewService handles crew member business logic
type CrewService struct {
	crewRepo          ports.CrewMemberRepository
	qualificationRepo ports.QualificationRepository
	assignmentRepo    ports.CrewAssignmentRepository
	logger            logger.Logger
}

// NewCrewService creates a new crew service
func NewCrewService(
	crewRepo ports.CrewMemberRepository,
	qualificationRepo ports.QualificationRepository,
	assignmentRepo ports.CrewAssignmentRepository,
	log logger.Logger,
) *CrewService {
	return &CrewService{
		crewRepo:          crewRepo,
		qualificationRepo: qualificationRepo,
		assignmentRepo:    assignmentRepo,
		logger:            log,
	}
}

// CreateCrewMember creates a new crew member
func (s *CrewService) CreateCrewMember(ctx context.Context, cmd CreateCrewMemberCommand) (*CrewMemberDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}
	if cmd.Email == "" {
		return nil, domain.ErrEmptyEmail
	}

	// Check if email already exists for tenant
	existing, err := s.crewRepo.FindByEmail(ctx, cmd.TenantID, cmd.Email)
	if err == nil && existing != nil {
		return nil, domain.ErrDuplicateEmail
	}

	member := &domain.CrewMember{
		ID:                 uuid.New().String(),
		TenantID:           cmd.TenantID,
		FirstName:          cmd.FirstName,
		LastName:           cmd.LastName,
		Email:              cmd.Email,
		Phone:              cmd.Phone,
		Role:               domain.CrewRole(cmd.Role),
		Status:             domain.CrewStatusActive,
		HourlyRate:         cmd.HourlyRate,
		DailyRate:          cmd.DailyRate,
		PreferredVehicleID: cmd.PreferredVehicleID,
		EmergencyContact:   cmd.EmergencyContact,
		Notes:              cmd.Notes,
		CreatedAt:          time.Now().UTC(),
		UpdatedAt:          time.Now().UTC(),
	}

	if cmd.Status != "" {
		member.Status = domain.CrewStatus(cmd.Status)
	}

	if err := s.crewRepo.Save(ctx, member); err != nil {
		s.logger.Error("failed to save crew member", err)
		return nil, err
	}

	return ToCrewMemberDTO(member), nil
}

// GetCrewMember retrieves a crew member by ID
func (s *CrewService) GetCrewMember(ctx context.Context, id string) (*CrewMemberDTO, error) {
	member, err := s.crewRepo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to find crew member", err, "id", id)
		return nil, err
	}
	if member == nil {
		return nil, domain.ErrCrewMemberNotFound
	}
	return ToCrewMemberDTO(member), nil
}

// UpdateCrewMember updates a crew member
func (s *CrewService) UpdateCrewMember(ctx context.Context, cmd UpdateCrewMemberCommand) (*CrewMemberDTO, error) {
	member, err := s.crewRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		s.logger.Error("failed to find crew member", err, "id", cmd.ID)
		return nil, err
	}
	if member == nil {
		return nil, domain.ErrCrewMemberNotFound
	}

	member.FirstName = cmd.FirstName
	member.LastName = cmd.LastName
	member.Email = cmd.Email
	member.Phone = cmd.Phone
	member.Role = domain.CrewRole(cmd.Role)
	member.Status = domain.CrewStatus(cmd.Status)
	member.HourlyRate = cmd.HourlyRate
	member.DailyRate = cmd.DailyRate
	member.PreferredVehicleID = cmd.PreferredVehicleID
	member.EmergencyContact = cmd.EmergencyContact
	member.Notes = cmd.Notes
	member.UpdatedAt = time.Now().UTC()

	if err := s.crewRepo.Save(ctx, member); err != nil {
		s.logger.Error("failed to update crew member", err)
		return nil, err
	}

	return ToCrewMemberDTO(member), nil
}

// DeleteCrewMember deletes a crew member
func (s *CrewService) DeleteCrewMember(ctx context.Context, id string) error {
	member, err := s.crewRepo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to find crew member", err, "id", id)
		return err
	}
	if member == nil {
		return domain.ErrCrewMemberNotFound
	}

	if err := s.crewRepo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete crew member", err)
		return err
	}

	return nil
}

// ListCrewMembers lists crew members for a tenant
func (s *CrewService) ListCrewMembers(ctx context.Context, tenantID string, page, perPage int) (*PaginatedResult, error) {
	members, total, err := s.crewRepo.List(ctx, tenantID, page, perPage)
	if err != nil {
		s.logger.Error("failed to list crew members", err)
		return nil, err
	}

	dtos := make([]interface{}, len(members))
	for i, member := range members {
		dtos[i] = ToCrewMemberDTO(member)
	}

	totalPages := (total + perPage - 1) / perPage
	return &PaginatedResult{
		Data:       dtos,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// CheckAvailability checks if crew member is available for a date range
func (s *CrewService) CheckAvailability(ctx context.Context, crewMemberID string, startDate, endDate time.Time) (*AvailabilityDTO, error) {
	member, err := s.crewRepo.FindByID(ctx, crewMemberID)
	if err != nil {
		s.logger.Error("failed to find crew member", err)
		return nil, err
	}
	if member == nil {
		return nil, domain.ErrCrewMemberNotFound
	}

	assignments, err := s.assignmentRepo.ListByCrewMember(ctx, crewMemberID)
	if err != nil {
		s.logger.Error("failed to list assignments", err)
		return nil, err
	}

	isAvailable := member.IsAvailable(startDate, endDate, assignments)

	return &AvailabilityDTO{
		CrewMemberID: crewMemberID,
		StartDate:    startDate.Format(time.RFC3339),
		EndDate:      endDate.Format(time.RFC3339),
		IsAvailable:  isAvailable,
	}, nil
}

// VerifyQualification checks if crew member has a specific qualification
func (s *CrewService) VerifyQualification(ctx context.Context, crewMemberID string, qualType string) (bool, error) {
	member, err := s.crewRepo.FindByID(ctx, crewMemberID)
	if err != nil {
		s.logger.Error("failed to find crew member", err)
		return false, err
	}
	if member == nil {
		return false, domain.ErrCrewMemberNotFound
	}

	qualifications, err := s.qualificationRepo.ListByCrewMember(ctx, crewMemberID)
	if err != nil {
		s.logger.Error("failed to list qualifications", err)
		return false, err
	}

	return member.HasQualification(domain.QualificationType(qualType), qualifications), nil
}

// GetDriverList retrieves drivers with their vehicle assignments and upcoming assignment dates
func (s *CrewService) GetDriverList(ctx context.Context, tenantID string) ([]*DriverDTO, error) {
	// Get all crew members for the tenant
	// We'll fetch all with high perPage to avoid pagination
	members, _, err := s.crewRepo.List(ctx, tenantID, 1, 1000)
	if err != nil {
		s.logger.Error("failed to list crew members", err)
		return nil, err
	}

	var drivers []*DriverDTO

	// Get all qualifications for filtering
	qualifications := make(map[string][]*domain.Qualification)
	for _, member := range members {
		quals, err := s.qualificationRepo.ListByCrewMember(ctx, member.ID)
		if err != nil {
			s.logger.Error("failed to list qualifications", err, "member_id", member.ID)
			continue
		}
		qualifications[member.ID] = quals
	}

	// Get all assignments to find upcoming ones
	for _, member := range members {
		// Check if member is a driver or has driver qualifications
		quals := qualifications[member.ID]
		isDriver := member.Role == domain.CrewRoleDriver || member.CanDrive(quals)

		if !isDriver {
			continue
		}

		// Get assignments for this driver
		assignments, err := s.assignmentRepo.ListByCrewMember(ctx, member.ID)
		if err != nil {
			s.logger.Error("failed to list assignments for member", err, "member_id", member.ID)
			continue
		}

		// Find upcoming assignments
		now := time.Now()
		var upcomingStart, upcomingEnd *time.Time
		for _, assignment := range assignments {
			if assignment.Status != domain.AssignmentStatusCancelled && assignment.EndDate.After(now) {
				if upcomingStart == nil || assignment.StartDate.Before(*upcomingStart) {
					upcomingStart = &assignment.StartDate
				}
				if upcomingEnd == nil || assignment.EndDate.After(*upcomingEnd) {
					upcomingEnd = &assignment.EndDate
				}
			}
		}

		driverDTO := &DriverDTO{
			MemberID:       member.ID,
			Name:           member.FirstName + " " + member.LastName,
			PreferredVehicleID: member.PreferredVehicleID,
			AssignmentStartDate: upcomingStart,
			AssignmentEndDate:   upcomingEnd,
		}

		drivers = append(drivers, driverDTO)
	}

	return drivers, nil
}
