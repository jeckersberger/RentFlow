package domain

import (
	"time"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Scan Action constants
// ---------------------------------------------------------------------------

const (
	ActionScan          = "scan"
	ActionCheckout      = "checkout"
	ActionCheckin       = "checkin"
	ActionInventoryScan = "inventory_scan"
	ActionAdhocBooking  = "adhoc_booking"
)

// validActions is the set of all allowed scan actions.
var validActions = map[string]bool{
	ActionScan:          true,
	ActionCheckout:      true,
	ActionCheckin:       true,
	ActionInventoryScan: true,
	ActionAdhocBooking:  true,
}

// ValidAction checks whether the given string is a valid scan action.
func ValidAction(a string) bool {
	return validActions[a]
}

// ---------------------------------------------------------------------------
// Domain models
// ---------------------------------------------------------------------------

// ScanEvent represents a single scan event recorded by a scanner device.
type ScanEvent struct {
	ID              uuid.UUID  `json:"id"`
	TenantID        uuid.UUID  `json:"tenant_id"`
	UserID          uuid.UUID  `json:"user_id"`
	DeviceID        string     `json:"device_id"`
	Barcode         string     `json:"barcode,omitempty"`
	RFIDTag         string     `json:"rfid_tag,omitempty"`
	EquipmentID     *uuid.UUID `json:"equipment_id,omitempty"`
	Action          string     `json:"action"`
	ProjectID       *uuid.UUID `json:"project_id,omitempty"`
	LocationID      *uuid.UUID `json:"location_id,omitempty"`
	ConditionRating *int       `json:"condition_rating,omitempty"`
	ConditionNotes  string     `json:"condition_notes,omitempty"`
	GPSLat          *float64   `json:"gps_lat,omitempty"`
	GPSLng          *float64   `json:"gps_lng,omitempty"`
	Timestamp       time.Time  `json:"timestamp"`
	SyncedAt        time.Time  `json:"synced_at"`
	CreatedAt       time.Time  `json:"created_at"`
}

// ScannerDevice represents a registered scanner device.
type ScannerDevice struct {
	ID            uuid.UUID  `json:"id"`
	TenantID      uuid.UUID  `json:"tenant_id"`
	DeviceID      string     `json:"device_id"`
	DeviceName    string     `json:"device_name"`
	DeviceType    string     `json:"device_type"`
	FCMToken      string     `json:"fcm_token,omitempty"`
	RingRequested bool       `json:"ring_requested"`
	LastSeen      *time.Time `json:"last_seen,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}
