package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/maintenance/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

type CreateScheduleRequest struct {
	EquipmentID  uuid.UUID `json:"equipment_id"`
	Name         string    `json:"name"`
	IntervalDays *int      `json:"interval_days"`
	NextDueAt    *string   `json:"next_due_at"`
	IsActive     *bool     `json:"is_active"`
	Notes        string    `json:"notes"`
}

type UpdateScheduleRequest struct {
	Name         *string `json:"name"`
	IntervalDays *int    `json:"interval_days"`
	NextDueAt    *string `json:"next_due_at"`
	IsActive     *bool   `json:"is_active"`
	Notes        *string `json:"notes"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

type ScheduleService struct {
	repo   domain.ScheduleRepository
	logger zerolog.Logger
}

func NewScheduleService(repo domain.ScheduleRepository, logger zerolog.Logger) *ScheduleService {
	return &ScheduleService{
		repo:   repo,
		logger: logger.With().Str("service", "schedule").Logger(),
	}
}

func (s *ScheduleService) Create(ctx context.Context, tenantID uuid.UUID, req CreateScheduleRequest) (*domain.MaintenanceSchedule, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.EquipmentID == uuid.Nil {
		return nil, fmt.Errorf("equipment_id is required")
	}

	intervalDays := 365
	if req.IntervalDays != nil {
		intervalDays = *req.IntervalDays
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	schedule := &domain.MaintenanceSchedule{
		ID:           uuid.New(),
		TenantID:     tenantID,
		EquipmentID:  req.EquipmentID,
		Name:         req.Name,
		IntervalDays: intervalDays,
		IsActive:     isActive,
		Notes:        req.Notes,
	}

	if req.NextDueAt != nil && *req.NextDueAt != "" {
		t, err := time.Parse(time.RFC3339, *req.NextDueAt)
		if err != nil {
			return nil, fmt.Errorf("invalid next_due_at format: %w", err)
		}
		schedule.NextDueAt = &t
	}

	if err := s.repo.Create(ctx, schedule); err != nil {
		return nil, fmt.Errorf("create schedule: %w", err)
	}

	s.logger.Info().Str("schedule_id", schedule.ID.String()).Msg("maintenance schedule created")
	return schedule, nil
}

func (s *ScheduleService) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.MaintenanceSchedule, error) {
	return s.repo.GetByID(ctx, id, tenantID)
}

func (s *ScheduleService) List(ctx context.Context, tenantID uuid.UUID, filter domain.ScheduleFilter) ([]*domain.MaintenanceSchedule, int64, error) {
	return s.repo.List(ctx, tenantID, filter)
}

func (s *ScheduleService) Update(ctx context.Context, id, tenantID uuid.UUID, req UpdateScheduleRequest) (*domain.MaintenanceSchedule, error) {
	existing, err := s.repo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.IntervalDays != nil {
		existing.IntervalDays = *req.IntervalDays
	}
	if req.NextDueAt != nil {
		if *req.NextDueAt == "" {
			existing.NextDueAt = nil
		} else {
			t, err := time.Parse(time.RFC3339, *req.NextDueAt)
			if err != nil {
				return nil, fmt.Errorf("invalid next_due_at format: %w", err)
			}
			existing.NextDueAt = &t
		}
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}
	if req.Notes != nil {
		existing.Notes = *req.Notes
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}
