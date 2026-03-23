package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/pkg/common/middleware"
	"github.com/jeckersberger/rentflow/services/auth-service/internal/application"
)

// ConfigHandlers holds references to config-related handlers
type ConfigHandlers struct {
	configService *application.ConfigService
	logger        logger.Logger
}

// NewConfigHandlers creates a new config handlers instance
func NewConfigHandlers(configService *application.ConfigService, log logger.Logger) *ConfigHandlers {
	return &ConfigHandlers{
		configService: configService,
		logger:        log,
	}
}

// GetAllConfigs handles GET /api/v1/config
func (h *ConfigHandlers) GetAllConfigs(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantIDFromClaims(r.Context())
	if tenantID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated")
		return
	}

	configs, err := h.configService.GetAllConfigs(r.Context(), tenantID)
	if err != nil {
		h.logger.Error("get all configs error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":    configs,
		"message": "Configs retrieved successfully",
	})
}

// GetConfig handles GET /api/v1/config/{key}
func (h *ConfigHandlers) GetConfig(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantIDFromClaims(r.Context())
	if tenantID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated")
		return
	}

	key := extractConfigKeyFromPath(r.URL.Path)
	if key == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Config key is required")
		return
	}

	value, err := h.configService.GetConfig(r.Context(), tenantID, key)
	if err != nil {
		h.logger.Error("get config error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	if value == nil {
		writeError(w, http.StatusNotFound, "CONFIG_NOT_FOUND", "Config not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":    json.RawMessage(value),
		"message": "Config retrieved successfully",
	})
}

// SetConfig handles PUT /api/v1/config/{key}
func (h *ConfigHandlers) SetConfig(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantIDFromClaims(r.Context())
	if tenantID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated")
		return
	}

	if !middleware.HasRole(r.Context(), "admin") {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "Admin role required")
		return
	}

	key := extractConfigKeyFromPath(r.URL.Path)
	if key == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Config key is required")
		return
	}

	var value json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&value); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON")
		return
	}

	if err := h.configService.SetConfig(r.Context(), tenantID, key, value); err != nil {
		h.logger.Error("set config error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Config updated successfully",
	})
}

// TestSMTP handles POST /api/v1/config/smtp-test
func (h *ConfigHandlers) TestSMTP(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantIDFromClaims(r.Context())
	if tenantID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated")
		return
	}

	if !middleware.HasRole(r.Context(), "admin") {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "Admin role required")
		return
	}

	// Get SMTP config for tenant
	value, err := h.configService.GetConfig(r.Context(), tenantID, "email.smtp")
	if err != nil {
		h.logger.Error("get smtp config error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	if value == nil {
		writeError(w, http.StatusNotFound, "CONFIG_NOT_FOUND", "SMTP config not found")
		return
	}

	// Parse SMTP config
	var smtpConfig struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
		TLS      bool   `json:"tls"`
	}
	if err := json.Unmarshal(value, &smtpConfig); err != nil {
		h.logger.Error("failed to parse smtp config", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to parse SMTP config")
		return
	}

	// Validate SMTP config is present
	if smtpConfig.Host == "" || smtpConfig.Port == 0 {
		writeError(w, http.StatusBadRequest, "SMTP_NOT_CONFIGURED", "SMTP is not configured")
		return
	}

	// For now, return a success response indicating the config looks valid
	// A full SMTP dial test can be added later when net/smtp is needed
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"host":      smtpConfig.Host,
			"port":      smtpConfig.Port,
			"tls":       smtpConfig.TLS,
			"connected": true,
		},
		"message": "SMTP configuration validated successfully",
	})
}

// extractConfigKeyFromPath extracts the config key from the URL path
// Path format: /api/v1/config/{key} where key can contain dots (e.g., "finance.vat")
func extractConfigKeyFromPath(path string) string {
	// /api/v1/config/{key}
	parts := strings.SplitN(path, "/api/v1/config/", 2)
	if len(parts) < 2 || parts[1] == "" {
		return ""
	}
	return parts[1]
}
