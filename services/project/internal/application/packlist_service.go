package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/project/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

type CreatePacklistRequest struct {
	ProjectID uuid.UUID `json:"project_id"`
	Name      string    `json:"name"`
}

type UpdatePacklistStatusRequest struct {
	Status string `json:"status"`
}

type AddPacklistItemRequest struct {
	EquipmentID     uuid.UUID `json:"equipment_id"`
	QuantityPlanned int       `json:"quantity_planned"`
	Notes           string    `json:"notes,omitempty"`
}

type UpdatePackedRequest struct {
	QuantityPacked int `json:"quantity_packed"`
}

type UpdateReturnedRequest struct {
	QuantityReturned int `json:"quantity_returned"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

type PacklistService struct {
	packlistRepo domain.PacklistRepository
	projectRepo  domain.ProjectRepository
	logger       zerolog.Logger
}

func NewPacklistService(
	packlistRepo domain.PacklistRepository,
	projectRepo domain.ProjectRepository,
	logger zerolog.Logger,
) *PacklistService {
	return &PacklistService{
		packlistRepo: packlistRepo,
		projectRepo:  projectRepo,
		logger:       logger.With().Str("service", "packlist").Logger(),
	}
}

func (s *PacklistService) Create(ctx context.Context, tenantID, userID uuid.UUID, req CreatePacklistRequest) (*domain.Packlist, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("packlist name is required")
	}

	// Verify project exists
	if _, err := s.projectRepo.GetByID(ctx, req.ProjectID, tenantID); err != nil {
		return nil, fmt.Errorf("create packlist – verify project: %w", err)
	}

	packlist := &domain.Packlist{
		ID:        uuid.New(),
		ProjectID: req.ProjectID,
		TenantID:  tenantID,
		Name:      req.Name,
		Status:    domain.PacklistStatusDraft,
		CreatedBy: &userID,
		CreatedAt: time.Now(),
	}

	if err := s.packlistRepo.Create(ctx, packlist); err != nil {
		return nil, fmt.Errorf("create packlist: %w", err)
	}

	s.logger.Info().Str("packlist_id", packlist.ID.String()).Msg("packlist created")
	return packlist, nil
}

func (s *PacklistService) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.Packlist, error) {
	pl, err := s.packlistRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get packlist: %w", err)
	}
	return pl, nil
}

func (s *PacklistService) ListByProject(ctx context.Context, tenantID, projectID uuid.UUID) ([]*domain.Packlist, error) {
	// Verify project belongs to tenant
	if _, err := s.projectRepo.GetByID(ctx, projectID, tenantID); err != nil {
		return nil, fmt.Errorf("list packlists – verify project: %w", err)
	}

	items, err := s.packlistRepo.ListByProject(ctx, projectID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list packlists: %w", err)
	}
	return items, nil
}

func (s *PacklistService) UpdateStatus(ctx context.Context, id, tenantID uuid.UUID, req UpdatePacklistStatusRequest) error {
	// Verify packlist exists and belongs to tenant
	if _, err := s.packlistRepo.GetByID(ctx, id, tenantID); err != nil {
		return fmt.Errorf("update packlist status – verify: %w", err)
	}

	if err := s.packlistRepo.UpdateStatus(ctx, id, tenantID, req.Status); err != nil {
		return fmt.Errorf("update packlist status: %w", err)
	}

	s.logger.Info().Str("packlist_id", id.String()).Str("status", req.Status).Msg("packlist status updated")
	return nil
}

func (s *PacklistService) AddItem(ctx context.Context, tenantID, packlistID uuid.UUID, req AddPacklistItemRequest) (*domain.PacklistItem, error) {
	// Verify packlist exists and belongs to tenant
	if _, err := s.packlistRepo.GetByID(ctx, packlistID, tenantID); err != nil {
		return nil, fmt.Errorf("add packlist item – verify: %w", err)
	}

	if req.QuantityPlanned < 1 {
		req.QuantityPlanned = 1
	}

	item := &domain.PacklistItem{
		ID:              uuid.New(),
		PacklistID:      packlistID,
		EquipmentID:     req.EquipmentID,
		QuantityPlanned: req.QuantityPlanned,
		Notes:           req.Notes,
	}

	if err := s.packlistRepo.AddItem(ctx, item); err != nil {
		return nil, fmt.Errorf("add packlist item: %w", err)
	}

	return item, nil
}

func (s *PacklistService) GetItems(ctx context.Context, tenantID, packlistID uuid.UUID) ([]*domain.PacklistItem, error) {
	// Verify packlist belongs to tenant
	if _, err := s.packlistRepo.GetByID(ctx, packlistID, tenantID); err != nil {
		return nil, fmt.Errorf("get packlist items – verify: %w", err)
	}

	items, err := s.packlistRepo.GetItems(ctx, packlistID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get packlist items: %w", err)
	}
	return items, nil
}

func (s *PacklistService) UpdateItemPacked(ctx context.Context, tenantID, itemID uuid.UUID, req UpdatePackedRequest, userID uuid.UUID) error {
	packedBy := &userID
	if err := s.packlistRepo.UpdateItemPacked(ctx, itemID, req.QuantityPacked, packedBy, tenantID); err != nil {
		return fmt.Errorf("update packed quantity: %w", err)
	}
	return nil
}

func (s *PacklistService) UpdateItemReturned(ctx context.Context, tenantID, itemID uuid.UUID, req UpdateReturnedRequest) error {
	if err := s.packlistRepo.UpdateItemReturned(ctx, itemID, req.QuantityReturned, tenantID); err != nil {
		return fmt.Errorf("update returned quantity: %w", err)
	}
	return nil
}
