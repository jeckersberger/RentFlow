package domain

import "fmt"

var (
	ErrPeerNotFound         = fmt.Errorf("peer not found")
	ErrEquipmentNotFound    = fmt.Errorf("equipment not found")
	ErrShareRequestNotFound = fmt.Errorf("share request not found")
	ErrInvalidPeerStatus    = fmt.Errorf("invalid peer status")
	ErrTenantIDRequired     = fmt.Errorf("tenant ID is required")
	ErrInvalidInput         = fmt.Errorf("invalid input")
	ErrUnauthorized         = fmt.Errorf("unauthorized")
	ErrPeerBlocked          = fmt.Errorf("peer is blocked")
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
