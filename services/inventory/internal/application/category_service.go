package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/inventory/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

// CreateCategoryRequest holds the data needed to create a new category.
type CreateCategoryRequest struct {
	Name      string     `json:"name"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty"`
	Icon      string     `json:"icon,omitempty"`
	Color     string     `json:"color,omitempty"`
	SortOrder int        `json:"sort_order"`
}

// UpdateCategoryRequest holds optional fields for patching a category.
type UpdateCategoryRequest struct {
	Name      *string `json:"name,omitempty"`
	Icon      *string `json:"icon,omitempty"`
	Color     *string `json:"color,omitempty"`
	SortOrder *int    `json:"sort_order,omitempty"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// CategoryService implements the application-level use cases for categories.
type CategoryService struct {
	categoryRepo domain.CategoryRepository
	logger       zerolog.Logger
}

// NewCategoryService constructs a new CategoryService.
func NewCategoryService(
	categoryRepo domain.CategoryRepository,
	logger zerolog.Logger,
) *CategoryService {
	return &CategoryService{
		categoryRepo: categoryRepo,
		logger:       logger.With().Str("service", "category").Logger(),
	}
}

// Create validates the request and persists a new category.
func (s *CategoryService) Create(
	ctx context.Context,
	tenantID uuid.UUID,
	req CreateCategoryRequest,
) (*domain.Category, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("category name is required")
	}

	now := time.Now()
	category := &domain.Category{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Name:      req.Name,
		ParentID:  req.ParentID,
		Icon:      req.Icon,
		Color:     req.Color,
		SortOrder: req.SortOrder,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.categoryRepo.Create(ctx, category); err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to create category")
		return nil, fmt.Errorf("create category: %w", err)
	}

	s.logger.Info().
		Str("category_id", category.ID.String()).
		Str("tenant_id", tenantID.String()).
		Msg("category created")

	return category, nil
}

// GetByID retrieves a single category by ID within a tenant scope.
func (s *CategoryService) GetByID(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
) (*domain.Category, error) {
	category, err := s.categoryRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("category_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to get category by id")
		return nil, fmt.Errorf("get category by id: %w", err)
	}
	return category, nil
}

// List returns the full category tree for a tenant.
func (s *CategoryService) List(
	ctx context.Context,
	tenantID uuid.UUID,
) ([]*domain.Category, error) {
	categories, err := s.categoryRepo.List(ctx, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to list categories")
		return nil, fmt.Errorf("list categories: %w", err)
	}
	return categories, nil
}

// Update patches an existing category with non-nil fields from the request.
func (s *CategoryService) Update(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
	req UpdateCategoryRequest,
) (*domain.Category, error) {
	existing, err := s.categoryRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("category_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to fetch category for update")
		return nil, fmt.Errorf("update category – fetch: %w", err)
	}

	if req.Name != nil {
		if *req.Name == "" {
			return nil, fmt.Errorf("category name must not be empty")
		}
		existing.Name = *req.Name
	}
	if req.Icon != nil {
		existing.Icon = *req.Icon
	}
	if req.Color != nil {
		existing.Color = *req.Color
	}
	if req.SortOrder != nil {
		existing.SortOrder = *req.SortOrder
	}

	existing.UpdatedAt = time.Now()

	if err := s.categoryRepo.Update(ctx, existing); err != nil {
		s.logger.Error().Err(err).
			Str("category_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to update category")
		return nil, fmt.Errorf("update category: %w", err)
	}

	s.logger.Info().
		Str("category_id", id.String()).
		Str("tenant_id", tenantID.String()).
		Msg("category updated")

	return existing, nil
}

// Delete removes a category after verifying it has no children and no
// associated equipment.
func (s *CategoryService) Delete(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
) error {
	hasChildren, err := s.categoryRepo.HasChildren(ctx, id, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("category_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to check category children")
		return fmt.Errorf("delete category – check children: %w", err)
	}
	if hasChildren {
		return domain.ErrCategoryHasChildren
	}

	hasEquipment, err := s.categoryRepo.HasEquipment(ctx, id, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("category_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to check category equipment")
		return fmt.Errorf("delete category – check equipment: %w", err)
	}
	if hasEquipment {
		return domain.ErrCategoryHasEquipment
	}

	if err := s.categoryRepo.Delete(ctx, id, tenantID); err != nil {
		s.logger.Error().Err(err).
			Str("category_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to delete category")
		return fmt.Errorf("delete category: %w", err)
	}

	s.logger.Info().
		Str("category_id", id.String()).
		Str("tenant_id", tenantID.String()).
		Msg("category deleted")

	return nil
}
