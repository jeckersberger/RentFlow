package application

import (
	"time"

	"github.com/jeckersberger/rentflow/services/project-service/internal/domain"
)

// Project Commands

type CreateProjectCommand struct {
	TenantID        string
	Name            string
	Description     string
	ClientName      string
	ClientEmail     string
	ClientPhone     string
	ClientAddress   AddressDTO
	VenueAddress    AddressDTO
	StartDate       time.Time
	EndDate         time.Time
	SetupDate       *time.Time
	TeardownDate    *time.Time
	Budget          float64
	Currency        string
	Notes           string
	Tags            []string
	CreatedByUserID string
}

type UpdateProjectCommand struct {
	ID            string
	TenantID      string
	Name          string
	Description   string
	ClientName    string
	ClientEmail   string
	ClientPhone   string
	ClientAddress AddressDTO
	VenueAddress  AddressDTO
	StartDate     time.Time
	EndDate       time.Time
	SetupDate     *time.Time
	TeardownDate  *time.Time
	Budget        float64
	Currency      string
	Notes         string
	Tags          []string
}

type ChangeProjectStatusCommand struct {
	ID       string
	TenantID string
	Status   domain.ProjectStatus
}

type SetProjectManagerCommand struct {
	ID        string
	TenantID  string
	ManagerID string
}

type DeleteProjectCommand struct {
	ID       string
	TenantID string
}

// Packlist Commands

type CreatePacklistCommand struct {
	TenantID  string
	ProjectID string
	Name      string
}

type UpdatePacklistCommand struct {
	ID       string
	TenantID string
	Name     string
}

type AddPacklistItemCommand struct {
	TenantID      string
	PacklistID    string
	EquipmentID   string
	EquipmentName string
	Quantity      int
	Notes         string
}

type RemovePacklistItemCommand struct {
	TenantID    string
	PacklistID  string
	EquipmentID string
}

type MarkItemPackedCommand struct {
	TenantID       string
	PacklistID     string
	EquipmentID    string
	QuantityPacked int
	Notes          string
}

type MarkItemReturnedCommand struct {
	TenantID         string
	PacklistID       string
	EquipmentID      string
	QuantityReturned int
	Notes            string
}

type ChangePacklistStatusCommand struct {
	ID       string
	TenantID string
	Status   domain.PacklistStatus
}

type DeletePacklistCommand struct {
	ID       string
	TenantID string
}

// Reservation Commands

type CreateReservationCommand struct {
	TenantID    string
	ProjectID   string
	EquipmentID string
	StartDate   time.Time
	EndDate     time.Time
}

type ConfirmReservationCommand struct {
	ID       string
	TenantID string
}

type CancelReservationCommand struct {
	ID       string
	TenantID string
	Reason   string
}

type DeleteReservationCommand struct {
	ID       string
	TenantID string
}
