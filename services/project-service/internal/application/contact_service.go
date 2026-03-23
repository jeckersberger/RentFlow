package application

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/project-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/project-service/internal/ports"
)

type ContactService struct {
	contactRepo ports.ContactRepository
	logger      logger.Logger
}

func NewContactService(
	contactRepo ports.ContactRepository,
	logger logger.Logger,
) *ContactService {
	return &ContactService{
		contactRepo: contactRepo,
		logger:      logger,
	}
}

func (s *ContactService) CreateContact(ctx context.Context, cmd CreateContactCommand) (*ContactDTO, error) {
	contact := domain.NewContact(uuid.New().String(), cmd.TenantID, domain.ContactType(cmd.Type))
	contact.CompanyName = cmd.CompanyName
	contact.FirstName = cmd.FirstName
	contact.LastName = cmd.LastName
	contact.Email = cmd.Email
	contact.Phone = cmd.Phone
	contact.Mobile = cmd.Mobile
	contact.Website = cmd.Website
	contact.Street = cmd.Street
	contact.HouseNumber = cmd.HouseNumber
	contact.Zip = cmd.Zip
	contact.City = cmd.City
	if cmd.Country != "" {
		contact.Country = cmd.Country
	}
	contact.VatID = cmd.VatID
	contact.Notes = cmd.Notes
	contact.Tags = cmd.Tags
	contact.CreatedBy = cmd.CreatedBy

	if err := contact.Validate(); err != nil {
		return nil, err
	}

	if err := s.contactRepo.Create(ctx, contact); err != nil {
		return nil, fmt.Errorf("failed to create contact: %w", err)
	}

	return ContactToDTO(contact), nil
}

func (s *ContactService) GetContact(ctx context.Context, tenantID, contactID string) (*ContactDTO, error) {
	contact, err := s.contactRepo.GetByID(ctx, tenantID, contactID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &domain.DomainError{
				Code:    "NOT_FOUND",
				Message: "contact not found",
			}
		}
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	return ContactToDTO(contact), nil
}

func (s *ContactService) ListContacts(ctx context.Context, query ListContactsQuery) (*PaginatedResult, error) {
	var result *ports.ContactListResult
	var err error

	if query.SearchTerm != "" {
		result, err = s.contactRepo.Search(ctx, query.TenantID, query.SearchTerm, query.Limit, query.Offset)
	} else {
		result, err = s.contactRepo.List(ctx, query.TenantID, query.Limit, query.Offset)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list contacts: %w", err)
	}

	dtos := make([]*ContactDTO, len(result.Items))
	for i, c := range result.Items {
		dtos[i] = ContactToDTO(c)
	}

	return &PaginatedResult{
		Data:   dtos,
		Total:  result.Total,
		Limit:  result.Limit,
		Offset: result.Offset,
	}, nil
}

func (s *ContactService) UpdateContact(ctx context.Context, cmd UpdateContactCommand) (*ContactDTO, error) {
	contact, err := s.contactRepo.GetByID(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &domain.DomainError{
				Code:    "NOT_FOUND",
				Message: "contact not found",
			}
		}
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	contact.Type = domain.ContactType(cmd.Type)
	contact.CompanyName = cmd.CompanyName
	contact.FirstName = cmd.FirstName
	contact.LastName = cmd.LastName
	contact.Email = cmd.Email
	contact.Phone = cmd.Phone
	contact.Mobile = cmd.Mobile
	contact.Website = cmd.Website
	contact.Street = cmd.Street
	contact.HouseNumber = cmd.HouseNumber
	contact.Zip = cmd.Zip
	contact.City = cmd.City
	if cmd.Country != "" {
		contact.Country = cmd.Country
	}
	contact.VatID = cmd.VatID
	contact.Notes = cmd.Notes
	contact.Tags = cmd.Tags
	contact.Update()

	if err := contact.Validate(); err != nil {
		return nil, err
	}

	if err := s.contactRepo.Update(ctx, contact); err != nil {
		return nil, fmt.Errorf("failed to update contact: %w", err)
	}

	return ContactToDTO(contact), nil
}

func (s *ContactService) DeleteContact(ctx context.Context, tenantID, contactID string) error {
	if err := s.contactRepo.Delete(ctx, tenantID, contactID); err != nil {
		if err == sql.ErrNoRows {
			return &domain.DomainError{
				Code:    "NOT_FOUND",
				Message: "contact not found",
			}
		}
		return fmt.Errorf("failed to delete contact: %w", err)
	}

	return nil
}
