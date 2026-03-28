package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/audit/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

type CreateAuditPolicyRequest struct {
	EntityType    string `json:"entity_type"`
	RetentionDays *int   `json:"retention_days"`
	LogReads      *bool  `json:"log_reads"`
	LogWrites     *bool  `json:"log_writes"`
	IsActive      *bool  `json:"is_active"`
}

type UpdateAuditPolicyRequest struct {
	EntityType    *string `json:"entity_type"`
	RetentionDays *int    `json:"retention_days"`
	LogReads      *bool   `json:"log_reads"`
	LogWrites     *bool   `json:"log_writes"`
	IsActive      *bool   `json:"is_active"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

type AuditPolicyService struct {
	repo   domain.AuditPolicyRepository
	logger zerolog.Logger
}

func NewAuditPolicyService(repo domain.AuditPolicyRepository, logger zerolog.Logger) *AuditPolicyService {
	return &AuditPolicyService{
		repo:   repo,
		logger: logger.With().Str("service", "audit_policy").Logger(),
	}
}

func (s *AuditPolicyService) Create(ctx context.Context, tenantID uuid.UUID, req CreateAuditPolicyRequest) (*domain.AuditPolicy, error) {
	if req.EntityType == "" {
		return nil, fmt.Errorf("entity_type is required")
	}

	retentionDays := 365
	if req.RetentionDays != nil {
		retentionDays = *req.RetentionDays
	}
	logReads := false
	if req.LogReads != nil {
		logReads = *req.LogReads
	}
	logWrites := true
	if req.LogWrites != nil {
		logWrites = *req.LogWrites
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	policy := &domain.AuditPolicy{
		ID:            uuid.New(),
		TenantID:      tenantID,
		EntityType:    req.EntityType,
		RetentionDays: retentionDays,
		LogReads:      logReads,
		LogWrites:     logWrites,
		IsActive:      isActive,
	}

	if err := s.repo.Create(ctx, policy); err != nil {
		return nil, fmt.Errorf("create audit policy: %w", err)
	}

	s.logger.Info().Str("policy_id", policy.ID.String()).Msg("audit policy created")
	return policy, nil
}

func (s *AuditPolicyService) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.AuditPolicy, error) {
	return s.repo.GetByID(ctx, id, tenantID)
}

func (s *AuditPolicyService) List(ctx context.Context, tenantID uuid.UUID, filter domain.AuditPolicyFilter) ([]*domain.AuditPolicy, int64, error) {
	return s.repo.List(ctx, tenantID, filter)
}

func (s *AuditPolicyService) Update(ctx context.Context, id, tenantID uuid.UUID, req UpdateAuditPolicyRequest) (*domain.AuditPolicy, error) {
	existing, err := s.repo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}

	if req.EntityType != nil {
		existing.EntityType = *req.EntityType
	}
	if req.RetentionDays != nil {
		existing.RetentionDays = *req.RetentionDays
	}
	if req.LogReads != nil {
		existing.LogReads = *req.LogReads
	}
	if req.LogWrites != nil {
		existing.LogWrites = *req.LogWrites
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}
