package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Filter
// ---------------------------------------------------------------------------

// ScanEventFilter holds optional criteria for listing scan events.
type ScanEventFilter struct {
	EquipmentID *uuid.UUID `json:"equipment_id,omitempty"`
	ProjectID   *uuid.UUID `json:"project_id,omitempty"`
	Action      *string    `json:"action,omitempty"`
	DeviceID    *string    `json:"device_id,omitempty"`
	Page        int        `json:"page"`
	PerPage     int        `json:"per_page"`
}

// ---------------------------------------------------------------------------
// Repository ports (driven / secondary adapters)
// ---------------------------------------------------------------------------

// ScanEventRepository defines persistence operations for ScanEvent aggregates.
type ScanEventRepository interface {
	Create(ctx context.Context, event *ScanEvent) error
	List(ctx context.Context, tenantID uuid.UUID, filter ScanEventFilter) ([]*ScanEvent, int64, error)
	ExistsByDedup(ctx context.Context, tenantID uuid.UUID, deviceID string, timestamp time.Time) (bool, error)
}

// ScannerDeviceRepository defines persistence operations for ScannerDevice aggregates.
type ScannerDeviceRepository interface {
	Upsert(ctx context.Context, device *ScannerDevice) error
	List(ctx context.Context, tenantID uuid.UUID) ([]*ScannerDevice, error)
	GetByDeviceID(ctx context.Context, tenantID uuid.UUID, deviceID string) (*ScannerDevice, error)
	GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*ScannerDevice, error)
	UpdateLastSeen(ctx context.Context, tenantID uuid.UUID, deviceID string) error
	SetRingRequested(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, ring bool) error
}
