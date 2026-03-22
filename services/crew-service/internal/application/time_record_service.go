package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/crew-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/crew-service/internal/ports"
)

// TimeRecordService handles time record business logic
type TimeRecordService struct {
	timeRecordRepo ports.TimeRecordRepository
	crewRepo       ports.CrewMemberRepository
	logger         logger.Logger
}

// NewTimeRecordService creates a new time record service
func NewTimeRecordService(
	timeRecordRepo ports.TimeRecordRepository,
	crewRepo ports.CrewMemberRepository,
	log logger.Logger,
) *TimeRecordService {
	return &TimeRecordService{
		timeRecordRepo: timeRecordRepo,
		crewRepo:       crewRepo,
		logger:         log,
	}
}

// StartTimeRecord starts a new time record
func (s *TimeRecordService) StartTimeRecord(ctx context.Context, cmd StartTimeRecordCommand) (*TimeRecordDTO, error) {
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

	// Parse date
	date, err := time.Parse("2006-01-02", cmd.Date)
	if err != nil {
		s.logger.Error("invalid date", err)
		return nil, domain.ErrInvalidDateRange
	}

	timeRecord := &domain.TimeRecord{
		ID:           uuid.New().String(),
		TenantID:     cmd.TenantID,
		CrewMemberID: cmd.CrewMemberID,
		AssignmentID: cmd.AssignmentID,
		Date:         date,
		StartTime:    time.Now().UTC(),
		Status:       domain.TimeRecordStatusRunning,
		Notes:        cmd.Notes,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if err := s.timeRecordRepo.Save(ctx, timeRecord); err != nil {
		s.logger.Error("failed to save time record", err)
		return nil, err
	}

	return ToTimeRecordDTO(timeRecord), nil
}

// StopTimeRecord stops a running time record
func (s *TimeRecordService) StopTimeRecord(ctx context.Context, cmd StopTimeRecordCommand) (*TimeRecordDTO, error) {
	timeRecord, err := s.timeRecordRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		s.logger.Error("failed to find time record", err)
		return nil, err
	}
	if timeRecord == nil {
		return nil, domain.ErrTimeRecordNotFound
	}

	if timeRecord.Status != domain.TimeRecordStatusRunning {
		return nil, domain.ErrTimeRecordAlreadyRunning
	}

	now := time.Now().UTC()
	timeRecord.EndTime = &now
	timeRecord.BreakMinutes = cmd.BreakMinutes
	timeRecord.OvertimeMinutes = cmd.OvertimeMinutes
	timeRecord.Status = domain.TimeRecordStatusCompleted
	timeRecord.Notes = cmd.Notes
	timeRecord.UpdatedAt = time.Now().UTC()

	if err := s.timeRecordRepo.Save(ctx, timeRecord); err != nil {
		s.logger.Error("failed to save time record", err)
		return nil, err
	}

	return ToTimeRecordDTO(timeRecord), nil
}

// GetTimeRecord retrieves a time record by ID
func (s *TimeRecordService) GetTimeRecord(ctx context.Context, id string) (*TimeRecordDTO, error) {
	timeRecord, err := s.timeRecordRepo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to find time record", err)
		return nil, err
	}
	if timeRecord == nil {
		return nil, domain.ErrTimeRecordNotFound
	}
	return ToTimeRecordDTO(timeRecord), nil
}

// ListTimeRecordsByCrewMember lists time records for a crew member
func (s *TimeRecordService) ListTimeRecordsByCrewMember(ctx context.Context, crewMemberID string) ([]*TimeRecordDTO, error) {
	records, err := s.timeRecordRepo.ListByCrewMember(ctx, crewMemberID)
	if err != nil {
		s.logger.Error("failed to list time records", err)
		return nil, err
	}

	dtos := make([]*TimeRecordDTO, len(records))
	for i, record := range records {
		dtos[i] = ToTimeRecordDTO(record)
	}

	return dtos, nil
}

// ListTimeRecordsByDate lists time records for a crew member on a specific date
func (s *TimeRecordService) ListTimeRecordsByDate(ctx context.Context, crewMemberID string, date time.Time) ([]*TimeRecordDTO, error) {
	records, err := s.timeRecordRepo.ListByCrewMemberAndDate(ctx, crewMemberID, date)
	if err != nil {
		s.logger.Error("failed to list time records", err)
		return nil, err
	}

	dtos := make([]*TimeRecordDTO, len(records))
	for i, record := range records {
		dtos[i] = ToTimeRecordDTO(record)
	}

	return dtos, nil
}

// ApproveTimeRecord approves a time record
func (s *TimeRecordService) ApproveTimeRecord(ctx context.Context, cmd ApproveTimeRecordCommand) (*TimeRecordDTO, error) {
	timeRecord, err := s.timeRecordRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		s.logger.Error("failed to find time record", err)
		return nil, err
	}
	if timeRecord == nil {
		return nil, domain.ErrTimeRecordNotFound
	}

	if timeRecord.Status != domain.TimeRecordStatusCompleted {
		return nil, domain.ErrInvalidStatus
	}

	timeRecord.Status = domain.TimeRecordStatusApproved
	timeRecord.UpdatedAt = time.Now().UTC()

	if err := s.timeRecordRepo.Save(ctx, timeRecord); err != nil {
		s.logger.Error("failed to save time record", err)
		return nil, err
	}

	return ToTimeRecordDTO(timeRecord), nil
}
