package application

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/services/reporting-service/internal/domain"
)

type CreateReportDefinitionRequest struct {
	Name            string                 `json:"name"`
	ReportType      domain.ReportType      `json:"report_type"`
	Description     string                 `json:"description"`
	Parameters      map[string]interface{} `json:"parameters"`
	ScheduleCron    *string                `json:"schedule_cron"`
	EmailRecipients []string               `json:"email_recipients"`
	Format          domain.ReportFormat    `json:"format"`
}

type UpdateReportDefinitionRequest struct {
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Parameters      map[string]interface{} `json:"parameters"`
	ScheduleCron    *string                `json:"schedule_cron"`
	EmailRecipients []string               `json:"email_recipients"`
	Format          domain.ReportFormat    `json:"format"`
	IsActive        bool                   `json:"is_active"`
}

type ReportDefinitionResponse struct {
	ID              uuid.UUID              `json:"id"`
	Name            string                 `json:"name"`
	ReportType      domain.ReportType      `json:"report_type"`
	Description     string                 `json:"description"`
	Parameters      map[string]interface{} `json:"parameters"`
	ScheduleCron    *string                `json:"schedule_cron"`
	EmailRecipients []string               `json:"email_recipients"`
	Format          domain.ReportFormat    `json:"format"`
	IsActive        bool                   `json:"is_active"`
	CreatedBy       uuid.UUID              `json:"created_by"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

type GenerateReportRequest struct {
	Parameters map[string]interface{} `json:"parameters"`
	PeriodStart *time.Time            `json:"period_start"`
	PeriodEnd   *time.Time            `json:"period_end"`
}

type ReportRunResponse struct {
	ID                 uuid.UUID                    `json:"id"`
	ReportDefinitionID uuid.UUID                    `json:"report_definition_id"`
	Status             domain.ReportRunStatus       `json:"status"`
	ParametersUsed     map[string]interface{}       `json:"parameters_used"`
	PeriodStart        *time.Time                   `json:"period_start"`
	PeriodEnd          *time.Time                   `json:"period_end"`
	FilePath           *string                      `json:"file_path"`
	FileSize           *int64                       `json:"file_size"`
	ErrorMessage       *string                      `json:"error_message"`
	StartedAt          *time.Time                   `json:"started_at"`
	CompletedAt        *time.Time                   `json:"completed_at"`
	CreatedAt          time.Time                    `json:"created_at"`
}

type KPIDashboardRequest struct {
	Period domain.Period `form:"period" json:"period"`
}

type KPITrendRequest struct {
	KPIType domain.KPIType `form:"kpi_type" json:"kpi_type"`
	Periods int            `form:"periods" json:"periods"`
}

func unmarshalJSON(data []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	if len(data) == 0 {
		return map[string]interface{}{}, nil
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func unmarshalStringSlice(data []byte) ([]string, error) {
	var result []string
	if len(data) == 0 {
		return []string{}, nil
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func NewReportDefinitionResponse(def *domain.ReportDefinition) (*ReportDefinitionResponse, error) {
	params, err := unmarshalJSON(def.Parameters)
	if err != nil {
		return nil, err
	}

	recipients, err := unmarshalStringSlice(def.EmailRecipients)
	if err != nil {
		return nil, err
	}

	return &ReportDefinitionResponse{
		ID:              def.ID,
		Name:            def.Name,
		ReportType:      def.ReportType,
		Description:     def.Description,
		Parameters:      params,
		ScheduleCron:    def.ScheduleCron,
		EmailRecipients: recipients,
		Format:          def.Format,
		IsActive:        def.IsActive,
		CreatedBy:       def.CreatedBy,
		CreatedAt:       def.CreatedAt,
		UpdatedAt:       def.UpdatedAt,
	}, nil
}

func NewReportRunResponse(run *domain.ReportRun) (*ReportRunResponse, error) {
	params, err := unmarshalJSON(run.ParametersUsed)
	if err != nil {
		return nil, err
	}

	return &ReportRunResponse{
		ID:                 run.ID,
		ReportDefinitionID: run.ReportDefinitionID,
		Status:             run.Status,
		ParametersUsed:     params,
		PeriodStart:        run.PeriodStart,
		PeriodEnd:          run.PeriodEnd,
		FilePath:           run.FilePath,
		FileSize:           run.FileSize,
		ErrorMessage:       run.ErrorMessage,
		StartedAt:          run.StartedAt,
		CompletedAt:        run.CompletedAt,
		CreatedAt:          run.CreatedAt,
	}, nil
}
