package domain

import (
	"context"

	"github.com/google/uuid"
)

// DocumentTemplateRepository defines persistence operations for DocumentTemplate aggregates.
type DocumentTemplateRepository interface {
	Create(ctx context.Context, tpl *DocumentTemplate) error
	GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*DocumentTemplate, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]*DocumentTemplate, error)
	Update(ctx context.Context, tpl *DocumentTemplate) error
}

// DocumentRepository defines persistence operations for Document aggregates.
type DocumentRepository interface {
	Create(ctx context.Context, doc *Document) error
	GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*Document, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]*Document, error)
}

// AttachmentRepository defines persistence operations for Attachment aggregates.
type AttachmentRepository interface {
	Create(ctx context.Context, att *Attachment) error
	GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*Attachment, error)
	ListByReference(ctx context.Context, tenantID uuid.UUID, referenceID uuid.UUID, referenceType string) ([]*Attachment, error)
	Delete(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error
}
