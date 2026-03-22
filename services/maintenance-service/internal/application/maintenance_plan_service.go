package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/maintenance-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/maintenance-service/internal/ports"
)

type MaintenancePlanService struct {
	planRepo ports.MaintenancePlanRepository
	taskRepo ports.MaintenanceTaskRepository
	logger   logger.Logger
}

func NewMaintenancePlanService(
	planRepo ports.MaintenancePlanRepository,
	taskRepo ports.MaintenanceTaskRepository,
	log logger.Logger,
) *MaintenancePlanService {
	return &MaintenancePlanService{
		planRepo: planRepo,
		taskRepo: taskRepo,
		logger:   log,
	}
}

func (s *MaintenancePlanService) CreatePlan(ctx context.Context, cmd CreateMaintenancePlanCommand) (*MaintenancePlanDTO, error) {
	if cmd.TenantID == "" || cmd.EquipmentID == "" || cmd.Name == "" {
		return nil, domain.ErrInvalidInput
	}

	var nextDueAt *time.Time
	if cmd.PlanType == string(domain.PlanTypeInterval) && cmd.IntervalDays != nil {
		due := time.Now().AddDate(0, 0, *cmd.IntervalDays)
		nextDueAt = &due
	}

	plan := &domain.MaintenancePlan{
		ID:                 uuid.New().String(),
		TenantID:           cmd.TenantID,
		EquipmentID:        cmd.EquipmentID,
		PlanType:           domain.PlanType(cmd.PlanType),
		IntervalDays:       cmd.IntervalDays,
		IntervalHours:      cmd.IntervalHours,
		Name:               cmd.Name,
		Description:        cmd.Description,
		ChecklistTemplateID: cmd.ChecklistTemplateID,
		NextDueAt:          nextDueAt,
		IsActive:           true,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.planRepo.Create(ctx, plan); err != nil {
		s.logger.Error("Failed to create maintenance plan", err)
		return nil, err
	}

	return PlanToDTO(plan), nil
}

func (s *MaintenancePlanService) GetPlan(ctx context.Context, tenantID, planID string) (*MaintenancePlanDTO, error) {
	if tenantID == "" || planID == "" {
		return nil, domain.ErrInvalidInput
	}

	plan, err := s.planRepo.GetByID(ctx, tenantID, planID)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, domain.ErrPlanNotFound
	}

	return PlanToDTO(plan), nil
}

func (s *MaintenancePlanService) ListPlans(ctx context.Context, tenantID string) ([]*MaintenancePlanDTO, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}

	plans, err := s.planRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		s.logger.Error("Failed to list maintenance plans", err)
		return nil, err
	}

	dtos := make([]*MaintenancePlanDTO, len(plans))
	for i, p := range plans {
		dtos[i] = PlanToDTO(p)
	}
	return dtos, nil
}

func (s *MaintenancePlanService) ListPlansByEquipment(ctx context.Context, tenantID, equipmentID string) ([]*MaintenancePlanDTO, error) {
	if tenantID == "" || equipmentID == "" {
		return nil, domain.ErrInvalidInput
	}

	plans, err := s.planRepo.GetByEquipment(ctx, tenantID, equipmentID)
	if err != nil {
		s.logger.Error("Failed to list plans by equipment", err)
		return nil, err
	}

	dtos := make([]*MaintenancePlanDTO, len(plans))
	for i, p := range plans {
		dtos[i] = PlanToDTO(p)
	}
	return dtos, nil
}

func (s *MaintenancePlanService) UpdatePlan(ctx context.Context, tenantID, planID string, cmd CreateMaintenancePlanCommand) (*MaintenancePlanDTO, error) {
	if tenantID == "" || planID == "" {
		return nil, domain.ErrInvalidInput
	}

	plan, err := s.planRepo.GetByID(ctx, tenantID, planID)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, domain.ErrPlanNotFound
	}

	plan.Name = cmd.Name
	plan.Description = cmd.Description
	plan.IntervalDays = cmd.IntervalDays
	plan.IntervalHours = cmd.IntervalHours
	plan.ChecklistTemplateID = cmd.ChecklistTemplateID
	plan.UpdatedAt = time.Now()

	if err := s.planRepo.Update(ctx, plan); err != nil {
		s.logger.Error("Failed to update maintenance plan", err)
		return nil, err
	}

	return PlanToDTO(plan), nil
}

func (s *MaintenancePlanService) DeletePlan(ctx context.Context, tenantID, planID string) error {
	if tenantID == "" || planID == "" {
		return domain.ErrInvalidInput
	}

	return s.planRepo.Delete(ctx, tenantID, planID)
}

func (s *MaintenancePlanService) GetDuePlans(ctx context.Context, tenantID string) ([]*MaintenancePlanDTO, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}

	plans, err := s.planRepo.ListDuePlans(ctx, tenantID)
	if err != nil {
		s.logger.Error("Failed to get due plans", err)
		return nil, err
	}

	dtos := make([]*MaintenancePlanDTO, len(plans))
	for i, p := range plans {
		dtos[i] = PlanToDTO(p)
	}
	return dtos, nil
}
