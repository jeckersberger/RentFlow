package domain

import "errors"

var (
	ErrScheduleNotFound = errors.New("maintenance schedule not found")
	ErrTaskNotFound     = errors.New("maintenance task not found")
	ErrNameRequired     = errors.New("name is required")
	ErrTitleRequired    = errors.New("title is required")
	ErrEquipmentRequired = errors.New("equipment_id is required")
	ErrTaskAlreadyCompleted = errors.New("task is already completed")
)
