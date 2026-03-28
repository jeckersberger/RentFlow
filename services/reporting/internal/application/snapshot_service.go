package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/reporting/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

type GenerateSnapshotRequest struct {
	Title       string          `json:"title"`
	Data        json.RawMessage `json:"data"`
	FilePath    string          `json:"file_path"`
	FileSize    int             `json:"file_size"`
	Format      string          `json:"format"`
	PeriodStart string          `json:"period_start"`
	PeriodEnd   string          `json:"period_end"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

type SnapshotService struct {
	snapRepo domain.ReportSnapshotRepository
	defRepo  domain.ReportDefinitionRepository
	logger   zerolog.Logger
}

func NewSnapshotService(snapRepo domain.ReportSnapshotRepository, defRepo domain.ReportDefinitionRepository, logger zerolog.Logger) *SnapshotService {
	return &SnapshotService{
		snapRepo: snapRepo,
		defRepo:  defRepo,
		logger:   logger.With().Str("service", "snapshot").Logger(),
	}
}

func (s *SnapshotService) Generate(ctx context.Context, definitionID, tenantID, userID uuid.UUID, req GenerateSnapshotRequest) (*domain.ReportSnapshot, error) {
	def, err := s.defRepo.GetByID(ctx, definitionID, tenantID)
	if err != nil {
		return nil, err
	}

	title := req.Title
	if title == "" {
		title = def.Name
	}
	format := req.Format
	if format == "" {
		format = def.Format
	}
	data := json.RawMessage("{}")
	if req.Data != nil {
		data = req.Data
	}

	snap := &domain.ReportSnapshot{
		ID:           uuid.New(),
		DefinitionID: definitionID,
		TenantID:     tenantID,
		Title:        title,
		Data:         data,
		FilePath:     req.FilePath,
		FileSize:     req.FileSize,
		Format:       format,
		PeriodStart:  req.PeriodStart,
		PeriodEnd:    req.PeriodEnd,
		GeneratedBy:  &userID,
	}

	if err := s.snapRepo.Create(ctx, snap); err != nil {
		return nil, fmt.Errorf("create snapshot: %w", err)
	}

	s.logger.Info().Str("snapshot_id", snap.ID.String()).Str("definition_id", definitionID.String()).Msg("report snapshot generated")
	return snap, nil
}

func (s *SnapshotService) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.ReportSnapshot, error) {
	return s.snapRepo.GetByID(ctx, id, tenantID)
}

func (s *SnapshotService) List(ctx context.Context, tenantID uuid.UUID, filter domain.SnapshotFilter) ([]*domain.ReportSnapshot, int64, error) {
	return s.snapRepo.List(ctx, tenantID, filter)
}
