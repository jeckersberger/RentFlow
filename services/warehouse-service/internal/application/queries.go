package application

import "time"

type LocationDTO struct {
	ID           string `json:"id"`
	TenantID     string `json:"tenant_id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	ParentID     *string `json:"parent_id"`
	Path         string `json:"path"`
	Capacity     int `json:"capacity"`
	CurrentCount int `json:"current_count"`
	SortOrder    int `json:"sort_order"`
	Barcode      string `json:"barcode"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type LocationTreeNode struct {
	*LocationDTO
	Children []*LocationTreeNode `json:"children,omitempty"`
}

type MovementDTO struct {
	ID             string `json:"id"`
	TenantID       string `json:"tenant_id"`
	EquipmentID    string `json:"equipment_id"`
	FromLocationID *string `json:"from_location_id"`
	ToLocationID   string `json:"to_location_id"`
	MovementType   string `json:"movement_type"`
	Quantity       int `json:"quantity"`
	Reason         string `json:"reason"`
	UserID         string `json:"user_id"`
	ProjectID      *string `json:"project_id"`
	Timestamp      string `json:"timestamp"`
	CreatedAt      string `json:"created_at"`
}

type InventoryCheckDTO struct {
	ID          string                      `json:"id"`
	TenantID    string                      `json:"tenant_id"`
	Name        string                      `json:"name"`
	LocationID  *string                     `json:"location_id"`
	Status      string                      `json:"status"`
	Items       []InventoryCheckItemDTO     `json:"items"`
	StartedAt   *string                     `json:"started_at"`
	CompletedAt *string                     `json:"completed_at"`
	CreatedAt   string                      `json:"created_at"`
	UpdatedAt   string                      `json:"updated_at"`
}

type InventoryCheckItemDTO struct {
	EquipmentID   string `json:"equipment_id"`
	ExpectedCount int `json:"expected_count"`
	ActualCount   int `json:"actual_count"`
	Status        string `json:"status"`
	Notes         string `json:"notes"`
	ScannedAt     *string `json:"scanned_at"`
}

type PaginatedResult struct {
	Data   interface{} `json:"data"`
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type HistoryQuery struct {
	EquipmentID *string
	FromLocation *string
	ToLocation   *string
	MovementType *string
	StartDate    *time.Time
	EndDate      *time.Time
	Limit        int
	Offset       int
}

type LocationContentsDTO struct {
	LocationID string      `json:"location_id"`
	Items      interface{} `json:"items"`
	Total      int `json:"total"`
}
