package application

import (
	"context"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/auth-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/auth-service/internal/ports"
)

// TenantService handles tenant-related business logic
type TenantService struct {
	tenantRepo ports.TenantRepository
	logger     logger.Logger
}

// NewTenantService creates a new tenant service
func NewTenantService(tenantRepo ports.TenantRepository, log logger.Logger) *TenantService {
	return &TenantService{
		tenantRepo: tenantRepo,
		logger:     log,
	}
}

// CreateTenant creates a new tenant
func (s *TenantService) CreateTenant(ctx context.Context, cmd CreateTenantCommand) (*TenantDTO, error) {
	// Generate tenant ID
	tenantID := generateID()

	// Create tenant aggregate
	tenant := domain.NewTenant(tenantID, cmd.Name, cmd.Slug)
	tenant.Settings = domain.TenantSettings{
		DefaultLanguage: cmd.DefaultLanguage,
		Currency:        cmd.Currency,
		TaxRate:         cmd.TaxRate,
		InvoicePrefix:   cmd.InvoicePrefix,
	}

	// Save tenant
	if err := s.tenantRepo.Save(ctx, tenant); err != nil {
		s.logger.Error("failed to save tenant", err, "name", cmd.Name)
		return nil, err
	}

	s.logger.Info("tenant created", "id", tenantID, "name", cmd.Name, "slug", cmd.Slug)

	return s.toTenantDTO(tenant), nil
}

// GetTenant gets a tenant by ID
func (s *TenantService) GetTenant(ctx context.Context, query GetTenantByIDQuery) (*TenantDTO, error) {
	tenant, err := s.tenantRepo.FindByID(ctx, query.TenantID)
	if err != nil {
		s.logger.Error("failed to find tenant", err, "id", query.TenantID)
		return nil, err
	}
	if tenant == nil {
		return nil, domain.ErrTenantNotFound
	}
	return s.toTenantDTO(tenant), nil
}

// GetTenantBySlug gets a tenant by slug
func (s *TenantService) GetTenantBySlug(ctx context.Context, slug string) (*TenantDTO, error) {
	tenant, err := s.tenantRepo.FindBySlug(ctx, slug)
	if err != nil {
		s.logger.Error("failed to find tenant", err, "slug", slug)
		return nil, err
	}
	if tenant == nil {
		return nil, domain.ErrTenantNotFound
	}
	return s.toTenantDTO(tenant), nil
}

// ListTenants lists all tenants
func (s *TenantService) ListTenants(ctx context.Context, query ListTenantsQuery) (*PaginatedResult, error) {
	tenants, total, err := s.tenantRepo.List(ctx, query.Page, query.PerPage)
	if err != nil {
		s.logger.Error("failed to list tenants", err)
		return nil, err
	}

	tenantDTOs := make([]TenantDTO, len(tenants))
	for i, tenant := range tenants {
		tenantDTOs[i] = *s.toTenantDTO(tenant)
	}

	totalPages := (total + query.PerPage - 1) / query.PerPage

	return &PaginatedResult{
		Data:       tenantDTOs,
		Page:       query.Page,
		PerPage:    query.PerPage,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// UpdateTenant updates a tenant
func (s *TenantService) UpdateTenant(ctx context.Context, cmd UpdateTenantCommand) (*TenantDTO, error) {
	// Get existing tenant
	tenant, err := s.tenantRepo.FindByID(ctx, cmd.TenantID)
	if err != nil {
		s.logger.Error("failed to find tenant", err, "id", cmd.TenantID)
		return nil, err
	}
	if tenant == nil {
		return nil, domain.ErrTenantNotFound
	}

	// Update tenant
	settings := domain.TenantSettings{
		DefaultLanguage: cmd.DefaultLanguage,
		Currency:        cmd.Currency,
		TaxRate:         cmd.TaxRate,
		InvoicePrefix:   cmd.InvoicePrefix,
	}
	tenant.UpdateTenant(cmd.Name, cmd.Slug, settings)

	// Save tenant
	if err := s.tenantRepo.Save(ctx, tenant); err != nil {
		s.logger.Error("failed to save tenant", err, "id", cmd.TenantID)
		return nil, err
	}

	s.logger.Info("tenant updated", "id", cmd.TenantID)

	return s.toTenantDTO(tenant), nil
}

// DeactivateTenant deactivates a tenant
func (s *TenantService) DeactivateTenant(ctx context.Context, tenantID string) error {
	// Get existing tenant
	tenant, err := s.tenantRepo.FindByID(ctx, tenantID)
	if err != nil {
		s.logger.Error("failed to find tenant", err, "id", tenantID)
		return err
	}
	if tenant == nil {
		return domain.ErrTenantNotFound
	}

	// Deactivate tenant
	tenant.Deactivate()

	// Save tenant
	if err := s.tenantRepo.Save(ctx, tenant); err != nil {
		s.logger.Error("failed to save tenant", err, "id", tenantID)
		return err
	}

	s.logger.Info("tenant deactivated", "id", tenantID)
	return nil
}

// Helper function
func (s *TenantService) toTenantDTO(tenant *domain.Tenant) *TenantDTO {
	return &TenantDTO{
		ID:              tenant.ID,
		Name:            tenant.Name,
		Slug:            tenant.Slug,
		Status:          string(tenant.Status),
		DefaultLanguage: tenant.Settings.DefaultLanguage,
		Currency:        tenant.Settings.Currency,
		TaxRate:         tenant.Settings.TaxRate,
		InvoicePrefix:   tenant.Settings.InvoicePrefix,
		CreatedAt:       tenant.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       tenant.UpdatedAt.Format(time.RFC3339),
	}
}
