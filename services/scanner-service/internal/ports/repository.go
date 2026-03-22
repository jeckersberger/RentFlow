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
}
