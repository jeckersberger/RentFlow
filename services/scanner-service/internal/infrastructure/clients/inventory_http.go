package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jeckersberger/rentflow/services/scanner-service/internal/ports"
)

// InventoryHTTPClient implements ports.InventoryServiceClient via HTTP calls
// to the inventory-service REST API.
type InventoryHTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewInventoryHTTPClient creates a new HTTP client for the inventory service.
// baseURL should be like "http://localhost:8002".
func NewInventoryHTTPClient(baseURL string) *InventoryHTTPClient {
	return &InventoryHTTPClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// equipmentResponse is the JSON shape returned by inventory-service equipment endpoints.
type equipmentResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	CategoryID string `json:"category_id"`
	Barcode    string `json:"barcode"`
	RfidTag    string `json:"rfid_tag"`
	Status     string `json:"status"`
	LocationID string `json:"location_id"`
}

func tenantCtx(ctx context.Context) string {
	// We pass tenantID explicitly, so this is unused — kept for interface compat.
	return ""
}

// ResolveBarcode calls GET /api/v1/equipment/lookup/barcode/{barcode} and returns the equipment ID.
func (c *InventoryHTTPClient) ResolveBarcode(ctx context.Context, tenantID, barcode string) (string, error) {
	url := fmt.Sprintf("%s/api/v1/equipment/lookup/barcode/%s", c.baseURL, barcode)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("X-Tenant-ID", tenantID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("inventory service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("inventory service returned %d: %s", resp.StatusCode, string(body))
	}

	var eq equipmentResponse
	if err := json.NewDecoder(resp.Body).Decode(&eq); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return eq.ID, nil
}

// UpdateEquipmentStatus calls PATCH /api/v1/equipment/{id}/status.
func (c *InventoryHTTPClient) UpdateEquipmentStatus(ctx context.Context, tenantID, equipmentID, status string) error {
	url := fmt.Sprintf("%s/api/v1/equipment/%s/status", c.baseURL, equipmentID)

	payload := fmt.Sprintf(`{"status":"%s"}`, status)
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, strings.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("X-Tenant-ID", tenantID)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("inventory service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("inventory service returned %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ResolveEquipment calls GET /api/v1/equipment/lookup/barcode/{code} or
// GET /api/v1/equipment/lookup/rfid/{code} depending on scanType,
// and returns the full EquipmentDetail.
func (c *InventoryHTTPClient) ResolveEquipment(ctx context.Context, tenantID, code, scanType string) (*ports.EquipmentDetail, error) {
	var url string
	switch scanType {
	case "rfid":
		url = fmt.Sprintf("%s/api/v1/equipment/lookup/rfid/%s", c.baseURL, code)
	default:
		// "barcode" or anything else defaults to barcode lookup
		url = fmt.Sprintf("%s/api/v1/equipment/lookup/barcode/%s", c.baseURL, code)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("X-Tenant-ID", tenantID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("inventory service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("inventory service returned %d: %s", resp.StatusCode, string(body))
	}

	var eq equipmentResponse
	if err := json.NewDecoder(resp.Body).Decode(&eq); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &ports.EquipmentDetail{
		ID:       eq.ID,
		Name:     eq.Name,
		Category: eq.CategoryID,
		Barcode:  eq.Barcode,
		RfidTag:  eq.RfidTag,
		Status:   eq.Status,
		Location: eq.LocationID,
	}, nil
}

// CheckOutEquipment calls POST /api/v1/equipment/{id}/check-out for each equipment ID.
func (c *InventoryHTTPClient) CheckOutEquipment(ctx context.Context, tenantID string, equipmentIDs []string, projectID, notes string) (*ports.CheckoutResult, error) {
	checkedOut := 0

	for _, eqID := range equipmentIDs {
		url := fmt.Sprintf("%s/api/v1/equipment/%s/check-out", c.baseURL, eqID)
		payload := fmt.Sprintf(`{"project_id":"%s"}`, projectID)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(payload))
		if err != nil {
			continue
		}
		req.Header.Set("X-Tenant-ID", tenantID)
		req.Header.Set("X-User-ID", "scanner-service")
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			checkedOut++
		}
	}

	return &ports.CheckoutResult{
		CheckedOut: checkedOut,
		Project: ports.ProjectRef{
			ID: projectID,
		},
	}, nil
}

// CheckInEquipment calls POST /api/v1/equipment/{id}/check-in for each equipment ID.
func (c *InventoryHTTPClient) CheckInEquipment(ctx context.Context, tenantID string, equipmentIDs []string, conditionRatings map[string]ports.ConditionRating) (*ports.CheckinResult, error) {
	checkedIn := 0
	damageReports := 0

	for _, eqID := range equipmentIDs {
		url := fmt.Sprintf("%s/api/v1/equipment/%s/check-in", c.baseURL, eqID)

		payload := "{}"
		if rating, ok := conditionRatings[eqID]; ok {
			if rating.DamageReported {
				damageReports++
			}
			payload = fmt.Sprintf(`{"notes":"%s"}`, rating.Notes)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(payload))
		if err != nil {
			continue
		}
		req.Header.Set("X-Tenant-ID", tenantID)
		req.Header.Set("X-User-ID", "scanner-service")
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			checkedIn++
		}
	}

	return &ports.CheckinResult{
		CheckedIn:     checkedIn,
		DamageReports: damageReports,
	}, nil
}
