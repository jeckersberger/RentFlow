package domain

import "errors"

var (
	ErrWorkflowNotFound           = errors.New("workflow definition not found")
	ErrWorkflowInstanceNotFound   = errors.New("workflow instance not found")
	ErrWorkflowStepNotFound       = errors.New("workflow step not found")
	ErrInvalidTriggerType         = errors.New("invalid trigger type")
	ErrInvalidActionType          = errors.New("invalid action type")
	ErrInvalidStatus              = errors.New("invalid workflow status")
	ErrWorkflowAlreadyCompleted   = errors.New("workflow instance already completed")
	ErrNoStepsInWorkflow          = errors.New("workflow has no steps defined")
	ErrInvalidStepConfiguration   = errors.New("invalid step configuration")
	ErrActionExecutionFailed      = errors.New("action execution failed")
	ErrWorkflowExecutionFailed    = errors.New("workflow execution failed")
	ErrUnknownActionType          = errors.New("unknown action type")
	ErrTemplateNotFound           = errors.New("template not found")
)
