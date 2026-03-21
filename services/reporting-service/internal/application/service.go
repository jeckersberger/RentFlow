package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/reporting-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/reporting-service/internal/ports"
)

type ReportDTO struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Parameters  map[string]interface{} `json:"parameters"`
	Status      string                 `json:"status"`
	GeneratedAt *time.Time             `json:"generated_at,omitempty"`
	FileRef     string                 `json:"file_ref,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

type GenerateReportCommand struct {
	TenantID   string                 `json:"tenant_id"`
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
}

type ReportService struct {
	repo   ports.ReportRepository
	logger *logger.Logger
}

func NewReportService(repo ports.ReportRepository, log *logger.Logger) *ReportService {
	return &ReportService{repo: repo, logger: log}
}

func (s *ReportService) GenerateReport(ctx context.Context, cmd GenerateReportCommand) (*ReportDTO, error) {
	if cmd.TenantID == "" || cmd.Type == "" {
		return nil, domain.ErrInvalidInput
	}

	paramsJSON, _ := json.Marshal(cmd.Parameters)
	var params map[string]interface{}
	json.Unmarshal(paramsJSON, &params)

	report := &domain.Report{
		ID:         fmt.Sprintf("report_%d", time.Now().UnixNano()),
		TenantID:   cmd.TenantID,
		Type:       domain.ReportType(cmd.Type),
		Parameters: params,
		Status:     domain.ReportStatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.repo.Create(ctx, report); err != nil {
		s.logger.Error("Failed to create report", err)
		return nil, err
	}

	return &ReportDTO{
		ID:         report.ID,
		Type:       string(report.Type),
		Parameters: report.Parameters,
		Status:     string(report.Status),
		CreatedAt:  report.CreatedAt,
	}, nil
}

func (s *ReportService) GetReport(ctx context.Context, tenantID, id string) (*ReportDTO, error) {
	report, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if report == nil {
		return nil, domain.ErrReportNotFound
	}

	return &ReportDTO{
		ID:          report.ID,
		Type:        string(report.Type),
		Parameters:  report.Parameters,
		Status:      string(report.Status),
		GeneratedAt: report.GeneratedAt,
		FileRef:     report.FileRef,
		CreatedAt:   report.CreatedAt,
	}, nil
}

func (s *ReportService) ListReports(ctx context.Context, tenantID string) ([]*ReportDTO, error) {
	reports, err := s.repo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*ReportDTO, len(reports))
	for i, r := range reports {
		dtos[i] = &ReportDTO{
			ID:          r.ID,
			Type:        string(r.Type),
			Parameters:  r.Parameters,
			Status:      string(r.Status),
			GeneratedAt: r.GeneratedAt,
			FileRef:     r.FileRef,
			CreatedAt:   r.CreatedAt,
		}
	}
	return dtos, nil
}

func (s *ReportService) GetKPIs(ctx context.Context, tenantID string) ([]domain.KPI, error) {
	kpis := []domain.KPI{
		{Name: "Total Revenue", Value: 150000, PreviousValue: 140000, ChangePercent: 7.14, Period: "MTD"},
		{Name: "Equipment Utilization", Value: 82.5, PreviousValue: 78.0, ChangePercent: 5.77, Period: "MTD"},
		{Name: "Inventory Value", Value: 450000, PreviousValue: 455000, ChangePercent: -1.10, Period: "Current"},
		{Name: "Active Projects", Value: 12, PreviousValue: 10, ChangePercent: 20.0, Period: "Active"},
	}
	return kpis, nil
}
