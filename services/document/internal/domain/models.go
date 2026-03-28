package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// DocumentTemplate represents a reusable document template.
type DocumentTemplate struct {
	ID        uuid.UUID        `json:"id"`
	TenantID  uuid.UUID        `json:"tenant_id"`
	Name      string           `json:"name"`
	Type      string           `json:"type"`
	Content   string           `json:"content"`
	Variables json.RawMessage  `json:"variables"`
	IsDefault bool             `json:"is_default"`
	IsActive  bool             `json:"is_active"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// Document represents a generated document.
type Document struct {
	ID            uuid.UUID  `json:"id"`
	TenantID      uuid.UUID  `json:"tenant_id"`
	TemplateID    *uuid.UUID `json:"template_id,omitempty"`
	Type          string     `json:"type"`
	Title         string     `json:"title"`
	ReferenceID   *uuid.UUID `json:"reference_id,omitempty"`
	ReferenceType string     `json:"reference_type,omitempty"`
	Content       string     `json:"content,omitempty"`
	FilePath      string     `json:"file_path,omitempty"`
	FileSize      int        `json:"file_size"`
	MimeType      string     `json:"mime_type"`
	Status        string     `json:"status"`
	CreatedBy     *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// Attachment represents a file attachment linked to a reference entity.
type Attachment struct {
	ID            uuid.UUID  `json:"id"`
	TenantID      uuid.UUID  `json:"tenant_id"`
	ReferenceID   uuid.UUID  `json:"reference_id"`
	ReferenceType string     `json:"reference_type"`
	FileName      string     `json:"file_name"`
	FilePath      string     `json:"file_path"`
	FileSize      int        `json:"file_size"`
	MimeType      string     `json:"mime_type,omitempty"`
	UploadedBy    *uuid.UUID `json:"uploaded_by,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}
