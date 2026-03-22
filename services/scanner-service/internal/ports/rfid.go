package ports

import (
	"context"
)

// RFIDScanner defines the interface for RFID scanner operations (e.g., Zebra FX9600)
type RFIDScanner interface {
	// Connect establishes connection to the RFID device
	Connect(ctx context.Context, deviceID string) error

	// Disconnect closes the connection to the RFID device
	Disconnect(ctx context.Context) error

	// StartSession initiates an RFID scan session with the specified parameters
	StartSession(ctx context.Context, sessionID string, power int) error

	// StopSession terminates the current RFID scan session
	StopSession(ctx context.Context) error

	// GetScans retrieves pending scan results from the RFID device
	GetScans(ctx context.Context) ([]RFIDScan, error)

	// SetPower adjusts the RF power output (0-300)
	SetPower(ctx context.Context, power int) error

	// GetStatus returns the current device status
	GetStatus(ctx context.Context) (RFIDStatus, error)
}

type RFIDScan struct {
	EPC       string // Electronic Product Code (barcode equivalent)
	RSSI      int    // Signal strength
	Timestamp int64  // Unix timestamp in milliseconds
	Antenna   int    // Which antenna received the scan (1-4)
}

type RFIDStatus struct {
	Connected      bool
	Active         bool
	Power          int
	AntennaCount   int
	FirmwareVersion string
}

// WebHIDScanner defines the interface for USB scanner via WebHID (browser-side)
// This is primarily frontend-facing, but backend provides API endpoint for events
type WebHIDScanner interface {
	// RegisterDevice associates a WebHID device with a session
	RegisterDevice(ctx context.Context, deviceID string, serialNumber string) error

	// ProcessScanEvent handles a scan event received from WebHID client
	ProcessScanEvent(ctx context.Context, deviceID string, barcode string) error
}
