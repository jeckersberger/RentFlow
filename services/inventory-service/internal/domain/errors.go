package domain

import "fmt"

// Domain errors
var (
	ErrEquipmentNotFound         = fmt.Errorf("equipment not found")
	ErrCategoryNotFound          = fmt.Errorf("category not found")
	ErrFlightcaseNotFound        = fmt.Errorf("flightcase not found")
	ErrBarcodeAlreadyExists      = fmt.Errorf("barcode already exists")
	ErrInvalidEquipmentStatus    = fmt.Errorf("invalid equipment status")
	ErrInvalidEquipmentCondition = fmt.Errorf("invalid equipment condition")
	ErrTenantIDRequired          = fmt.Errorf("tenant ID is required")
	ErrInvalidInput              = fmt.Errorf("invalid input")
	ErrUnauthorized              = fmt.Errorf("unauthorized")
)

type DomainError struct {
	Code    string
	Message string
	Err     error
}

func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func NewDomainError(code, message string, err error) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
