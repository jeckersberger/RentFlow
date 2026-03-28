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

type CreateContactRequest struct {
	CustomerID uuid.UUID `json:"customer_id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Email      string    `json:"email,omitempty"`
	Phone      string    `json:"phone,omitempty"`
	Mobile     string    `json:"mobile,omitempty"`
	Position   string    `json:"position,omitempty"`
	IsPrimary  bool      `json:"is_primary"`
	Notes      string    `json:"notes,omitempty"`
}

type UpdateContactRequest struct {
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Email     *string `json:"email,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	Mobile    *string `json:"mobile,omitempty"`
	Position  *string `json:"position,omitempty"`
	IsPrimary *bool   `json:"is_primary,omitempty"`
	Notes     *string `json:"notes,omitempty"`
}

type CreateContactNoteRequest struct {
	Content string `json:"content"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

type ContactService struct {
	contactRepo domain.ContactRepository
	noteRepo    domain.ContactNoteRepository
	logger      zerolog.Logger
}

func NewContactService(contactRepo domain.ContactRepository, noteRepo domain.ContactNoteRepository, logger zerolog.Logger) *ContactService {
	return &ContactService{
		contactRepo: contactRepo,
		noteRepo:    noteRepo,
		logger:      logger.With().Str("service", "contact").Logger(),
	}
}

func (s *ContactService) Create(ctx context.Context, tenantID uuid.UUID, req CreateContactRequest) (*domain.Contact, error) {
	if req.FirstName == "" || req.LastName == "" {
		return nil, fmt.Errorf("first_name and last_name are required")
	}

	now := time.Now()
	contact := &domain.Contact{
		ID:         uuid.New(),
		TenantID:   tenantID,
		CustomerID: req.CustomerID,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Email:      req.Email,
		Phone:      req.Phone,
		Mobile:     req.Mobile,
		Position:   req.Position,
		IsPrimary:  req.IsPrimary,
		Notes:      req.Notes,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.contactRepo.Create(ctx, contact); err != nil {
		s.logger.Error().Err(err).Str("tenant_id", tenantID.String()).Msg("failed to create contact")
		return nil, fmt.Errorf("create contact: %w", err)
	}

	s.logger.Info().Str("contact_id", contact.ID.String()).Msg("contact created")
	return contact, nil
}

func (s *ContactService) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.Contact, error) {
	contact, err := s.contactRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get contact: %w", err)
	}
	return contact, nil
}

func (s *ContactService) ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]*domain.Contact, error) {
	items, err := s.contactRepo.ListByCustomer(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("list contacts: %w", err)
	}
	return items, nil
}

func (s *ContactService) Update(ctx context.Context, id, tenantID uuid.UUID, req UpdateContactRequest) (*domain.Contact, error) {
	existing, err := s.contactRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, fmt.Errorf("update contact - fetch: %w", err)
	}

	if req.FirstName != nil {
		existing.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		existing.LastName = *req.LastName
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
	if req.Position != nil {
		existing.Position = *req.Position
	}
	if req.IsPrimary != nil {
		existing.IsPrimary = *req.IsPrimary
	}
	if req.Notes != nil {
		existing.Notes = *req.Notes
	}

	existing.UpdatedAt = time.Now()

	if err := s.contactRepo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("update contact: %w", err)
	}

	s.logger.Info().Str("contact_id", id.String()).Msg("contact updated")
	return existing, nil
}

func (s *ContactService) Delete(ctx context.Context, id, tenantID uuid.UUID) error {
	if err := s.contactRepo.Delete(ctx, id, tenantID); err != nil {
		return fmt.Errorf("delete contact: %w", err)
	}

	s.logger.Info().Str("contact_id", id.String()).Msg("contact deleted")
	return nil
}

// --- Contact Notes ---

func (s *ContactService) AddNote(ctx context.Context, tenantID uuid.UUID, contactID uuid.UUID, userID *uuid.UUID, req CreateContactNoteRequest) (*domain.ContactNote, error) {
	if req.Content == "" {
		return nil, fmt.Errorf("note content is required")
	}

	note := &domain.ContactNote{
		ID:        uuid.New(),
		TenantID:  tenantID,
		ContactID: contactID,
		Content:   req.Content,
		CreatedBy: userID,
		CreatedAt: time.Now(),
	}

	if err := s.noteRepo.Create(ctx, note); err != nil {
		return nil, fmt.Errorf("add contact note: %w", err)
	}

	s.logger.Info().Str("note_id", note.ID.String()).Str("contact_id", contactID.String()).Msg("contact note added")
	return note, nil
}

func (s *ContactService) ListNotes(ctx context.Context, contactID uuid.UUID) ([]*domain.ContactNote, error) {
	items, err := s.noteRepo.ListByContact(ctx, contactID)
	if err != nil {
		return nil, fmt.Errorf("list contact notes: %w", err)
	}
	return items, nil
}
