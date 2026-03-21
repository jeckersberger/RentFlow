package application

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/services/project-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/project-service/internal/ports"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
)

type CustomerService struct {
	customerRepo ports.CustomerRepository
	logger       logger.Logger
}

func NewCustomerService(
	customerRepo ports.CustomerRepository,
	logger logger.Logger,
) *CustomerService {
	return &CustomerService{
		customerRepo: customerRepo,
		logger:       logger,
	}
}

func (s *CustomerService) CreateCustomer(ctx context.Context, cmd CreateCustomerCommand) (*CustomerDTO, error) {
	customer := domain.NewCustomer(uuid.New().String(), cmd.TenantID, cmd.Name)
	customer.Email = cmd.Email
	customer.Phone = cmd.Phone
	customer.TaxID = cmd.TaxID
	customer.Notes = cmd.Notes

	if cmd.AddressCity != "" {
		if err := customer.SetAddress(cmd.AddressStreet, cmd.AddressCity, cmd.AddressPostcode, cmd.AddressCountry); err != nil {
			return nil, fmt.Errorf("invalid address: %w", err)
		}
	}

	if err := customer.Validate(); err != nil {
		return nil, err
	}

	if err := s.customerRepo.Create(ctx, customer); err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	return CustomerToDTO(customer), nil
}

func (s *CustomerService) GetCustomer(ctx context.Context, tenantID, customerID string) (*CustomerDTO, error) {
	customer, err := s.customerRepo.GetByID(ctx, tenantID, customerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &domain.DomainError{
				Code:    "NOT_FOUND",
				Message: "customer not found",
			}
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return CustomerToDTO(customer), nil
}

func (s *CustomerService) ListCustomers(ctx context.Context, query ListCustomersQuery) (*PaginatedResult, error) {
	result, err := s.customerRepo.List(ctx, query.TenantID, query.Limit, query.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list customers: %w", err)
	}

	dtos := make([]*CustomerDTO, len(result.Items))
	for i, c := range result.Items {
		dtos[i] = CustomerToDTO(c)
	}

	return &PaginatedResult{
		Data:   dtos,
		Total:  result.Total,
		Limit:  result.Limit,
		Offset: result.Offset,
	}, nil
}

func (s *CustomerService) SearchCustomers(ctx context.Context, query SearchCustomersQuery) (*PaginatedResult, error) {
	result, err := s.customerRepo.Search(ctx, query.TenantID, query.SearchTerm, query.Limit, query.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search customers: %w", err)
	}

	dtos := make([]*CustomerDTO, len(result.Items))
	for i, c := range result.Items {
		dtos[i] = CustomerToDTO(c)
	}

	return &PaginatedResult{
		Data:   dtos,
		Total:  result.Total,
		Limit:  result.Limit,
		Offset: result.Offset,
	}, nil
}

func (s *CustomerService) UpdateCustomer(ctx context.Context, cmd UpdateCustomerCommand) (*CustomerDTO, error) {
	customer, err := s.customerRepo.GetByID(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &domain.DomainError{
				Code:    "NOT_FOUND",
				Message: "customer not found",
			}
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	if err := customer.UpdateBasicInfo(cmd.Name, cmd.Email, cmd.Phone); err != nil {
		return nil, err
	}

	if cmd.AddressCity != "" {
		if err := customer.SetAddress(cmd.AddressStreet, cmd.AddressCity, cmd.AddressPostcode, cmd.AddressCountry); err != nil {
			return nil, fmt.Errorf("invalid address: %w", err)
		}
	}

	customer.SetTaxID(cmd.TaxID)
	customer.SetNotes(cmd.Notes)

	if err := s.customerRepo.Update(ctx, customer); err != nil {
		return nil, fmt.Errorf("failed to update customer: %w", err)
	}

	return CustomerToDTO(customer), nil
}

func (s *CustomerService) DeleteCustomer(ctx context.Context, tenantID, customerID string) error {
	if err := s.customerRepo.Delete(ctx, tenantID, customerID); err != nil {
		if err == sql.ErrNoRows {
			return &domain.DomainError{
				Code:    "NOT_FOUND",
				Message: "customer not found",
			}
		}
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	return nil
}
