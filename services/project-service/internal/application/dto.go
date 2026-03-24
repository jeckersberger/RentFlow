package application

import (
	"time"

	"github.com/jeckersberger/rentflow/services/project-service/internal/domain"
)

type ProjectDTO struct {
	ID              string     `json:"id"`
	TenantID        string     `json:"tenant_id"`
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	ClientName      string     `json:"client_name"`
	ClientEmail     string     `json:"client_email"`
	ClientPhone     string     `json:"client_phone"`
	ClientAddress   AddressDTO `json:"client_address"`
	VenueAddress    AddressDTO `json:"venue_address"`
	Status          string     `json:"status"`
	StartDate       time.Time  `json:"start_date"`
	EndDate         time.Time  `json:"end_date"`
	SetupDate       *time.Time `json:"setup_date,omitempty"`
	TeardownDate    *time.Time `json:"teardown_date,omitempty"`
	ProjectManager  string     `json:"project_manager"`
	Budget          float64    `json:"budget"`
	Currency        string     `json:"currency"`
	Notes           string     `json:"notes"`
	Tags            []string   `json:"tags"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	CreatedByUserID string     `json:"created_by_user_id"`
}

type AddressDTO struct {
	Street      string `json:"street"`
	City        string `json:"city"`
	State       string `json:"state"`
	PostalCode  string `json:"postal_code"`
	Country     string `json:"country"`
	Coordinates string `json:"coordinates,omitempty"`
}

type PacklistDTO struct {
	ID        string            `json:"id"`
	ProjectID string            `json:"project_id"`
	TenantID  string            `json:"tenant_id"`
	Name      string            `json:"name"`
	Items     []PacklistItemDTO `json:"items"`
	Status    string            `json:"status"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type PacklistItemDTO struct {
	ID               string `json:"id"`
	EquipmentID      string `json:"equipment_id"`
	EquipmentName    string `json:"equipment_name"`
	Quantity         int    `json:"quantity"`
	QuantityPacked   int    `json:"quantity_packed"`
	QuantityReturned int    `json:"quantity_returned"`
	Notes            string `json:"notes"`
	StorageLocation  string `json:"storage_location"`
	Status           string `json:"status"`
}

type ReservationDTO struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	ProjectID   string    `json:"project_id"`
	EquipmentID string    `json:"equipment_id"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Packing List JSON response grouped by storage location
type PackingListDTO struct {
	ProjectName  string                    `json:"project_name"`
	ProjectDates string                    `json:"project_dates"`
	Locations    []PackingListLocationDTO  `json:"locations"`
	TotalItems   int                       `json:"total_items"`
}

type PackingListLocationDTO struct {
	Location string                  `json:"location"`
	Items    []PackingListItemDTO    `json:"items"`
}

type PackingListItemDTO struct {
	Name           string `json:"name"`
	Quantity       int    `json:"quantity"`
	QuantityPacked int    `json:"quantity_packed"`
	SKU            string `json:"sku,omitempty"`
	Barcode        string `json:"barcode,omitempty"`
	Status         string `json:"status"`
	PacklistName   string `json:"packlist_name,omitempty"`
}

type ReservationConflictDTO struct {
	ReservationID string    `json:"reservation_id"`
	ProjectID     string    `json:"project_id"`
	ProjectName   string    `json:"project_name,omitempty"`
	EquipmentID   string    `json:"equipment_id"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	Status        string    `json:"status"`
}

type PaginatedResult struct {
	Data   interface{} `json:"data"`
	Total  int64       `json:"total"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
}

// DTO conversion functions

func ProjectToDTO(p *domain.Project) *ProjectDTO {
	return &ProjectDTO{
		ID:          p.ID,
		TenantID:    p.TenantID,
		Name:        p.Name,
		Description: p.Description,
		ClientName:  p.ClientName,
		ClientEmail: p.ClientEmail,
		ClientPhone: p.ClientPhone,
		ClientAddress: AddressDTO{
			Street:      p.ClientAddress.Street,
			City:        p.ClientAddress.City,
			State:       p.ClientAddress.State,
			PostalCode:  p.ClientAddress.PostalCode,
			Country:     p.ClientAddress.Country,
			Coordinates: p.ClientAddress.Coordinates,
		},
		VenueAddress: AddressDTO{
			Street:      p.VenueAddress.Street,
			City:        p.VenueAddress.City,
			State:       p.VenueAddress.State,
			PostalCode:  p.VenueAddress.PostalCode,
			Country:     p.VenueAddress.Country,
			Coordinates: p.VenueAddress.Coordinates,
		},
		Status:          string(p.Status),
		StartDate:       p.StartDate,
		EndDate:         p.EndDate,
		SetupDate:       p.SetupDate,
		TeardownDate:    p.TeardownDate,
		ProjectManager:  p.ProjectManager,
		Budget:          p.Budget,
		Currency:        p.Currency,
		Notes:           p.Notes,
		Tags:            p.Tags,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
		CreatedByUserID: p.CreatedByUserID,
	}
}

func PacklistToDTO(p *domain.Packlist) *PacklistDTO {
	items := make([]PacklistItemDTO, len(p.Items))
	for i, item := range p.Items {
		items[i] = PacklistItemDTO{
			ID:               item.ID,
			EquipmentID:      item.EquipmentID,
			EquipmentName:    item.EquipmentName,
			Quantity:         item.Quantity,
			QuantityPacked:   item.QuantityPacked,
			QuantityReturned: item.QuantityReturned,
			Notes:            item.Notes,
			StorageLocation:  item.StorageLocation,
			Status:           string(item.Status),
		}
	}

	return &PacklistDTO{
		ID:        p.ID,
		ProjectID: p.ProjectID,
		TenantID:  p.TenantID,
		Name:      p.Name,
		Items:     items,
		Status:    string(p.Status),
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func ReservationToDTO(r *domain.Reservation) *ReservationDTO {
	return &ReservationDTO{
		ID:          r.ID,
		TenantID:    r.TenantID,
		ProjectID:   r.ProjectID,
		EquipmentID: r.EquipmentID,
		StartDate:   r.StartDate,
		EndDate:     r.EndDate,
		Status:      string(r.Status),
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

type CustomerDTO struct {
	ID              string    `json:"id"`
	TenantID        string    `json:"tenant_id"`
	Name            string    `json:"name"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone"`
	AddressStreet   string    `json:"address_street"`
	AddressCity     string    `json:"address_city"`
	AddressPostcode string    `json:"address_postcode"`
	AddressCountry  string    `json:"address_country"`
	TaxID           string    `json:"tax_id"`
	Notes           string    `json:"notes"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ContactDTO struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Type        string    `json:"type"`
	CompanyName string    `json:"company_name"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	Mobile      string    `json:"mobile"`
	Website     string    `json:"website"`
	Street      string    `json:"street"`
	HouseNumber string    `json:"house_number"`
	Zip         string    `json:"zip"`
	City        string    `json:"city"`
	Country     string    `json:"country"`
	VatID       string    `json:"vat_id"`
	Notes       string    `json:"notes"`
	Tags        []string  `json:"tags"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedBy   string    `json:"created_by,omitempty"`
}

func ContactToDTO(c *domain.Contact) *ContactDTO {
	return &ContactDTO{
		ID:          c.ID,
		TenantID:    c.TenantID,
		Type:        string(c.Type),
		CompanyName: c.CompanyName,
		FirstName:   c.FirstName,
		LastName:    c.LastName,
		Email:       c.Email,
		Phone:       c.Phone,
		Mobile:      c.Mobile,
		Website:     c.Website,
		Street:      c.Street,
		HouseNumber: c.HouseNumber,
		Zip:         c.Zip,
		City:        c.City,
		Country:     c.Country,
		VatID:       c.VatID,
		Notes:       c.Notes,
		Tags:        c.Tags,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
		CreatedBy:   c.CreatedBy,
	}
}

func CustomerToDTO(c *domain.Customer) *CustomerDTO {
	return &CustomerDTO{
		ID:              c.ID,
		TenantID:        c.TenantID,
		Name:            c.Name,
		Email:           c.Email,
		Phone:           c.Phone,
		AddressStreet:   c.AddressStreet,
		AddressCity:     c.AddressCity,
		AddressPostcode: c.AddressPostcode,
		AddressCountry:  c.AddressCountry,
		TaxID:           c.TaxID,
		Notes:           c.Notes,
		CreatedAt:       c.CreatedAt,
		UpdatedAt:       c.UpdatedAt,
	}
}
