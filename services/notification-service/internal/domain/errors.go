package domain

import "errors"

var (
	ErrNotificationNotFound = errors.New("notification not found")
	ErrChannelNotFound      = errors.New("channel not found")
	ErrPreferenceNotFound   = errors.New("preference not found")
	ErrInvalidInput         = errors.New("invalid input")
	ErrDatabaseError        = errors.New("database error")
)
