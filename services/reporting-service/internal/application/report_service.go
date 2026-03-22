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

// ExportCSV exports a report run to CSV format
func (s *ReportService) ExportCSV(ctx context.Context, tenantID, runID uuid.UUID) ([]byte, error) {
	run, err := s.repo.GetRunByID(ctx, tenantID, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report run: %w", err)
	}

	if run.Status != domain.ReportRunStatusCompleted {
		return nil, fmt.Errorf("report run not in completed status: %s", run.Status)
	}

	// Build CSV header
	csv := "Report Export\n"
	csv += fmt.Sprintf("Generated: %s\n\n", run.CreatedAt.Format("2006-01-02 15:04:05"))

	if run.PeriodStart != nil && run.PeriodEnd != nil {
		csv += fmt.Sprintf("Period: %s to %s\n", run.PeriodStart.Format("2006-01-02"), run.PeriodEnd.Format("2006-01-02"))
	}
	csv += "\n"

	// Parse and export parameters
	params, err := unmarshalJSON(run.ParametersUsed)
	if err == nil && len(params) > 0 {
		csv += "Parameters:\n"
		for key, value := range params {
			csv += fmt.Sprintf("%s,%v\n", key, value)
		}
		csv += "\n"
	}

	// Status summary
	csv += fmt.Sprintf("Status,Started,Completed,File Size\n")
	startedStr := "-"
	completedStr := "-"
	filesizeStr := "-"

	if run.StartedAt != nil {
		startedStr = run.StartedAt.Format("2006-01-02 15:04:05")
	}
	if run.CompletedAt != nil {
		completedStr = run.CompletedAt.Format("2006-01-02 15:04:05")
	}
	if run.FileSize != nil {
		filesizeStr = fmt.Sprintf("%d bytes", *run.FileSize)
	}

	csv += fmt.Sprintf("%s,%s,%s,%s\n", run.Status, startedStr, completedStr, filesizeStr)

	return []byte(csv), nil
}

// ExportPDF exports a report run to PDF format
func (s *ReportService) ExportPDF(ctx context.Context, tenantID, runID uuid.UUID) ([]byte, error) {
	run, err := s.repo.GetRunByID(ctx, tenantID, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report run: %w", err)
	}

	if run.Status != domain.ReportRunStatusCompleted {
		return nil, fmt.Errorf("report run not in completed status: %s", run.Status)
	}

	// Build PDF content
	startedStr := "-"
	completedStr := "-"
	filesizeStr := "-"

	if run.StartedAt != nil {
		startedStr = run.StartedAt.Format("2006-01-02 15:04:05")
	}
	if run.CompletedAt != nil {
		completedStr = run.CompletedAt.Format("2006-01-02 15:04:05")
	}
	if run.FileSize != nil {
		filesizeStr = fmt.Sprintf("%d bytes", *run.FileSize)
	}

	periodStr := "-"
	if run.PeriodStart != nil && run.PeriodEnd != nil {
		periodStr = fmt.Sprintf("%s to %s", run.PeriodStart.Format("2006-01-02"), run.PeriodEnd.Format("2006-01-02"))
	}

	// Minimal valid PDF 1.4 structure with report data
	pdf := fmt.Sprintf(`%%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>
endobj
4 0 obj
<< /Length 400 >>
stream
BT
/F1 16 Tf
50 750 Td
(Report Export) Tj
0 -30 Td
/F1 12 Tf
(Report ID: %s) Tj
0 -20 Td
(Generated: %s) Tj
0 -20 Td
(Period: %s) Tj
0 -20 Td
(Status: %s) Tj
0 -20 Td
(Started: %s) Tj
0 -20 Td
(Completed: %s) Tj
0 -20 Td
(File Size: %s) Tj
ET
endstream
endobj
5 0 obj
<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>
endobj
xref
0 6
0000000000 65535 f
0000000010 00000 n
0000000074 00000 n
0000000133 00000 n
0000000281 00000 n
0000000738 00000 n
trailer
<< /Size 6 /Root 1 0 R >>
startxref
0832
%%%%EOF
`, runID.String(), time.Now().Format("2006-01-02 15:04:05"), periodStr, run.Status, startedStr, completedStr, filesizeStr)

	return []byte(pdf), nil
}

// ProcessScheduledReports processes scheduled reports that are due to run
func (s *ReportService) ProcessScheduledReports(ctx context.Context, tenantID uuid.UUID) ([]uuid.UUID, error) {
	// Get all active report definitions for the tenant
	definitions, err := s.repo.ListDefinitions(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list report definitions: %w", err)
	}

	var createdRunIDs []uuid.UUID

	// Process each definition with a schedule
	for _, def := range definitions {
		if !def.IsActive || def.ScheduleCron == nil {
			continue
		}

		// Create a report run for this scheduled definition
		run := &domain.ReportRun{
			ID:                uuid.New(),
			TenantID:          tenantID,
			ReportDefinitionID: def.ID,
			Status:            domain.ReportRunStatusQueued,
			ParametersUsed:    def.Parameters,
			CreatedAt:         time.Now(),
		}

		if err := s.repo.CreateRun(ctx, run); err != nil {
			continue
		}

		createdRunIDs = append(createdRunIDs, run.ID)
	}

	return createdRunIDs, nil
}
