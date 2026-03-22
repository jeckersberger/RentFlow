package ports

import (
	"context"

	"github.com/jeckersberger/rentflow/services/document-service/internal/domain"
)

type DocumentRepository interface {
	Create(ctx context.Context, doc *domain.Document) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.Document, error)
	ListByTenant(ctx context.Context, tenantID string) ([]*domain.Document, error)
	ListByType(ctx context.Context, tenantID string, docType domain.DocumentType) ([]*domain.Document, error)
	Update(ctx context.Context, doc *domain.Document) error
	Archive(ctx context.Context, tenantID, docID string) error
	ListAllForChecksumChain(ctx context.Context, tenantID string) ([]*domain.Document, error)
}

type DocumentVersionRepository interface {
	Create(ctx context.Context, version *domain.DocumentVersion) error
	GetByID(ctx context.Context, versionID string) (*domain.DocumentVersion, error)
	ListByDocument(ctx context.Context, docID string) ([]*domain.DocumentVersion, error)
	GetByVersion(ctx context.Context, docID string, versionNumber int) (*domain.DocumentVersion, error)
}

type SignatureRepository interface {
	Create(ctx context.Context, sig *domain.Signature) error
	GetByID(ctx context.Context, sigID string) (*domain.Signature, error)
	ListByDocument(ctx context.Context, docID string) ([]*domain.Signature, error)
	Update(ctx context.Context, sig *domain.Signature) error
	GetByDocumentAndEmail(ctx context.Context, docID, email string) (*domain.Signature, error)
}
