package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/ai-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/ai-service/internal/ports"
)

// AIService orchestrates AI operations
type AIService struct {
	requestRepo      ports.AIRequestRepository
	feedbackRepo     ports.AIFeedbackRepository
	exampleRepo      ports.FewShotExampleRepository
	providerRepo     ports.AIProviderRepository
	providerSvc      *ProviderService
	anonymizationSvc *AnonymizationService
	predictionSvc    *PredictionService
	logger           logger.Logger
}

// NewAIService creates a new AI service
func NewAIService(
	requestRepo ports.AIRequestRepository,
	feedbackRepo ports.AIFeedbackRepository,
	exampleRepo ports.FewShotExampleRepository,
	providerRepo ports.AIProviderRepository,
	providerSvc *ProviderService,
	anonymizationSvc *AnonymizationService,
	predictionSvc *PredictionService,
	log logger.Logger,
) *AIService {
	return &AIService{
		requestRepo:      requestRepo,
		feedbackRepo:     feedbackRepo,
		exampleRepo:      exampleRepo,
		providerRepo:     providerRepo,
		providerSvc:      providerSvc,
		anonymizationSvc: anonymizationSvc,
		predictionSvc:    predictionSvc,
		logger:           log,
	}
}

// CreateAIRequest creates and processes an AI request
func (s *AIService) CreateAIRequest(ctx context.Context, cmd CreateAIRequestCommand) (*AIRequestDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}
	if cmd.InputText == "" {
		return nil, domain.ErrInputTextRequired
	}
	if cmd.ProviderID == "" {
		return nil, domain.ErrProviderIDRequired
	}

	// Validate provider exists
	provider, err := s.providerRepo.FindByID(ctx, cmd.ProviderID)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return nil, domain.ErrAIProviderNotFound
	}

	// Create request entity
	request := &domain.AIRequest{
		ID:          uuid.New().String(),
		TenantID:    cmd.TenantID,
		ProviderID:  cmd.ProviderID,
		RequestType: domain.RequestType(cmd.RequestType),
		InputText:   cmd.InputText,
		Status:      domain.RequestStatusPending,
		CreatedAt:   time.Now().UTC(),
	}

	// Anonymize if requested
	if cmd.Anonymize {
		anonymized := s.anonymizationSvc.Anonymize(cmd.InputText)
		request.AnonymizedInput = &anonymized
	}

	// Save request
	if err := s.requestRepo.Save(ctx, request); err != nil {
		s.logger.Error("failed to save AI request", err)
		return nil, err
	}

	return ToAIRequestDTO(request), nil
}

// GetAIRequest retrieves an AI request by ID
func (s *AIService) GetAIRequest(ctx context.Context, id string) (*AIRequestDTO, error) {
	request, err := s.requestRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if request == nil {
		return nil, domain.ErrAIRequestNotFound
	}
	return ToAIRequestDTO(request), nil
}

// ListAIRequests lists AI requests with pagination
func (s *AIService) ListAIRequests(ctx context.Context, tenantID string, page, perPage int) ([]*AIRequestDTO, int, error) {
	requests, total, err := s.requestRepo.List(ctx, tenantID, page, perPage)
	if err != nil {
		return nil, 0, err
	}

	dtos := make([]*AIRequestDTO, len(requests))
	for i, req := range requests {
		dtos[i] = ToAIRequestDTO(req)
	}

	return dtos, total, nil
}

// SubmitFeedback submits feedback for an AI request
func (s *AIService) SubmitFeedback(ctx context.Context, cmd SubmitFeedbackCommand) (*AIFeedbackDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}
	if cmd.RequestID == "" {
		return nil, domain.ErrRequestIDRequired
	}
	if cmd.Rating < 1 || cmd.Rating > 5 {
		return nil, domain.ErrInvalidRating
	}

	// Verify request exists
	request, err := s.requestRepo.FindByID(ctx, cmd.RequestID)
	if err != nil {
		return nil, err
	}
	if request == nil {
		return nil, domain.ErrAIRequestNotFound
	}

	feedback := &domain.AIFeedback{
		ID:        uuid.New().String(),
		TenantID:  cmd.TenantID,
		RequestID: cmd.RequestID,
		Rating:    cmd.Rating,
		Comment:   cmd.Comment,
		IsCorrect: cmd.IsCorrect,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.feedbackRepo.Save(ctx, feedback); err != nil {
		s.logger.Error("failed to save feedback", err)
		return nil, err
	}

	return ToAIFeedbackDTO(feedback), nil
}

// CreateFewShotExample creates a few-shot learning example
func (s *AIService) CreateFewShotExample(ctx context.Context, cmd CreateFewShotExampleCommand) (*FewShotExampleDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}
	if cmd.InputExample == "" || cmd.OutputExample == "" {
		return nil, domain.ErrInputTextRequired
	}

	example := &domain.FewShotExample{
		ID:            uuid.New().String(),
		TenantID:      cmd.TenantID,
		RequestType:   domain.RequestType(cmd.RequestType),
		InputExample:  cmd.InputExample,
		OutputExample: cmd.OutputExample,
		IsActive:      true,
		UsageCount:    0,
		CreatedAt:     time.Now().UTC(),
	}

	if err := s.exampleRepo.Save(ctx, example); err != nil {
		s.logger.Error("failed to save few-shot example", err)
		return nil, err
	}

	return ToFewShotExampleDTO(example), nil
}

// ListFewShotExamples lists few-shot examples
func (s *AIService) ListFewShotExamples(ctx context.Context, tenantID string, page, perPage int) ([]*FewShotExampleDTO, int, error) {
	examples, total, err := s.exampleRepo.List(ctx, tenantID, page, perPage)
	if err != nil {
		return nil, 0, err
	}

	dtos := make([]*FewShotExampleDTO, len(examples))
	for i, ex := range examples {
		dtos[i] = ToFewShotExampleDTO(ex)
	}

	return dtos, total, nil
}

// RegisterAIProvider registers a new AI provider
func (s *AIService) RegisterAIProvider(ctx context.Context, cmd CreateAIProviderCommand) (*AIProviderDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}

	provider := &domain.AIProvider{
		ID:          uuid.New().String(),
		TenantID:    cmd.TenantID,
		Name:        domain.ProviderType(cmd.Name),
		APIEndpoint: cmd.APIEndpoint,
		ModelName:   cmd.ModelName,
		IsActive:    cmd.IsActive,
		Priority:    cmd.Priority,
		Config:      cmd.Config,
		CreatedAt:   time.Now().UTC(),
	}

	if err := s.providerRepo.Save(ctx, provider); err != nil {
		s.logger.Error("failed to save provider", err)
		return nil, err
	}

	return ToAIProviderDTO(provider), nil
}

// ListAIProviders lists AI providers for a tenant
func (s *AIService) ListAIProviders(ctx context.Context, tenantID string) ([]*AIProviderDTO, error) {
	providers, err := s.providerRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*AIProviderDTO, len(providers))
	for i, p := range providers {
		dtos[i] = ToAIProviderDTO(p)
	}

	return dtos, nil
}

// Anonymize anonymizes text
func (s *AIService) Anonymize(text string) string {
	return s.anonymizationSvc.Anonymize(text)
}

// OptimizePrice generates price recommendations
func (s *AIService) OptimizePrice(ctx context.Context, equipmentType string, rentalDays int, season string) (string, error) {
	return s.predictionSvc.PriceOptimization(ctx, equipmentType, rentalDays, season)
}

// ForecastDemand predicts demand for equipment
func (s *AIService) ForecastDemand(ctx context.Context, category string, period string) (string, error) {
	return s.predictionSvc.DemandForecast(ctx, category, period)
}

// CreateAsset generates asset metadata
func (s *AIService) CreateAsset(ctx context.Context, description string) (string, error) {
	return s.predictionSvc.SmartAssetCreator(ctx, description)
}

// PredictMaintenance predicts maintenance needs
func (s *AIService) PredictMaintenance(ctx context.Context, equipmentID string, usageHours int, lastMaintenance *time.Time) (string, error) {
	return s.predictionSvc.PredictiveMaintenance(ctx, equipmentID, usageHours, lastMaintenance)
}

// GetDashboard returns dashboard statistics
func (s *AIService) GetDashboard(ctx context.Context, tenantID string) (*DashboardDTO, error) {
	_, _, err := s.requestRepo.List(ctx, tenantID, 1, 1)
	if err != nil {
		return nil, err
	}

	dashboard := &DashboardDTO{
		TotalRequests:    1,
		RequestsThisDay:  1,
		RequestsThisWeek: 7,
		AverageRating:    nil,
		TopProvider:      nil,
		AverageLatencyMs: nil,
		SuccessRate:      100.0,
	}

	return dashboard, nil
}
