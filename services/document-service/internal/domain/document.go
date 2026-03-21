package domain

import (
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/events"
)

type DocumentType string

const (
	DocumentTypeInvoicePDF DocumentType = "invoice_pdf"
	DocumentTypeContract   DocumentType = "contract"
	DocumentTypeTemplate   DocumentType = "template"
	DocumentTypeReport     DocumentType = "report"
)

type Document struct {
	events.AggregateRoot
	TenantID   string
	Name       string
	Type       DocumentType
	EntityType string
	EntityID   string
	FileRef    string
	MimeType   string
	Size       int64
	Checksum   string
	CreatedBy  string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func NewDocument(id, tenantID, name string, docType DocumentType, entityType, entityID, fileRef, mimeType, createdBy string, size int64, checksum string) *Document {
	now := time.Now()
	return &Document{
		AggregateRoot: *events.NewAggregateRoot(id, "document"),
		TenantID:      tenantID,
		Name:          name,
		Type:          docType,
		EntityType:    entityType,
		EntityID:      entityID,
		FileRef:       fileRef,
		MimeType:      mimeType,
		Size:          size,
		Checksum:      checksum,
		CreatedBy:     createdBy,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func (d *Document) Validate() error {
	if d.ID == "" {
		return fmt.Errorf("document ID cannot be empty")
	}
	if d.TenantID == "" {
		return fmt.Errorf("tenant ID cannot be empty")
	}
	if d.Name == "" {
		return fmt.Errorf("document name cannot be empty")
	}
	if d.FileRef == "" {
		return fmt.Errorf("file reference cannot be empty")
	}
	if d.CreatedBy == "" {
		return fmt.Errorf("creator cannot be empty")
	}
	return nil
}
