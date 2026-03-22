package domain

import (
	"errors"
	"time"
)

var (
	ErrDocumentNotFound       = errors.New("document not found")
	ErrInvalidInput           = errors.New("invalid input")
	ErrTenantIDRequired       = errors.New("tenant ID required")
	ErrChecksumMismatch       = errors.New("checksum mismatch - integrity compromised")
	ErrInvalidDocumentType    = errors.New("invalid document type")
	ErrSignatureNotFound      = errors.New("signature not found")
	ErrDocumentAlreadySigned  = errors.New("document already signed")
	ErrVersionNotFound        = errors.New("version not found")
)

type DocumentStatus string

const (
	DocumentStatusDraft            DocumentStatus = "draft"
	DocumentStatusGenerated        DocumentStatus = "generated"
	DocumentStatusSent             DocumentStatus = "sent"
	DocumentStatusSigned           DocumentStatus = "signed"
	DocumentStatusArchived         DocumentStatus = "archived"
)

type DocumentType string

const (
	DocumentTypeOffer              DocumentType = "offer"
	DocumentTypeOrderConfirmation  DocumentType = "order_confirmation"
	DocumentTypeDeliveryNote       DocumentType = "delivery_note"
	DocumentTypePickupNote         DocumentType = "pickup_note"
	DocumentTypeInvoice            DocumentType = "invoice"
	DocumentTypeRentalContract     DocumentType = "rental_contract"
	DocumentTypeDamageReport       DocumentType = "damage_report"
)

func IsValidDocumentType(dt string) bool {
	switch DocumentType(dt) {
	case DocumentTypeOffer, DocumentTypeOrderConfirmation, DocumentTypeDeliveryNote,
		DocumentTypePickupNote, DocumentTypeInvoice, DocumentTypeRentalContract,
		DocumentTypeDamageReport:
		return true
	}
	return false
}

type Document struct {
	ID               string                 `json:"id"`
	TenantID         string                 `json:"tenant_id"`
	DocumentType     DocumentType           `json:"document_type"`
	ReferenceID      string                 `json:"reference_id"`
	DocumentNumber   string                 `json:"document_number"`
	Title            string                 `json:"title"`
	Status           DocumentStatus         `json:"status"`
	CurrentVersion   int                    `json:"current_version"`
	TemplateID       string                 `json:"template_id"`
	Metadata         map[string]interface{} `json:"metadata"`
	ChecksumSHA256   string                 `json:"checksum_sha256"`
	PreviousChecksum string                 `json:"previous_checksum"`
	CreatedBy        string                 `json:"created_by"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

type DocumentVersion struct {
	ID                 string    `json:"id"`
	DocumentID         string    `json:"document_id"`
	VersionNumber      int       `json:"version_number"`
	FilePath           string    `json:"file_path"`
	FileSize           int       `json:"file_size"`
	MimeType           string    `json:"mime_type"`
	ChecksumSHA256     string    `json:"checksum_sha256"`
	ChangesDescription string    `json:"changes_description"`
	CreatedBy          string    `json:"created_by"`
	CreatedAt          time.Time `json:"created_at"`
}

type Signature struct {
	ID            string    `json:"id"`
	DocumentID    string    `json:"document_id"`
	SignerName    string    `json:"signer_name"`
	SignerEmail   string    `json:"signer_email"`
	SignerRole    string    `json:"signer_role"`
	SignatureData string    `json:"signature_data"`
	SignedAt      *time.Time `json:"signed_at"`
	IPAddress     string    `json:"ip_address"`
	UserAgent     string    `json:"user_agent"`
	Verified      bool      `json:"verified"`
	CreatedAt     time.Time `json:"created_at"`
}
