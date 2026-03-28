package domain

import "errors"

var (
	ErrPredictionNotFound = errors.New("ai prediction not found")
	ErrSuggestionNotFound = errors.New("ai suggestion not found")
	ErrTypeRequired       = errors.New("type is required")
	ErrTitleRequired      = errors.New("title is required")
)
