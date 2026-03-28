package domain

import "errors"

// Sentinel errors for the warehouse domain.
var (
	ErrWarehouseNotFound   = errors.New("warehouse not found")
	ErrZoneNotFound        = errors.New("zone not found")
	ErrRackNotFound        = errors.New("rack not found")
	ErrLocationNotFound    = errors.New("location not found")
	ErrMovementNotFound    = errors.New("movement not found")
	ErrCheckNotFound       = errors.New("inventory check not found")
	ErrDuplicateCode       = errors.New("duplicate location code")
	ErrCheckAlreadyDone    = errors.New("inventory check already completed")
	ErrCheckItemNotFound   = errors.New("inventory check item not found")
)
