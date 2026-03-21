package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Clean up environment before test
	os.Clearenv()

	cfg := Load("test-service")

	if cfg.ServiceName != "test-service" {
		t.Errorf("expected ServiceName 'test-service', got '%s'", cfg.ServiceName)
	}

	if cfg.ServicePort != 8080 {
		t.Errorf("expected default ServicePort 8080, got %d", cfg.ServicePort)
	}

	if cfg.Environment != "development" {
		t.Errorf("expected default Environment 'development', got '%s'", cfg.Environment)
	}
}

func TestLoadWithEnvironmentVariables(t *testing.T) {
	// Clean up and set environment variables
	os.Clearenv()
	os.Setenv("SERVICE_PORT", "9000")
	os.Setenv("ENVIRONMENT", "production")
	os.Setenv("DB_HOST", "db.example.com")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_NAME", "rentflow_prod")
	os.Setenv("DB_USER", "admin")
	os.Setenv("DB_PASSWORD", "secret123")
	os.Setenv("JWT_SECRET", "my-jwt-secret")
	os.Setenv("JWT_ISSUER", "rentflow-app")
	os.Setenv("JWT_EXPIRY_MINUTES", "120")

	defer os.Clearenv()

	cfg := Load("my-service")

	tests := []struct {
		name     string
		expected interface{}
		actual   interface{}
	}{
		{"ServicePort", 9000, cfg.ServicePort},
		{"Environment", "production", cfg.Environment},
		{"DBHost", "db.example.com", cfg.DBHost},
		{"DBPort", 5433, cfg.DBPort},
		{"DBName", "rentflow_prod", cfg.DBName},
		{"DBUser", "admin", cfg.DBUser},
		{"DBPassword", "secret123", cfg.DBPassword},
		{"JWTSecret", "my-jwt-secret", cfg.JWTSecret},
		{"JWTIssuer", "rentflow-app", cfg.JWTIssuer},
		{"JWTExpiryMinutes", 120, cfg.JWTExpiryMinutes},
	}

	for _, tt := range tests {
		if tt.expected != tt.actual {
			t.Errorf("TestLoadWithEnvironmentVariables %s: expected %v, got %v", tt.name, tt.expected, tt.actual)
		}
	}
}

func TestLoadRedisConfiguration(t *testing.T) {
	os.Clearenv()
	os.Setenv("REDIS_HOST", "redis.local")
	os.Setenv("REDIS_PORT", "6380")
	os.Setenv("REDIS_PASSWORD", "redispass")
	os.Setenv("REDIS_DB", "1")

	defer os.Clearenv()

	cfg := Load("test")

	if cfg.RedisHost != "redis.local" {
		t.Errorf("expected RedisHost 'redis.local', got '%s'", cfg.RedisHost)
	}

	if cfg.RedisPort != 6380 {
		t.Errorf("expected RedisPort 6380, got %d", cfg.RedisPort)
	}

	if cfg.RedisPassword != "redispass" {
		t.Errorf("expected RedisPassword 'redispass', got '%s'", cfg.RedisPassword)
	}

	if cfg.RedisDB != 1 {
		t.Errorf("expected RedisDB 1, got %d", cfg.RedisDB)
	}
}

func TestLoadMultiTenantConfiguration(t *testing.T) {
	os.Clearenv()
	os.Setenv("MULTI_TENANT_ENABLED", "true")

	defer os.Clearenv()

	cfg := Load("test")

	if !cfg.MultiTenantEnabled {
		t.Errorf("expected MultiTenantEnabled true, got false")
	}
}

func TestLoadNASConfiguration(t *testing.T) {
	os.Clearenv()
	os.Setenv("NAS_ENABLED", "true")
	os.Setenv("NAS_MOUNT_PATH", "/mnt/storage")
	os.Setenv("NAS_TYPE", "nfs")

	defer os.Clearenv()

	cfg := Load("test")

	if !cfg.NASEnabled {
		t.Errorf("expected NASEnabled true, got false")
	}

	if cfg.NASMountPath != "/mnt/storage" {
		t.Errorf("expected NASMountPath '/mnt/storage', got '%s'", cfg.NASMountPath)
	}

	if cfg.NASType != "nfs" {
		t.Errorf("expected NASType 'nfs', got '%s'", cfg.NASType)
	}
}

func TestLoadCORSConfiguration(t *testing.T) {
	os.Clearenv()
	os.Setenv("CORS_ALLOWED_ORIGINS", "https://example.com, https://app.example.com")
	os.Setenv("CORS_ALLOWED_METHODS", "GET,POST,PUT,DELETE")
	os.Setenv("CORS_ALLOWED_HEADERS", "Content-Type,Authorization,X-Custom")
	os.Setenv("CORS_MAX_AGE", "7200")

	defer os.Clearenv()

	cfg := Load("test")

	if len(cfg.CORSAllowedOrigins) != 2 {
		t.Errorf("expected 2 CORS origins, got %d", len(cfg.CORSAllowedOrigins))
	}

	if cfg.CORSAllowedOrigins[0] != "https://example.com" {
		t.Errorf("expected first origin 'https://example.com', got '%s'", cfg.CORSAllowedOrigins[0])
	}

	if len(cfg.CORSAllowedMethods) != 4 {
		t.Errorf("expected 4 CORS methods, got %d", len(cfg.CORSAllowedMethods))
	}

	if cfg.CORSMaxAge != 7200 {
		t.Errorf("expected CORSMaxAge 7200, got %d", cfg.CORSMaxAge)
	}
}

func TestConnectionString(t *testing.T) {
	cfg := &Config{
		DBUser:    "testuser",
		DBPassword: "testpass",
		DBHost:    "localhost",
		DBPort:    5432,
		DBName:    "testdb",
		DBSSLMode: "disable",
	}

	expected := "postgresql://testuser:testpass@localhost:5432/testdb?sslmode=disable"
	actual := cfg.ConnectionString()

	if actual != expected {
		t.Errorf("expected ConnectionString '%s', got '%s'", expected, actual)
	}
}

func TestConnectionStringWithSSL(t *testing.T) {
	cfg := &Config{
		DBUser:    "produser",
		DBPassword: "prodpass",
		DBHost:    "db.prod.example.com",
		DBPort:    5432,
		DBName:    "rentflow_prod",
		DBSSLMode: "require",
	}

	expected := "postgresql://produser:prodpass@db.prod.example.com:5432/rentflow_prod?sslmode=require"
	actual := cfg.ConnectionString()

	if actual != expected {
		t.Errorf("expected ConnectionString '%s', got '%s'", expected, actual)
	}
}

func TestRedisAddr(t *testing.T) {
	cfg := &Config{
		RedisHost: "redis.local",
		RedisPort: 6379,
	}

	expected := "redis.local:6379"
	actual := cfg.RedisAddr()

	if actual != expected {
		t.Errorf("expected RedisAddr '%s', got '%s'", expected, actual)
	}
}

func TestGetEnvInt_InvalidValue(t *testing.T) {
	os.Clearenv()
	os.Setenv("TEST_PORT", "invalid")
	defer os.Clearenv()

	result := getEnvInt("TEST_PORT", 8080)

	if result != 8080 {
		t.Errorf("expected default 8080 for invalid int, got %d", result)
	}
}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"true_string", "true", true},
		{"one", "1", true},
		{"yes", "yes", true},
		{"false_string", "false", false},
		{"zero", "0", false},
		{"no", "no", false},
	}

	for _, tt := range tests {
		os.Clearenv()
		os.Setenv("TEST_BOOL", tt.value)
		result := getEnvBool("TEST_BOOL", false)
		if result != tt.expected {
			t.Errorf("getEnvBool(%s): expected %v, got %v", tt.value, tt.expected, result)
		}
	}

	os.Clearenv()
}

func TestParseEnvList(t *testing.T) {
	os.Clearenv()
	os.Setenv("TEST_LIST", "item1, item2, item3")
	defer os.Clearenv()

	result := parseEnvList("TEST_LIST", []string{})

	if len(result) != 3 {
		t.Errorf("expected 3 items, got %d", len(result))
	}

	if result[0] != "item1" || result[1] != "item2" || result[2] != "item3" {
		t.Errorf("expected [item1, item2, item3], got %v", result)
	}
}

func TestParseEnvList_WithSpaces(t *testing.T) {
	os.Clearenv()
	os.Setenv("TEST_LIST", " item1 , item2 , item3 ")
	defer os.Clearenv()

	result := parseEnvList("TEST_LIST", []string{})

	if len(result) != 3 {
		t.Errorf("expected 3 items after trim, got %d", len(result))
	}

	if result[0] != "item1" || result[1] != "item2" || result[2] != "item3" {
		t.Errorf("expected trimmed items, got %v", result)
	}
}

func TestParseEnvList_EmptyValue(t *testing.T) {
	os.Clearenv()
	os.Setenv("TEST_LIST", "")
	defer os.Clearenv()

	defaultList := []string{"default1", "default2"}
	result := parseEnvList("TEST_LIST", defaultList)

	if len(result) != 2 {
		t.Errorf("expected default list with 2 items, got %d", len(result))
	}

	if result[0] != "default1" || result[1] != "default2" {
		t.Errorf("expected default list, got %v", result)
	}
}

func TestParseEnvList_NotSet(t *testing.T) {
	os.Clearenv()
	defer os.Clearenv()

	defaultList := []string{"default1", "default2"}
	result := parseEnvList("NONEXISTENT", defaultList)

	if len(result) != 2 {
		t.Errorf("expected default list with 2 items, got %d", len(result))
	}
}

func TestLoadDebugMode(t *testing.T) {
	os.Clearenv()
	os.Setenv("DEBUG", "true")
	defer os.Clearenv()

	cfg := Load("test")

	if !cfg.DebugMode {
		t.Errorf("expected DebugMode true, got false")
	}
}

func TestLoadLogLevel(t *testing.T) {
	os.Clearenv()
	os.Setenv("LOG_LEVEL", "debug")
	defer os.Clearenv()

	cfg := Load("test")

	if cfg.LogLevel != "debug" {
		t.Errorf("expected LogLevel 'debug', got '%s'", cfg.LogLevel)
	}
}
