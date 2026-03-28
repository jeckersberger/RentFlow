package domain

import "errors"

// Sentinel errors for the inventory domain.
var (
	ErrEquipmentNotFound     = errors.New("equipment not found")
	ErrCategoryNotFound      = errors.New("category not found")
	ErrEquipmentTypeNotFound = errors.New("equipment type not found")
	ErrFlightcaseNotFound    = errors.New("flightcase not found")
	ErrDuplicateBarcode      = errors.New("duplicate barcode")
	ErrDuplicateRFID         = errors.New("duplicate rfid tag")
	ErrInvalidStatus         = errors.New("invalid equipment status")
	ErrInvalidCondition      = errors.New("invalid equipment condition")
	ErrCategoryHasChildren   = errors.New("category has child categories")
	ErrCategoryHasEquipment  = errors.New("category has associated equipment")
	ErrEquipmentNotAvailable = errors.New("equipment is not available")
)
