package application

import (
	"time"

	"github.com/jeckersberger/rentflow/services/inventory-service/internal/domain"
)

type EquipmentDTO struct {
	ID              string            `json:"id"`
	TenantID        string            `json:"tenant_id"`
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	CategoryID      string            `json:"category_id"`
	SKU             string            `json:"sku"`
	SerialNumber    string            `json:"serial_number"`
	Barcode         string            `json:"barcode"`
	Status          string            `json:"status"`
	Condition       string            `json:"condition"`
	PurchaseDate    *time.Time        `json:"purchase_date,omitempty"`
	PurchasePrice   float64           `json:"purchase_price"`
	RentalPriceDay  float64           `json:"rental_price_day"`
	RentalPriceWeek float64           `json:"rental_price_week"`
	Weight          float64           `json:"weight"`
	Dimensions      DimensionsDTO     `json:"dimensions"`
	LocationID      string            `json:"location_id"`
	ImageRefs       []string          `json:"image_refs"`
	Tags            []string          `json:"tags"`
	CustomFields    map[string]string `json:"custom_fields"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	CreatedByUserID string            `json:"created_by_user_id"`
}

type DimensionsDTO struct {
	Length float64 `json:"length"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Unit   string  `json:"unit"`
}

type CategoryDTO struct {
	ID              string    `json:"id"`
	TenantID        string    `json:"tenant_id"`
	Name            string    `json:"name"`
	ParentID        *string   `json:"parent_id,omitempty"`
	Icon            string    `json:"icon"`
	Color           string    `json:"color"`
	SortOrder       int       `json:"sort_order"`
	CreatedAt       time.Time `json:"created_at"`
	CreatedByUserID string    `json:"created_by_user_id"`
}

type FlightcaseDTO struct {
	ID              string              `json:"id"`
	TenantID        string              `json:"tenant_id"`
	Name            string              `json:"name"`
	Description     string              `json:"description"`
	Barcode         string              `json:"barcode"`
	Contents        []FlightcaseItemDTO `json:"contents"`
	Weight          float64             `json:"weight"`
	LocationID      string              `json:"location_id"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	CreatedByUserID string              `json:"created_by_user_id"`
}

type FlightcaseItemDTO struct {
	EquipmentID string    `json:"equipment_id"`
	Quantity    int       `json:"quantity"`
	AddedAt     time.Time `json:"added_at"`
}

// Create Equipment DTO from domain object
func EquipmentToDTO(eq *domain.Equipment) *EquipmentDTO {
	return &EquipmentDTO{
		ID:              eq.ID,
		TenantID:        eq.TenantID,
		Name:            eq.Name,
		Description:     eq.Description,
		CategoryID:      eq.CategoryID,
		SKU:             eq.SKU,
		SerialNumber:    eq.SerialNumber,
		Barcode:         eq.Barcode,
		Status:          string(eq.Status),
		Condition:       string(eq.Condition),
		PurchaseDate:    eq.PurchaseDate,
		PurchasePrice:   eq.PurchasePrice,
		RentalPriceDay:  eq.RentalPriceDay,
		RentalPriceWeek: eq.RentalPriceWeek,
		Weight:          eq.Weight,
		Dimensions: DimensionsDTO{
			Length: eq.Dimensions.Length,
			Width:  eq.Dimensions.Width,
			Height: eq.Dimensions.Height,
			Unit:   eq.Dimensions.Unit,
		},
		LocationID:      eq.LocationID,
		ImageRefs:       eq.ImageRefs,
		Tags:            eq.Tags,
		CustomFields:    eq.CustomFields,
		CreatedAt:       eq.CreatedAt,
		UpdatedAt:       eq.UpdatedAt,
		CreatedByUserID: eq.CreatedByUserID,
	}
}

// Create Category DTO from domain object
func CategoryToDTO(cat *domain.Category) *CategoryDTO {
	return &CategoryDTO{
		ID:              cat.ID,
		TenantID:        cat.TenantID,
		Name:            cat.Name,
		ParentID:        cat.ParentID,
		Icon:            cat.Icon,
		Color:           cat.Color,
		SortOrder:       cat.SortOrder,
		CreatedAt:       cat.CreatedAt,
		CreatedByUserID: cat.CreatedByUserID,
	}
}

// Create Flightcase DTO from domain object
func FlightcaseToDTO(fc *domain.Flightcase) *FlightcaseDTO {
	items := make([]FlightcaseItemDTO, len(fc.Contents))
	for i, item := range fc.Contents {
		items[i] = FlightcaseItemDTO{
			EquipmentID: item.EquipmentID,
			Quantity:    item.Quantity,
			AddedAt:     item.AddedAt,
		}
	}

	return &FlightcaseDTO{
		ID:              fc.ID,
		TenantID:        fc.TenantID,
		Name:            fc.Name,
		Description:     fc.Description,
		Barcode:         fc.Barcode,
		Contents:        items,
		Weight:          fc.Weight,
		LocationID:      fc.LocationID,
		CreatedAt:       fc.CreatedAt,
		UpdatedAt:       fc.UpdatedAt,
		CreatedByUserID: fc.CreatedByUserID,
	}
}
