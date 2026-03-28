package domain

import (
	"context"

	"github.com/google/uuid"
)

// CustomerRepository defines persistence operations for customers.
type CustomerRepository interface {
	Create(ctx context.Context, customer *Customer) error
	GetByID(ctx context.Context, id, tenantID uuid.UUID) (*Customer, error)
	List(ctx context.Context, tenantID uuid.UUID, filter CustomerFilter) ([]*Customer, int64, error)
	Update(ctx context.Context, customer *Customer) error
	Delete(ctx context.Context, id, tenantID uuid.UUID) error
	Search(ctx context.Context, tenantID uuid.UUID, query string, page, perPage int) ([]*Customer, int64, error)
}

// ContactRepository defines persistence operations for contacts.
type ContactRepository interface {
	Create(ctx context.Context, contact *Contact) error
	GetByID(ctx context.Context, id, tenantID uuid.UUID) (*Contact, error)
	ListByCustomer(ctx context.Context, customerID uuid.UUID, tenantID uuid.UUID) ([]*Contact, error)
	Update(ctx context.Context, contact *Contact) error
	Delete(ctx context.Context, id, tenantID uuid.UUID) error
}

// ContactNoteRepository defines persistence operations for contact notes.
type ContactNoteRepository interface {
	Create(ctx context.Context, note *ContactNote) error
	ListByContact(ctx context.Context, contactID uuid.UUID, tenantID uuid.UUID) ([]*ContactNote, error)
}
