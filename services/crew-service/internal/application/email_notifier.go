package application

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
)

// sendEmailAsync fires an HTTP POST to the notification-service in a goroutine.
// It is fire-and-forget: failures are logged but never block the caller.
func sendEmailAsync(serviceURL, endpoint string, body interface{}, log logger.Logger) {
	go func() {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			log.Error("failed to marshal email notification body", err)
			return
		}
		resp, err := http.Post(serviceURL+endpoint, "application/json", bytes.NewReader(jsonBody))
		if err != nil {
			log.Error("failed to send email notification", err, "endpoint", endpoint)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			log.Warn("email notification returned error", "endpoint", endpoint, "status", resp.StatusCode)
		}
	}()
}
