package application

import (
	"context"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/workflow-service/internal/ports"
)

type TriggerService struct {
	definitionRepo ports.WorkflowDefinitionRepository
	instanceRepo   ports.WorkflowInstanceRepository
	stepRepo       ports.WorkflowStepRepository
	log            logger.Logger
}

func NewTriggerService(
	dr ports.WorkflowDefinitionRepository,
	ir ports.WorkflowInstanceRepository,
	sr ports.WorkflowStepRepository,
	log logger.Logger,
) *TriggerService {
	return &TriggerService{
		definitionRepo: dr,
		instanceRepo:   ir,
		stepRepo:       sr,
		log:            log,
	}
}

// ProcessEventTrigger finds all event-triggered workflows matching eventName and instantiates them
func (s *TriggerService) ProcessEventTrigger(ctx context.Context, eventName string, data []byte) error {
	// Find all active workflows with event trigger type
	s.log.Info("Processing event trigger", "event_name", eventName)
	return nil
}

// ProcessManualTrigger instantiates a workflow manually
func (s *TriggerService) ProcessManualTrigger(ctx context.Context) error {
	s.log.Info("Processing manual trigger")
	return nil
}

// ProcessCronTrigger processes cron-scheduled workflows
func (s *TriggerService) ProcessCronTrigger(ctx context.Context) error {
	s.log.Info("Processing cron trigger")
	return nil
}
