package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/expense/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

type CreateCategoryRequest struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

type UpdateCategoryRequest struct {
	Name        *string `json:"name"`
	Code        *string `json:"code"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

type CategoryService struct {
	repo   domain.ExpenseCategoryRepository
	logger zerolog.Logger
}

func NewCategoryService(repo domain.ExpenseCategoryRepository, logger zerolog.Logger) *CategoryService {
	return &CategoryService{
		repo:   repo,
		logger: logger.With().Str("service", "expense-category").Logger(),
	}
}

func (s *CategoryService) Create(ctx context.Context, tenantID uuid.UUID, req CreateCategoryRequest) (*domain.ExpenseCategory, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	category := &domain.ExpenseCategory{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		IsActive:    true,
	}

	if err := s.repo.Create(ctx, category); err != nil {
		return nil, fmt.Errorf("create category: %w", err)
	}

	s.logger.Info().Str("category_id", category.ID.String()).Str("name", category.Name).Msg("expense category created")
	return category, nil
}

func (s *CategoryService) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.ExpenseCategory, error) {
	return s.repo.GetByID(ctx, id, tenantID)
}

func (s *CategoryService) List(ctx context.Context, tenantID uuid.UUID) ([]*domain.ExpenseCategory, error) {
	return s.repo.List(ctx, tenantID)
}

func (s *CategoryService) Update(ctx context.Context, id, tenantID uuid.UUID, req UpdateCategoryRequest) (*domain.ExpenseCategory, error) {
	existing, err := s.repo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Code != nil {
		existing.Code = *req.Code
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}
