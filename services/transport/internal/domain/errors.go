package domain

import "errors"

// Sentinel errors for the transport domain.
var (
	ErrVehicleNotFound       = errors.New("vehicle not found")
	ErrOrderNotFound         = errors.New("transport order not found")
	ErrItemNotFound          = errors.New("transport item not found")
	ErrInvalidOrderType      = errors.New("invalid order type")
	ErrInvalidStatus         = errors.New("invalid order status")
	ErrMissingVehicleName    = errors.New("vehicle name is required")
	ErrMissingEquipmentID    = errors.New("equipment_id is required")
	ErrOrderAlreadyCompleted = errors.New("order is already completed")
)
