package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/auth/internal/domain"
)

// ConfigService handles tenant configuration key-value management.
type ConfigService struct {
	configRepo domain.TenantConfigRepository
	logger     zerolog.Logger
}

// NewConfigService creates a new ConfigService.
func NewConfigService(configRepo domain.TenantConfigRepository, logger zerolog.Logger) *ConfigService {
	return &ConfigService{
		configRepo: configRepo,
		logger:     logger,
	}
}

// GetAll returns all config entries for a tenant.
func (s *ConfigService) GetAll(ctx context.Context, tenantID uuid.UUID) ([]*domain.TenantConfig, error) {
	configs, err := s.configRepo.GetAll(ctx, tenantID)
	if err != nil {
		s.logger.Error().Err(err).Str("tenant_id", tenantID.String()).Msg("failed to get all configs")
		return nil, fmt.Errorf("config retrieval failed: %w", err)
	}
	return configs, nil
}

// Get returns a single config entry by tenant and key.
func (s *ConfigService) Get(ctx context.Context, tenantID uuid.UUID, key string) (*domain.TenantConfig, error) {
	config, err := s.configRepo.Get(ctx, tenantID, key)
	if err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Str("key", key).
			Msg("failed to get config")
		return nil, fmt.Errorf("config retrieval failed: %w", err)
	}
	return config, nil
}

// Set creates or updates a config value for a tenant.
func (s *ConfigService) Set(ctx context.Context, tenantID uuid.UUID, key string, value json.RawMessage) error {
	if key == "" {
		return fmt.Errorf("config key must not be empty")
	}

	if err := s.configRepo.Set(ctx, tenantID, key, value); err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Str("key", key).
			Msg("failed to set config")
		return fmt.Errorf("config update failed: %w", err)
	}

	s.logger.Info().
		Str("tenant_id", tenantID.String()).
		Str("key", key).
		Msg("config updated")
	return nil
}

// Delete removes a config entry for a tenant.
func (s *ConfigService) Delete(ctx context.Context, tenantID uuid.UUID, key string) error {
	if err := s.configRepo.Delete(ctx, tenantID, key); err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Str("key", key).
			Msg("failed to delete config")
		return fmt.Errorf("config deletion failed: %w", err)
	}

	s.logger.Info().
		Str("tenant_id", tenantID.String()).
		Str("key", key).
		Msg("config deleted")
	return nil
}
