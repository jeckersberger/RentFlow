package ports

import (
	"context"

	"github.com/jeckersberger/rentflow/services/document-service/internal/domain"
)

type DocumentRepository interface {
	Create(ctx context.Context, document *domain.Document) error
	GetByID(ctx context.Context, tenantID, documentID string) (*domain.Document, error)
	List(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Document, int64, error)
	ListByEntity(ctx context.Context, tenantID, entityType, entityID string, limit, offset int) ([]*domain.Document, int64, error)
	Delete(ctx context.Context, tenantID, documentID string) error
}

type TemplateRepository interface {
	Create(ctx context.Context, template *domain.Template) error
	Update(ctx context.Context, template *domain.Template) error
	GetByID(ctx context.Context, tenantID, templateID string) (*domain.Template, error)
	List(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Template, int64, error)
	Delete(ctx context.Context, tenantID, templateID string) error
}
