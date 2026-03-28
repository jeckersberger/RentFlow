package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all service configuration loaded from environment variables.
type Config struct {
	DBHost            string
	DBPort            int
	DBUser            string
	DBPassword        string
	DBName            string
	RedisURL          string
	JWTPrivateKeyPath string
	JWTPublicKeyPath  string
	JWTAccessExpiry   time.Duration
	JWTRefreshExpiry  time.Duration
	Port              string
	LogLevel          string
	CORSOrigins       []string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            getEnvInt("DB_PORT", 5432),
		DBUser:            getEnv("DB_USER", "cratedesk"),
		DBPassword:        getEnv("DB_PASSWORD", ""),
		DBName:            getEnv("DB_NAME", "cratedesk"),
		RedisURL:          getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTPrivateKeyPath: getEnv("JWT_PRIVATE_KEY_PATH", "/etc/cratedesk/keys/private.pem"),
		JWTPublicKeyPath:  getEnv("JWT_PUBLIC_KEY_PATH", "/etc/cratedesk/keys/public.pem"),
		JWTAccessExpiry:   getEnvDuration("JWT_ACCESS_EXPIRY", 15*time.Minute),
		JWTRefreshExpiry:  getEnvDuration("JWT_REFRESH_EXPIRY", 7*24*time.Hour),
		Port:              getEnv("PORT", "8080"),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		CORSOrigins:       getEnvSlice("CORS_ORIGINS", []string{"http://localhost:5173"}),
	}
}

// DSN returns the PostgreSQL connection string.
func (c *Config) DSN() string {
	return "postgres://" + c.DBUser + ":" + c.DBPassword +
		"@" + c.DBHost + ":" + strconv.Itoa(c.DBPort) +
		"/" + c.DBName + "?sslmode=disable"
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

func getEnvSlice(key string, fallback []string) []string {
	if v := os.Getenv(key); v != "" {
		return strings.Split(v, ",")
	}
	return fallback
}
