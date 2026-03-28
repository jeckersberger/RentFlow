package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/warehouse/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

// CreateInventoryCheckRequest holds the data needed to start an inventory check.
type CreateInventoryCheckRequest struct {
	ZoneID *uuid.UUID `json:"zone_id,omitempty"`
	Notes  string     `json:"notes,omitempty"`
}

// ScanItemRequest holds the data needed to register a scanned equipment item.
type ScanItemRequest struct {
	EquipmentID uuid.UUID `json:"equipment_id"`
	Notes       string    `json:"notes,omitempty"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// InventoryCheckService implements the application-level use cases for inventory checks.
type InventoryCheckService struct {
	checkRepo domain.InventoryCheckRepository
	logger    zerolog.Logger
}

// NewInventoryCheckService constructs a new InventoryCheckService.
func NewInventoryCheckService(
	checkRepo domain.InventoryCheckRepository,
	logger zerolog.Logger,
) *InventoryCheckService {
	return &InventoryCheckService{
		checkRepo: checkRepo,
		logger:    logger.With().Str("service", "inventory_check").Logger(),
	}
}

// Create starts a new inventory check.
func (s *InventoryCheckService) Create(
	ctx context.Context,
	tenantID uuid.UUID,
	userID uuid.UUID,
	req CreateInventoryCheckRequest,
) (*domain.InventoryCheck, error) {
	now := time.Now()
	check := &domain.InventoryCheck{
		ID:        uuid.New(),
		TenantID:  tenantID,
		ZoneID:    req.ZoneID,
		Status:    domain.CheckStatusInProgress,
		StartedBy: &userID,
		StartedAt: now,
		Notes:     req.Notes,
	}

	if err := s.checkRepo.Create(ctx, check); err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to create inventory check")
		return nil, fmt.Errorf("create inventory check: %w", err)
	}

	s.logger.Info().
		Str("check_id", check.ID.String()).
		Str("tenant_id", tenantID.String()).
		Msg("inventory check started")

	return check, nil
}

// List returns a paginated list of inventory checks for a tenant.
func (s *InventoryCheckService) List(
	ctx context.Context,
	tenantID uuid.UUID,
	page int,
	perPage int,
) ([]*domain.InventoryCheck, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	items, total, err := s.checkRepo.List(ctx, tenantID, page, perPage)
	if err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to list inventory checks")
		return nil, 0, fmt.Errorf("list inventory checks: %w", err)
	}
	return items, total, nil
}

// Complete finalises an inventory check, calculating actual_count and discrepancy_count.
func (s *InventoryCheckService) Complete(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
) (*domain.InventoryCheck, error) {
	check, err := s.checkRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("check_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to fetch inventory check for completion")
		return nil, fmt.Errorf("complete inventory check – fetch: %w", err)
	}

	if check.Status == domain.CheckStatusCompleted {
		return nil, domain.ErrCheckAlreadyDone
	}

	expected, err := s.checkRepo.CountExpected(ctx, id)
	if err != nil {
		s.logger.Error().Err(err).
			Str("check_id", id.String()).
			Msg("failed to count expected items")
		return nil, fmt.Errorf("complete inventory check – count expected: %w", err)
	}

	actual, err := s.checkRepo.CountFound(ctx, id)
	if err != nil {
		s.logger.Error().Err(err).
			Str("check_id", id.String()).
			Msg("failed to count found items")
		return nil, fmt.Errorf("complete inventory check – count found: %w", err)
	}

	now := time.Now()
	discrepancy := expected - actual
	if discrepancy < 0 {
		discrepancy = -discrepancy
	}

	check.Status = domain.CheckStatusCompleted
	check.CompletedAt = &now
	check.ExpectedCount = expected
	check.ActualCount = actual
	check.DiscrepancyCount = discrepancy

	if err := s.checkRepo.Complete(ctx, check); err != nil {
		s.logger.Error().Err(err).
			Str("check_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to complete inventory check")
		return nil, fmt.Errorf("complete inventory check: %w", err)
	}

	s.logger.Info().
		Str("check_id", id.String()).
		Str("tenant_id", tenantID.String()).
		Int("expected", expected).
		Int("actual", actual).
		Int("discrepancy", discrepancy).
		Msg("inventory check completed")

	return check, nil
}

// ScanItem registers a scanned equipment item in an inventory check.
func (s *InventoryCheckService) ScanItem(
	ctx context.Context,
	checkID uuid.UUID,
	tenantID uuid.UUID,
	req ScanItemRequest,
) error {
	if req.EquipmentID == uuid.Nil {
		return fmt.Errorf("equipment_id is required")
	}

	// Verify the check exists and is in progress.
	check, err := s.checkRepo.GetByID(ctx, checkID, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("check_id", checkID.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to fetch inventory check for scan")
		return fmt.Errorf("scan item – fetch check: %w", err)
	}
	if check.Status == domain.CheckStatusCompleted {
		return domain.ErrCheckAlreadyDone
	}

	if err := s.checkRepo.ScanItem(ctx, checkID, req.EquipmentID); err != nil {
		s.logger.Error().Err(err).
			Str("check_id", checkID.String()).
			Str("equipment_id", req.EquipmentID.String()).
			Msg("failed to scan item")
		return fmt.Errorf("scan item: %w", err)
	}

	s.logger.Info().
		Str("check_id", checkID.String()).
		Str("equipment_id", req.EquipmentID.String()).
		Msg("item scanned")

	return nil
}

// GetResult returns the inventory check together with all its items.
func (s *InventoryCheckService) GetResult(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
) (*domain.InventoryCheckResult, error) {
	check, err := s.checkRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("check_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to fetch inventory check for result")
		return nil, fmt.Errorf("get result – fetch check: %w", err)
	}

	items, err := s.checkRepo.GetItems(ctx, id)
	if err != nil {
		s.logger.Error().Err(err).
			Str("check_id", id.String()).
			Msg("failed to fetch inventory check items")
		return nil, fmt.Errorf("get result – fetch items: %w", err)
	}

	return &domain.InventoryCheckResult{
		Check: check,
		Items: items,
	}, nil
}
