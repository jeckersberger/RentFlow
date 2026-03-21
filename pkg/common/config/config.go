package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all application configuration
type Config struct {
	// Service configuration
	ServiceName string
	ServicePort int
	Environment string

	// Database configuration
	DBHost     string
	DBPort     int
	DBName     string
	DBUser     string
	DBPassword string
	DBSSLMode  string // disable, require, prefer, allow

	// KurrentDB configuration
	KurrentDBURL string

	// Redis configuration
	RedisHost     string
	RedisPort     int
	RedisPassword string
	RedisDB       int

	// JWT configuration
	JWTSecret        string
	JWTIssuer        string
	JWTExpiryMinutes int

	// Multi-tenancy configuration
	MultiTenantEnabled bool

	// NAS (Network Attached Storage) configuration
	NASEnabled   bool
	NASMountPath string
	NASType      string // "smb", "nfs", "local"

	// CORS configuration
	CORSAllowedOrigins []string
	CORSAllowedMethods []string
	CORSAllowedHeaders []string
	CORSMaxAge         int

	// Logging configuration
	LogLevel string

	// Feature flags
	DebugMode bool
}

// Load loads configuration from environment variables
// Environment variables take priority over YAML config (if implemented)
func Load(serviceName string) *Config {
	return &Config{
		ServiceName:        serviceName,
		ServicePort:        getEnvInt("SERVICE_PORT", 8080),
		Environment:        getEnv("ENVIRONMENT", "development"),
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBPort:             getEnvInt("DB_PORT", 5432),
		DBName:             getEnv("DB_NAME", serviceName),
		DBUser:             getEnv("DB_USER", "postgres"),
		DBPassword:         getEnv("DB_PASSWORD", "password"),
		DBSSLMode:          getEnv("DB_SSL_MODE", "disable"),
		KurrentDBURL:       getEnv("KURRENTDB_URL", "http://localhost:8000"),
		RedisHost:          getEnv("REDIS_HOST", "localhost"),
		RedisPort:          getEnvInt("REDIS_PORT", 6379),
		RedisPassword:      getEnv("REDIS_PASSWORD", ""),
		RedisDB:            getEnvInt("REDIS_DB", 0),
		JWTSecret:          getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		JWTIssuer:          getEnv("JWT_ISSUER", "rentflow"),
		JWTExpiryMinutes:   getEnvInt("JWT_EXPIRY_MINUTES", 60),
		MultiTenantEnabled: getEnvBool("MULTI_TENANT_ENABLED", true),
		NASEnabled:         getEnvBool("NAS_ENABLED", false),
		NASMountPath:       getEnv("NAS_MOUNT_PATH", "/mnt/nas"),
		NASType:            getEnv("NAS_TYPE", "smb"),
		CORSAllowedOrigins: parseEnvList("CORS_ALLOWED_ORIGINS", []string{"*"}),
		CORSAllowedMethods: parseEnvList("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		CORSAllowedHeaders: parseEnvList("CORS_ALLOWED_HEADERS", []string{"Content-Type", "Authorization"}),
		CORSMaxAge:         getEnvInt("CORS_MAX_AGE", 3600),
		LogLevel:           getEnv("LOG_LEVEL", "info"),
		DebugMode:          getEnvBool("DEBUG", false),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}

// ConnectionString returns the PostgreSQL connection string
func (c *Config) ConnectionString() string {
	return fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode,
	)
}

// RedisAddr returns the Redis address in host:port format
func (c *Config) RedisAddr() string {
	return fmt.Sprintf("%s:%d", c.RedisHost, c.RedisPort)
}

// parseEnvList parses a comma-separated environment variable into a slice
func parseEnvList(key string, defaultValue []string) []string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		var items []string
		for _, item := range strings.Split(value, ",") {
			if trimmed := strings.TrimSpace(item); trimmed != "" {
				items = append(items, trimmed)
			}
		}
		if len(items) > 0 {
			return items
		}
	}
	return defaultValue
}
