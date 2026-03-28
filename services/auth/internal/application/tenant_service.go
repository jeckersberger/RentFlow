package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/auth/internal/domain"
)

// UpdateTenantRequest holds optional fields for patching a tenant.
type UpdateTenantRequest struct {
	Name           *string          `json:"name,omitempty"`
	Email          *string          `json:"email,omitempty"`
	Phone          *string          `json:"phone,omitempty"`
	AddressStreet  *string          `json:"address_street,omitempty"`
	AddressCity    *string          `json:"address_city,omitempty"`
	AddressZip     *string          `json:"address_zip,omitempty"`
	AddressCountry *string          `json:"address_country,omitempty"`
	TaxNumber      *string          `json:"tax_number,omitempty"`
	VatID          *string          `json:"vat_id,omitempty"`
	IBAN           *string          `json:"iban,omitempty"`
	BIC            *string          `json:"bic,omitempty"`
	BankName       *string          `json:"bank_name,omitempty"`
	LogoURL        *string          `json:"logo_url,omitempty"`
	Currency       *string          `json:"currency,omitempty"`
	Locale         *string          `json:"locale,omitempty"`
	Timezone       *string          `json:"timezone,omitempty"`
	InvoicePrefix  *string          `json:"invoice_prefix,omitempty"`
	Settings       *json.RawMessage `json:"settings,omitempty"`
}

// TenantService handles tenant-related business logic.
type TenantService struct {
	tenantRepo domain.TenantRepository
	logger     zerolog.Logger
}

// NewTenantService creates a new TenantService.
func NewTenantService(tenantRepo domain.TenantRepository, logger zerolog.Logger) *TenantService {
	return &TenantService{
		tenantRepo: tenantRepo,
		logger:     logger,
	}
}

// GetCurrent returns the tenant for the given ID.
func (s *TenantService) GetCurrent(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error) {
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		s.logger.Error().Err(err).Str("tenant_id", tenantID.String()).Msg("failed to get tenant")
		return nil, domain.ErrTenantNotFound
	}
	return tenant, nil
}

// Update patches non-nil fields on an existing tenant and persists the changes.
func (s *TenantService) Update(ctx context.Context, tenantID uuid.UUID, req UpdateTenantRequest) (*domain.Tenant, error) {
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		s.logger.Error().Err(err).Str("tenant_id", tenantID.String()).Msg("tenant not found for update")
		return nil, domain.ErrTenantNotFound
	}

	if req.Name != nil {
		tenant.Name = *req.Name
	}
	if req.Email != nil {
		tenant.Email = *req.Email
	}
	if req.Phone != nil {
		tenant.Phone = *req.Phone
	}
	if req.AddressStreet != nil {
		tenant.AddressStreet = *req.AddressStreet
	}
	if req.AddressCity != nil {
		tenant.AddressCity = *req.AddressCity
	}
	if req.AddressZip != nil {
		tenant.AddressZip = *req.AddressZip
	}
	if req.AddressCountry != nil {
		tenant.AddressCountry = *req.AddressCountry
	}
	if req.TaxNumber != nil {
		tenant.TaxNumber = *req.TaxNumber
	}
	if req.VatID != nil {
		tenant.VatID = *req.VatID
	}
	if req.IBAN != nil {
		tenant.IBAN = *req.IBAN
	}
	if req.BIC != nil {
		tenant.BIC = *req.BIC
	}
	if req.BankName != nil {
		tenant.BankName = *req.BankName
	}
	if req.LogoURL != nil {
		tenant.LogoURL = *req.LogoURL
	}
	if req.Currency != nil {
		tenant.Currency = *req.Currency
	}
	if req.Locale != nil {
		tenant.Locale = *req.Locale
	}
	if req.Timezone != nil {
		tenant.Timezone = *req.Timezone
	}
	if req.InvoicePrefix != nil {
		tenant.InvoicePrefix = *req.InvoicePrefix
	}
	if req.Settings != nil {
		tenant.Settings = *req.Settings
	}
	tenant.UpdatedAt = time.Now()

	if err := s.tenantRepo.Update(ctx, tenant); err != nil {
		s.logger.Error().Err(err).Str("tenant_id", tenantID.String()).Msg("failed to update tenant")
		return nil, fmt.Errorf("tenant update failed: %w", err)
	}

	s.logger.Info().Str("tenant_id", tenantID.String()).Msg("tenant updated")
	return tenant, nil
}
