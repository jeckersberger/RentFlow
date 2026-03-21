package application

import (
	"time"

	"github.com/jeckersberger/rentflow/services/document-service/internal/domain"
)

type DocumentDTO struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	EntityType string    `json:"entity_type"`
	EntityID   string    `json:"entity_id"`
	FileRef    string    `json:"file_ref"`
	MimeType   string    `json:"mime_type"`
	Size       int64     `json:"size"`
	Checksum   string    `json:"checksum"`
	CreatedBy  string    `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type TemplateDTO struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Content   string    `json:"content"`
	Variables []string  `json:"variables"`
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy string    `json:"created_by"`
}

type DocumentListResult struct {
	Items  []*DocumentDTO `json:"items"`
	Total  int64          `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

type TemplateListResult struct {
	Items  []*TemplateDTO `json:"items"`
	Total  int64          `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

func DocumentToDTO(doc *domain.Document) *DocumentDTO {
	return &DocumentDTO{
		ID:         doc.ID,
		TenantID:   doc.TenantID,
		Name:       doc.Name,
		Type:       string(doc.Type),
		EntityType: doc.EntityType,
		EntityID:   doc.EntityID,
		FileRef:    doc.FileRef,
		MimeType:   doc.MimeType,
		Size:       doc.Size,
		Checksum:   doc.Checksum,
		CreatedBy:  doc.CreatedBy,
		CreatedAt:  doc.CreatedAt,
		UpdatedAt:  doc.UpdatedAt,
	}
}

func TemplateToDTO(tpl *domain.Template) *TemplateDTO {
	return &TemplateDTO{
		ID:        tpl.ID,
		TenantID:  tpl.TenantID,
		Name:      tpl.Name,
		Type:      string(tpl.Type),
		Content:   tpl.Content,
		Variables: tpl.Variables,
		IsDefault: tpl.IsDefault,
		CreatedAt: tpl.CreatedAt,
		UpdatedAt: tpl.UpdatedAt,
		CreatedBy: tpl.CreatedBy,
	}
}
