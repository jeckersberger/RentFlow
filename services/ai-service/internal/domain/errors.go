package domain

import "errors"

// Domain errors
var (
	ErrAIRequestNotFound     = errors.New("ai request not found")
	ErrAIFeedbackNotFound    = errors.New("ai feedback not found")
	ErrFewShotExampleNotFound = errors.New("few-shot example not found")
	ErrAIProviderNotFound    = errors.New("ai provider not found")
	ErrProviderNotAvailable  = errors.New("ai provider not available")
	ErrInvalidRequestType    = errors.New("invalid request type")
	ErrInvalidStatus         = errors.New("invalid status")
	ErrTenantIDRequired      = errors.New("tenant ID is required")
	ErrInputTextRequired     = errors.New("input text is required")
	ErrProviderIDRequired    = errors.New("provider ID is required")
	ErrRequestIDRequired     = errors.New("request ID is required")
	ErrInvalidRating         = errors.New("rating must be between 1 and 5")
	ErrNoProviders           = errors.New("no active providers configured")
	ErrAnonymizationFailed   = errors.New("anonymization failed")
	ErrAIProviderError       = errors.New("ai provider returned error")
)
