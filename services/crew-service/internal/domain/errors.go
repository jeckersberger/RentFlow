package domain

import "errors"

// Domain errors
var (
	ErrCrewMemberNotFound      = errors.New("crew member not found")
	ErrCrewMemberAlreadyExists = errors.New("crew member already exists")
	ErrQualificationNotFound   = errors.New("qualification not found")
	ErrAssignmentNotFound      = errors.New("assignment not found")
	ErrConflictDetected        = errors.New("assignment conflict detected")
	ErrQualificationExpired    = errors.New("qualification is expired")
	ErrInvalidStatus           = errors.New("invalid status")
	ErrInvalidRole             = errors.New("invalid role")
	ErrTenantIDRequired        = errors.New("tenant ID is required")
	ErrCrewMemberIDRequired    = errors.New("crew member ID is required")
	ErrTimeRecordNotFound      = errors.New("time record not found")
	ErrTimeRecordAlreadyRunning = errors.New("time record is already running")
	ErrInvalidDateRange        = errors.New("invalid date range")
	ErrEmptyEmail              = errors.New("email is required")
	ErrDuplicateEmail          = errors.New("email already exists for this tenant")
	ErrBookingNotFound         = errors.New("booking request not found")
	ErrBookingAlreadyResponded = errors.New("booking request has already been responded to")
)
