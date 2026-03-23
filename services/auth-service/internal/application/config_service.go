package application

import (
	"context"
	"encoding/json"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/auth-service/internal/ports"
)

// ConfigService handles tenant configuration business logic
type ConfigService struct {
	configRepo ports.ConfigRepository
	logger     logger.Logger
}

// NewConfigService creates a new config service
func NewConfigService(configRepo ports.ConfigRepository, log logger.Logger) *ConfigService {
	return &ConfigService{
		configRepo: configRepo,
		logger:     log,
	}
}

// GetConfig retrieves a specific config value for a tenant
func (s *ConfigService) GetConfig(ctx context.Context, tenantID, key string) (json.RawMessage, error) {
	value, err := s.configRepo.GetConfig(ctx, tenantID, key)
	if err != nil {
		s.logger.Error("failed to get config", err, "tenantID", tenantID, "key", key)
		return nil, err
	}
	return value, nil
}

// SetConfig sets a config value for a tenant
func (s *ConfigService) SetConfig(ctx context.Context, tenantID, key string, value json.RawMessage) error {
	if err := s.configRepo.SetConfig(ctx, tenantID, key, value); err != nil {
		s.logger.Error("failed to set config", err, "tenantID", tenantID, "key", key)
		return err
	}

	s.logger.Info("config updated", "tenantID", tenantID, "key", key)
	return nil
}

// GetAllConfigs retrieves all config values for a tenant
func (s *ConfigService) GetAllConfigs(ctx context.Context, tenantID string) (map[string]json.RawMessage, error) {
	configs, err := s.configRepo.GetAllConfigs(ctx, tenantID)
	if err != nil {
		s.logger.Error("failed to get all configs", err, "tenantID", tenantID)
		return nil, err
	}
	return configs, nil
}
