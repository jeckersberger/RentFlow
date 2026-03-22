package application

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"strconv"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/notification-service/internal/domain"
)

// ChannelDriver interface for sending notifications
type ChannelDriver interface {
	Send(ctx context.Context, notification *domain.Notification, config map[string]interface{}) error
	Type() string
}

// SMTPDriver sends email notifications
type SMTPDriver struct {
	host     string
	port     int
	username string
	password string
	from     string
	log      logger.Logger
}

func NewSMTPDriver(host string, port int, username, password, from string, log logger.Logger) *SMTPDriver {
	return &SMTPDriver{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
		log:      log,
	}
}

func (d *SMTPDriver) Type() string {
	return "email"
}

func (d *SMTPDriver) Send(ctx context.Context, n *domain.Notification, config map[string]interface{}) error {
	// If SMTP host is empty, gracefully fallback
	if d.host == "" {
		d.log.Warn("SMTP not configured, skipping email notification", "notificationID", n.ID)
		return nil
	}

	// Extract recipient email from config
	recipient, ok := config["email"].(string)
	if !ok || recipient == "" {
		d.log.Warn("Email recipient not found in config", "notificationID", n.ID)
		return fmt.Errorf("email recipient not configured")
	}

	// Build email message
	subject := n.Title
	body := n.Body

	message := fmt.Sprintf(
		"To: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s",
		recipient,
		subject,
		body,
	)

	// Send email using SMTP
	addr := d.host + ":" + strconv.Itoa(d.port)
	auth := smtp.PlainAuth("", d.username, d.password, d.host)

	err := smtp.SendMail(addr, auth, d.from, []string{recipient}, []byte(message))
	if err != nil {
		d.log.Error("Failed to send email", "error", err, "notificationID", n.ID, "recipient", recipient)
		return fmt.Errorf("failed to send email: %w", err)
	}

	d.log.Info("Email sent successfully", "notificationID", n.ID, "recipient", recipient)
	return nil
}

// VAPIDPushDriver sends Web Push notifications
type VAPIDPushDriver struct {
	vapidPublicKey  string
	vapidPrivateKey string
	log             logger.Logger
}

func NewVAPIDPushDriver(publicKey, privateKey string, log logger.Logger) *VAPIDPushDriver {
	return &VAPIDPushDriver{
		vapidPublicKey:  publicKey,
		vapidPrivateKey: privateKey,
		log:             log,
	}
}

func (d *VAPIDPushDriver) Type() string {
	return "webpush"
}

func (d *VAPIDPushDriver) Send(ctx context.Context, n *domain.Notification, config map[string]interface{}) error {
	// If VAPID keys are not configured, gracefully fallback
	if d.vapidPublicKey == "" || d.vapidPrivateKey == "" {
		d.log.Warn("VAPID keys not configured, skipping web push", "notificationID", n.ID)
		return nil
	}

	// Extract push subscription details from config
	endpoint, ok := config["endpoint"].(string)
	if !ok || endpoint == "" {
		d.log.Warn("Push endpoint not found in config", "notificationID", n.ID)
		return fmt.Errorf("push endpoint not configured")
	}

	p256dh, ok := config["p256dh"].(string)
	if !ok || p256dh == "" {
		d.log.Warn("p256dh not found in config", "notificationID", n.ID)
		return fmt.Errorf("p256dh not configured")
	}

	auth, ok := config["auth"].(string)
	if !ok || auth == "" {
		d.log.Warn("auth not found in config", "notificationID", n.ID)
		return fmt.Errorf("auth not configured")
	}

	// Build push payload
	payload := map[string]interface{}{
		"title": n.Title,
		"body":  n.Body,
		"data":  n.Data,
		"badge": "/assets/badge.png",
		"icon":  "/assets/icon.png",
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		d.log.Error("Failed to marshal push payload", "error", err, "notificationID", n.ID)
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HTTP POST request to push service
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(payloadJSON))
	if err != nil {
		d.log.Error("Failed to create push request", "error", err, "notificationID", n.ID)
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers for web push
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "aes128gcm")

	// In a production implementation, this would include proper VAPID JWT encoding
	// For now, add basic auth headers from subscription
	req.Header.Set("Crypto-Key", "p256ecdsa="+d.vapidPublicKey)

	// Make the request
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		d.log.Error("Failed to send web push", "error", err, "notificationID", n.ID, "endpoint", endpoint)
		return fmt.Errorf("failed to send push: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		d.log.Error("Web push returned error status", "status", resp.StatusCode, "notificationID", n.ID, "response", string(respBody))
		return fmt.Errorf("push service returned status %d", resp.StatusCode)
	}

	d.log.Info("Web push sent successfully", "notificationID", n.ID, "endpoint", truncateEndpoint(endpoint))
	return nil
}

// InAppDriver for in-app notifications
type InAppDriver struct {
	log logger.Logger
}

func NewInAppDriver(log logger.Logger) *InAppDriver {
	return &InAppDriver{
		log: log,
	}
}

func (d *InAppDriver) Type() string {
	return "in_app"
}

func (d *InAppDriver) Send(ctx context.Context, n *domain.Notification, config map[string]interface{}) error {
	// In production, this would:
	// 1. Publish to Redis channel for WebSocket delivery
	// 2. Store in a message queue
	// 3. Deliver via WebSocket to connected clients

	// For now, log the in-app notification as delivered
	d.log.Info("In-app notification marked as delivered",
		"notificationID", n.ID,
		"userID", n.UserID,
		"title", n.Title,
	)

	return nil
}

// truncateEndpoint returns a shortened endpoint URL for logging
func truncateEndpoint(endpoint string) string {
	if len(endpoint) > 50 {
		return endpoint[:50] + "..."
	}
	return endpoint
}
