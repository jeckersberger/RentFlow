package domain

import "fmt"

var (
	ErrAuditEntryNotFound = fmt.Errorf("audit entry not found")
	ErrIntegrityViolation = fmt.Errorf("integrity violation detected")
	ErrTenantIDRequired   = fmt.Errorf("tenant ID is required")
	ErrInvalidInput       = fmt.Errorf("invalid input")
)
