package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/customer/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

type CreateCustomerRequest struct {
	CompanyName            string `json:"company_name"`
	CustomerNumber         string `json:"customer_number,omitempty"`
	Email                  string `json:"email,omitempty"`
	Phone                  string `json:"phone,omitempty"`
	Mobile                 string `json:"mobile,omitempty"`
	Website                string `json:"website,omitempty"`
	BillingAddressStreet   string `json:"billing_address_street,omitempty"`
	BillingAddressCity     string `json:"billing_address_city,omitempty"`
	BillingAddressZip      string `json:"billing_address_zip,omitempty"`
	BillingAddressCountry  string `json:"billing_address_country,omitempty"`
	ShippingAddressStreet  string `json:"shipping_address_street,omitempty"`
	ShippingAddressCity    string `json:"shipping_address_city,omitempty"`
	ShippingAddressZip     string `json:"shipping_address_zip,omitempty"`
	ShippingAddressCountry string `json:"shipping_address_country,omitempty"`
	TaxID                  string `json:"tax_id,omitempty"`
	Notes                  string `json:"notes,omitempty"`
}

type UpdateCustomerRequest struct {
	CompanyName            *string `json:"company_name,omitempty"`
	CustomerNumber         *string `json:"customer_number,omitempty"`
	Email                  *string `json:"email,omitempty"`
	Phone                  *string `json:"phone,omitempty"`
	Mobile                 *string `json:"mobile,omitempty"`
	Website                *string `json:"website,omitempty"`
	BillingAddressStreet   *string `json:"billing_address_street,omitempty"`
	BillingAddressCity     *string `json:"billing_address_city,omitempty"`
	BillingAddressZip      *string `json:"billing_address_zip,omitempty"`
	BillingAddressCountry  *string `json:"billing_address_country,omitempty"`
	ShippingAddressStreet  *string `json:"shipping_address_street,omitempty"`
	ShippingAddressCity    *string `json:"shipping_address_city,omitempty"`
	ShippingAddressZip     *string `json:"shipping_address_zip,omitempty"`
	ShippingAddressCountry *string `json:"shipping_address_country,omitempty"`
	TaxID                  *string `json:"tax_id,omitempty"`
	Notes                  *string `json:"notes,omitempty"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

type CustomerService struct {
	customerRepo domain.CustomerRepository
	logger       zerolog.Logger
}

func NewCustomerService(customerRepo domain.CustomerRepository, logger zerolog.Logger) *CustomerService {
	return &CustomerService{
		customerRepo: customerRepo,
		logger:       logger.With().Str("service", "customer").Logger(),
	}
}

func (s *CustomerService) Create(ctx context.Context, tenantID uuid.UUID, req CreateCustomerRequest) (*domain.Customer, error) {
	if req.CompanyName == "" {
		return nil, fmt.Errorf("company name is required")
	}

	now := time.Now()
	customer := &domain.Customer{
		ID:                     uuid.New(),
		TenantID:               tenantID,
		CompanyName:            req.CompanyName,
		CustomerNumber:         req.CustomerNumber,
		Email:                  req.Email,
		Phone:                  req.Phone,
		Mobile:                 req.Mobile,
		Website:                req.Website,
		BillingAddressStreet:   req.BillingAddressStreet,
		BillingAddressCity:     req.BillingAddressCity,
		BillingAddressZip:      req.BillingAddressZip,
		BillingAddressCountry:  req.BillingAddressCountry,
		ShippingAddressStreet:  req.ShippingAddressStreet,
		ShippingAddressCity:    req.ShippingAddressCity,
		ShippingAddressZip:     req.ShippingAddressZip,
		ShippingAddressCountry: req.ShippingAddressCountry,
		TaxID:                  req.TaxID,
		Notes:                  req.Notes,
		IsActive:               true,
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	if err := s.customerRepo.Create(ctx, customer); err != nil {
		s.logger.Error().Err(err).Str("tenant_id", tenantID.String()).Msg("failed to create customer")
		return nil, fmt.Errorf("create customer: %w", err)
	}

	s.logger.Info().Str("customer_id", customer.ID.String()).Str("tenant_id", tenantID.String()).Msg("customer created")
	return customer, nil
}

func (s *CustomerService) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.Customer, error) {
	customer, err := s.customerRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get customer: %w", err)
	}
	return customer, nil
}

func (s *CustomerService) List(ctx context.Context, tenantID uuid.UUID, filter domain.CustomerFilter) ([]*domain.Customer, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 {
		filter.PerPage = 20
	}

	items, total, err := s.customerRepo.List(ctx, tenantID, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("list customers: %w", err)
	}
	return items, total, nil
}

func (s *CustomerService) Update(ctx context.Context, id, tenantID uuid.UUID, req UpdateCustomerRequest) (*domain.Customer, error) {
	existing, err := s.customerRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, fmt.Errorf("update customer - fetch: %w", err)
	}

	if req.CompanyName != nil {
		existing.CompanyName = *req.CompanyName
	}
	if req.CustomerNumber != nil {
		existing.CustomerNumber = *req.CustomerNumber
	}
	if req.Email != nil {
		existing.Email = *req.Email
	}
	if req.Phone != nil {
		existing.Phone = *req.Phone
	}
	if req.Mobile != nil {
		existing.Mobile = *req.Mobile
	}
	if req.Website != nil {
		existing.Website = *req.Website
	}
	if req.BillingAddressStreet != nil {
		existing.BillingAddressStreet = *req.BillingAddressStreet
	}
	if req.BillingAddressCity != nil {
		existing.BillingAddressCity = *req.BillingAddressCity
	}
	if req.BillingAddressZip != nil {
		existing.BillingAddressZip = *req.BillingAddressZip
	}
	if req.BillingAddressCountry != nil {
		existing.BillingAddressCountry = *req.BillingAddressCountry
	}
	if req.ShippingAddressStreet != nil {
		existing.ShippingAddressStreet = *req.ShippingAddressStreet
	}
	if req.ShippingAddressCity != nil {
		existing.ShippingAddressCity = *req.ShippingAddressCity
	}
	if req.ShippingAddressZip != nil {
		existing.ShippingAddressZip = *req.ShippingAddressZip
	}
	if req.ShippingAddressCountry != nil {
		existing.ShippingAddressCountry = *req.ShippingAddressCountry
	}
	if req.TaxID != nil {
		existing.TaxID = *req.TaxID
	}
	if req.Notes != nil {
		existing.Notes = *req.Notes
	}

	existing.UpdatedAt = time.Now()

	if err := s.customerRepo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("update customer: %w", err)
	}

	s.logger.Info().Str("customer_id", id.String()).Str("tenant_id", tenantID.String()).Msg("customer updated")
	return existing, nil
}

func (s *CustomerService) Delete(ctx context.Context, id, tenantID uuid.UUID) error {
	if err := s.customerRepo.Delete(ctx, id, tenantID); err != nil {
		return fmt.Errorf("delete customer: %w", err)
	}

	s.logger.Info().Str("customer_id", id.String()).Msg("customer deleted (soft)")
	return nil
}

func (s *CustomerService) Search(ctx context.Context, tenantID uuid.UUID, query string, page, perPage int) ([]*domain.Customer, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	items, total, err := s.customerRepo.Search(ctx, tenantID, query, page, perPage)
	if err != nil {
		return nil, 0, fmt.Errorf("search customers: %w", err)
	}
	return items, total, nil
}
