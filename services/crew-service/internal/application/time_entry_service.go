package application

import (
	"context"
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/crew-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/crew-service/internal/ports"
)

type TimeEntryService struct {
	timeRepo ports.TimeEntryRepository
	logger   logger.Logger
}

func NewTimeEntryService(timeRepo ports.TimeEntryRepository, logger logger.Logger) *TimeEntryService {
	return &TimeEntryService{
		timeRepo: timeRepo,
		logger:   logger,
	}
}

func (s *TimeEntryService) CreateTimeEntry(ctx context.Context, cmd CreateTimeEntryCommand) (*TimeEntryDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	teID := fmt.Sprintf("te_%d", hashString(cmd.TenantID+cmd.CrewMemberID+cmd.Date.String()))
	entryType := domain.TimeEntryType(cmd.Type)

	te := domain.NewTimeEntry(teID, cmd.TenantID, cmd.CrewMemberID, cmd.ProjectID, cmd.Date, cmd.StartTime, cmd.EndTime, entryType)
	te.Notes = cmd.Notes

	if err := te.Validate(); err != nil {
		return nil, domain.NewDomainError("VALIDATION_ERROR", err.Error(), nil)
	}

	if err := s.timeRepo.CreateTimeEntry(ctx, te); err != nil {
		return nil, domain.NewDomainError("CREATE_FAILED", "failed to create time entry", err)
	}

	return TimeEntryToDTO(te), nil
}

func (s *TimeEntryService) GetTimeEntry(ctx context.Context, tenantID, timeEntryID string) (*TimeEntryDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	te, err := s.timeRepo.GetTimeEntry(ctx, tenantID, timeEntryID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "time entry not found", err)
	}

	return TimeEntryToDTO(te), nil
}

func (s *TimeEntryService) ListTimeEntries(ctx context.Context, tenantID string, limit, offset int) ([]*TimeEntryDTO, int64, error) {
	if tenantID == "" {
		return nil, 0, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	entries, total, err := s.timeRepo.ListTimeEntries(ctx, tenantID, limit, offset)
	if err != nil {
		return nil, 0, domain.NewDomainError("LIST_FAILED", "failed to list time entries", err)
	}

	dtos := make([]*TimeEntryDTO, len(entries))
	for i, e := range entries {
		dtos[i] = TimeEntryToDTO(e)
	}

	return dtos, total, nil
}

func (s *TimeEntryService) UpdateTimeEntry(ctx context.Context, cmd UpdateTimeEntryCommand) (*TimeEntryDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	te, err := s.timeRepo.GetTimeEntry(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "time entry not found", err)
	}

	if cmd.BreakMinutes >= 0 {
		if err := te.SetBreak(cmd.BreakMinutes); err != nil {
			return nil, domain.NewDomainError("UPDATE_FAILED", err.Error(), nil)
		}
	}
	if cmd.Notes != "" {
		te.Notes = cmd.Notes
	}

	if err := s.timeRepo.UpdateTimeEntry(ctx, te); err != nil {
		return nil, domain.NewDomainError("UPDATE_FAILED", "failed to update time entry", err)
	}

	return TimeEntryToDTO(te), nil
}

func (s *TimeEntryService) ApproveTimeEntry(ctx context.Context, tenantID, timeEntryID string) (*TimeEntryDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	te, err := s.timeRepo.GetTimeEntry(ctx, tenantID, timeEntryID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "time entry not found", err)
	}

	if err := te.Approve(); err != nil {
		return nil, domain.NewDomainError("APPROVE_FAILED", err.Error(), nil)
	}

	if err := s.timeRepo.UpdateTimeEntry(ctx, te); err != nil {
		return nil, domain.NewDomainError("UPDATE_FAILED", "failed to approve time entry", err)
	}

	return TimeEntryToDTO(te), nil
}

func (s *TimeEntryService) RejectTimeEntry(ctx context.Context, tenantID, timeEntryID string) (*TimeEntryDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	te, err := s.timeRepo.GetTimeEntry(ctx, tenantID, timeEntryID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "time entry not found", err)
	}

	if err := te.Reject(); err != nil {
		return nil, domain.NewDomainError("REJECT_FAILED", err.Error(), nil)
	}

	if err := s.timeRepo.UpdateTimeEntry(ctx, te); err != nil {
		return nil, domain.NewDomainError("UPDATE_FAILED", "failed to reject time entry", err)
	}

	return TimeEntryToDTO(te), nil
}

func (s *TimeEntryService) GetSummary(ctx context.Context, tenantID, crewMemberID string, from, to time.Time) (*TimeEntrySummary, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	entries, _, err := s.timeRepo.ListByCrewMemberAndPeriod(ctx, tenantID, crewMemberID, from, to)
	if err != nil {
		return nil, domain.NewDomainError("QUERY_FAILED", "failed to get summary", err)
	}

	summary := &TimeEntrySummary{
		CrewMemberID: crewMemberID,
	}

	for _, e := range entries {
		summary.TotalHours += e.TotalHours
		summary.Entries++
		if e.Status == domain.TimeStatusApproved {
			summary.ApprovedHours += e.TotalHours
		}
	}

	return summary, nil
}
