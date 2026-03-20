package config

import (
	"fmt"
	"os"
	"strconv"
)

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

	// KurrentDB configuration
	KurrentDBURL string

	// Logging configuration
	LogLevel string

	// Feature flags
	DebugMode bool
}

// Load loads configuration from environment variables
func Load(serviceName string) *Config {
	return &Config{
		ServiceName:  serviceName,
		ServicePort:  getEnvInt("SERVICE_PORT", 8080),
		Environment:  getEnv("ENVIRONMENT", "development"),
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       getEnvInt("DB_PORT", 5432),
		DBName:       getEnv("DB_NAME", serviceName),
		DBUser:       getEnv("DB_USER", "postgres"),
		DBPassword:   getEnv("DB_PASSWORD", "password"),
		KurrentDBURL: getEnv("KURRENTDB_URL", "http://localhost:8000"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
		DebugMode:    getEnvBool("DEBUG", false),
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
		"postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}
