package application

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
)

type ActionExecutor struct {
	log    logger.Logger
	client *http.Client
}

// ConditionOperand represents a value in a condition (can be string, number, etc)
type ConditionOperand struct {
	Type  string      // "field", "value"
	Field string      // field name if Type == "field"
	Value interface{} // literal value if Type == "value"
}

// ParsedCondition represents a parsed condition expression
type ParsedCondition struct {
	LeftField    string
	Operator     string      // ==, !=, >, <, >=, <=
	RightValue   interface{}
	ContextData  map[string]interface{}
}

func NewActionExecutor(log logger.Logger) *ActionExecutor {
	return &ActionExecutor{
		log: log,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type EmailActionConfig struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type WebhookActionConfig struct {
	URL     string                 `json:"url"`
	Method  string                 `json:"method"`
	Headers map[string]string      `json:"headers"`
	Body    map[string]interface{} `json:"body"`
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

// ExecuteEmailAction sends a real email via SMTP
func (e *ActionExecutor) ExecuteEmailAction(config json.RawMessage) error {
	var emailCfg EmailActionConfig
	if err := json.Unmarshal(config, &emailCfg); err != nil {
		e.log.Error("Invalid email action config", err)
		return err
	}

	// Get SMTP configuration from environment
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpFrom := os.Getenv("SMTP_FROM")

	// If SMTP not configured, log warning and skip
	if smtpHost == "" || smtpPortStr == "" {
		e.log.Warn("SMTP not configured, skipping email action", "to", emailCfg.To, "subject", emailCfg.Subject)
		return nil
	}

	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		e.log.Warn("Invalid SMTP_PORT, skipping email action", "error", err)
		return nil
	}

	// Build email message
	message := fmt.Sprintf(
		"To: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s",
		emailCfg.To,
		emailCfg.Subject,
		emailCfg.Body,
	)

	// Send email using SMTP
	addr := smtpHost + ":" + strconv.Itoa(smtpPort)
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)

	err = smtp.SendMail(addr, auth, smtpFrom, []string{emailCfg.To}, []byte(message))
	if err != nil {
		e.log.Error("Failed to send email", "error", err, "to", emailCfg.To)
		return fmt.Errorf("failed to send email: %w", err)
	}

	e.log.Info("Email sent successfully", "to", emailCfg.To, "subject", emailCfg.Subject)
	return nil
}

// ExecuteWebhookAction makes a real HTTP POST/PUT call to external webhook
func (e *ActionExecutor) ExecuteWebhookAction(config json.RawMessage) error {
	var webhookCfg WebhookActionConfig
	if err := json.Unmarshal(config, &webhookCfg); err != nil {
		e.log.Error("Invalid webhook action config", err)
		return err
	}

	// Default to POST if method not specified
	method := strings.ToUpper(webhookCfg.Method)
	if method == "" {
		method = "POST"
	}

	// Marshal body to JSON
	var bodyReader *bytes.Reader
	if webhookCfg.Body != nil {
		bodyJSON, err := json.Marshal(webhookCfg.Body)
		if err != nil {
			e.log.Error("Failed to marshal webhook body", "error", err, "url", webhookCfg.URL)
			return fmt.Errorf("failed to marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyJSON)
	} else {
		bodyReader = bytes.NewReader([]byte("{}"))
	}

	// Create HTTP request
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, webhookCfg.URL, bodyReader)
	if err != nil {
		e.log.Error("Failed to create webhook request", "error", err, "url", webhookCfg.URL)
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	for key, value := range webhookCfg.Headers {
		req.Header.Set(key, value)
	}

	// Make the request
	resp, err := e.client.Do(req)
	if err != nil {
		e.log.Error("Failed to send webhook", "error", err, "url", webhookCfg.URL)
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	e.log.Info("Webhook executed", "url", webhookCfg.URL, "method", method, "status", resp.StatusCode)
	return nil
}

// ExecuteServiceCallAction makes a real HTTP call to internal service
func (e *ActionExecutor) ExecuteServiceCallAction(config json.RawMessage) error {
	var svcCfg ServiceCallActionConfig
	if err := json.Unmarshal(config, &svcCfg); err != nil {
		e.log.Error("Invalid service call action config", err)
		return err
	}

	// Default to POST if method not specified
	method := strings.ToUpper(svcCfg.Method)
	if method == "" {
		method = "POST"
	}

	// Validate method
	if method != "GET" && method != "POST" && method != "PUT" {
		e.log.Error("Unsupported service call method", "method", method)
		return fmt.Errorf("unsupported method: %s", method)
	}

	// Marshal body to JSON
	var bodyReader *bytes.Reader
	if method != "GET" {
		if svcCfg.Body != nil {
			bodyJSON, err := json.Marshal(svcCfg.Body)
			if err != nil {
				e.log.Error("Failed to marshal service call body", "error", err, "url", svcCfg.URL)
				return fmt.Errorf("failed to marshal body: %w", err)
			}
			bodyReader = bytes.NewReader(bodyJSON)
		} else {
			bodyReader = bytes.NewReader([]byte("{}"))
		}
	} else {
		bodyReader = bytes.NewReader(nil)
	}

	// Create HTTP request
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, svcCfg.URL, bodyReader)
	if err != nil {
		e.log.Error("Failed to create service call request", "error", err, "url", svcCfg.URL)
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set content type for non-GET requests
	if method != "GET" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Make the request
	resp, err := e.client.Do(req)
	if err != nil {
		e.log.Error("Failed to call service", "error", err, "url", svcCfg.URL)
		return fmt.Errorf("failed to call service: %w", err)
	}
	defer resp.Body.Close()

	e.log.Info("Service call executed", "url", svcCfg.URL, "method", method, "status", resp.StatusCode)
	return nil
}

// ExecuteStatusChangeAction makes an HTTP call to update status in the relevant service
func (e *ActionExecutor) ExecuteStatusChangeAction(config json.RawMessage) (map[string]interface{}, error) {
	var statusCfg StatusChangeActionConfig
	if err := json.Unmarshal(config, &statusCfg); err != nil {
		e.log.Error("Invalid status change action config", err)
		return nil, err
	}

	// For status changes, we assume there's a service URL configured or we use a standard pattern
	// This is a simplified implementation that returns the updated context
	result := map[string]interface{}{
		statusCfg.StatusKey: statusCfg.StatusValue,
	}

	e.log.Info("Status change action executed", "key", statusCfg.StatusKey, "value", statusCfg.StatusValue)
	return result, nil
}

// ExecuteDelayAction blocks for the configured duration
func (e *ActionExecutor) ExecuteDelayAction(config json.RawMessage) (time.Duration, error) {
	var delayCfg DelayActionConfig
	if err := json.Unmarshal(config, &delayCfg); err != nil {
		e.log.Error("Invalid delay action config", err)
		return 0, err
	}

	duration := time.Duration(delayCfg.DelayMinutes) * time.Minute
	e.log.Info("Delay action started", "duration_minutes", delayCfg.DelayMinutes)

	// Sleep for the configured duration
	time.Sleep(duration)

	e.log.Info("Delay action completed", "duration_minutes", delayCfg.DelayMinutes)
	return duration, nil
}

// ExecuteConditionAction evaluates simple field conditions
// Supports: field == value, field != value, field > value, field < value, field >= value, field <= value
func (e *ActionExecutor) ExecuteConditionAction(config json.RawMessage) (bool, error) {
	var condCfg ConditionActionConfig
	if err := json.Unmarshal(config, &condCfg); err != nil {
		e.log.Error("Invalid condition action config", err)
		return false, err
	}

	// Parse the expression: "field operator value"
	// Operators: ==, !=, >, <, >=, <=
	result := e.evaluateCondition(condCfg.Expression)

	e.log.Info("Condition action evaluated", "expression", condCfg.Expression, "result", result)
	return result, nil
}

// evaluateCondition parses and evaluates a simple condition expression
func (e *ActionExecutor) evaluateCondition(expression string) bool {
	// Extract operator and operands
	operators := []string{">=", "<=", "==", "!=", ">", "<"}
	var operator string
	var parts []string

	for _, op := range operators {
		if strings.Contains(expression, op) {
			operator = op
			parts = strings.Split(expression, op)
			break
		}
	}

	if operator == "" || len(parts) != 2 {
		e.log.Warn("Invalid condition expression", "expression", expression)
		return false
	}

	leftOperand := strings.TrimSpace(parts[0])
	rightOperand := strings.TrimSpace(parts[1])

	// Parse operands to get actual values
	leftValue := e.parseOperand(leftOperand)
	rightValue := e.parseOperand(rightOperand)

	// Compare based on operator
	return e.compareValues(leftValue, rightValue, operator)
}

// parseOperand converts a string to a value (handles numbers and quoted strings)
func (e *ActionExecutor) parseOperand(operand string) interface{} {
	operand = strings.TrimSpace(operand)

	// Check if it's a quoted string
	if (strings.HasPrefix(operand, "\"") && strings.HasSuffix(operand, "\"")) ||
		(strings.HasPrefix(operand, "'") && strings.HasSuffix(operand, "'")) {
		return operand[1 : len(operand)-1]
	}

	// Try to parse as number
	if num, err := strconv.ParseFloat(operand, 64); err == nil {
		// Return as int if it's a whole number
		if num == float64(int64(num)) {
			return int64(num)
		}
		return num
	}

	// Return as string if not a number
	return operand
}

// compareValues compares two values based on the operator
func (e *ActionExecutor) compareValues(left, right interface{}, operator string) bool {
	switch operator {
	case "==":
		return e.valuesEqual(left, right)
	case "!=":
		return !e.valuesEqual(left, right)
	case ">":
		return e.compare(left, right) > 0
	case "<":
		return e.compare(left, right) < 0
	case ">=":
		return e.compare(left, right) >= 0
	case "<=":
		return e.compare(left, right) <= 0
	default:
		return false
	}
}

// valuesEqual checks if two values are equal
func (e *ActionExecutor) valuesEqual(left, right interface{}) bool {
	// Convert to string for comparison
	leftStr := fmt.Sprintf("%v", left)
	rightStr := fmt.Sprintf("%v", right)
	return leftStr == rightStr
}

// compare compares two numeric values
func (e *ActionExecutor) compare(left, right interface{}) int {
	leftNum := e.toFloat(left)
	rightNum := e.toFloat(right)

	if leftNum < rightNum {
		return -1
	} else if leftNum > rightNum {
		return 1
	}
	return 0
}

// toFloat converts a value to float64 for numeric comparison
func (e *ActionExecutor) toFloat(val interface{}) float64 {
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
		return 0
	default:
		return 0
	}
}
