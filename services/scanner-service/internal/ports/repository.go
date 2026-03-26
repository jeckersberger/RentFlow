package ports

import (
	"context"

	"github.com/jeckersberger/rentflow/services/scanner-service/internal/domain"
)

type ScanEventRepository interface {
	Create(ctx context.Context, event *domain.ScanEvent) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.ScanEvent, error)
	List(ctx context.Context, tenantID string, query *ScanListQuery) (*ScanListResult, error)
	Update(ctx context.Context, event *domain.ScanEvent) error
	GetByBarcode(ctx context.Context, tenantID, barcode string) (*domain.ScanEvent, error)
}

type ScanListQuery struct {
	EquipmentID *string
	ProjectID   *string
	UserID      *string
	DeviceID    *string
	SessionID   *string
	Status      *string
	Limit       int
	Offset      int
}

type ScanListResult struct {
	Items  []*domain.ScanEvent
	Total  int
	Limit  int
	Offset int
}

type DeviceRepository interface {
	Create(ctx context.Context, device *domain.Device) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.Device, error)
	GetBySerial(ctx context.Context, tenantID, serial string) (*domain.Device, error)
	List(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Device, int, error)
	Update(ctx context.Context, device *domain.Device) error
	Delete(ctx context.Context, tenantID, id string) error
}

type ScanSessionRepository interface {
	Create(ctx context.Context, session *domain.ScanSession) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.ScanSession, error)
	Update(ctx context.Context, session *domain.ScanSession) error
	List(ctx context.Context, tenantID string, limit, offset int) ([]*domain.ScanSession, int, error)
	GetByUserAndContext(ctx context.Context, tenantID, userID string, context domain.ScanContext) (*domain.ScanSession, error)
}

type OfflineQueueRepository interface {
	Create(ctx context.Context, item *domain.OfflineQueueItem) error
	GetByID(ctx context.Context, id string) (*domain.OfflineQueueItem, error)
	GetPending(ctx context.Context, tenantID string, limit int) ([]*domain.OfflineQueueItem, error)
	Update(ctx context.Context, item *domain.OfflineQueueItem) error
	DeleteByID(ctx context.Context, id string) error
	GetCount(ctx context.Context, tenantID string, status string) (int, error)
}

// ScannerDeviceRepository manages scanner devices for the "Find My Scanner" feature.
// Uses the scanner_devices table (separate from the legacy devices table).
type ScannerDeviceRepository interface {
	Upsert(ctx context.Context, device *ScannerDevice) error
	List(ctx context.Context, tenantID string) ([]*ScannerDevice, error)
	GetByDeviceID(ctx context.Context, tenantID, deviceID string) (*ScannerDevice, error)
	GetByID(ctx context.Context, id string) (*ScannerDevice, error)
	SetRingRequested(ctx context.Context, id string, requested bool) error
	UpdateLastSeen(ctx context.Context, tenantID, deviceID string) error
}

// ScannerDevice represents a registered scanner device (app instance)
type ScannerDevice struct {
	ID            string  `json:"id"`
	TenantID      string  `json:"tenant_id"`
	DeviceID      string  `json:"device_id"`
	DeviceName    string  `json:"device_name"`
	DeviceType    string  `json:"device_type"`
	FCMToken      *string `json:"fcm_token"`
	RingRequested bool    `json:"ring_requested"`
	LastSeen      *string `json:"last_seen"`
	CreatedAt     string  `json:"created_at"`
}

type InventoryServiceClient interface {
	ResolveBarcode(ctx context.Context, tenantID, barcode string) (string, error)
	UpdateEquipmentStatus(ctx context.Context, tenantID, equipmentID, status string) error
	// Scanner App contract: resolve barcode/RFID to full equipment detail
	ResolveEquipment(ctx context.Context, tenantID, code, scanType string) (*EquipmentDetail, error)
	// Scanner App contract: batch checkout
	CheckOutEquipment(ctx context.Context, tenantID string, equipmentIDs []string, projectID, notes string) (*CheckoutResult, error)
	// Scanner App contract: batch checkin with condition ratings
	CheckInEquipment(ctx context.Context, tenantID string, equipmentIDs []string, conditionRatings map[string]ConditionRating) (*CheckinResult, error)
}

// EquipmentDetail is the full equipment object returned by /scanner/scan
type EquipmentDetail struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Category    string       `json:"category"`
	Barcode     string       `json:"barcode"`
	RfidTag     string       `json:"rfid_tag"`
	Status      string       `json:"status"`
	Location    string       `json:"location"`
	ImageURL    string       `json:"image_url"`
	LastProject *ProjectRef  `json:"last_project"`
}

type ProjectRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ConditionRating struct {
	Rating         int    `json:"rating"`
	Notes          string `json:"notes"`
	DamageReported bool   `json:"damage_reported"`
}

type CheckoutResult struct {
	CheckedOut int        `json:"checked_out"`
	Project    ProjectRef `json:"project"`
}

type CheckinResult struct {
	CheckedIn     int `json:"checked_in"`
	DamageReports int `json:"damage_reports"`
}
