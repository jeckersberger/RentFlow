package domain

import (
	"context"

	"github.com/google/uuid"
)

type InvoiceRepository interface {
	Create(ctx context.Context, invoice *Invoice) error
	GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*Invoice, error)
	List(ctx context.Context, tenantID uuid.UUID, filter InvoiceFilter) ([]*Invoice, int64, error)
	Update(ctx context.Context, invoice *Invoice) error
	UpdateStatus(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, status string) error
	UpdateAmountPaid(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, amount int64) error
	GetLastHash(ctx context.Context, tenantID uuid.UUID) (string, error)
	Search(ctx context.Context, tenantID uuid.UUID, query string, page int, perPage int) ([]*Invoice, int64, error)
}

type InvoiceItemRepository interface {
	Create(ctx context.Context, item *InvoiceItem) error
	ListByInvoice(ctx context.Context, invoiceID uuid.UUID, tenantID uuid.UUID) ([]*InvoiceItem, error)
	Delete(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error
	DeleteAllByInvoice(ctx context.Context, invoiceID uuid.UUID, tenantID uuid.UUID) error
}

type NumberSequenceRepository interface {
	NextNumber(ctx context.Context, tenantID uuid.UUID, prefix string, year int) (int64, error)
	EnsureSequence(ctx context.Context, tenantID uuid.UUID, prefix string, year int) error
}

type PaymentRepository interface {
	Create(ctx context.Context, payment *Payment) error
	ListByInvoice(ctx context.Context, invoiceID uuid.UUID, tenantID uuid.UUID) ([]*Payment, error)
}
