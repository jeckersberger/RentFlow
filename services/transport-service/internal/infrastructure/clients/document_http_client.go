package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/transport-service/internal/ports"
)

// DocumentHTTPClient implements DocumentClient interface using HTTP
type DocumentHTTPClient struct {
	baseURL    string
	httpClient *http.Client
	logger     logger.Logger
}

// NewDocumentHTTPClient creates a new document service HTTP client
func NewDocumentHTTPClient(baseURL string, log logger.Logger) *DocumentHTTPClient {
	return &DocumentHTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		logger: log,
	}
}

// GenerateDeliveryNote generates a delivery note document for a tour
func (d *DocumentHTTPClient) GenerateDeliveryNote(ctx context.Context, tenantID string, tourID string, equipmentItems []ports.DeliveryNoteItem) (*ports.DeliveryNoteResult, error) {
	url := fmt.Sprintf("%s/api/v1/documents/from-project/%s", d.baseURL, tourID)

	requestBody := map[string]interface{}{
		"document_type": "delivery_note",
		"items": equipmentItems,
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		d.logger.Error("Failed to marshal request body for GenerateDeliveryNote", err, "tourID", tourID)
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payload))
	if err != nil {
		d.logger.Error("Failed to create request for GenerateDeliveryNote", err, "tourID", tourID)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID)

	resp, err := d.httpClient.Do(req)
	if err != nil {
		d.logger.Error("Failed to call document-service", err, "tourID", tourID)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		d.logger.Error("Document-service returned error", fmt.Errorf("status %d", resp.StatusCode), "tourID", tourID)
		return nil, fmt.Errorf("document-service error: status %d", resp.StatusCode)
	}

	var resultData struct {
		DeliveryNoteNumber string `json:"delivery_note_number"`
		DocumentURL        string `json:"document_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&resultData); err != nil {
		d.logger.Error("Failed to decode document-service response", err, "tourID", tourID)
		return nil, err
	}

	return &ports.DeliveryNoteResult{
		DeliveryNoteNumber: resultData.DeliveryNoteNumber,
		DocumentURL:        resultData.DocumentURL,
	}, nil
}

// NoopDocumentClient is a no-op implementation that returns mock data
type NoopDocumentClient struct {
	logger logger.Logger
}

// NewNoopDocumentClient creates a new no-op document client for fallback scenarios
func NewNoopDocumentClient(log logger.Logger) *NoopDocumentClient {
	return &NoopDocumentClient{logger: log}
}

// GenerateDeliveryNote returns a mock delivery note result
func (n *NoopDocumentClient) GenerateDeliveryNote(ctx context.Context, tenantID string, tourID string, equipmentItems []ports.DeliveryNoteItem) (*ports.DeliveryNoteResult, error) {
	n.logger.Info("NoopDocumentClient: GenerateDeliveryNote called (returning mock data)", "tourID", tourID, "itemCount", len(equipmentItems))

	mockNumber := fmt.Sprintf("DN-MOCK-%s", tourID[:8])

	return &ports.DeliveryNoteResult{
		DeliveryNoteNumber: mockNumber,
		DocumentURL:        fmt.Sprintf("https://documents.mock/delivery-notes/%s", mockNumber),
	}, nil
}
