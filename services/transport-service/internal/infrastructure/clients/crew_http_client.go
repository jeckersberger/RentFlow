package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/transport-service/internal/ports"
)

// CrewHTTPClient implements CrewClient interface using HTTP
type CrewHTTPClient struct {
	baseURL    string
	httpClient *http.Client
	logger     logger.Logger
}

// NewCrewHTTPClient creates a new crew service HTTP client
func NewCrewHTTPClient(baseURL string, log logger.Logger) *CrewHTTPClient {
	return &CrewHTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: log,
	}
}

// ValidateDriver verifies that a driver exists and is valid
func (c *CrewHTTPClient) ValidateDriver(ctx context.Context, tenantID, driverID string) (*ports.DriverInfo, error) {
	url := fmt.Sprintf("%s/api/v1/crew/members/%s", c.baseURL, driverID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		c.logger.Error("Failed to create request for ValidateDriver", err, "driverID", driverID)
		return nil, err
	}

	req.Header.Set("X-Tenant-ID", tenantID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to call crew-service", err, "driverID", driverID)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("Crew-service returned error", fmt.Errorf("status %d", resp.StatusCode), "driverID", driverID)
		return nil, fmt.Errorf("crew-service error: status %d", resp.StatusCode)
	}

	var driverData struct {
		ID               string   `json:"id"`
		Name             string   `json:"name"`
		Phone            string   `json:"phone"`
		LicenseType      string   `json:"license_type"`
		QualificationIDs []string `json:"qualification_ids"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&driverData); err != nil {
		c.logger.Error("Failed to decode crew-service response", err, "driverID", driverID)
		return nil, err
	}

	return &ports.DriverInfo{
		ID:               driverData.ID,
		Name:             driverData.Name,
		Phone:            driverData.Phone,
		LicenseType:      driverData.LicenseType,
		QualificationIDs: driverData.QualificationIDs,
	}, nil
}

// GetDriverAvailability checks if a driver is available on a specific date
func (c *CrewHTTPClient) GetDriverAvailability(ctx context.Context, tenantID, driverID string, date time.Time) (bool, error) {
	url := fmt.Sprintf("%s/api/v1/crew/members/%s/availability", c.baseURL, driverID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		c.logger.Error("Failed to create request for GetDriverAvailability", err, "driverID", driverID)
		return false, err
	}

	q := req.URL.Query()
	q.Add("date", date.Format("2006-01-02"))
	req.URL.RawQuery = q.Encode()

	req.Header.Set("X-Tenant-ID", tenantID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to call crew-service availability", err, "driverID", driverID)
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("Crew-service availability returned error", fmt.Errorf("status %d", resp.StatusCode), "driverID", driverID)
		return false, fmt.Errorf("crew-service error: status %d", resp.StatusCode)
	}

	var availData struct {
		Available bool `json:"available"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&availData); err != nil {
		c.logger.Error("Failed to decode availability response", err, "driverID", driverID)
		return false, err
	}

	return availData.Available, nil
}

// NoopCrewClient is a no-op implementation that returns mock data
type NoopCrewClient struct {
	logger logger.Logger
}

// NewNoopCrewClient creates a new no-op crew client for fallback scenarios
func NewNoopCrewClient(log logger.Logger) *NoopCrewClient {
	return &NoopCrewClient{logger: log}
}

// ValidateDriver returns mock driver data
func (n *NoopCrewClient) ValidateDriver(ctx context.Context, tenantID, driverID string) (*ports.DriverInfo, error) {
	n.logger.Info("NoopCrewClient: ValidateDriver called (returning mock data)", "driverID", driverID)
	return &ports.DriverInfo{
		ID:          driverID,
		Name:        "Mock Driver",
		Phone:       "+49 000 0000000",
		LicenseType: "C",
		QualificationIDs: []string{},
	}, nil
}

// GetDriverAvailability returns mock availability
func (n *NoopCrewClient) GetDriverAvailability(ctx context.Context, tenantID, driverID string, date time.Time) (bool, error) {
	n.logger.Info("NoopCrewClient: GetDriverAvailability called (returning true)", "driverID", driverID, "date", date)
	return true, nil
}
