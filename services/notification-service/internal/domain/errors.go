package domain

import "fmt"

var (
	ErrNotificationNotFound = fmt.Errorf("notification not found")
	ErrPreferenceNotFound   = fmt.Errorf("preference not found")
	ErrTenantIDRequired     = fmt.Errorf("tenant ID is required")
	ErrInvalidInput         = fmt.Errorf("invalid input")
	ErrUnauthorized         = fmt.Errorf("unauthorized")
)
