package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ProjectRepository defines persistence operations for projects.
type ProjectRepository interface {
	Create(ctx context.Context, project *Project) error
	GetByID(ctx context.Context, id, tenantID uuid.UUID) (*Project, error)
	List(ctx context.Context, tenantID uuid.UUID, filter ProjectFilter) ([]*Project, int64, error)
	Update(ctx context.Context, project *Project) error
	Delete(ctx context.Context, id, tenantID uuid.UUID) error
	UpdateStatus(ctx context.Context, id, tenantID uuid.UUID, status string) error
	Search(ctx context.Context, tenantID uuid.UUID, query string, page, perPage int) ([]*Project, int64, error)
	GetByDateRange(ctx context.Context, tenantID uuid.UUID, start, end time.Time) ([]*Project, error)
}

// ProjectEquipmentRepository defines persistence operations for project equipment assignments.
type ProjectEquipmentRepository interface {
	Add(ctx context.Context, pe *ProjectEquipment) error
	Remove(ctx context.Context, id uuid.UUID) error
	ListByProject(ctx context.Context, projectID uuid.UUID) ([]*ProjectEquipment, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ProjectEquipment, error)
}

// PacklistRepository defines persistence operations for packlists and their items.
type PacklistRepository interface {
	Create(ctx context.Context, packlist *Packlist) error
	GetByID(ctx context.Context, id, tenantID uuid.UUID) (*Packlist, error)
	ListByProject(ctx context.Context, projectID, tenantID uuid.UUID) ([]*Packlist, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]*Packlist, error)
	UpdateStatus(ctx context.Context, id, tenantID uuid.UUID, status string) error
	AddItem(ctx context.Context, item *PacklistItem) error
	GetItems(ctx context.Context, packlistID, tenantID uuid.UUID) ([]*PacklistItem, error)
	UpdateItemPacked(ctx context.Context, itemID uuid.UUID, qty int, packedBy *uuid.UUID, tenantID uuid.UUID) error
	UpdateItemReturned(ctx context.Context, itemID uuid.UUID, qty int, tenantID uuid.UUID) error
}

// ReservationRepository defines persistence operations for equipment reservations.
type ReservationRepository interface {
	Create(ctx context.Context, reservation *Reservation) error
	GetByID(ctx context.Context, id, tenantID uuid.UUID) (*Reservation, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]*Reservation, error)
	ListByEquipmentAndDateRange(ctx context.Context, equipmentID uuid.UUID, start, end time.Time) ([]*Reservation, error)
	UpdateStatus(ctx context.Context, id, tenantID uuid.UUID, status string) error
}
