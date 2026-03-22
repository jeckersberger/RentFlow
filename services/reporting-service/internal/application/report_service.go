package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/services/reporting-service/internal/domain"
)

type ReportRepository interface {
	CreateDefinition(ctx context.Context, def *domain.ReportDefinition) error
	UpdateDefinition(ctx context.Context, def *domain.ReportDefinition) error
	GetDefinitionByID(ctx context.Context, tenantID, defID uuid.UUID) (*domain.ReportDefinition, error)
	ListDefinitions(ctx context.Context, tenantID uuid.UUID) ([]domain.ReportDefinition, error)
	CreateRun(ctx context.Context, run *domain.ReportRun) error
	UpdateRun(ctx context.Context, run *domain.ReportRun) error
	GetRunByID(ctx context.Context, tenantID, runID uuid.UUID) (*domain.ReportRun, error)
	ListRuns(ctx context.Context, tenantID, defID uuid.UUID) ([]domain.ReportRun, error)
}

type ReportService struct {
	repo ReportRepository
}

func NewReportService(repo ReportRepository) *ReportService {
	return &ReportService{repo: repo}
}

func (s *ReportService) CreateDefinition(ctx context.Context, cmd *CreateReportDefinitionCommand) (*domain.ReportDefinition, error) {
	def := &domain.ReportDefinition{
		ID:              uuid.New(),
		TenantID:        cmd.TenantID,
		Name:            cmd.Name,
		ReportType:      domain.ReportType(cmd.ReportType),
		Description:     cmd.Description,
		Parameters:      cmd.Parameters,
		ScheduleCron:    cmd.ScheduleCron,
		EmailRecipients: cmd.EmailRecipients,
		Format:          domain.ReportFormat(cmd.Format),
		IsActive:        true,
		CreatedBy:       cmd.CreatedBy,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.repo.CreateDefinition(ctx, def); err != nil {
		return nil, err
	}

	return def, nil
}

func (s *ReportService) UpdateDefinition(ctx context.Context, cmd *UpdateReportDefinitionCommand) (*domain.ReportDefinition, error) {
	def, err := s.repo.GetDefinitionByID(ctx, cmd.TenantID, cmd.ReportDefinitionID)
	if err != nil {
		return nil, err
	}

	def.Name = cmd.Name
	def.Description = cmd.Description
	def.Parameters = cmd.Parameters
	def.ScheduleCron = cmd.ScheduleCron
	def.EmailRecipients = cmd.EmailRecipients
	def.Format = domain.ReportFormat(cmd.Format)
	def.IsActive = cmd.IsActive
	def.UpdatedAt = time.Now()

	if err := s.repo.UpdateDefinition(ctx, def); err != nil {
		return nil, err
	}

	return def, nil
}

func (s *ReportService) GetDefinition(ctx context.Context, tenantID, defID uuid.UUID) (*domain.ReportDefinition, error) {
	return s.repo.GetDefinitionByID(ctx, tenantID, defID)
}

func (s *ReportService) ListDefinitions(ctx context.Context, tenantID uuid.UUID) ([]domain.ReportDefinition, error) {
	return s.repo.ListDefinitions(ctx, tenantID)
}

func (s *ReportService) GenerateReport(ctx context.Context, cmd *GenerateReportCommand) (*domain.ReportRun, error) {
	def, err := s.repo.GetDefinitionByID(ctx, cmd.TenantID, cmd.ReportDefinitionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report definition: %w", err)
	}

	run := &domain.ReportRun{
		ID:                uuid.New(),
		TenantID:          cmd.TenantID,
		ReportDefinitionID: cmd.ReportDefinitionID,
		Status:            domain.ReportRunStatusQueued,
		ParametersUsed:    cmd.Parameters,
		CreatedAt:         time.Now(),
	}

	if cmd.PeriodStart != nil {
		ps, err := time.Parse(time.RFC3339, *cmd.PeriodStart)
		if err == nil {
			run.PeriodStart = &ps
		}
	}

	if cmd.PeriodEnd != nil {
		pe, err := time.Parse(time.RFC3339, *cmd.PeriodEnd)
		if err == nil {
			run.PeriodEnd = &pe
		}
	}

	if err := s.repo.CreateRun(ctx, run); err != nil {
		return nil, fmt.Errorf("failed to create report run: %w", err)
	}

	_ = def

	return run, nil
}

func (s *ReportService) GetRun(ctx context.Context, tenantID, runID uuid.UUID) (*domain.ReportRun, error) {
	return s.repo.GetRunByID(ctx, tenantID, runID)
}

func (s *ReportService) ListRuns(ctx context.Context, tenantID, defID uuid.UUID) ([]domain.ReportRun, error) {
	return s.repo.ListRuns(ctx, tenantID, defID)
}

func (s *ReportService) ListAllRuns(ctx context.Context, tenantID uuid.UUID) ([]domain.ReportRun, error) {
	return s.repo.ListRuns(ctx, tenantID, uuid.Nil)
}

func marshalJSON(data map[string]interface{}) (json.RawMessage, error) {
	if len(data) == 0 {
		return json.RawMessage(`{}`), nil
	}
	return json.Marshal(data)
}

func marshalStringSlice(data []string) (json.RawMessage, error) {
	if len(data) == 0 {
		return json.RawMessage(`[]`), nil
	}
	return json.Marshal(data)
}
