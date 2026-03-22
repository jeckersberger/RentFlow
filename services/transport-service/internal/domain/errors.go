package domain

import "fmt"

var (
	ErrVehicleNotFound      = fmt.Errorf("vehicle not found")
	ErrTourNotFound         = fmt.Errorf("tour not found")
	ErrTourEquipmentNotFound = fmt.Errorf("tour equipment not found")
	ErrDriverLogNotFound    = fmt.Errorf("driver log not found")
	ErrInvalidStatus        = fmt.Errorf("invalid status")
	ErrTenantIDRequired     = fmt.Errorf("tenant ID is required")
	ErrInvalidInput         = fmt.Errorf("invalid input")
	ErrUnauthorized         = fmt.Errorf("unauthorized")
	ErrCapacityExceeded     = fmt.Errorf("vehicle capacity exceeded")
	ErrEquipmentNotFound    = fmt.Errorf("equipment not found")
)
