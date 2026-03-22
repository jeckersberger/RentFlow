package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
)

// EquipmentHTTPClient calls the equipment-service via HTTP
type EquipmentHTTPClient struct {
	baseURL string
	client  *http.Client
	logger  logger.Logger
}

// NewEquipmentHTTPClient creates a new equipment HTTP client
func NewEquipmentHTTPClient(baseURL string, client *http.Client, log logger.Logger) *EquipmentHTTPClient {
	if client == nil {
		client = &http.Client{}
	}
	return &EquipmentHTTPClient{
		baseURL: baseURL,
		client:  client,
		logger:  log,
	}
}

type equipmentStatusRequest struct {
	Status string `json:"status"`
}

// LockForMaintenance sets equipment status to "maintenance"
func (e *EquipmentHTTPClient) LockForMaintenance(ctx context.Context, tenantID, equipmentID, taskID string) error {
	body := equipmentStatusRequest{Status: "maintenance"}
	payload, _ := json.Marshal(body)

	url := fmt.Sprintf("%s/api/v1/equipment/%s/status", e.baseURL, equipmentID)
	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(payload))
	if err != nil {
		e.logger.Error("Failed to create request for lock", err, "equipment_id", equipmentID)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID)

	resp, err := e.client.Do(req)
	if err != nil {
		// Graceful fallback: log warning but don't fail
		e.logger.Warn("Equipment service unreachable, continuing without lock", "equipment_id", equipmentID, "error", err.Error())
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		e.logger.Warn("Failed to lock equipment", "equipment_id", equipmentID, "status", resp.StatusCode)
		// Still return nil to allow maintenance workflow to continue
		return nil
	}

	e.logger.Info("Equipment locked for maintenance", "equipment_id", equipmentID, "task_id", taskID)
	return nil
}

// UnlockFromMaintenance sets equipment status to "available"
func (e *EquipmentHTTPClient) UnlockFromMaintenance(ctx context.Context, tenantID, equipmentID, taskID string) error {
	body := equipmentStatusRequest{Status: "available"}
	payload, _ := json.Marshal(body)

	url := fmt.Sprintf("%s/api/v1/equipment/%s/status", e.baseURL, equipmentID)
	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(payload))
	if err != nil {
		e.logger.Error("Failed to create request for unlock", err, "equipment_id", equipmentID)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID)

	resp, err := e.client.Do(req)
	if err != nil {
		// Graceful fallback: log warning but don't fail
		e.logger.Warn("Equipment service unreachable, continuing without unlock", "equipment_id", equipmentID, "error", err.Error())
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		e.logger.Warn("Failed to unlock equipment", "equipment_id", equipmentID, "status", resp.StatusCode)
		// Still return nil to allow maintenance workflow to continue
		return nil
	}

	e.logger.Info("Equipment unlocked from maintenance", "equipment_id", equipmentID, "task_id", taskID)
	return nil
}

// IsLockedForMaintenance checks if equipment is in maintenance status
func (e *EquipmentHTTPClient) IsLockedForMaintenance(ctx context.Context, tenantID, equipmentID string) (bool, error) {
	url := fmt.Sprintf("%s/api/v1/equipment/%s/status", e.baseURL, equipmentID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		e.logger.Error("Failed to create request for status check", err, "equipment_id", equipmentID)
		return false, err
	}

	req.Header.Set("X-Tenant-ID", tenantID)

	resp, err := e.client.Do(req)
	if err != nil {
		// Graceful fallback: assume not locked
		e.logger.Warn("Equipment service unreachable, assuming not locked", "equipment_id", equipmentID, "error", err.Error())
		return false, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		e.logger.Warn("Failed to check equipment status", "equipment_id", equipmentID, "status", resp.StatusCode)
		return false, nil
	}

	var statusResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&statusResp)
	status, ok := statusResp["status"].(string)
	if !ok {
		return false, nil
	}

	return status == "maintenance", nil
}
