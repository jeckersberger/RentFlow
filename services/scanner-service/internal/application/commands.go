package application

import "github.com/jeckersberger/rentflow/services/scanner-service/internal/domain"

type ProcessScanCommand struct {
	TenantID   string
	Barcode    string
	ScanType   domain.ScanType
	UserID     string
	DeviceID   string
	DeviceType domain.DeviceType
	ProjectID  *string
	LocationID *string
	Latitude   *float64
	Longitude  *float64
	Notes      string
}

type BatchScanCommand struct {
	TenantID string
	Scans    []ProcessScanCommand
}

type SyncOfflineCommand struct {
	TenantID string
	Scans    []ProcessScanCommand
}

type RegisterDeviceCommand struct {
	TenantID   string
	Name       string
	Type       domain.DeviceType
	Serial     string
	Location   string
	CreatedByUserID string
}
