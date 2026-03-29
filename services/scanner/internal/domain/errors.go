package domain

import "errors"

// Sentinel errors for the scanner domain.
var (
	ErrScanEventNotFound = errors.New("scan event not found")
	ErrDeviceNotFound    = errors.New("scanner device not found")
	ErrDuplicateDevice   = errors.New("device already registered for this tenant")
	ErrInvalidAction     = errors.New("invalid scan action")
	ErrMissingIdentifier = errors.New("barcode or rfid_tag is required")
	ErrMissingProjectID  = errors.New("project_id is required for checkout")
	ErrMissingAction     = errors.New("action is required")
	ErrMissingDeviceID   = errors.New("device_id is required")
	ErrMissingTimestamp  = errors.New("timestamp is required")
	ErrSessionNotFound   = errors.New("scan session not found")
	ErrSessionAlreadyEnd = errors.New("scan session already completed")
	ErrInvalidSessionType = errors.New("invalid session type")
)
