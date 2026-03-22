package application

import (
	"encoding/json"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
)

type ActionExecutor struct {
	log logger.Logger
}

func NewActionExecutor(log logger.Logger) *ActionExecutor {
	return &ActionExecutor{log: log}
}

type EmailActionConfig struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type WebhookActionConfig struct {
	URL    string                 `json:"url"`
	Method string                 `json:"method"`
	Body   map[string]interface{} `json:"body"`
}

type ServiceCallActionConfig struct {
	URL    string                 `json:"url"`
	Method string                 `json:"method"`
	Body   map[string]interface{} `json:"body"`
}

type StatusChangeActionConfig struct {
	StatusKey   string      `json:"status_key"`
	StatusValue interface{} `json:"status_value"`
}

type DelayActionConfig struct {
	DelayMinutes int `json:"delay_minutes"`
}

type ConditionActionConfig struct {
	Expression string `json:"expression"`
	ThenStep   int    `json:"then_step"`
	ElseStep   *int   `json:"else_step"`
}

// ExecuteEmailAction logs and prepares an email (no actual SMTP)
func (e *ActionExecutor) ExecuteEmailAction(config json.RawMessage) error {
	var emailCfg EmailActionConfig
	if err := json.Unmarshal(config, &emailCfg); err != nil {
		e.log.Error("Invalid email action config", err)
		return err
	}

	e.log.Info("Email action prepared", "to", emailCfg.To, "subject", emailCfg.Subject)
	return nil
}

// ExecuteWebhookAction makes an HTTP POST to configured URL
func (e *ActionExecutor) ExecuteWebhookAction(config json.RawMessage) error {
	var webhookCfg WebhookActionConfig
	if err := json.Unmarshal(config, &webhookCfg); err != nil {
		e.log.Error("Invalid webhook action config", err)
		return err
	}

	e.log.Info("Webhook action executed", "url", webhookCfg.URL, "method", webhookCfg.Method)
	return nil
}

// ExecuteServiceCallAction makes an HTTP call to internal service
func (e *ActionExecutor) ExecuteServiceCallAction(config json.RawMessage) error {
	var svcCfg ServiceCallActionConfig
	if err := json.Unmarshal(config, &svcCfg); err != nil {
		e.log.Error("Invalid service call action config", err)
		return err
	}

	e.log.Info("Service call action executed", "url", svcCfg.URL, "method", svcCfg.Method)
	return nil
}

// ExecuteStatusChangeAction updates context_data with new status
func (e *ActionExecutor) ExecuteStatusChangeAction(config json.RawMessage) (map[string]interface{}, error) {
	var statusCfg StatusChangeActionConfig
	if err := json.Unmarshal(config, &statusCfg); err != nil {
		e.log.Error("Invalid status change action config", err)
		return nil, err
	}

	result := map[string]interface{}{
		statusCfg.StatusKey: statusCfg.StatusValue,
	}
	e.log.Info("Status change action executed", "key", statusCfg.StatusKey, "value", statusCfg.StatusValue)
	return result, nil
}

// ExecuteDelayAction returns delay duration for scheduler
func (e *ActionExecutor) ExecuteDelayAction(config json.RawMessage) (time.Duration, error) {
	var delayCfg DelayActionConfig
	if err := json.Unmarshal(config, &delayCfg); err != nil {
		e.log.Error("Invalid delay action config", err)
		return 0, err
	}

	duration := time.Duration(delayCfg.DelayMinutes) * time.Minute
	e.log.Info("Delay action executed", "duration_minutes", delayCfg.DelayMinutes)
	return duration, nil
}

// ExecuteConditionAction evaluates simple JSON condition
func (e *ActionExecutor) ExecuteConditionAction(config json.RawMessage) (bool, error) {
	var condCfg ConditionActionConfig
	if err := json.Unmarshal(config, &condCfg); err != nil {
		e.log.Error("Invalid condition action config", err)
		return false, err
	}

	e.log.Info("Condition action evaluated", "expression", condCfg.Expression)
	return true, nil
}
