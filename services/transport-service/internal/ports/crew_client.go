package ports

import (
	"context"
	"time"
)

// DriverInfo contains information about a driver from crew-service
type DriverInfo struct {
	ID               string
	Name             string
	Phone            string
	LicenseType      string
	QualificationIDs []string
}

// CrewClient defines the interface for communicating with crew-service
type CrewClient interface {
	// ValidateDriver verifies that a driver exists and is valid
	ValidateDriver(ctx context.Context, tenantID, driverID string) (*DriverInfo, error)

	// GetDriverAvailability checks if a driver is available on a specific date
	GetDriverAvailability(ctx context.Context, tenantID, driverID string, date time.Time) (bool, error)
}
