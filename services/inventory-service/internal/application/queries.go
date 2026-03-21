package application

import "github.com/jeckersberger/rentflow/services/inventory-service/internal/domain"

type ListEquipmentQuery struct {
	TenantID   string
	Status     *domain.EquipmentStatus
	CategoryID *string
	LocationID *string
	Limit      int
	Offset     int
}

type SearchEquipmentQuery struct {
	TenantID   string
	SearchTerm string
	Limit      int
	Offset     int
}

type GetEquipmentQuery struct {
	ID       string
	TenantID string
}

type GetEquipmentByBarcodeQuery struct {
	TenantID string
	Barcode  string
}

type ListCategoriesQuery struct {
	TenantID string
}

type GetCategoryQuery struct {
	ID       string
	TenantID string
}

type ListFlightcasesQuery struct {
	TenantID string
	Limit    int
	Offset   int
}

type GetFlightcaseQuery struct {
	ID       string
	TenantID string
}

type GetFlightcaseByBarcodeQuery struct {
	TenantID string
	Barcode  string
}

type PaginatedResult struct {
	Data   interface{} `json:"data"`
	Total  int64       `json:"total"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
}
