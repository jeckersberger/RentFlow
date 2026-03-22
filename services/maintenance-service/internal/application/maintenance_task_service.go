package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/maintenance-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/maintenance-service/internal/ports"
)

type MaintenanceTaskService struct {
	taskRepo ports.MaintenanceTaskRepository
	planRepo ports.MaintenancePlanRepository
	logger   logger.Logger
}

func NewMaintenanceTaskService(
	taskRepo ports.MaintenanceTaskRepository,
	planRepo ports.MaintenancePlanRepository,
	log logger.Logger,
) *MaintenanceTaskService {
	return &MaintenanceTaskService{
		taskRepo: taskRepo,
		planRepo: planRepo,
		logger:   log,
	}
}

func (s *MaintenanceTaskService) CreateTask(ctx context.Context, cmd CreateMaintenanceTaskCommand) (*MaintenanceTaskDTO, error) {
	if cmd.TenantID == "" || cmd.PlanID == "" || cmd.EquipmentID == "" {
		return nil, domain.ErrInvalidInput
	}

	priority := domain.PriorityMedium
	if cmd.Priority != "" {
		priority = domain.TaskPriority(cmd.Priority)
	}

	task := &domain.MaintenanceTask{
		ID:          uuid.New().String(),
		TenantID:    cmd.TenantID,
		PlanID:      cmd.PlanID,
		EquipmentID: cmd.EquipmentID,
		AssignedTo:  cmd.AssignedTo,
		Status:      domain.TaskStatusPlanned,
		Priority:    priority,
		ScheduledAt: cmd.ScheduledAt,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		s.logger.Error("Failed to create maintenance task", err)
		return nil, err
	}

	return TaskToDTO(task), nil
}

func (s *MaintenanceTaskService) GetTask(ctx context.Context, tenantID, taskID string) (*MaintenanceTaskDTO, error) {
	if tenantID == "" || taskID == "" {
		return nil, domain.ErrInvalidInput
	}

	task, err := s.taskRepo.GetByID(ctx, tenantID, taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, domain.ErrTaskNotFound
	}

	return TaskToDTO(task), nil
}

func (s *MaintenanceTaskService) ListTasks(ctx context.Context, tenantID string) ([]*MaintenanceTaskDTO, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}

	tasks, err := s.taskRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		s.logger.Error("Failed to list tasks", err)
		return nil, err
	}

	dtos := make([]*MaintenanceTaskDTO, len(tasks))
	for i, t := range tasks {
		dtos[i] = TaskToDTO(t)
	}
	return dtos, nil
}

func (s *MaintenanceTaskService) ListTasksByPlan(ctx context.Context, tenantID, planID string) ([]*MaintenanceTaskDTO, error) {
	if tenantID == "" || planID == "" {
		return nil, domain.ErrInvalidInput
	}

	tasks, err := s.taskRepo.ListByPlan(ctx, tenantID, planID)
	if err != nil {
		s.logger.Error("Failed to list tasks by plan", err)
		return nil, err
	}

	dtos := make([]*MaintenanceTaskDTO, len(tasks))
	for i, t := range tasks {
		dtos[i] = TaskToDTO(t)
	}
	return dtos, nil
}

func (s *MaintenanceTaskService) StartTask(ctx context.Context, cmd StartMaintenanceTaskCommand) (*MaintenanceTaskDTO, error) {
	if cmd.TenantID == "" || cmd.TaskID == "" {
		return nil, domain.ErrInvalidInput
	}

	task, err := s.taskRepo.GetByID(ctx, cmd.TenantID, cmd.TaskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, domain.ErrTaskNotFound
	}

	if task.Status != domain.TaskStatusPlanned {
		return nil, domain.ErrInvalidTaskStatus
	}

	now := time.Now()
	task.Status = domain.TaskStatusInProgress
	task.StartedAt = &now
	task.UpdatedAt = now

	if err := s.taskRepo.Update(ctx, task); err != nil {
		s.logger.Error("Failed to start task", err)
		return nil, err
	}

	return TaskToDTO(task), nil
}

func (s *MaintenanceTaskService) CompleteTask(ctx context.Context, cmd CompleteMaintenanceTaskCommand) (*MaintenanceTaskDTO, error) {
	if cmd.TenantID == "" || cmd.TaskID == "" {
		return nil, domain.ErrInvalidInput
	}

	task, err := s.taskRepo.GetByID(ctx, cmd.TenantID, cmd.TaskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, domain.ErrTaskNotFound
	}

	if task.Status != domain.TaskStatusInProgress && task.Status != domain.TaskStatusPlanned {
		return nil, domain.ErrInvalidTaskStatus
	}

	now := time.Now()
	task.Status = domain.TaskStatusCompleted
	task.CompletedAt = &now
	if task.StartedAt == nil {
		task.StartedAt = &now
	}
	task.Notes = cmd.Notes
	task.UpdatedAt = now

	if cmd.ChecklistData != nil {
		task.ChecklistData = &domain.ChecklistData{
			Items: []domain.ChecklistItemResult{},
		}
	}

	if err := s.taskRepo.Update(ctx, task); err != nil {
		s.logger.Error("Failed to complete task", err)
		return nil, err
	}

	return TaskToDTO(task), nil
}

func (s *MaintenanceTaskService) GetDueTasks(ctx context.Context, tenantID string) ([]*MaintenanceTaskDTO, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}

	tasks, err := s.taskRepo.ListDueTasks(ctx, tenantID)
	if err != nil {
		s.logger.Error("Failed to get due tasks", err)
		return nil, err
	}

	dtos := make([]*MaintenanceTaskDTO, len(tasks))
	for i, t := range tasks {
		dtos[i] = TaskToDTO(t)
	}
	return dtos, nil
}

func (s *MaintenanceTaskService) GetOverdueTasks(ctx context.Context, tenantID string) ([]*MaintenanceTaskDTO, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}

	tasks, err := s.taskRepo.ListOverdueTasks(ctx, tenantID)
	if err != nil {
		s.logger.Error("Failed to get overdue tasks", err)
		return nil, err
	}

	dtos := make([]*MaintenanceTaskDTO, len(tasks))
	for i, t := range tasks {
		dtos[i] = TaskToDTO(t)
	}
	return dtos, nil
}

func (s *MaintenanceTaskService) CancelTask(ctx context.Context, tenantID, taskID string) (*MaintenanceTaskDTO, error) {
	if tenantID == "" || taskID == "" {
		return nil, domain.ErrInvalidInput
	}

	task, err := s.taskRepo.GetByID(ctx, tenantID, taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, domain.ErrTaskNotFound
	}

	task.Status = domain.TaskStatusCancelled
	task.UpdatedAt = time.Now()

	if err := s.taskRepo.Update(ctx, task); err != nil {
		s.logger.Error("Failed to cancel task", err)
		return nil, err
	}

	return TaskToDTO(task), nil
}
