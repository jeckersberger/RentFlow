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
		ID:        id,
		TenantID:  tenantID,
		Barcode:   barcode,
		ScanType:  scanType,
		UserID:    userID,
		DeviceID:  deviceID,
		DeviceType: deviceType,
		Timestamp: time.Now(),
		Status:    ScanPending,
		CreatedAt: time.Now(),
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

type Device struct {
	ID       string
	TenantID string
	Name     string
	Type     DeviceType
	Serial   string
	Active   bool
	Location string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewDevice(id, tenantID, name string, deviceType DeviceType, serial string) *Device {
	return &Device{
		ID:       id,
		TenantID: tenantID,
		Name:     name,
		Type:     deviceType,
		Serial:   serial,
		Active:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
