package application

import (
	"context"
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/maintenance-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/maintenance-service/internal/ports"
)

type MaintenanceRecordService struct {
	repo   ports.MaintenanceRecordRepository
	logger logger.Logger
}

func NewMaintenanceRecordService(repo ports.MaintenanceRecordRepository, log logger.Logger) *MaintenanceRecordService {
	return &MaintenanceRecordService{
		repo:   repo,
		logger: log,
	}
}

func (s *MaintenanceRecordService) CreateRecord(ctx context.Context, cmd CreateMaintenanceRecordCommand) (*MaintenanceRecordDTO, error) {
	if cmd.TenantID == "" || cmd.EquipmentID == "" {
		return nil, domain.ErrTenantIDRequired
	}

	record := &domain.MaintenanceRecord{
		ID:            fmt.Sprintf("maint_%d", time.Now().UnixNano()),
		TenantID:      cmd.TenantID,
		EquipmentID:   cmd.EquipmentID,
		Type:          domain.MaintenanceType(cmd.Type),
		Status:        domain.MaintenanceStatusScheduled,
		ScheduledDate: cmd.ScheduledDate,
		Technician:    cmd.Technician,
		Cost:          cmd.Cost,
		Notes:         cmd.Notes,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.repo.Create(ctx, record); err != nil {
		s.logger.Error("Failed to create maintenance record", err)
		return nil, err
	}

	return RecordToDTO(record), nil
}

func (s *MaintenanceRecordService) GetRecord(ctx context.Context, tenantID, id string) (*MaintenanceRecordDTO, error) {
	if tenantID == "" || id == "" {
		return nil, domain.ErrInvalidInput
	}

	record, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, domain.ErrRecordNotFound
	}

	return RecordToDTO(record), nil
}

func (s *MaintenanceRecordService) ListRecords(ctx context.Context, tenantID string) ([]*MaintenanceRecordDTO, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}

	records, err := s.repo.ListByTenant(ctx, tenantID)
	if err != nil {
		s.logger.Error("Failed to list maintenance records", err)
		return nil, err
	}

	dtos := make([]*MaintenanceRecordDTO, len(records))
	for i, r := range records {
		dtos[i] = RecordToDTO(r)
	}
	return dtos, nil
}

func (s *MaintenanceRecordService) ListByEquipment(ctx context.Context, tenantID, equipmentID string) ([]*MaintenanceRecordDTO, error) {
	if tenantID == "" || equipmentID == "" {
		return nil, domain.ErrInvalidInput
	}

	records, err := s.repo.ListByEquipment(ctx, tenantID, equipmentID)
	if err != nil {
		s.logger.Error("Failed to list maintenance records by equipment", err)
		return nil, err
	}

	dtos := make([]*MaintenanceRecordDTO, len(records))
	for i, r := range records {
		dtos[i] = RecordToDTO(r)
	}
	return dtos, nil
}

func (s *MaintenanceRecordService) ListOverdue(ctx context.Context, tenantID string) ([]*MaintenanceRecordDTO, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}

	records, err := s.repo.ListOverdue(ctx, tenantID)
	if err != nil {
		s.logger.Error("Failed to list overdue maintenance records", err)
		return nil, err
	}

	dtos := make([]*MaintenanceRecordDTO, len(records))
	for i, r := range records {
		dtos[i] = RecordToDTO(r)
	}
	return dtos, nil
}

func (s *MaintenanceRecordService) CompleteRecord(ctx context.Context, cmd CompleteMaintenanceCommand) (*MaintenanceRecordDTO, error) {
	if cmd.TenantID == "" || cmd.RecordID == "" {
		return nil, domain.ErrInvalidInput
	}

	record, err := s.repo.GetByID(ctx, cmd.TenantID, cmd.RecordID)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, domain.ErrRecordNotFound
	}

	record.Status = domain.MaintenanceStatusCompleted
	record.CompletedDate = &cmd.CompletedDate
	record.Notes = cmd.Notes
	record.CertificateRef = cmd.CertificateRef
	record.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, record); err != nil {
		s.logger.Error("Failed to complete maintenance record", err)
		return nil, err
	}

	return RecordToDTO(record), nil
}

func (s *MaintenanceRecordService) DeleteRecord(ctx context.Context, tenantID, id string) error {
	if tenantID == "" || id == "" {
		return domain.ErrInvalidInput
	}

	return s.repo.Delete(ctx, tenantID, id)
}
