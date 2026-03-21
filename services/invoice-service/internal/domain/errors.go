package domain

import "fmt"

// Domain errors
var (
	ErrInvoiceNotFound        = fmt.Errorf("invoice not found")
	ErrQuoteNotFound          = fmt.Errorf("quote not found")
	ErrDunningNotFound        = fmt.Errorf("dunning entry not found")
	ErrTenantIDRequired       = fmt.Errorf("tenant ID is required")
	ErrInvalidInput           = fmt.Errorf("invalid input")
	ErrUnauthorized           = fmt.Errorf("unauthorized")
	ErrInvalidStatus          = fmt.Errorf("invalid status")
	ErrInvalidTaxRate         = fmt.Errorf("invalid tax rate")
	ErrNoGapAllowed           = fmt.Errorf("invoice numbering has gaps")
	ErrCannotModifyFinalized  = fmt.Errorf("cannot modify finalized invoice")
	ErrInvalidTransition      = fmt.Errorf("invalid status transition")
	ErrZeroAmount             = fmt.Errorf("invoice total must be greater than zero")
	ErrNoItems                = fmt.Errorf("invoice must have at least one item")
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
