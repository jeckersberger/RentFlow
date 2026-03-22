package application

import (
	"time"

	"github.com/jeckersberger/rentflow/services/document-service/internal/domain"
)

type CreateDocumentRequest struct {
	DocumentType DocumentTypeDTO             `json:"document_type"`
	ReferenceID  string                      `json:"reference_id"`
	DocumentNumber string                   `json:"document_number"`
	Title        string                      `json:"title"`
	TemplateID   string                      `json:"template_id,omitempty"`
	Metadata     map[string]interface{}      `json:"metadata,omitempty"`
}

type DocumentTypeDTO string

type DocumentResponse struct {
	ID               string                 `json:"id"`
	DocumentType     string                 `json:"document_type"`
	ReferenceID      string                 `json:"reference_id"`
	DocumentNumber   string                 `json:"document_number"`
	Title            string                 `json:"title"`
	Status           string                 `json:"status"`
	CurrentVersion   int                    `json:"current_version"`
	TemplateID       string                 `json:"template_id"`
	Metadata         map[string]interface{} `json:"metadata"`
	ChecksumSHA256   string                 `json:"checksum_sha256"`
	CreatedBy        string                 `json:"created_by"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

type GenerateFromTemplateRequest struct {
	TemplateID string                 `json:"template_id"`
	Data       map[string]interface{} `json:"data"`
}

type RequestSignatureRequest struct {
	SignerName  string `json:"signer_name"`
	SignerEmail string `json:"signer_email"`
	SignerRole  string `json:"signer_role"`
}

type SubmitSignatureRequest struct {
	SignatureData string `json:"signature_data"`
	IPAddress     string `json:"ip_address"`
	UserAgent     string `json:"user_agent"`
}

type SignatureResponse struct {
	ID          string     `json:"id"`
	DocumentID  string     `json:"document_id"`
	SignerName  string     `json:"signer_name"`
	SignerEmail string     `json:"signer_email"`
	SignerRole  string     `json:"signer_role"`
	SignedAt    *time.Time `json:"signed_at"`
	Verified    bool       `json:"verified"`
	CreatedAt   time.Time  `json:"created_at"`
}

type DocumentVersionResponse struct {
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

type ChecksumChainVerificationResponse struct {
	Status         string   `json:"status"`
	DocumentCount  int      `json:"document_count"`
	IntegrityValid bool     `json:"integrity_valid"`
	Errors         []string `json:"errors"`
}

type UploadScanRequest struct {
	ReferenceID string `json:"reference_id"`
	ScanType    string `json:"scan_type"` // document_type: invoice, contract, etc.
}

type PublicSignatureResponse struct {
	ID          string     `json:"id"`
	DocumentID  string     `json:"document_id"`
	SignerName  string     `json:"signer_name"`
	SignerEmail string     `json:"signer_email"`
	SignedAt    *time.Time `json:"signed_at"`
	Verified    bool       `json:"verified"`
}

type PublicSignatureSubmitRequest struct {
	SignerName    string `json:"signer_name"`
	SignatureData string `json:"signature_data"`
	IPAddress     string `json:"ip_address"`
	UserAgent     string `json:"user_agent"`
}

func DocumentToResponse(doc *domain.Document) *DocumentResponse {
	return &DocumentResponse{
		ID:               doc.ID,
		DocumentType:     string(doc.DocumentType),
		ReferenceID:      doc.ReferenceID,
		DocumentNumber:   doc.DocumentNumber,
		Title:            doc.Title,
		Status:           string(doc.Status),
		CurrentVersion:   doc.CurrentVersion,
		TemplateID:       doc.TemplateID,
		Metadata:         doc.Metadata,
		ChecksumSHA256:   doc.ChecksumSHA256,
		CreatedBy:        doc.CreatedBy,
		CreatedAt:        doc.CreatedAt,
		UpdatedAt:        doc.UpdatedAt,
	}
}

func DocumentVersionToResponse(ver *domain.DocumentVersion) *DocumentVersionResponse {
	return &DocumentVersionResponse{
		ID:                 ver.ID,
		DocumentID:         ver.DocumentID,
		VersionNumber:      ver.VersionNumber,
		FilePath:           ver.FilePath,
		FileSize:           ver.FileSize,
		MimeType:           ver.MimeType,
		ChecksumSHA256:     ver.ChecksumSHA256,
		ChangesDescription: ver.ChangesDescription,
		CreatedBy:          ver.CreatedBy,
		CreatedAt:          ver.CreatedAt,
	}
}

func SignatureToResponse(sig *domain.Signature) *SignatureResponse {
	return &SignatureResponse{
		ID:          sig.ID,
		DocumentID:  sig.DocumentID,
		SignerName:  sig.SignerName,
		SignerEmail: sig.SignerEmail,
		SignerRole:  sig.SignerRole,
		SignedAt:    sig.SignedAt,
		Verified:    sig.Verified,
		CreatedAt:   sig.CreatedAt,
	}
}
