package errors

import "fmt"

// AppError represents an application-level error with context
type AppError struct {
	Code    string
	Message string
	Err     error
	Details map[string]interface{}
}

// New creates a new AppError
func New(code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// WithError wraps an underlying error
func (e *AppError) WithError(err error) *AppError {
	e.Err = err
	return e
}

// WithDetail adds a detail to the error
func (e *AppError) WithDetail(key string, value interface{}) *AppError {
	e.Details[key] = value
	return e
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Common error codes
const (
	CodeValidation      = "VALIDATION_ERROR"
	CodeNotFound        = "NOT_FOUND"
	CodeConflict        = "CONFLICT"
	CodeInternalServer  = "INTERNAL_SERVER_ERROR"
	CodeUnauthorized    = "UNAUTHORIZED"
	CodeForbidden       = "FORBIDDEN"
	CodeBadRequest      = "BAD_REQUEST"
	CodeDatabaseError   = "DATABASE_ERROR"
	CodeEventSourcingErr = "EVENT_SOURCING_ERROR"
)

// IsCode checks if an AppError has a specific code
func IsCode(err error, code string) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == code
	}
	return false
}
