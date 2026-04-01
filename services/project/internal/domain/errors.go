package domain

import "errors"

var (
	ErrProjectNotFound          = errors.New("project not found")
	ErrInvalidProjectStatus     = errors.New("invalid project status")
	ErrInvalidStatusTransition  = errors.New("invalid status transition")
	ErrInvalidPEStatus          = errors.New("invalid project equipment status")
	ErrPacklistNotFound             = errors.New("packlist not found")
	ErrPacklistItemNotFound         = errors.New("packlist item not found")
	ErrInvalidPacklistItemStatus    = errors.New("invalid packlist item status")
	ErrInvalidItemStatusTransition  = errors.New("invalid item status transition")
	ErrReservationNotFound          = errors.New("reservation not found")
	ErrReservationConflict      = errors.New("reservation conflict")
	ErrProjectHasEquipment      = errors.New("project has assigned equipment")
)
