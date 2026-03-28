package domain

import "errors"

// Sentinel errors for the crew domain.
var (
	ErrCrewMemberNotFound    = errors.New("crew member not found")
	ErrAssignmentNotFound    = errors.New("crew assignment not found")
	ErrQualificationNotFound = errors.New("crew qualification not found")
	ErrFirstNameRequired     = errors.New("first_name and last_name are required")
	ErrQualificationName     = errors.New("name is required")
)
