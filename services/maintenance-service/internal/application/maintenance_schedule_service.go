package application

import (
	"context"
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/maintenance-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/maintenance-service/internal/ports"
)

type MaintenanceScheduleService struct {
	scheduleRepo ports.MaintenanceScheduleRepository
	recordRepo   ports.MaintenanceRecordRepository
	logger       logger.Logger
}

func NewMaintenanceScheduleService(
	scheduleRepo ports.MaintenanceScheduleRepository,
	recordRepo ports.MaintenanceRecordRepository,
	log logger.Logger,
) *MaintenanceScheduleService {
	return &MaintenanceScheduleService{
		scheduleRepo: scheduleRepo,
		recordRepo:   recordRepo,
		logger:       log,
	}
}

func (s *MaintenanceScheduleService) CreateSchedule(ctx context.Context, cmd CreateScheduleCommand) (*MaintenanceScheduleDTO, error) {
	if cmd.TenantID == "" || cmd.EquipmentID == "" || cmd.IntervalDays == 0 {
		return nil, domain.ErrInvalidInput
	}

	nextDue := time.Now().AddDate(0, 0, cmd.IntervalDays)

	schedule := &domain.MaintenanceSchedule{
		ID:           fmt.Sprintf("sched_%d", time.Now().UnixNano()),
		TenantID:     cmd.TenantID,
		EquipmentID:  cmd.EquipmentID,
		IntervalDays: cmd.IntervalDays,
		NextDue:      nextDue,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.scheduleRepo.Create(ctx, schedule); err != nil {
		s.logger.Error("Failed to create maintenance schedule", err)
		return nil, err
	}

	return ScheduleToDTO(schedule), nil
}

func (s *MaintenanceScheduleService) GetSchedule(ctx context.Context, tenantID, id string) (*MaintenanceScheduleDTO, error) {
	if tenantID == "" || id == "" {
		return nil, domain.ErrInvalidInput
	}

	schedule, err := s.scheduleRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if schedule == nil {
		return nil, domain.ErrScheduleNotFound
	}

	return ScheduleToDTO(schedule), nil
}

func (s *MaintenanceScheduleService) GetScheduleByEquipment(ctx context.Context, tenantID, equipmentID string) (*MaintenanceScheduleDTO, error) {
	if tenantID == "" || equipmentID == "" {
		return nil, domain.ErrInvalidInput
	}

	schedule, err := s.scheduleRepo.GetByEquipment(ctx, tenantID, equipmentID)
	if err != nil {
		return nil, err
	}
	if schedule == nil {
		return nil, domain.ErrScheduleNotFound
	}

	return ScheduleToDTO(schedule), nil
}

func (s *MaintenanceScheduleService) ListSchedules(ctx context.Context, tenantID string) ([]*MaintenanceScheduleDTO, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}

	schedules, err := s.scheduleRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		s.logger.Error("Failed to list maintenance schedules", err)
		return nil, err
	}

	dtos := make([]*MaintenanceScheduleDTO, len(schedules))
	for i, s := range schedules {
		dtos[i] = ScheduleToDTO(s)
	}
	return dtos, nil
}

func (s *MaintenanceScheduleService) UpdateSchedule(ctx context.Context, cmd UpdateScheduleCommand) (*MaintenanceScheduleDTO, error) {
	if cmd.TenantID == "" || cmd.ScheduleID == "" {
		return nil, domain.ErrInvalidInput
	}

	schedule, err := s.scheduleRepo.GetByID(ctx, cmd.TenantID, cmd.ScheduleID)
	if err != nil {
		return nil, err
	}
	if schedule == nil {
		return nil, domain.ErrScheduleNotFound
	}

	schedule.IntervalDays = cmd.IntervalDays
	schedule.NextDue = time.Now().AddDate(0, 0, cmd.IntervalDays)
	schedule.UpdatedAt = time.Now()

	if err := s.scheduleRepo.Update(ctx, schedule); err != nil {
		s.logger.Error("Failed to update maintenance schedule", err)
		return nil, err
	}

	return ScheduleToDTO(schedule), nil
}

func (s *MaintenanceScheduleService) DeleteSchedule(ctx context.Context, tenantID, id string) error {
	if tenantID == "" || id == "" {
		return domain.ErrInvalidInput
	}

	return s.scheduleRepo.Delete(ctx, tenantID, id)
}
