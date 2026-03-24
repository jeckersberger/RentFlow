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
	ListWithFilter(ctx context.Context, tenantID, filter string, limit, offset int) (*ProjectListResult, error)
	Delete(ctx context.Context, tenantID, projectID string) error
	Search(ctx context.Context, tenantID, term string, limit, offset int) (*ProjectListResult, error)
	ListByDateRange(ctx context.Context, tenantID, startDate, endDate string) ([]*domain.Project, error)
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
	FindOverlapping(ctx context.Context, tenantID, equipmentID string, startDate, endDate string, excludeReservationID string) ([]*domain.Reservation, error)
	Delete(ctx context.Context, tenantID, reservationID string) error
}

type CustomerRepository interface {
	Create(ctx context.Context, customer *domain.Customer) error
	GetByID(ctx context.Context, tenantID, customerID string) (*domain.Customer, error)
	List(ctx context.Context, tenantID string, limit, offset int) (*CustomerListResult, error)
	Update(ctx context.Context, customer *domain.Customer) error
	Delete(ctx context.Context, tenantID, customerID string) error
	Search(ctx context.Context, tenantID, term string, limit, offset int) (*CustomerListResult, error)
}

type ContactRepository interface {
	Create(ctx context.Context, contact *domain.Contact) error
	GetByID(ctx context.Context, tenantID, contactID string) (*domain.Contact, error)
	List(ctx context.Context, tenantID string, limit, offset int) (*ContactListResult, error)
	Update(ctx context.Context, contact *domain.Contact) error
	Delete(ctx context.Context, tenantID, contactID string) error
	Search(ctx context.Context, tenantID, term string, limit, offset int) (*ContactListResult, error)
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

type CustomerListResult struct {
	Items  []*domain.Customer
	Total  int64
	Limit  int
	Offset int
}

type ContactListResult struct {
	Items  []*domain.Contact
	Total  int64
	Limit  int
	Offset int
}
