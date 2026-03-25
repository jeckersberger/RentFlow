package domain

import (
	"fmt"
	"time"
)

type ScanType string

const (
	ScanCheckIn   ScanType = "check_in"
	ScanCheckOut  ScanType = "check_out"
	ScanInventory ScanType = "inventory"
	ScanMovement  ScanType = "movement"
	ScanReturn    ScanType = "return"
)

type ScanContext string

const (
	ContextCheckOut       ScanContext = "check_out"
	ContextCheckIn        ScanContext = "check_in"
	ContextWarehouseStore ScanContext = "lager_einraeumen"
	ContextInventory      ScanContext = "inventur"
)

type DeviceType string

const (
	DeviceUSB      DeviceType = "usb"
	DeviceHandheld DeviceType = "handheld"
	DeviceCamera   DeviceType = "camera"
	DeviceRFID     DeviceType = "rfid"
)

type ScanStatus string

const (
	ScanPending   ScanStatus = "pending"
	ScanProcessed ScanStatus = "processed"
	ScanFailed    ScanStatus = "failed"
	ScanSynced    ScanStatus = "synced"
)

type ScanEvent struct {
	ID          string
	TenantID    string
	Barcode     string
	ScanType    ScanType
	EquipmentID string
	ProjectID   *string
	LocationID  *string
	SessionID   string
	UserID      string
	DeviceID    string
	DeviceType  DeviceType
	Timestamp   time.Time
	Latitude    *float64
	Longitude   *float64
	Notes       string
	Status      ScanStatus
	CreatedAt   time.Time
}

func NewScanEvent(
	id string,
	tenantID string,
	barcode string,
	scanType ScanType,
	userID string,
	deviceID string,
	deviceType DeviceType,
) *ScanEvent {
	return &ScanEvent{
		ID:         id,
		TenantID:   tenantID,
		Barcode:    barcode,
		ScanType:   scanType,
		UserID:     userID,
		DeviceID:   deviceID,
		DeviceType: deviceType,
		Timestamp:  time.Now(),
		Status:     ScanPending,
		CreatedAt:  time.Now(),
	}
}

func (s *ScanEvent) Validate() error {
	if s.TenantID == "" {
		return fmt.Errorf("tenant ID is required")
	}
	if s.Barcode == "" {
		return fmt.Errorf("barcode is required")
	}
	if s.UserID == "" {
		return fmt.Errorf("user ID is required")
	}
	if s.DeviceID == "" {
		return fmt.Errorf("device ID is required")
	}
	return nil
}

func (s *ScanEvent) MarkProcessed(equipmentID string) {
	s.EquipmentID = equipmentID
	s.Status = ScanProcessed
}

func (s *ScanEvent) MarkFailed() {
	s.Status = ScanFailed
}

func (s *ScanEvent) MarkSynced() {
	s.Status = ScanSynced
}

type ScanSession struct {
	ID              string
	TenantID        string
	UserID          string
	Context         ScanContext
	ProjectID       *string
	StartedAt       time.Time
	EndedAt         *time.Time
	DeviceType      DeviceType
	DeviceID        string
	TotalScans      int
	SuccessfulScans int
	FailedScans     int
	SignatureData   string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func NewScanSession(
	id string,
	tenantID string,
	userID string,
	context ScanContext,
	deviceType DeviceType,
	deviceID string,
) *ScanSession {
	return &ScanSession{
		ID:           id,
		TenantID:     tenantID,
		UserID:       userID,
		Context:      context,
		DeviceType:   deviceType,
		DeviceID:     deviceID,
		StartedAt:    time.Now(),
		TotalScans:   0,
		SuccessfulScans: 0,
		FailedScans:    0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func (s *ScanSession) End() {
	now := time.Now()
	s.EndedAt = &now
	s.UpdatedAt = now
}

func (s *ScanSession) IncrementTotal() {
	s.TotalScans++
	s.UpdatedAt = time.Now()
}

func (s *ScanSession) IncrementSuccessful() {
	s.SuccessfulScans++
	s.UpdatedAt = time.Now()
}

func (s *ScanSession) IncrementFailed() {
	s.FailedScans++
	s.UpdatedAt = time.Now()
}

type OfflineQueueItem struct {
	ID        string
	TenantID  string
	DeviceID  string
	Payload   string // JSON
	CreatedAt time.Time
	SyncedAt  *time.Time
	SyncStatus string // pending, syncing, synced, failed
}

func NewOfflineQueueItem(
	id string,
	tenantID string,
	deviceID string,
	payload string,
) *OfflineQueueItem {
	return &OfflineQueueItem{
		ID:         id,
		TenantID:   tenantID,
		DeviceID:   deviceID,
		Payload:    payload,
		CreatedAt:  time.Now(),
		SyncStatus: "pending",
	}
}

func (o *OfflineQueueItem) MarkSynced() {
	now := time.Now()
	o.SyncedAt = &now
	o.SyncStatus = "synced"
}

func (o *OfflineQueueItem) MarkFailed() {
	o.SyncStatus = "failed"
}

type Device struct {
	ID        string
	TenantID  string
	Name      string
	Type      DeviceType
	Serial    string
	Active    bool
	Location  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewDevice(id, tenantID, name string, deviceType DeviceType, serial string) *Device {
	return &Device{
		ID:        id,
		TenantID:  tenantID,
		Name:      name,
		Type:      deviceType,
		Serial:    serial,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
