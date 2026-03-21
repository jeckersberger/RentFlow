package ports

import (
	"context"

	"github.com/jeckersberger/rentflow/services/invoice-service/internal/domain"
)

// InvoiceRepository interface for invoice persistence
type InvoiceRepository interface {
	Create(ctx context.Context, invoice *domain.Invoice) error
	GetByID(ctx context.Context, tenantID, invoiceID string) (*domain.Invoice, error)
	GetByNumber(ctx context.Context, tenantID, invoiceNumber string) (*domain.Invoice, error)
	List(ctx context.Context, query *InvoiceListQuery) (*InvoiceListResult, error)
	Update(ctx context.Context, invoice *domain.Invoice) error
	Delete(ctx context.Context, tenantID, invoiceID string) error
	GetNextSequenceNumber(ctx context.Context, tenantID string) (int, error)
	GetOverdueInvoices(ctx context.Context, tenantID string) ([]*domain.Invoice, error)
}

// QuoteRepository interface for quote persistence
type QuoteRepository interface {
	Create(ctx context.Context, quote *domain.Quote) error
	GetByID(ctx context.Context, tenantID, quoteID string) (*domain.Quote, error)
	GetByNumber(ctx context.Context, tenantID, quoteNumber string) (*domain.Quote, error)
	List(ctx context.Context, query *QuoteListQuery) (*QuoteListResult, error)
	Update(ctx context.Context, quote *domain.Quote) error
	Delete(ctx context.Context, tenantID, quoteID string) error
	GetNextSequenceNumber(ctx context.Context, tenantID string) (int, error)
}

// CreditNoteRepository interface for credit note persistence
type CreditNoteRepository interface {
	Create(ctx context.Context, creditNote *domain.CreditNote) error
	GetByID(ctx context.Context, tenantID, creditNoteID string) (*domain.CreditNote, error)
	Update(ctx context.Context, creditNote *domain.CreditNote) error
	Delete(ctx context.Context, tenantID, creditNoteID string) error
}

// DunningRepository interface for dunning persistence
type DunningRepository interface {
	Create(ctx context.Context, dunning *domain.DunningEntry) error
	GetByID(ctx context.Context, tenantID, dunningID string) (*domain.DunningEntry, error)
	List(ctx context.Context, tenantID, invoiceID string) ([]*domain.DunningEntry, error)
	Update(ctx context.Context, dunning *domain.DunningEntry) error
}

// NumberSequenceRepository manages sequential invoice/quote numbers
type NumberSequenceRepository interface {
	GetNextNumber(ctx context.Context, tenantID, sequenceType string) (int, error)
	ResetSequence(ctx context.Context, tenantID, sequenceType string) error
}

// InvoiceListQuery for filtering invoices
type InvoiceListQuery struct {
	TenantID   string
	Status     *domain.InvoiceStatus
	ClientName *string
	FromDate   *string
	ToDate     *string
	Limit      int
	Offset     int
}

// InvoiceListResult for paginated invoice results
type InvoiceListResult struct {
	Items  []*domain.Invoice
	Total  int64
	Limit  int
	Offset int
}

// QuoteListQuery for filtering quotes
type QuoteListQuery struct {
	TenantID   string
	Status     *domain.QuoteStatus
	ClientName *string
	FromDate   *string
	ToDate     *string
	Limit      int
	Offset     int
}

// QuoteListResult for paginated quote results
type QuoteListResult struct {
	Items  []*domain.Quote
	Total  int64
	Limit  int
	Offset int
}

// PDFGenerator generiert PDFs aus HTML-Vorlagen.
// Abstrahiert die konkrete Implementierung (chromedp, wkhtmltopdf, etc.)
type PDFGenerator interface {
	GeneratePDF(ctx context.Context, html string) ([]byte, error)
}

// EmailSender versendet E-Mails mit optionalem Anhang.
// Abstrahiert den konkreten Transportweg (SMTP, SES, etc.)
type EmailSender interface {
	Send(to, subject, bodyHTML string) error
	SendWithAttachment(to, subject, bodyHTML string, attachment []byte, attachmentName string) error
}
