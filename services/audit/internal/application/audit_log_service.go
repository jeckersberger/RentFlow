package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/audit/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

type CreateAuditLogRequest struct {
	UserID     *uuid.UUID       `json:"user_id"`
	Action     string           `json:"action"`
	EntityType string           `json:"entity_type"`
	EntityID   *uuid.UUID       `json:"entity_id"`
	OldData    json.RawMessage  `json:"old_data"`
	NewData    json.RawMessage  `json:"new_data"`
	IPAddress  string           `json:"ip_address"`
	UserAgent  string           `json:"user_agent"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

type AuditLogService struct {
	repo   domain.AuditLogRepository
	logger zerolog.Logger
}

func NewAuditLogService(repo domain.AuditLogRepository, logger zerolog.Logger) *AuditLogService {
	return &AuditLogService{
		repo:   repo,
		logger: logger.With().Str("service", "audit_log").Logger(),
	}
}

func (s *AuditLogService) Create(ctx context.Context, tenantID uuid.UUID, req CreateAuditLogRequest) (*domain.AuditLog, error) {
	if req.Action == "" {
		return nil, fmt.Errorf("action is required")
	}
	if req.EntityType == "" {
		return nil, fmt.Errorf("entity_type is required")
	}

	entry := &domain.AuditLog{
		ID:         uuid.New(),
		TenantID:   tenantID,
		UserID:     req.UserID,
		Action:     req.Action,
		EntityType: req.EntityType,
		EntityID:   req.EntityID,
		OldData:    req.OldData,
		NewData:    req.NewData,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
	}

	if err := s.repo.Create(ctx, entry); err != nil {
		return nil, fmt.Errorf("create audit log: %w", err)
	}

	s.logger.Info().Str("audit_log_id", entry.ID.String()).Str("action", entry.Action).Msg("audit log created")
	return entry, nil
}

func (s *AuditLogService) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.AuditLog, error) {
	return s.repo.GetByID(ctx, id, tenantID)
}

func (s *AuditLogService) List(ctx context.Context, tenantID uuid.UUID, filter domain.AuditLogFilter) ([]*domain.AuditLog, int64, error) {
	return s.repo.List(ctx, tenantID, filter)
}
