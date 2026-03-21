package application

import "time"

type ScanHistoryQuery struct {
	TenantID    string
	EquipmentID *string
	ProjectID   *string
	UserID      *string
	DeviceID    *string
	ScanType    *string
	Status      *string
	StartDate   *time.Time
	EndDate     *time.Time
	Limit       int
	Offset      int
}

type PaginatedResult struct {
	Data   interface{} `json:"data"`
	Total  int         `json:"total"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
}

type ScanEventDTO struct {
	ID          string   `json:"id"`
	TenantID    string   `json:"tenant_id"`
	Barcode     string   `json:"barcode"`
	ScanType    string   `json:"scan_type"`
	EquipmentID string   `json:"equipment_id"`
	ProjectID   *string  `json:"project_id"`
	LocationID  *string  `json:"location_id"`
	UserID      string   `json:"user_id"`
	DeviceID    string   `json:"device_id"`
	DeviceType  string   `json:"device_type"`
	Timestamp   string   `json:"timestamp"`
	Latitude    *float64 `json:"latitude"`
	Longitude   *float64 `json:"longitude"`
	Notes       string   `json:"notes"`
	Status      string   `json:"status"`
	CreatedAt   string   `json:"created_at"`
}

type DeviceDTO struct {
	ID        string `json:"id"`
	TenantID  string `json:"tenant_id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Serial    string `json:"serial"`
	Active    bool   `json:"active"`
	Location  string `json:"location"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
