package domain

import "errors"

// Sentinel errors for the crew domain.
var (
	ErrCrewMemberNotFound    = errors.New("crew member not found")
	ErrAssignmentNotFound    = errors.New("crew assignment not found")
	ErrQualificationNotFound = errors.New("crew qualification not found")
	ErrFirstNameRequired     = errors.New("first_name and last_name are required")
	ErrQualificationName     = errors.New("name is required")

	// Time tracking errors.
	ErrTimeEntryNotFound    = errors.New("time entry not found")
	ErrAlreadyCheckedIn     = errors.New("crew member already has an active check-in")
	ErrNoActiveCheckIn      = errors.New("crew member has no active check-in")
	ErrCrewMemberIDRequired = errors.New("crew_member_id is required")

	// Availability errors.
	ErrAvailabilityNotFound = errors.New("availability entry not found")
	ErrInvalidDateFormat    = errors.New("invalid date format, expected YYYY-MM-DD")
	ErrInvalidAvailStatus   = errors.New("invalid status, expected: available, unavailable, or on_request")

	// Availability block errors.
	ErrAvailabilityBlockNotFound = errors.New("availability block not found")
	ErrInvalidBlockType          = errors.New("invalid block_type, expected: vacation, sick, training, blocked, or other")
	ErrEndBeforeStart            = errors.New("end_date must not be before start_date")
	ErrBlockConflict             = errors.New("crew member already has an overlapping availability block in this period")

	// Skill-matching errors.
	ErrNoSkillsRequested = errors.New("required_skills must not be empty")
	ErrInvalidSkillCount = errors.New("each required skill must have count >= 1")
	ErrDateFromRequired  = errors.New("date_from is required (YYYY-MM-DD)")
	ErrDateToRequired    = errors.New("date_to is required (YYYY-MM-DD)")
)
