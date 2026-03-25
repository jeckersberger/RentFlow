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
	SessionID   string   `json:"session_id"`
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

// ConditionRatingPayload is the JSON payload for per-equipment condition during checkin
type ConditionRatingPayload struct {
	Rating         int    `json:"rating"`
	Notes          string `json:"notes"`
	DamageReported bool   `json:"damage_reported"`
}

// BulkAction represents a single action inside POST /api/v1/scanner/bulk
type BulkAction struct {
	Type             string                        `json:"type"` // "checkout" or "checkin"
	EquipmentIDs     []string                      `json:"equipment_ids"`
	ProjectID        string                        `json:"project_id,omitempty"`
	Notes            string                        `json:"notes,omitempty"`
	ConditionRatings map[string]ConditionRatingPayload `json:"condition_ratings,omitempty"`
	Timestamp        string                        `json:"timestamp,omitempty"`
}

// AdhocBookingResult is the response for POST /api/v1/scanner/adhoc-booking
type AdhocBookingResult struct {
	Success     bool   `json:"success"`
	EquipmentID string `json:"equipment_id"`
	ProjectID   string `json:"project_id"`
	CheckedOut  int    `json:"checked_out"`
	Notes       string `json:"notes"`
}

// BulkResult is the response for POST /api/v1/scanner/bulk
type BulkResult struct {
	Processed int               `json:"processed"`
	Failed    int               `json:"failed"`
	Results   []BulkActionResult `json:"results"`
}

// BulkActionResult is the per-action result inside BulkResult
type BulkActionResult struct {
	Type    string      `json:"type"`
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}
