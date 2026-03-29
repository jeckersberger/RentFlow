package domain

import "errors"

var (
	ErrProjectNotFound          = errors.New("project not found")
	ErrInvalidProjectStatus     = errors.New("invalid project status")
	ErrInvalidStatusTransition  = errors.New("invalid status transition")
	ErrInvalidPEStatus          = errors.New("invalid project equipment status")
	ErrPacklistNotFound         = errors.New("packlist not found")
	ErrReservationNotFound      = errors.New("reservation not found")
	ErrReservationConflict      = errors.New("reservation conflict")
	ErrProjectHasEquipment      = errors.New("project has assigned equipment")
)
