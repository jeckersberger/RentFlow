package ports

import (
	"context"

	"github.com/jeckersberger/rentflow/services/project-service/internal/domain"
)

type ProjectRepository interface {
	Create(ctx context.Context, project *domain.Project) error
	Update(ctx context.Context, project *domain.Project) error
	GetByID(ctx context.Context, tenantID, projectID string) (*domain.Project, error)
	List(ctx context.Context, tenantID string, limit, offset int) (*ProjectListResult, error)
	Delete(ctx context.Context, tenantID, projectID string) error
	Search(ctx context.Context, tenantID, term string, limit, offset int) (*ProjectListResult, error)
}

type PacklistRepository interface {
	Create(ctx context.Context, packlist *domain.Packlist) error
	Update(ctx context.Context, packlist *domain.Packlist) error
	GetByID(ctx context.Context, tenantID, packlistID string) (*domain.Packlist, error)
	ListByProjectID(ctx context.Context, tenantID, projectID string, limit, offset int) (*PacklistListResult, error)
	Delete(ctx context.Context, tenantID, packlistID string) error
}

type ReservationRepository interface {
	Create(ctx context.Context, reservation *domain.Reservation) error
	Update(ctx context.Context, reservation *domain.Reservation) error
	GetByID(ctx context.Context, tenantID, reservationID string) (*domain.Reservation, error)
	ListByProjectID(ctx context.Context, tenantID, projectID string) ([]*domain.Reservation, error)
	ListByEquipmentID(ctx context.Context, tenantID, equipmentID string, startDate, endDate string) ([]*domain.Reservation, error)
	Delete(ctx context.Context, tenantID, reservationID string) error
}

type ProjectListResult struct {
	Items  []*domain.Project
	Total  int64
	Limit  int
	Offset int
}

type PacklistListResult struct {
	Items  []*domain.Packlist
	Total  int64
	Limit  int
	Offset int
}
