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

type CreateTaskRequest struct {
	ScheduleID  *uuid.UUID `json:"schedule_id"`
	EquipmentID uuid.UUID  `json:"equipment_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Priority    string     `json:"priority"`
	AssignedTo  *uuid.UUID `json:"assigned_to"`
	DueDate     string     `json:"due_date"`
	Cost        *int64     `json:"cost"`
	Notes       string     `json:"notes"`
}

type UpdateTaskRequest struct {
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	Status      *string    `json:"status"`
	Priority    *string    `json:"priority"`
	AssignedTo  *uuid.UUID `json:"assigned_to"`
	DueDate     *string    `json:"due_date"`
	Cost        *int64     `json:"cost"`
	Notes       *string    `json:"notes"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

type TaskService struct {
	taskRepo domain.TaskRepository
	logRepo  domain.LogRepository
	logger   zerolog.Logger
}

func NewTaskService(taskRepo domain.TaskRepository, logRepo domain.LogRepository, logger zerolog.Logger) *TaskService {
	return &TaskService{
		taskRepo: taskRepo,
		logRepo:  logRepo,
		logger:   logger.With().Str("service", "task").Logger(),
	}
}

func (s *TaskService) Create(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, req CreateTaskRequest) (*domain.MaintenanceTask, error) {
	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if req.EquipmentID == uuid.Nil {
		return nil, fmt.Errorf("equipment_id is required")
	}

	priority := "normal"
	if req.Priority != "" {
		priority = req.Priority
	}
	var cost int64
	if req.Cost != nil {
		cost = *req.Cost
	}

	task := &domain.MaintenanceTask{
		ID:          uuid.New(),
		TenantID:    tenantID,
		ScheduleID:  req.ScheduleID,
		EquipmentID: req.EquipmentID,
		Title:       req.Title,
		Description: req.Description,
		Status:      "pending",
		Priority:    priority,
		AssignedTo:  req.AssignedTo,
		DueDate:     req.DueDate,
		Cost:        cost,
		Notes:       req.Notes,
		CreatedBy:   &userID,
	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}

	// Log creation.
	_ = s.logRepo.Create(ctx, &domain.MaintenanceLog{
		ID:          uuid.New(),
		TaskID:      task.ID,
		TenantID:    tenantID,
		Action:      "created",
		PerformedBy: &userID,
		Notes:       "Task created",
	})

	s.logger.Info().Str("task_id", task.ID.String()).Msg("maintenance task created")
	return task, nil
}

func (s *TaskService) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.MaintenanceTask, error) {
	return s.taskRepo.GetByID(ctx, id, tenantID)
}

func (s *TaskService) List(ctx context.Context, tenantID uuid.UUID, filter domain.TaskFilter) ([]*domain.MaintenanceTask, int64, error) {
	return s.taskRepo.List(ctx, tenantID, filter)
}

func (s *TaskService) Update(ctx context.Context, id, tenantID uuid.UUID, userID uuid.UUID, req UpdateTaskRequest) (*domain.MaintenanceTask, error) {
	existing, err := s.taskRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		existing.Title = *req.Title
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.Status != nil {
		existing.Status = *req.Status
	}
	if req.Priority != nil {
		existing.Priority = *req.Priority
	}
	if req.AssignedTo != nil {
		existing.AssignedTo = req.AssignedTo
	}
	if req.DueDate != nil {
		existing.DueDate = *req.DueDate
	}
	if req.Cost != nil {
		existing.Cost = *req.Cost
	}
	if req.Notes != nil {
		existing.Notes = *req.Notes
	}

	if err := s.taskRepo.Update(ctx, existing); err != nil {
		return nil, err
	}

	// Log update.
	_ = s.logRepo.Create(ctx, &domain.MaintenanceLog{
		ID:          uuid.New(),
		TaskID:      existing.ID,
		TenantID:    tenantID,
		Action:      "updated",
		PerformedBy: &userID,
		Notes:       "Task updated",
	})

	return existing, nil
}

func (s *TaskService) Complete(ctx context.Context, id, tenantID uuid.UUID, userID uuid.UUID) (*domain.MaintenanceTask, error) {
	existing, err := s.taskRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}

	if existing.Status == "completed" {
		return nil, domain.ErrTaskAlreadyCompleted
	}

	now := time.Now()
	existing.Status = "completed"
	existing.CompletedAt = &now

	if err := s.taskRepo.Update(ctx, existing); err != nil {
		return nil, err
	}

	// Log completion.
	_ = s.logRepo.Create(ctx, &domain.MaintenanceLog{
		ID:          uuid.New(),
		TaskID:      existing.ID,
		TenantID:    tenantID,
		Action:      "completed",
		PerformedBy: &userID,
		Notes:       "Task completed",
	})

	s.logger.Info().Str("task_id", existing.ID.String()).Msg("maintenance task completed")
	return existing, nil
}

func (s *TaskService) ListLogs(ctx context.Context, taskID, tenantID uuid.UUID, filter domain.LogFilter) ([]*domain.MaintenanceLog, int64, error) {
	return s.logRepo.ListByTask(ctx, taskID, tenantID, filter)
}
