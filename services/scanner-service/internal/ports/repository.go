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
