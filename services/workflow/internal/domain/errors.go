package domain

import "errors"

// Sentinel errors for the workflow domain.
var (
	ErrDefinitionNotFound = errors.New("workflow definition not found")
	ErrInstanceNotFound   = errors.New("workflow instance not found")
	ErrInvalidAction      = errors.New("invalid workflow action")
	ErrInvalidType        = errors.New("invalid workflow type")
	ErrInvalidStatus      = errors.New("invalid workflow status")
	ErrMissingName        = errors.New("name is required")
	ErrMissingDefinition  = errors.New("definition_id is required")
	ErrInstanceCompleted  = errors.New("workflow instance already completed")
)
