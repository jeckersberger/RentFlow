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
	TenantID        string
	Name            string
	Type            domain.DeviceType
	Serial          string
	Location        string
	CreatedByUserID string
}

type StartScanSessionCommand struct {
	TenantID   string
	UserID     string
	Context    string // check_out, check_in, lager_einraeumen, inventur
	ProjectID  *string
	DeviceType domain.DeviceType
	DeviceID   string
}

type EndScanSessionCommand struct {
	TenantID  string
	SessionID string
}

type ProcessSessionScanCommand struct {
	SessionID   string
	TenantID    string
	Barcode     string
	DeviceID    string
	DeviceType  domain.DeviceType
	Latitude    *float64
	Longitude   *float64
	Notes       string
}

type SyncOfflineQueueCommand struct {
	TenantID string
	Items    []struct {
		ID       string `json:"id"`
		Barcode  string `json:"barcode"`
		ScanType string `json:"scan_type"`
		UserID   string `json:"user_id"`
		DeviceID string `json:"device_id"`
		ProjectID *string `json:"project_id"`
		Latitude *float64 `json:"latitude"`
		Longitude *float64 `json:"longitude"`
		Notes    string `json:"notes"`
	}
}
