package domain

import "fmt"

var (
	ErrReportNotFound   = fmt.Errorf("report not found")
	ErrTenantIDRequired = fmt.Errorf("tenant ID is required")
	ErrInvalidInput     = fmt.Errorf("invalid input")
)
