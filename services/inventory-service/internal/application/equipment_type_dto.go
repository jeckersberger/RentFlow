package application

import (
	"time"

	"github.com/jeckersberger/rentflow/services/inventory-service/internal/domain"
)

// Equipment Type DTO

type EquipmentTypeDTO struct {
	ID               string            `json:"id"`
	TenantID         string            `json:"tenant_id"`
	Name             string            `json:"name"`
	Description      string            `json:"description"`
	CategoryID       string            `json:"category_id"`
	Manufacturer     string            `json:"manufacturer"`
	Model            string            `json:"model"`
	SKUPrefix        string            `json:"sku_prefix"`
	RentalPriceDay   float64           `json:"rental_price_day"`
	RentalPriceWeek  float64           `json:"rental_price_week"`
	ReplacementValue float64           `json:"replacement_value"`
	Weight           float64           `json:"weight"`
	Dimensions       DimensionsDTO     `json:"dimensions"`
	ImageURL         string            `json:"image_url"`
	Tags             []string          `json:"tags"`
	CustomFields     map[string]string `json:"custom_fields"`
	ItemCount        int64             `json:"item_count"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

func EquipmentTypeToDTO(et *domain.EquipmentType) *EquipmentTypeDTO {
	return &EquipmentTypeDTO{
		ID:               et.ID,
		TenantID:         et.TenantID,
		Name:             et.Name,
		Description:      et.Description,
		CategoryID:       et.CategoryID,
		Manufacturer:     et.Manufacturer,
		Model:            et.Model,
		SKUPrefix:        et.SKUPrefix,
		RentalPriceDay:   et.RentalPriceDay,
		RentalPriceWeek:  et.RentalPriceWeek,
		ReplacementValue: et.ReplacementValue,
		Weight:           et.Weight,
		Dimensions: DimensionsDTO{
			Length: et.Dimensions.Length,
			Width:  et.Dimensions.Width,
			Height: et.Dimensions.Height,
			Unit:   et.Dimensions.Unit,
		},
		ImageURL:     et.ImageURL,
		Tags:         et.Tags,
		CustomFields: et.CustomFields,
		CreatedAt:    et.CreatedAt,
		UpdatedAt:    et.UpdatedAt,
	}
}

// Equipment Type Commands

type CreateEquipmentTypeCommand struct {
	TenantID         string            `json:"tenant_id"`
	Name             string            `json:"name"`
	Description      string            `json:"description"`
	CategoryID       string            `json:"category_id"`
	Manufacturer     string            `json:"manufacturer"`
	Model            string            `json:"model"`
	SKUPrefix        string            `json:"sku_prefix"`
	RentalPriceDay   float64           `json:"rental_price_day"`
	RentalPriceWeek  float64           `json:"rental_price_week"`
	ReplacementValue float64           `json:"replacement_value"`
	Weight           float64           `json:"weight"`
	Dimensions       DimensionsDTO     `json:"dimensions"`
	ImageURL         string            `json:"image_url"`
	Tags             []string          `json:"tags"`
	CustomFields     map[string]string `json:"custom_fields"`
}

type UpdateEquipmentTypeCommand struct {
	ID               string            `json:"id"`
	TenantID         string            `json:"tenant_id"`
	Name             string            `json:"name"`
	Description      string            `json:"description"`
	CategoryID       string            `json:"category_id"`
	Manufacturer     string            `json:"manufacturer"`
	Model            string            `json:"model"`
	SKUPrefix        string            `json:"sku_prefix"`
	RentalPriceDay   float64           `json:"rental_price_day"`
	RentalPriceWeek  float64           `json:"rental_price_week"`
	ReplacementValue float64           `json:"replacement_value"`
	Weight           float64           `json:"weight"`
	Dimensions       DimensionsDTO     `json:"dimensions"`
	ImageURL         string            `json:"image_url"`
	Tags             []string          `json:"tags"`
	CustomFields     map[string]string `json:"custom_fields"`
}

type CreateItemsFromTypeCommand struct {
	TenantID        string `json:"tenant_id"`
	TypeID          string `json:"type_id"`
	Quantity        int    `json:"quantity"`
	CreatedByUserID string `json:"created_by_user_id"`
}
