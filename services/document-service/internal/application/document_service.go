package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/document-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/document-service/internal/ports"
)

type DocumentService struct {
	docRepo       ports.DocumentRepository
	versionRepo   ports.DocumentVersionRepository
	sigRepo       ports.SignatureRepository
	checksumSvc   *ChecksumService
	logger        logger.Logger
}

func NewDocumentService(
	docRepo ports.DocumentRepository,
	versionRepo ports.DocumentVersionRepository,
	sigRepo ports.SignatureRepository,
	checksumSvc *ChecksumService,
	log logger.Logger,
) *DocumentService {
	return &DocumentService{
		docRepo:       docRepo,
		versionRepo:   versionRepo,
		sigRepo:       sigRepo,
		checksumSvc:   checksumSvc,
		logger:        log,
	}
}

// CreateDocument creates a new document
func (s *DocumentService) CreateDocument(
	ctx context.Context,
	tenantID, userID string,
	req CreateDocumentRequest,
) (*DocumentResponse, error) {
	if tenantID == "" || userID == "" {
		return nil, domain.ErrTenantIDRequired
	}
	if req.DocumentNumber == "" || req.Title == "" {
		return nil, domain.ErrInvalidInput
	}
	if !domain.IsValidDocumentType(string(req.DocumentType)) {
		return nil, domain.ErrInvalidDocumentType
	}

	doc := &domain.Document{
		ID:             uuid.New().String(),
		TenantID:       tenantID,
		DocumentType:   domain.DocumentType(req.DocumentType),
		ReferenceID:    req.ReferenceID,
		DocumentNumber: req.DocumentNumber,
		Title:          req.Title,
		Status:         domain.DocumentStatusDraft,
		CurrentVersion: 1,
		TemplateID:     req.TemplateID,
		Metadata:       req.Metadata,
		CreatedBy:      userID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Generate initial checksum
	doc.ChecksumSHA256 = s.checksumSvc.GenerateChecksum(doc)

	if err := s.docRepo.Create(ctx, doc); err != nil {
		s.logger.Error("Failed to create document", err)
		return nil, err
	}

	s.logger.Info("Document created", "id", doc.ID, "tenant", tenantID)
	return DocumentToResponse(doc), nil
}

// GetDocument retrieves a document by ID
func (s *DocumentService) GetDocument(
	ctx context.Context,
	tenantID, docID string,
) (*DocumentResponse, error) {
	if tenantID == "" || docID == "" {
		return nil, domain.ErrTenantIDRequired
	}

	doc, err := s.docRepo.GetByID(ctx, tenantID, docID)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, domain.ErrDocumentNotFound
	}

	return DocumentToResponse(doc), nil
}

// ListDocuments retrieves all documents for a tenant
func (s *DocumentService) ListDocuments(
	ctx context.Context,
	tenantID string,
) ([]*DocumentResponse, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}

	docs, err := s.docRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		s.logger.Error("Failed to list documents", err)
		return nil, err
	}

	responses := make([]*DocumentResponse, len(docs))
	for i, doc := range docs {
		responses[i] = DocumentToResponse(doc)
	}

	return responses, nil
}

// ListDocumentsByType retrieves documents of a specific type for a tenant
func (s *DocumentService) ListDocumentsByType(
	ctx context.Context,
	tenantID string,
	docType string,
) ([]*DocumentResponse, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}
	if !domain.IsValidDocumentType(docType) {
		return nil, domain.ErrInvalidDocumentType
	}

	docs, err := s.docRepo.ListByType(ctx, tenantID, domain.DocumentType(docType))
	if err != nil {
		s.logger.Error("Failed to list documents by type", err)
		return nil, err
	}

	responses := make([]*DocumentResponse, len(docs))
	for i, doc := range docs {
		responses[i] = DocumentToResponse(doc)
	}

	return responses, nil
}

// GenerateFromTemplate generates a document from a template
func (s *DocumentService) GenerateFromTemplate(
	ctx context.Context,
	tenantID, userID, docID string,
	req GenerateFromTemplateRequest,
) (*DocumentResponse, error) {
	if tenantID == "" || docID == "" {
		return nil, domain.ErrTenantIDRequired
	}

	// Get the document
	doc, err := s.docRepo.GetByID(ctx, tenantID, docID)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, domain.ErrDocumentNotFound
	}

	// Update document status
	doc.Status = domain.DocumentStatusGenerated
	doc.UpdatedAt = time.Now()

	// Update checksum (previous checksum is set)
	oldChecksum := doc.ChecksumSHA256
	doc.PreviousChecksum = oldChecksum
	doc.ChecksumSHA256 = s.checksumSvc.GenerateChecksum(doc)

	if err := s.docRepo.Update(ctx, doc); err != nil {
		s.logger.Error("Failed to update document", err)
		return nil, err
	}

	// Create a version
	version := &domain.DocumentVersion{
		ID:                 uuid.New().String(),
		DocumentID:         docID,
		VersionNumber:      doc.CurrentVersion,
		FilePath:           fmt.Sprintf("/documents/%s/%s-v%d.pdf", tenantID, docID, doc.CurrentVersion),
		MimeType:           "application/pdf",
		ChecksumSHA256:     doc.ChecksumSHA256,
		ChangesDescription: "Generated from template",
		CreatedBy:          userID,
		CreatedAt:          time.Now(),
	}

	if err := s.versionRepo.Create(ctx, version); err != nil {
		s.logger.Error("Failed to create document version", err)
		return nil, err
	}

	s.logger.Info("Document generated from template", "id", docID, "tenant", tenantID)
	return DocumentToResponse(doc), nil
}

// ArchiveDocument archives a document
func (s *DocumentService) ArchiveDocument(
	ctx context.Context,
	tenantID, docID string,
) error {
	if tenantID == "" || docID == "" {
		return domain.ErrTenantIDRequired
	}

	// Verify document exists
	doc, err := s.docRepo.GetByID(ctx, tenantID, docID)
	if err != nil {
		return err
	}
	if doc == nil {
		return domain.ErrDocumentNotFound
	}

	if err := s.docRepo.Archive(ctx, tenantID, docID); err != nil {
		s.logger.Error("Failed to archive document", err)
		return err
	}

	s.logger.Info("Document archived", "id", docID, "tenant", tenantID)
	return nil
}

// GetVersions retrieves all versions of a document
func (s *DocumentService) GetVersions(
	ctx context.Context,
	tenantID, docID string,
) ([]*DocumentVersionResponse, error) {
	if tenantID == "" || docID == "" {
		return nil, domain.ErrTenantIDRequired
	}

	// Verify document exists
	doc, err := s.docRepo.GetByID(ctx, tenantID, docID)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, domain.ErrDocumentNotFound
	}

	versions, err := s.versionRepo.ListByDocument(ctx, docID)
	if err != nil {
		s.logger.Error("Failed to fetch versions", err)
		return nil, err
	}

	responses := make([]*DocumentVersionResponse, len(versions))
	for i, ver := range versions {
		responses[i] = DocumentVersionToResponse(ver)
	}

	return responses, nil
}

// GenerateDeliveryNote generates a delivery note from a project
func (s *DocumentService) GenerateDeliveryNote(
	ctx context.Context,
	tenantID, userID, projectID string,
) (*DocumentResponse, error) {
	if tenantID == "" || projectID == "" {
		return nil, domain.ErrTenantIDRequired
	}

	// Create delivery note document
	doc := &domain.Document{
		ID:             uuid.New().String(),
		TenantID:       tenantID,
		DocumentType:   domain.DocumentTypeDeliveryNote,
		ReferenceID:    projectID,
		DocumentNumber: fmt.Sprintf("DN-%d", time.Now().UnixNano()),
		Title:          "Delivery Note",
		Status:         domain.DocumentStatusGenerated,
		CurrentVersion: 1,
		Metadata:       make(map[string]interface{}),
		CreatedBy:      userID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Generate checksum
	doc.ChecksumSHA256 = s.checksumSvc.GenerateChecksum(doc)

	if err := s.docRepo.Create(ctx, doc); err != nil {
		s.logger.Error("Failed to create delivery note", err)
		return nil, err
	}

	s.logger.Info("Delivery note generated", "id", doc.ID, "projectId", projectID, "tenant", tenantID)
	return DocumentToResponse(doc), nil
}
