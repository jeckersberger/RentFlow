package application

import (
	"bytes"
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

// CheckInRequest holds the data needed to start a time tracking entry.
type CheckInRequest struct {
	CrewMemberID uuid.UUID  `json:"crew_member_id"`
	AssignmentID *uuid.UUID `json:"assignment_id,omitempty"`
	ProjectID    *uuid.UUID `json:"project_id,omitempty"`
	Notes        string     `json:"notes,omitempty"`
}

// CheckOutRequest holds the data needed to end a time tracking entry.
type CheckOutRequest struct {
	CrewMemberID uuid.UUID `json:"crew_member_id"`
	BreakMinutes int       `json:"break_minutes,omitempty"`
	Notes        string    `json:"notes,omitempty"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// TimeTrackingService implements the application-level use cases for time tracking.
type TimeTrackingService struct {
	timeEntryRepo domain.TimeEntryRepository
	memberRepo    domain.CrewMemberRepository
	logger        zerolog.Logger
}

// NewTimeTrackingService constructs a new TimeTrackingService.
func NewTimeTrackingService(
	timeEntryRepo domain.TimeEntryRepository,
	memberRepo domain.CrewMemberRepository,
	logger zerolog.Logger,
) *TimeTrackingService {
	return &TimeTrackingService{
		timeEntryRepo: timeEntryRepo,
		memberRepo:    memberRepo,
		logger:        logger.With().Str("service", "time_tracking").Logger(),
	}
}

// CheckIn creates a new active time entry for a crew member.
func (s *TimeTrackingService) CheckIn(
	ctx context.Context,
	tenantID uuid.UUID,
	req CheckInRequest,
) (*domain.TimeEntry, error) {
	if req.CrewMemberID == uuid.Nil {
		return nil, domain.ErrCrewMemberIDRequired
	}

	// Verify crew member exists.
	if _, err := s.memberRepo.GetByID(ctx, req.CrewMemberID, tenantID); err != nil {
		return nil, fmt.Errorf("check-in – member lookup: %w", err)
	}

	// Verify no active entry exists.
	existing, err := s.timeEntryRepo.GetActiveByMember(ctx, req.CrewMemberID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("check-in – active lookup: %w", err)
	}
	if existing != nil {
		return nil, domain.ErrAlreadyCheckedIn
	}

	entry := &domain.TimeEntry{
		ID:           uuid.New(),
		TenantID:     tenantID,
		CrewMemberID: req.CrewMemberID,
		AssignmentID: req.AssignmentID,
		ProjectID:    req.ProjectID,
		CheckInTime:  time.Now(),
		BreakMinutes: 0,
		Notes:        req.Notes,
		Status:       domain.TimeEntryStatusActive,
	}

	if err := s.timeEntryRepo.Create(ctx, entry); err != nil {
		s.logger.Error().Err(err).
			Str("crew_member_id", req.CrewMemberID.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to create time entry")
		return nil, fmt.Errorf("check-in: %w", err)
	}

	s.logger.Info().
		Str("time_entry_id", entry.ID.String()).
		Str("crew_member_id", req.CrewMemberID.String()).
		Str("tenant_id", tenantID.String()).
		Msg("crew member checked in")

	return entry, nil
}

// CheckOut ends the active time entry for a crew member.
func (s *TimeTrackingService) CheckOut(
	ctx context.Context,
	tenantID uuid.UUID,
	req CheckOutRequest,
) (*domain.TimeEntry, error) {
	if req.CrewMemberID == uuid.Nil {
		return nil, domain.ErrCrewMemberIDRequired
	}

	// Find active entry.
	entry, err := s.timeEntryRepo.GetActiveByMember(ctx, req.CrewMemberID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("check-out – active lookup: %w", err)
	}
	if entry == nil {
		return nil, domain.ErrNoActiveCheckIn
	}

	now := time.Now()
	entry.CheckOutTime = &now
	entry.BreakMinutes = req.BreakMinutes
	entry.Status = domain.TimeEntryStatusCompleted

	// Calculate duration in minutes (total minus break).
	totalMinutes := int(now.Sub(entry.CheckInTime).Minutes())
	netMinutes := totalMinutes - req.BreakMinutes
	if netMinutes < 0 {
		netMinutes = 0
	}
	entry.DurationMinutes = &netMinutes

	if req.Notes != "" {
		entry.Notes = req.Notes
	}

	if err := s.timeEntryRepo.Update(ctx, entry); err != nil {
		s.logger.Error().Err(err).
			Str("time_entry_id", entry.ID.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to update time entry for check-out")
		return nil, fmt.Errorf("check-out: %w", err)
	}

	s.logger.Info().
		Str("time_entry_id", entry.ID.String()).
		Str("crew_member_id", req.CrewMemberID.String()).
		Str("tenant_id", tenantID.String()).
		Int("duration_minutes", netMinutes).
		Msg("crew member checked out")

	return entry, nil
}

// GetActiveEntry returns the currently active time entry for a member, or nil.
func (s *TimeTrackingService) GetActiveEntry(
	ctx context.Context,
	tenantID uuid.UUID,
	memberID uuid.UUID,
) (*domain.TimeEntry, error) {
	entry, err := s.timeEntryRepo.GetActiveByMember(ctx, memberID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get active entry: %w", err)
	}
	return entry, nil
}

// ListByMember returns time entries for a specific member within a date range.
func (s *TimeTrackingService) ListByMember(
	ctx context.Context,
	tenantID uuid.UUID,
	memberID uuid.UUID,
	from time.Time,
	to time.Time,
) ([]*domain.TimeEntry, error) {
	entries, err := s.timeEntryRepo.ListByMember(ctx, memberID, tenantID, from, to)
	if err != nil {
		return nil, fmt.Errorf("list entries by member: %w", err)
	}
	return entries, nil
}

// ListByProject returns time entries for a specific project.
func (s *TimeTrackingService) ListByProject(
	ctx context.Context,
	tenantID uuid.UUID,
	projectID uuid.UUID,
) ([]*domain.TimeEntry, error) {
	entries, err := s.timeEntryRepo.ListByProject(ctx, projectID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list entries by project: %w", err)
	}
	return entries, nil
}

// ExportCSV generates a CSV export of time entries within a date range.
func (s *TimeTrackingService) ExportCSV(
	ctx context.Context,
	tenantID uuid.UUID,
	from time.Time,
	to time.Time,
) ([]byte, error) {
	entries, err := s.timeEntryRepo.ListByTenant(ctx, tenantID, from, to)
	if err != nil {
		return nil, fmt.Errorf("export csv: %w", err)
	}

	var buf bytes.Buffer
	// CSV header
	buf.WriteString("id,crew_member_id,project_id,check_in_time,check_out_time,duration_minutes,break_minutes,status,notes\n")

	for _, e := range entries {
		projectID := ""
		if e.ProjectID != nil {
			projectID = e.ProjectID.String()
		}
		checkOutTime := ""
		if e.CheckOutTime != nil {
			checkOutTime = e.CheckOutTime.Format(time.RFC3339)
		}
		durationMinutes := ""
		if e.DurationMinutes != nil {
			durationMinutes = fmt.Sprintf("%d", *e.DurationMinutes)
		}

		line := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%d,%s,\"%s\"\n",
			e.ID.String(),
			e.CrewMemberID.String(),
			projectID,
			e.CheckInTime.Format(time.RFC3339),
			checkOutTime,
			durationMinutes,
			e.BreakMinutes,
			e.Status,
			e.Notes,
		)
		buf.WriteString(line)
	}

	return buf.Bytes(), nil
}
