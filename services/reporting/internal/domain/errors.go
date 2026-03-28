package domain

import "errors"

var (
	ErrDefinitionNotFound = errors.New("report definition not found")
	ErrSnapshotNotFound   = errors.New("report snapshot not found")
	ErrWidgetNotFound     = errors.New("dashboard widget not found")
	ErrNameRequired       = errors.New("name is required")
	ErrTitleRequired      = errors.New("title is required")
)
