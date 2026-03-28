package domain

import "errors"

var (
	ErrNotificationNotFound = errors.New("notification not found")
	ErrTitleRequired        = errors.New("title is required")
	ErrTypeRequired         = errors.New("type is required")
)
