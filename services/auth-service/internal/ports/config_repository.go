package ports

import (
	"context"
	"encoding/json"
)

// ConfigRepository defines the interface for tenant config data access
type ConfigRepository interface {
	// GetConfig retrieves a specific config value for a tenant
	GetConfig(ctx context.Context, tenantID, key string) (json.RawMessage, error)

	// SetConfig sets a config value for a tenant
	SetConfig(ctx context.Context, tenantID, key string, value json.RawMessage) error

	// GetAllConfigs retrieves all config values for a tenant
	GetAllConfigs(ctx context.Context, tenantID string) (map[string]json.RawMessage, error)

	// DeleteConfig deletes a config value for a tenant
	DeleteConfig(ctx context.Context, tenantID, key string) error
}
