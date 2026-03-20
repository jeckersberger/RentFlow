# Shared Libraries (pkg/common) — RentFlow

**Stand:** 20. März 2026
**Zielgruppe:** Alle 17 Go-Services

---

## 1. pkg/common/config — Konfigurationsmanagement

```go
// pkg/common/config/config.go
package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	EventStore EventStoreConfig
	Redis     RedisConfig
	Auth      AuthConfig
	Logging   LoggingConfig
	Features  FeaturesConfig
}

type ServerConfig struct {
	Addr         string        // ":8002"
	Port         int           // 8002
	Timeout      time.Duration // 15s
	ReadTimeout  time.Duration // 15s
	WriteTimeout time.Duration // 15s
	IdleTimeout  time.Duration // 60s
	Environment  string        // "development", "staging", "production"
	ServiceName  string        // "inventory-service"
}

type DatabaseConfig struct {
	DSN             string        // PostgreSQL Connection String
	MaxConnections  int           // 20
	MinConnections  int           // 5
	ConnectionTTL   time.Duration // 5m
	StatementCache  int           // 512
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string        // "disable", "require", "verify-full"
	Schema          string        // "inventory_schema"
	MigrationDir    string        // "./migrations"
	AutoMigrate     bool          // true in dev, false in prod
}

type EventStoreConfig struct {
	URL         string        // "esdb://kurrentdb:2113"
	Credentials Credentials
	Timeout     time.Duration // 5s
	MaxRetries  int           // 3
	CertPath    string        // Optional: mTLS
	KeyPath     string
}

type Credentials struct {
	Username string
	Password string
}

type RedisConfig struct {
	Addr           string        // "localhost:6379"
	Password       string
	DB             int
	PoolSize       int           // 10
	MinIdleConns   int           // 5
	MaxRetries     int           // 3
	ReadTimeout    time.Duration // 3s
	WriteTimeout   time.Duration // 3s
	PoolTimeout    time.Duration // 4s
	IdleTimeout    time.Duration // 5m
	MaxConnAge     time.Duration // 0 (unlimited)
}

type AuthConfig struct {
	JWTSecret         string        // Signing Secret
	JWTPublicKey      string        // Public Key für RS256
	JWTPrivateKey     string        // Private Key für RS256
	JWTAlgorithm      string        // "HS256", "RS256"
	JWTExpiration     time.Duration // 24h
	RefreshTokenTTL   time.Duration // 7d
	TokenIssuer       string        // "rentflow"
	TokenAudience     string        // "rentflow-api"
	CookieSecure      bool          // true in prod
	CookieHttpOnly    bool          // true
	CookieSameSite    string        // "Strict", "Lax", "None"
	MFAEnabled        bool
	PasswordMinLength int           // 12
}

type LoggingConfig struct {
	Level      string        // "debug", "info", "warn", "error"
	Format     string        // "json", "text"
	Output     string        // "stdout", "file"
	FilePath   string        // "./logs/service.log"
	MaxSize    int           // 100 (MB)
	MaxBackups int           // 3
	MaxAge     int           // 30 (days)
}

type FeaturesConfig struct {
	EnableRelationalSink bool
	EnableCaching        bool
	EnableMetrics        bool
	EnableTracing        bool
}

// LoadConfig aus YAML + Environment Overrides
func LoadConfig(configPath string, envPrefix string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetEnvPrefix(envPrefix)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.Server.Port < 1024 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Server.Port)
	}
	if c.Database.Host == "" {
		return fmt.Errorf("database host required")
	}
	if c.Auth.JWTExpiration == 0 {
		return fmt.Errorf("JWT expiration required")
	}
	return nil
}

// config.yaml Template
/*
server:
  port: 8002
  environment: development
  service_name: inventory-service
  timeout: 15s
  read_timeout: 15s
  write_timeout: 15s
  idle_timeout: 60s

database:
  host: localhost
  port: 5432
  user: rentflow
  password: ${DB_PASSWORD}
  database: rentflow
  ssl_mode: disable
  schema: inventory_schema
  max_connections: 20
  min_connections: 5
  connection_ttl: 5m
  auto_migrate: true

event_store:
  url: "esdb://localhost:2113"
  credentials:
    username: ${ESDB_USER}
    password: ${ESDB_PASSWORD}
  timeout: 5s
  max_retries: 3

redis:
  addr: "localhost:6379"
  db: 0
  pool_size: 10
  min_idle_conns: 5

auth:
  jwt_secret: ${JWT_SECRET}
  jwt_algorithm: HS256
  jwt_expiration: 24h
  token_issuer: rentflow
  mfa_enabled: false

logging:
  level: debug
  format: json
  output: stdout

features:
  enable_relational_sink: true
  enable_caching: true
  enable_metrics: true
*/
```

---

## 2. pkg/common/auth — JWT & RBAC

```go
// pkg/common/auth/jwt.go
package auth

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID      string   `json:"user_id"`
	TenantID    string   `json:"tenant_id"`
	Email       string   `json:"email"`
	Roles       []string `json:"roles"` // ["admin", "user"]
	Permissions []string `json:"permissions"`
	IssuedAt    int64    `json:"iat"`
	ExpiresAt   int64    `json:"exp"`
}

type TokenManager struct {
	signingKey interface{}
	publicKey  interface{}
	algorithm  string
	issuer     string
	audience   string
	expiration time.Duration
}

func NewTokenManager(
	signingKey interface{},
	publicKey interface{},
	algorithm string,
	issuer string,
	audience string,
	expiration time.Duration,
) *TokenManager {
	return &TokenManager{
		signingKey: signingKey,
		publicKey:  publicKey,
		algorithm:  algorithm,
		issuer:     issuer,
		audience:   audience,
		expiration: expiration,
	}
}

// GenerateToken erstellt einen neuen JWT
func (tm *TokenManager) GenerateToken(claims *Claims) (string, error) {
	now := time.Now()
	claims.IssuedAt = now.Unix()
	claims.ExpiresAt = now.Add(tm.expiration).Unix()

	var signingMethod jwt.SigningMethod
	switch tm.algorithm {
	case "HS256":
		signingMethod = jwt.SigningMethodHS256
	case "RS256":
		signingMethod = jwt.SigningMethodRS256
	default:
		return "", fmt.Errorf("unsupported algorithm: %s", tm.algorithm)
	}

	token := jwt.NewWithClaims(signingMethod, jwt.MapClaims{
		"user_id":     claims.UserID,
		"tenant_id":   claims.TenantID,
		"email":       claims.Email,
		"roles":       claims.Roles,
		"permissions": claims.Permissions,
		"iat":         claims.IssuedAt,
		"exp":         claims.ExpiresAt,
		"iss":         tm.issuer,
		"aud":         tm.audience,
	})

	return token.SignedString(tm.signingKey)
}

// ValidateToken verifiziert JWT und extrahiert Claims
func (tm *TokenManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verifiziere Algorithm
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return tm.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("token parsing failed: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token structure")
	}

	// Extract Claims
	claims := &Claims{
		UserID:    mapClaims["user_id"].(string),
		TenantID:  mapClaims["tenant_id"].(string),
		Email:     mapClaims["email"].(string),
		ExpiresAt: int64(mapClaims["exp"].(float64)),
	}

	// Roles
	if roles, ok := mapClaims["roles"].([]interface{}); ok {
		for _, r := range roles {
			claims.Roles = append(claims.Roles, r.(string))
		}
	}

	return claims, nil
}

// Context Helper
type ctxKey string

const claimsKey ctxKey = "claims"

func WithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(claimsKey).(*Claims)
	return claims, ok
}

// RBAC Helpers
func (c *Claims) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}

func (c *Claims) HasPermission(permission string) bool {
	for _, p := range c.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

func (c *Claims) HasAnyRole(roles ...string) bool {
	for _, role := range roles {
		if c.HasRole(role) {
			return true
		}
	}
	return false
}
```

---

## 3. pkg/common/middleware — HTTP Middleware

```go
// pkg/common/middleware/middleware.go
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"rentflow/pkg/common/auth"
)

// RequestID für Korrelation
type requestIDKey struct{}

func GenerateRequestID() string {
	return uuid.New().String()
}

func RequestIDMiddleware(logger *zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = GenerateRequestID()
			}

			ctx := context.WithValue(r.Context(), requestIDKey{}, requestID)
			w.Header().Set("X-Request-ID", requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// TenantMiddleware extrahiert Tenant aus Claims
func TenantMiddleware(logger *zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := auth.ClaimsFromContext(r.Context())
			if !ok {
				http.Error(w, "missing claims", http.StatusUnauthorized)
				return
			}

			// Set PostgreSQL app.current_tenant_id für RLS
			ctx := r.Context()
			ctx = context.WithValue(ctx, "tenant_id", claims.TenantID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// CorrelationIDMiddleware für distributed tracing
func CorrelationIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := r.Header.Get("X-Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.New().String()
		}

		ctx := context.WithValue(r.Context(), "correlation_id", correlationID)
		w.Header().Set("X-Correlation-ID", correlationID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoggingMiddleware mit strukturiertem Logging
func LoggingMiddleware(logger *zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap ResponseWriter für Status-Code
			wrapped := &wrappedResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Add logger to context
			ctx := logger.WithContext(r.Context())
			next.ServeHTTP(wrapped, r.WithContext(ctx))

			duration := time.Since(start)
			logger.Info().
				Str("method", r.Method).
				Str("path", r.RequestURI).
				Int("status", wrapped.statusCode).
				Duration("duration_ms", duration).
				Str("remote_addr", r.RemoteAddr).
				Str("user_agent", r.UserAgent()).
				Msg("HTTP Request")
		})
	}
}

type wrappedResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// CircuitBreakerMiddleware für externe Service Calls
type CircuitBreakerMiddleware struct {
	failureThreshold int
	successThreshold int
	timeout          time.Duration
	state            string // "closed", "open", "half-open"
	failures         int
	successes        int
	lastFailureTime   time.Time
}

func NewCircuitBreakerMiddleware(
	failureThreshold, successThreshold int,
	timeout time.Duration,
) *CircuitBreakerMiddleware {
	return &CircuitBreakerMiddleware{
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		timeout:          timeout,
		state:            "closed",
	}
}

func (cb *CircuitBreakerMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cb.state == "open" {
			if time.Since(cb.lastFailureTime) > cb.timeout {
				cb.state = "half-open"
				cb.successes = 0
			} else {
				http.Error(w, "service unavailable", http.StatusServiceUnavailable)
				return
			}
		}

		wrapped := &wrappedResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)

		if wrapped.statusCode >= 500 {
			cb.failures++
			cb.lastFailureTime = time.Now()
			if cb.failures >= cb.failureThreshold {
				cb.state = "open"
			}
		} else {
			if cb.state == "half-open" {
				cb.successes++
				if cb.successes >= cb.successThreshold {
					cb.state = "closed"
					cb.failures = 0
				}
			}
		}
	})
}

// RateLimitMiddleware
type RateLimiter interface {
	Allow(clientID string) bool
}

func RateLimitMiddleware(limiter RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, _ := auth.ClaimsFromContext(r.Context())
			clientID := claims.TenantID

			if !limiter.Allow(clientID) {
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
```

---

## 4. pkg/common/errors — Custom Error Types

```go
// pkg/common/errors/errors.go
package errors

import (
	"fmt"
	"net/http"
)

type ErrorCode string

const (
	ErrUnauthorized      ErrorCode = "UNAUTHORIZED"
	ErrForbidden         ErrorCode = "FORBIDDEN"
	ErrNotFound          ErrorCode = "NOT_FOUND"
	ErrConflict          ErrorCode = "CONFLICT"
	ErrValidation        ErrorCode = "VALIDATION_ERROR"
	ErrInternalServer    ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrTimeout           ErrorCode = "TIMEOUT"
	ErrDomainError       ErrorCode = "DOMAIN_ERROR"
)

type Error struct {
	Code      ErrorCode `json:"code"`
	Message   string    `json:"message"`
	Details   string    `json:"details,omitempty"`
	ErrorType string    `json:"error_type"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Code, e.ErrorType, e.Message)
}

func (e *Error) HTTPStatus() int {
	switch e.Code {
	case ErrUnauthorized:
		return http.StatusUnauthorized
	case ErrForbidden:
		return http.StatusForbidden
	case ErrNotFound:
		return http.StatusNotFound
	case ErrConflict:
		return http.StatusConflict
	case ErrValidation:
		return http.StatusBadRequest
	case ErrServiceUnavailable:
		return http.StatusServiceUnavailable
	case ErrTimeout:
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}

// Factory Functions
func NewNotFound(resource, id string) *Error {
	return &Error{
		Code:      ErrNotFound,
		Message:   fmt.Sprintf("%s with id %s not found", resource, id),
		ErrorType: "ResourceNotFound",
	}
}

func NewValidationError(field, reason string) *Error {
	return &Error{
		Code:      ErrValidation,
		Message:   fmt.Sprintf("Invalid %s: %s", field, reason),
		ErrorType: "ValidationError",
	}
}

func NewDomainError(aggregate, rule string) *Error {
	return &Error{
		Code:      ErrDomainError,
		Message:   fmt.Sprintf("Domain rule violated: %s.%s", aggregate, rule),
		ErrorType: "DomainError",
	}
}

func IsDomainError(err error) bool {
	e, ok := err.(*Error)
	return ok && e.Code == ErrDomainError
}
```

---

## 5. pkg/common/events — KurrentDB Client Wrapper

```go
// pkg/common/events/client.go
package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	esdb "github.com/EventStore/EventStore-Client-Go/v4"
)

type Client struct {
	conn   *esdb.Client
	config *esdb.Configuration
}

type EventData struct {
	ID            string      `json:"id"`
	Type          string      `json:"type"`
	AggregateID   string      `json:"aggregate_id"`
	Version       int64       `json:"version"`
	Data          interface{} `json:"data"`
	Metadata      Metadata    `json:"metadata"`
	CreatedAt     time.Time   `json:"created_at"`
}

type Metadata struct {
	TenantID      string `json:"tenant_id"`
	UserID        string `json:"user_id"`
	CorrelationID string `json:"correlation_id"`
	CausationID   string `json:"causation_id"`
}

type EventHandler func(ctx context.Context, event *EventData) error

func NewClient(connectionString string, creds *Credentials) (*Client, error) {
	settings, err := esdb.ParseConnectionString(connectionString)
	if err != nil {
		return nil, fmt.Errorf("invalid connection string: %w", err)
	}

	if creds != nil {
		settings.Username = creds.Username
		settings.Password = creds.Password
	}

	conn, err := esdb.NewClient(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return &Client{
		conn:   conn,
		config: settings,
	}, nil
}

// AppendEvent — Schreibe ein Event zu einem Stream
func (c *Client) AppendEvent(
	ctx context.Context,
	streamID string,
	event interface{},
	metadata Metadata,
	expectedVersion int64,
) error {
	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	metadataData, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	opts := esdb.AppendToStreamOptions{
		ExpectedRevision: esdb.Revision{Value: uint64(expectedVersion)},
	}

	appendResp, err := c.conn.AppendToStream(ctx, streamID, opts, &esdb.EventData{
		ContentType: esdb.JsonContentType,
		EventType:   fmt.Sprintf("%T", event),
		Data:        eventData,
		Metadata:    metadataData,
	})

	if err != nil {
		if esdb.IsWrongExpectedVersionError(err) {
			return fmt.Errorf("concurrency conflict: %w", err)
		}
		return fmt.Errorf("failed to append: %w", err)
	}

	// TODO: log appendResp.NextExpectedVersion
	_ = appendResp

	return nil
}

// ReadStream — Lese alle Events aus einem Stream
func (c *Client) ReadStream(
	ctx context.Context,
	streamID string,
	fromVersion int64,
) ([]*EventData, error) {
	opts := esdb.ReadStreamOptions{
		From: esdb.Revision{Value: uint64(fromVersion)},
	}

	stream, err := c.conn.ReadStream(ctx, streamID, opts, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to read stream: %w", err)
	}
	defer stream.Close()

	var events []*EventData

	for {
		event := stream.Next()
		if event == nil {
			break
		}

		if event.Err != nil {
			return nil, fmt.Errorf("stream read error: %w", event.Err)
		}

		ed := &EventData{
			ID:        event.Event.EventID.String(),
			Type:      event.Event.EventType,
			Version:   int64(event.Event.EventNumber.Value),
			Data:      event.Event.Data,
			CreatedAt: event.Event.CreatedDate,
		}

		if err := json.Unmarshal(event.Event.Metadata, &ed.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		events = append(events, ed)
	}

	return events, nil
}

// Subscribe — Subscribiere auf einen Stream (Catch-up Subscription)
func (c *Client) Subscribe(
	ctx context.Context,
	streamName string,
	handler EventHandler,
) error {
	opts := esdb.SubscribeToStreamOptions{}

	sub, err := c.conn.SubscribeToStream(ctx, streamName, opts)
	if err != nil {
		return fmt.Errorf("subscription failed: %w", err)
	}
	defer sub.Close()

	for {
		message := sub.Next()

		switch m := message.(type) {
		case *esdb.EventAppeared:
			ed := &EventData{
				ID:        m.Event.EventID.String(),
				Type:      m.Event.EventType,
				Version:   int64(m.Event.EventNumber.Value),
				Data:      m.Event.Data,
				CreatedAt: m.Event.CreatedDate,
			}

			if err := json.Unmarshal(m.Event.Metadata, &ed.Metadata); err != nil {
				return fmt.Errorf("failed to unmarshal metadata: %w", err)
			}

			if err := handler(ctx, ed); err != nil {
				return fmt.Errorf("handler failed: %w", err)
			}

		case *esdb.SubscriptionConfirmed:
			// Subscription established
		case *esdb.SubscriptionDropped:
			return fmt.Errorf("subscription dropped: %v", m.Error)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
}

// Ping — Health Check
func (c *Client) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.conn.Ping(ctx)
}

// Close — Cleanup
func (c *Client) Close() error {
	return c.conn.Close()
}
```

---

## 6. pkg/common/database — PostgreSQL Helper

```go
// pkg/common/database/postgres.go
package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"rentflow/pkg/common/config"
)

type PostgresConnection struct {
	pool *pgxpool.Pool
}

func NewPostgresConnection(cfg *config.DatabaseConfig) (*PostgresConnection, error) {
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s&application_name=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		cfg.SSLMode,
		"rentflow-service",
	)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pool config: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.MaxConnections)
	poolConfig.MinConns = int32(cfg.MinConnections)
	poolConfig.MaxConnLifetime = cfg.ConnectionTTL

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	return &PostgresConnection{pool: pool}, nil
}

// SetTenantContext — Setze app.current_tenant_id für RLS
func (p *PostgresConnection) SetTenantContext(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, "tenant_id", tenantID)
}

// ExecuteWithTenant — Query mit Tenant RLS
func (p *PostgresConnection) ExecuteWithTenant(
	ctx context.Context,
	query string,
	args ...interface{},
) error {
	tenantID := ctx.Value("tenant_id").(string)

	conn, err := p.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "SET app.current_tenant_id = $1", tenantID); err != nil {
		return err
	}

	_, err = conn.Exec(ctx, query, args...)
	return err
}

// QueryWithTenant — Query mit RLS
func (p *PostgresConnection) QueryWithTenant(
	ctx context.Context,
	query string,
	args ...interface{},
) (*pgx.Rows, error) {
	tenantID := ctx.Value("tenant_id").(string)

	conn, err := p.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "SET app.current_tenant_id = $1", tenantID); err != nil {
		return nil, err
	}

	return conn.Query(ctx, query, args...)
}

func (p *PostgresConnection) Close() error {
	p.pool.Close()
	return nil
}
```

---

Continued in Teil 3 (API Gateway, Inter-Service Communication, Deployment)...