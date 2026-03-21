package domain

import "fmt"

var (
	ErrVehicleNotFound   = fmt.Errorf("vehicle not found")
	ErrTourNotFound      = fmt.Errorf("tour not found")
	ErrInvalidStatus     = fmt.Errorf("invalid status")
	ErrTenantIDRequired  = fmt.Errorf("tenant ID is required")
	ErrInvalidInput      = fmt.Errorf("invalid input")
	ErrUnauthorized      = fmt.Errorf("unauthorized")
)
