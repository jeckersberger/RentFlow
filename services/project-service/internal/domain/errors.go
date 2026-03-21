package domain

import "fmt"

// Domain errors
var (
	ErrProjectNotFound        = fmt.Errorf("project not found")
	ErrPacklistNotFound       = fmt.Errorf("packlist not found")
	ErrReservationNotFound    = fmt.Errorf("reservation not found")
	ErrInvalidProjectStatus   = fmt.Errorf("invalid project status")
	ErrInvalidPacklistStatus  = fmt.Errorf("invalid packlist status")
	ErrTenantIDRequired       = fmt.Errorf("tenant ID is required")
	ErrInvalidInput           = fmt.Errorf("invalid input")
	ErrUnauthorized           = fmt.Errorf("unauthorized")
	ErrReservationConflict    = fmt.Errorf("reservation conflicts with existing reservations")
	ErrInvalidStatusTransition = fmt.Errorf("invalid status transition")
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
