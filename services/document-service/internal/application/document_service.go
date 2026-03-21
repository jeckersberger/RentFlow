package application

import (
	"context"
	"fmt"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/document-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/document-service/internal/ports"
)

type DocumentService struct {
	docRepo ports.DocumentRepository
	logger  *logger.Logger
}

func NewDocumentService(
	docRepo ports.DocumentRepository,
	logger *logger.Logger,
) *DocumentService {
	return &DocumentService{
		docRepo: docRepo,
		logger:  logger,
	}
}

func (s *DocumentService) CreateDocument(ctx context.Context, cmd CreateDocumentCommand) (*DocumentDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if cmd.FileRef == "" {
		return nil, domain.NewDomainError("FILE_REF_REQUIRED", "file reference is required", nil)
	}

	docID := fmt.Sprintf("doc_%d", hashString(cmd.TenantID+cmd.FileRef))
	docType := domain.DocumentType(cmd.Type)

	doc := domain.NewDocument(docID, cmd.TenantID, cmd.Name, docType, cmd.EntityType, cmd.EntityID, cmd.FileRef, cmd.MimeType, cmd.CreatedBy, cmd.Size, cmd.Checksum)

	if err := doc.Validate(); err != nil {
		return nil, domain.NewDomainError("VALIDATION_ERROR", err.Error(), nil)
	}

	if err := s.docRepo.Create(ctx, doc); err != nil {
		return nil, domain.NewDomainError("CREATE_FAILED", "failed to create document", err)
	}

	return DocumentToDTO(doc), nil
}

func (s *DocumentService) GetDocument(ctx context.Context, tenantID, documentID string) (*DocumentDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	doc, err := s.docRepo.GetByID(ctx, tenantID, documentID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "document not found", err)
	}

	return DocumentToDTO(doc), nil
}

func (s *DocumentService) ListDocuments(ctx context.Context, tenantID string, limit, offset int) (*DocumentListResult, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	docs, total, err := s.docRepo.List(ctx, tenantID, limit, offset)
	if err != nil {
		return nil, domain.NewDomainError("LIST_FAILED", "failed to list documents", err)
	}

	items := make([]*DocumentDTO, len(docs))
	for i, doc := range docs {
		items[i] = DocumentToDTO(doc)
	}

	return &DocumentListResult{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func (s *DocumentService) ListByEntity(ctx context.Context, tenantID, entityType, entityID string, limit, offset int) (*DocumentListResult, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	docs, total, err := s.docRepo.ListByEntity(ctx, tenantID, entityType, entityID, limit, offset)
	if err != nil {
		return nil, domain.NewDomainError("LIST_FAILED", "failed to list documents", err)
	}

	items := make([]*DocumentDTO, len(docs))
	for i, doc := range docs {
		items[i] = DocumentToDTO(doc)
	}

	return &DocumentListResult{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func (s *DocumentService) DeleteDocument(ctx context.Context, tenantID, documentID string) error {
	if tenantID == "" {
		return domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	if err := s.docRepo.Delete(ctx, tenantID, documentID); err != nil {
		return domain.NewDomainError("DELETE_FAILED", "failed to delete document", err)
	}

	return nil
}
