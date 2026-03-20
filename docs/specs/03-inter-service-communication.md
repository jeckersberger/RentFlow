# Inter-Service Communication & API Gateway — RentFlow

**Stand:** 20. März 2026
**Zielgruppe:** Go-Entwickler, DevOps, Architekten

---

## 1. Synchrone Calls — HTTP Client mit Circuit Breaker

### 1.1 Service Client Pattern

```go
// pkg/common/client/service_client.go
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"rentflow/pkg/common/errors"
)

type ServiceClient struct {
	baseURL      string
	httpClient   *http.Client
	circuitBreaker *CircuitBreaker
	logger       *zerolog.Logger
}

type CircuitBreaker struct {
	maxFailures      int
	resetTimeout     time.Duration
	state            string // "closed", "open", "half-open"
	failureCount     int
	lastFailureTime  time.Time
	successCount     int
}

func NewServiceClient(
	baseURL string,
	timeout time.Duration,
	logger *zerolog.Logger,
) *ServiceClient {
	return &ServiceClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		circuitBreaker: &CircuitBreaker{
			maxFailures:  5,
			resetTimeout: 30 * time.Second,
			state:        "closed",
		},
		logger: logger,
	}
}

// GET mit Retry & Circuit Breaker
func (sc *ServiceClient) Get(
	ctx context.Context,
	path string,
	headers map[string]string,
	result interface{},
) error {
	return sc.doWithRetry(ctx, "GET", path, headers, nil, result)
}

// POST mit Retry & Circuit Breaker
func (sc *ServiceClient) Post(
	ctx context.Context,
	path string,
	headers map[string]string,
	body interface{},
	result interface{},
) error {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}
	return sc.doWithRetry(ctx, "POST", path, headers, bodyBytes, result)
}

// doWithRetry — Exponential Backoff + Circuit Breaker
func (sc *ServiceClient) doWithRetry(
	ctx context.Context,
	method, path string,
	headers map[string]string,
	body []byte,
	result interface{},
) error {
	// Check Circuit Breaker
	if sc.circuitBreaker.state == "open" {
		if time.Since(sc.circuitBreaker.lastFailureTime) > sc.circuitBreaker.resetTimeout {
			sc.circuitBreaker.state = "half-open"
			sc.circuitBreaker.successCount = 0
		} else {
			return errors.NewError(
				errors.ErrServiceUnavailable,
				fmt.Sprintf("service %s is unavailable", sc.baseURL),
			)
		}
	}

	// Retry Logic
	maxAttempts := 3
	backoff := 100 * time.Millisecond

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err := sc.do(ctx, method, path, headers, body, result)
		if err == nil {
			// Success — Circuit Breaker reset
			if sc.circuitBreaker.state == "half-open" {
				sc.circuitBreaker.successCount++
				if sc.circuitBreaker.successCount >= 3 {
					sc.circuitBreaker.state = "closed"
					sc.circuitBreaker.failureCount = 0
				}
			}
			return nil
		}

		// Check if error is retryable
		if !isRetryable(err) {
			return err
		}

		if attempt < maxAttempts {
			select {
			case <-time.After(backoff):
				backoff *= 2
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	// All retries failed
	sc.circuitBreaker.failureCount++
	sc.circuitBreaker.lastFailureTime = time.Now()
	if sc.circuitBreaker.failureCount >= sc.circuitBreaker.maxFailures {
		sc.circuitBreaker.state = "open"
	}

	return fmt.Errorf("all retry attempts failed for %s %s", method, path)
}

func (sc *ServiceClient) do(
	ctx context.Context,
	method, path string,
	headers map[string]string,
	body []byte,
	result interface{},
) error {
	url := sc.baseURL + path

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set Headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", ctx.Value("request_id").(string))
	req.Header.Set("X-Correlation-ID", ctx.Value("correlation_id").(string))

	// Forward Authorization
	if auth := ctx.Value("authorization"); auth != nil {
		req.Header.Set("Authorization", auth.(string))
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if body != nil {
		req.Body = io.NopCloser(bytes.NewReader(body))
	}

	resp, err := sc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read Response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check Status Code
	if resp.StatusCode >= 400 {
		var errResp errors.Error
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			return &errResp
		}
		return fmt.Errorf("service returned %d: %s", resp.StatusCode, string(respBody))
	}

	// Unmarshal Result
	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

func isRetryable(err error) bool {
	// Don't retry client errors (4xx)
	if svcErr, ok := err.(*errors.Error); ok {
		return svcErr.Code != errors.ErrValidation &&
			svcErr.Code != errors.ErrUnauthorized &&
			svcErr.Code != errors.ErrForbidden
	}
	// Retry network errors
	return true
}
```

### 1.2 Dependency Injection für Service Clients

```go
// cmd/server/wire.go — Service Client Providers

func newInventoryServiceClient(cfg *config.Config) *client.ServiceClient {
	return client.NewServiceClient(
		"http://inventory-service:8002",
		5*time.Second,
		logger,
	)
}

func newProjectServiceClient(cfg *config.Config) *client.ServiceClient {
	return client.NewServiceClient(
		"http://project-service:8003",
		5*time.Second,
		logger,
	)
}

// In Application Service:
type ScannerService struct {
	inventoryClient *client.ServiceClient
	projectClient   *client.ServiceClient
}

func NewScannerService(
	inventoryClient *client.ServiceClient,
	projectClient *client.ServiceClient,
) *ScannerService {
	return &ScannerService{
		inventoryClient: inventoryClient,
		projectClient:   projectClient,
	}
}

// Beispiel: Scanner reserviert Equipment
func (s *ScannerService) ReserveEquipment(
	ctx context.Context,
	equipmentID string,
	quantity int,
) error {
	// Synchroner Call zu inventory-service
	req := struct {
		EquipmentID string `json:"equipment_id"`
		Quantity    int    `json:"quantity"`
	}{
		EquipmentID: equipmentID,
		Quantity:    quantity,
	}

	var resp struct {
		Reserved int `json:"reserved"`
	}

	return s.inventoryClient.Post(ctx, "/reserve", nil, req, &resp)
}
```

---

## 2. Asynchrone Events — KurrentDB Subscriptions

### 2.1 Event-Driven Choreography Pattern

```
Beispiel: Projekt erstellen → Equipment reservieren → Packliste generieren

┌─────────────────┐
│  project-service │
│  CreateProject  │
└────────┬────────┘
         │ Publishes: ProjectCreated
         │           (stream: project-{uuid})
         ▼
    KurrentDB
    (Event Store)
         │
    ┌────┴─────┐
    │           │
    ▼           ▼
+─────────────────────+  +──────────────────+
│ inventory-service   │  │ notification-svc │
│ Subscribes:         │  │ Subscribes:      │
│ $ce-project         │  │ $ce-project      │
│ → Reserves Equipment│  │ → Send Alerts    │
│ → Publishes         │  └──────────────────+
│   EquipmentReserved │
└─────────────────────+
         │
         ▼
    KurrentDB
    (event: EquipmentReserved)
         │
         ▼
+──────────────────────+
│ document-service    │
│ Subscribes:         │
│ EquipmentReserved   │
│ → Generate Packing  │
│   List              │
└──────────────────────+
```

**Advantages of Event Choreography:**
- Lose Kopplung zwischen Services
- Asynchrone Verarbeitung
- Bessere Fehlertoleranz (Retry via Dead Letter Queue)
- Natural Event Audit Trail

### 2.2 Subscription Implementation

```go
// pkg/common/events/subscriber.go
package events

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Subscriber struct {
	eventClient *Client
	pgPool      *pgxpool.Pool
	logger      *zerolog.Logger
	retryConfig RetryConfig
}

type RetryConfig struct {
	MaxAttempts int
	InitialBackoff time.Duration
	MaxBackoff time.Duration
}

func NewSubscriber(
	eventClient *Client,
	pgPool *pgxpool.Pool,
	logger *zerolog.Logger,
) *Subscriber {
	return &Subscriber{
		eventClient: eventClient,
		pgPool:      pgPool,
		logger:      logger,
		retryConfig: RetryConfig{
			MaxAttempts: 3,
			InitialBackoff: 100 * time.Millisecond,
			MaxBackoff: 30 * time.Second,
		},
	}
}

type EventHandler func(ctx context.Context, event *EventData) error

// StartSubscription mit Restart-Loop
func (s *Subscriber) StartSubscription(
	ctx context.Context,
	streamName string,
	handler EventHandler,
) {
	backoff := time.Second

	for {
		select {
		case <-ctx.Done():
			s.logger.Info().Msg("Subscription context cancelled")
			return
		default:
		}

		s.logger.Info().Str("stream", streamName).Msg("Starting subscription")

		err := s.eventClient.Subscribe(ctx, streamName, func(ctx context.Context, event *EventData) error {
			return s.handleWithRetry(ctx, event, handler)
		})

		if err != nil {
			s.logger.Error().Err(err).Msg("Subscription failed, reconnecting...")
			time.Sleep(backoff)
			backoff *= 2
			if backoff > 30*time.Second {
				backoff = 30 * time.Second
			}
			continue
		}

		backoff = time.Second // Reset on successful reconnect
	}
}

// handleWithRetry mit Idempotenz-Check
func (s *Subscriber) handleWithRetry(
	ctx context.Context,
	event *EventData,
	handler EventHandler,
) error {
	// 1. Check if already processed (Idempotenz)
	query := `
	SELECT event_id FROM processed_events
	WHERE tenant_id = $1 AND event_id = $2
	`

	var existingID string
	err := s.pgPool.QueryRow(ctx, query, event.Metadata.TenantID, event.ID).Scan(&existingID)
	if err == nil {
		s.logger.Debug().Str("event_id", event.ID).Msg("Event already processed, skipping")
		return nil
	}

	// 2. Process with Retry
	var lastErr error
	backoff := s.retryConfig.InitialBackoff

	for attempt := 1; attempt <= s.retryConfig.MaxAttempts; attempt++ {
		err := handler(ctx, event)
		if err == nil {
			// 3. Mark as processed
			markQuery := `
			INSERT INTO processed_events
			(tenant_id, event_id, event_type, processed_at)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT DO NOTHING
			`
			s.pgPool.Exec(ctx, markQuery,
				event.Metadata.TenantID,
				event.ID,
				event.Type,
				time.Now(),
			)
			return nil
		}

		lastErr = err
		s.logger.Warn().
			Err(err).
			Str("event_id", event.ID).
			Int("attempt", attempt).
			Msg("Event handler failed, retrying...")

		if attempt < s.retryConfig.MaxAttempts {
			time.Sleep(backoff)
			backoff *= 2
			if backoff > s.retryConfig.MaxBackoff {
				backoff = s.retryConfig.MaxBackoff
			}
		}
	}

	// 4. All retries failed → Dead Letter Queue
	s.sendToDeadLetterQueue(ctx, event, lastErr)
	return lastErr
}

func (s *Subscriber) sendToDeadLetterQueue(
	ctx context.Context,
	event *EventData,
	err error,
) {
	dlqQuery := `
	INSERT INTO dead_letters
	(event_id, event_type, tenant_id, error_msg, correlation_id, created_at)
	VALUES ($1, $2, $3, $4, $5, $6)
	`

	s.pgPool.Exec(ctx, dlqQuery,
		event.ID,
		event.Type,
		event.Metadata.TenantID,
		err.Error(),
		event.Metadata.CorrelationID,
		time.Now(),
	)

	s.logger.Error().
		Str("event_id", event.ID).
		Str("correlation_id", event.Metadata.CorrelationID).
		Err(err).
		Msg("Event moved to dead letter queue")
}
```

### 2.3 Konkrete Handler Beispiele

**inventory-service Handler für ProjectCreated:**

```go
// internal/application/handlers.go
func (s *InventoryService) HandleProjectCreated(
	ctx context.Context,
	event *events.EventData,
) error {
	// 1. Unmarshal Event
	var projectCreated struct {
		ProjectID string `json:"project_id"`
		TenantID  string `json:"tenant_id"`
		Name      string `json:"name"`
		StartDate time.Time `json:"start_date"`
		EndDate   time.Time `json:"end_date"`
	}

	if err := json.Unmarshal(event.Data, &projectCreated); err != nil {
		return fmt.Errorf("failed to unmarshal: %w", err)
	}

	// 2. Business Logic: Auto-Reserve Equipment based on Projekt Type
	// (Beispiel: 'theatre' project gets standard equipment bundle)

	// 3. Publish EquipmentReserved Event(s)
	for _, equipID := range getBundleEquipment(projectCreated) {
		err := s.eventStore.AppendEvent(ctx,
			fmt.Sprintf("equipment-%s", equipID),
			&domain.EquipmentReserved{
				EventID:       uuid.New().String(),
				EquipmentID:   equipID,
				TenantID:      projectCreated.TenantID,
				ProjectID:     projectCreated.ProjectID,
				Quantity:      1,
				ReservedAt:    time.Now(),
				CorrelationID: event.Metadata.CorrelationID,
			},
			events.Metadata{
				TenantID:      projectCreated.TenantID,
				CorrelationID: event.Metadata.CorrelationID,
			},
			-1,
		)
		if err != nil {
			return fmt.Errorf("failed to reserve equipment: %w", err)
		}
	}

	return nil
}
```

---

## 3. Saga Pattern — Verteilte Transaktionen

### 3.1 Orchestration Pattern (Centralized Saga)

```
project-service ist der SAGA ORCHESTRATOR
       ↓
   1. ProjectCreated Event
       ↓
   2. Call: inventory-service/reserve (Synchron)
       ↓
   IF reservation OK:
       ↓
   3. Call: document-service/generate-packlist
       ↓
   IF generation OK:
       ↓
   4. Publish: ProjectSetupCompleted
       ↓
   ELSE:
       ↓
   4. Publish: ProjectSetupFailed → Rollback
```

**Synchrone Saga Orchestration in project-service:**

```go
// internal/application/saga/project_setup_saga.go
package saga

import (
	"context"
	"fmt"
	"rentflow/pkg/common/client"
)

type ProjectSetupSaga struct {
	inventoryClient  *client.ServiceClient
	documentClient   *client.ServiceClient
	notificationSvc  *NotificationService
}

type ProjectSetupCommand struct {
	ProjectID   string
	TenantID    string
	Name        string
	ClientName  string
	EquipmentIDs []string
}

func (s *ProjectSetupSaga) Execute(
	ctx context.Context,
	cmd ProjectSetupCommand,
) error {
	// Step 1: Reserve Equipment
	reserveReq := struct {
		EquipmentIDs []string `json:"equipment_ids"`
		ProjectID    string   `json:"project_id"`
	}{
		EquipmentIDs: cmd.EquipmentIDs,
		ProjectID:    cmd.ProjectID,
	}

	var reserveResp struct {
		Reserved int `json:"reserved"`
	}

	err := s.inventoryClient.Post(ctx, "/reserve", nil, reserveReq, &reserveResp)
	if err != nil {
		// Compensating Transaction: Cleanup
		s.compensate(ctx, "reservation_failed", cmd)
		return fmt.Errorf("reservation failed: %w", err)
	}

	// Step 2: Generate Packlist
	packlistReq := struct {
		ProjectID string `json:"project_id"`
		Items     int    `json:"items"`
	}{
		ProjectID: cmd.ProjectID,
		Items:     reserveResp.Reserved,
	}

	var packlistResp struct {
		PacklistID string `json:"packlist_id"`
	}

	err = s.documentClient.Post(ctx, "/packlists", nil, packlistReq, &packlistResp)
	if err != nil {
		// Compensating Transaction
		s.compensate(ctx, "packlist_failed", cmd)
		return fmt.Errorf("packlist generation failed: %w", err)
	}

	// Step 3: Success — Publish ProjectSetupCompleted Event
	setupCompleted := struct {
		ProjectID  string `json:"project_id"`
		PacklistID string `json:"packlist_id"`
		Status     string `json:"status"`
	}{
		ProjectID:  cmd.ProjectID,
		PacklistID: packlistResp.PacklistID,
		Status:     "setup_completed",
	}

	return s.eventStore.AppendEvent(ctx, fmt.Sprintf("project-%s", cmd.ProjectID),
		setupCompleted,
		events.Metadata{TenantID: cmd.TenantID, CorrelationID: ctx.Value("correlation_id").(string)},
		-1)
}

// Compensating Transactions (Rollback)
func (s *ProjectSetupSaga) compensate(
	ctx context.Context,
	reason string,
	cmd ProjectSetupCommand,
) error {
	// Call inventory-service to release reservations
	releaseReq := struct {
		ProjectID string `json:"project_id"`
		Reason    string `json:"reason"`
	}{
		ProjectID: cmd.ProjectID,
		Reason:    reason,
	}

	return s.inventoryClient.Post(ctx, "/release", nil, releaseReq, nil)
}
```

### 3.2 Event Choreography Pattern (Decentralized)

```
ProjectCreated
    ↓
inventory-service:
    Subscribes: ProjectCreated
    → Reserve Equipment
    → Publish: EquipmentReserved
    ↓
document-service:
    Subscribes: EquipmentReserved
    → Generate Packlist
    → Publish: PacklistGenerated
    ↓
notification-service:
    Subscribes: PacklistGenerated
    → Send Notification
```

**inventory-service (Event Choreography):**

```go
// Inherent: Lose Kopplung, aber schwerer zu debuggen
// Vorteil: Keine zentrale Abhängigkeit

func (s *InventoryService) OnProjectCreated(ctx context.Context, event *events.EventData) error {
	// 1. Reserve equipment
	// 2. Publish EquipmentReserved
	// → No direct knowledge of document-service
}
```

---

## 4. Correlation-ID Propagation

```go
// Alle Events und Requests müssen correlation_id weitergeben

// In Event:
type Metadata struct {
	TenantID      string `json:"tenant_id"`
	UserID        string `json:"user_id"`
	CorrelationID string `json:"correlation_id"` // Propagated durch alle Services
	CausationID   string `json:"causation_id"`   // Direkter Parent-Event
}

// In HTTP Headers:
req.Header.Set("X-Correlation-ID", ctx.Value("correlation_id").(string))

// Tracing Query:
// SELECT * FROM all_logs WHERE correlation_id = 'xyz-123'
// → Sieht alle Events/Logs für eine Business-Operation über alle Services
```

---

## 5. Traefik v3 — API Gateway Configuration

### 5.1 Static Configuration

**traefik.yml:**

```yaml
# Global Config
global:
  checkNewVersion: false
  sendAnonymousUsage: false

# EntryPoints
entryPoints:
  web:
    address: ":80"
    http:
      redirections:
        entrypoint:
          to: websecure
          scheme: https
  websecure:
    address: ":443"
    http:
      tls:
        certResolver: letsencrypt
        options: default

api:
  insecure: false
  dashboard: true  # https://traefik.local:8080

# Metrics
metrics:
  prometheus:
    addEntryPointsLabels: true
    addServicesLabels: true

# Let's Encrypt
certificatesResolvers:
  letsencrypt:
    acme:
      email: admin@rentflow.local
      storage: acme.json
      httpChallenge:
        entryPoint: web

# Providers
providers:
  docker:
    endpoint: unix:///var/run/docker.sock
    exposedByDefault: false
    swarmMode: false
    network: rentflow_default

# Logging
log:
  level: INFO
  format: json

accessLog:
  format: json
  filePath: /var/log/traefik/access.log
```

### 5.2 Dynamic Configuration via Docker Labels

**docker-compose.yml Beispiel:**

```yaml
version: "3.9"

services:
  traefik:
    image: traefik:v3.0
    ports:
      - "80:80"
      - "443:443"
      - "8080:8080"  # Dashboard
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./traefik.yml:/traefik.yml
      - ./acme.json:/acme.json
    networks:
      - rentflow_network

  # INVENTORY SERVICE
  inventory-service:
    image: rentflow/inventory-service:latest
    ports:
      - "8002:8002"
    environment:
      INVENTORY_SERVER_PORT: 8002
      INVENTORY_DATABASE_HOST: postgres
    depends_on:
      - postgres
      - kurrentdb
    networks:
      - rentflow_network
    labels:
      traefik.enable: "true"
      # Router Definition
      traefik.http.routers.inventory.rule: "Host(`api.rentflow.local`) && PathPrefix(`/inventory`)"
      traefik.http.routers.inventory.entrypoints: "websecure"
      traefik.http.routers.inventory.tls: "true"
      # Service Definition
      traefik.http.services.inventory.loadbalancer.server.port: "8002"
      traefik.http.services.inventory.loadbalancer.healthcheck.path: "/health"
      traefik.http.services.inventory.loadbalancer.healthcheck.interval: "10s"
      # Middleware
      traefik.http.routers.inventory.middlewares: "auth@docker,ratelimit@docker,cors@docker"

  # PROJECT SERVICE
  project-service:
    image: rentflow/project-service:latest
    ports:
      - "8003:8003"
    environment:
      PROJECT_SERVER_PORT: 8003
    depends_on:
      - postgres
      - kurrentdb
    networks:
      - rentflow_network
    labels:
      traefik.enable: "true"
      traefik.http.routers.project.rule: "Host(`api.rentflow.local`) && PathPrefix(`/projects`)"
      traefik.http.routers.project.entrypoints: "websecure"
      traefik.http.routers.project.tls: "true"
      traefik.http.services.project.loadbalancer.server.port: "8003"
      traefik.http.services.project.loadbalancer.healthcheck.path: "/health"
      traefik.http.services.project.loadbalancer.healthcheck.interval: "10s"

  # POSTGRES
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: rentflow
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: rentflow
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - rentflow_network

  # KURRENTDB (ehemals EventStoreDB)
  kurrentdb:
    image: eventstore/eventstore:latest
    environment:
      EVENTSTORE_MEM_DB: "true"
      EVENTSTORE_EXT_IP: kurrentdb
    ports:
      - "2113:2113"
    networks:
      - rentflow_network

  # REDIS
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    networks:
      - rentflow_network

volumes:
  postgres_data:

networks:
  rentflow_network:
    driver: bridge
```

### 5.3 Middleware Definition

**Middleware in Traefik:**

```yaml
# traefik-middleware.yml (alternativ zu Labels)

---
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: auth-middleware
spec:
  forwardAuth:
    address: http://auth-service:8001/verify
    trustForwardHeader: true

---
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: ratelimit-middleware
spec:
  rateLimit:
    average: 100
    burst: 50
    period: 1s

---
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: cors-middleware
spec:
  headers:
    accessControlAllowOriginList:
      - https://rentflow.local
      - https://app.rentflow.local
    accessControlAllowMethods:
      - GET
      - POST
      - PUT
      - DELETE
      - OPTIONS
    accessControlAllowHeaders:
      - Content-Type
      - Authorization
      - X-Request-ID
      - X-Correlation-ID
    accessControlMaxAge: 3600

---
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: circuit-breaker-middleware
spec:
  circuitBreaker:
    expression: "NetworkErrorRatio() > 0.5 || ResponseCodeRatio(500, 600, 0, 600) > 0.5"
```

### 5.4 TLS Configuration

**Self-Signed für Development:**

```bash
# Generate Self-Signed Cert
openssl req -x509 -newkey rsa:4096 -nodes \
  -out cert.pem -keyout key.pem -days 365 \
  -subj "/CN=rentflow.local"

# In traefik.yml
entryPoints:
  websecure:
    address: ":443"
    http:
      tls:
        certificates:
          - certFile: /path/to/cert.pem
            keyFile: /path/to/key.pem
```

**Let's Encrypt für Production (via ACME):**

```yaml
# Bereits in Static Config
certificatesResolvers:
  letsencrypt:
    acme:
      email: admin@rentflow.local
      storage: /data/acme.json
      httpChallenge:
        entryPoint: web
```

---

Continued in Teil 4 (Deployment, Kubernetes, Production Setup)...