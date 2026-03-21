package domain

import "errors"

// Domain errors
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserLocked         = errors.New("user account is locked")
	ErrEmailExists        = errors.New("email already exists")
	ErrWeakPassword       = errors.New("password does not meet requirements")
	ErrTenantNotFound     = errors.New("tenant not found")
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrInvalidTenantID    = errors.New("invalid tenant ID")
)
