package application

import (
	"time"

	"github.com/jeckersberger/rentflow/services/project-service/internal/domain"
)

// Project Commands

type CreateProjectCommand struct {
	TenantID        string     `json:"tenant_id"`
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	ClientName      string     `json:"client_name"`
	ClientEmail     string     `json:"client_email"`
	ClientPhone     string     `json:"client_phone"`
	ClientAddress   AddressDTO `json:"client_address"`
	VenueAddress    AddressDTO `json:"venue_address"`
	StartDate       time.Time  `json:"start_date"`
	EndDate         time.Time  `json:"end_date"`
	SetupDate       *time.Time `json:"setup_date"`
	TeardownDate    *time.Time `json:"teardown_date"`
	Budget          float64    `json:"budget"`
	Currency        string     `json:"currency"`
	Notes           string     `json:"notes"`
	Tags            []string   `json:"tags"`
	CreatedByUserID string     `json:"created_by_user_id"`
}

type UpdateProjectCommand struct {
	ID            string     `json:"id"`
	TenantID      string     `json:"tenant_id"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	ClientName    string     `json:"client_name"`
	ClientEmail   string     `json:"client_email"`
	ClientPhone   string     `json:"client_phone"`
	ClientAddress AddressDTO `json:"client_address"`
	VenueAddress  AddressDTO `json:"venue_address"`
	StartDate     time.Time  `json:"start_date"`
	EndDate       time.Time  `json:"end_date"`
	SetupDate     *time.Time `json:"setup_date"`
	TeardownDate  *time.Time `json:"teardown_date"`
	Budget        float64    `json:"budget"`
	Currency      string     `json:"currency"`
	Notes         string     `json:"notes"`
	Tags          []string   `json:"tags"`
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

// Customer Commands

type CreateCustomerCommand struct {
	TenantID        string
	Name            string
	Email           string
	Phone           string
	AddressStreet   string
	AddressCity     string
	AddressPostcode string
	AddressCountry  string
	TaxID           string
	Notes           string
}

type UpdateCustomerCommand struct {
	ID              string
	TenantID        string
	Name            string
	Email           string
	Phone           string
	AddressStreet   string
	AddressCity     string
	AddressPostcode string
	AddressCountry  string
	TaxID           string
	Notes           string
}

type DeleteCustomerCommand struct {
	ID       string
	TenantID string
}

// Copy Project Command
type CopyProjectCommand struct {
	ID              string
	TenantID        string
	NewStartDate    time.Time
	NewEndDate      time.Time
	CreatedByUserID string
}
