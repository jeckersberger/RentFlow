package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/document/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

// CreateDocumentRequest holds the data needed to create (generate) a document.
type CreateDocumentRequest struct {
	TemplateID    *uuid.UUID `json:"template_id,omitempty"`
	Type          string     `json:"type"`
	Title         string     `json:"title"`
	ReferenceID   *uuid.UUID `json:"reference_id,omitempty"`
	ReferenceType string     `json:"reference_type,omitempty"`
	Content       string     `json:"content,omitempty"`
	FilePath      string     `json:"file_path,omitempty"`
	FileSize      int        `json:"file_size,omitempty"`
	MimeType      string     `json:"mime_type,omitempty"`
	Status        string     `json:"status,omitempty"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// DocumentService implements the application-level use cases for documents.
type DocumentService struct {
	documentRepo domain.DocumentRepository
	logger       zerolog.Logger
}

// NewDocumentService constructs a new DocumentService.
func NewDocumentService(
	documentRepo domain.DocumentRepository,
	logger zerolog.Logger,
) *DocumentService {
	return &DocumentService{
		documentRepo: documentRepo,
		logger:       logger.With().Str("service", "document").Logger(),
	}
}

// Create creates a new document.
func (s *DocumentService) Create(
	ctx context.Context,
	tenantID uuid.UUID,
	createdBy *uuid.UUID,
	req CreateDocumentRequest,
) (*domain.Document, error) {
	if req.Title == "" {
		return nil, domain.ErrTitleRequired
	}
	if req.Type == "" {
		return nil, domain.ErrTypeRequired
	}

	mimeType := req.MimeType
	if mimeType == "" {
		mimeType = "application/pdf"
	}

	status := req.Status
	if status == "" {
		status = "draft"
	}

	doc := &domain.Document{
		ID:            uuid.New(),
		TenantID:      tenantID,
		TemplateID:    req.TemplateID,
		Type:          req.Type,
		Title:         req.Title,
		ReferenceID:   req.ReferenceID,
		ReferenceType: req.ReferenceType,
		Content:       req.Content,
		FilePath:      req.FilePath,
		FileSize:      req.FileSize,
		MimeType:      mimeType,
		Status:        status,
		CreatedBy:     createdBy,
	}

	if err := s.documentRepo.Create(ctx, doc); err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to create document")
		return nil, fmt.Errorf("create document: %w", err)
	}

	s.logger.Info().
		Str("document_id", doc.ID.String()).
		Str("tenant_id", tenantID.String()).
		Msg("document created")

	return doc, nil
}

// GetByID retrieves a document by ID.
func (s *DocumentService) GetByID(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
) (*domain.Document, error) {
	doc, err := s.documentRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("document_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to get document")
		return nil, fmt.Errorf("get document: %w", err)
	}
	return doc, nil
}

// List returns all documents for a tenant.
func (s *DocumentService) List(
	ctx context.Context,
	tenantID uuid.UUID,
) ([]*domain.Document, error) {
	docs, err := s.documentRepo.List(ctx, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to list documents")
		return nil, fmt.Errorf("list documents: %w", err)
	}
	return docs, nil
}
