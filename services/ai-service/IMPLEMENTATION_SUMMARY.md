# AI-Service Microservice - Implementation Summary

## Overview
Complete Go microservice for the RentFlow project implementing AI capabilities with multi-provider architecture, anonymization, and predictive analytics.

**Location:** `/sessions/beautiful-adoring-keller/mnt/rentflow/services/ai-service/`  
**Port:** 8013  
**Pattern:** Hexagonal Architecture (same as crew-service)

---

## File Structure

```
services/ai-service/
├── cmd/server/main.go                                  (Entry point, server setup)
├── go.mod                                               (Go module definition)
├── go.sum                                               (Dependency lock file)
├── Dockerfile                                           (Multi-stage build)
├── migrations/
│   └── 001_ai_service.sql                              (Database schema)
└── internal/
    ├── domain/
    │   ├── entities.go       (AIRequest, AIFeedback, FewShotExample, AIProvider)
    │   └── errors.go         (Domain-specific errors)
    ├── ports/
    │   └── repositories.go   (Repository interfaces)
    ├── application/
    │   ├── commands.go                      (Command DTOs)
    │   ├── dto.go                           (Response DTOs)
    │   ├── ai_service.go                    (Main orchestrator service)
    │   ├── provider_service.go              (Multi-provider support: Claude, GPT-4o, Gemini, Mistral, Ollama)
    │   ├── anonymization_service.go         (7 PII patterns: Name, Email, IBAN, Phone, Address, TaxID, CustomerID)
    │   └── prediction_service.go            (Price optimization, demand forecast, asset creation, predictive maintenance)
    ├── infrastructure/
    │   └── repositories/
    │       ├── ai_request_postgres.go       (AIRequest repository)
    │       ├── ai_feedback_postgres.go      (AIFeedback repository)
    │       ├── few_shot_postgres.go         (FewShotExample repository)
    │       └── ai_provider_postgres.go      (AIProvider repository)
    └── adapters/http/
        ├── handlers.go                      (HTTP request handlers)
        └── router.go                        (Route setup)
```

---

## Database Schema (001_ai_service.sql)

### Tables
1. **ai_providers**
   - Multi-provider support (Claude, GPT-4o, Gemini, Mistral, Ollama)
   - Per-tenant configuration with priority fallback
   - JSONB config field for provider-specific settings

2. **ai_requests**
   - Tracks all AI processing requests
   - Supports 5 request types: price_optimization, demand_forecast, asset_creator, predictive_maintenance, general
   - Stores anonymized input, output, token usage, and latency
   - Status tracking: pending, processing, completed, failed

3. **ai_feedback**
   - User feedback on AI responses (1-5 rating)
   - Correctness tracking for few-shot learning
   - Comments for qualitative feedback

4. **few_shot_examples**
   - Few-shot learning examples per request type
   - Usage tracking for analytics
   - Average rating calculation

### Indices
Performance indices on:
- tenant_id, provider_id, created_at, status
- request_id, is_active
- request_type with is_active filter

---

## Hexagonal Architecture Layers

### 1. Domain Layer
**Location:** `internal/domain/`

**Entities:**
- `AIRequest`: Core domain model for AI requests
- `AIFeedback`: User feedback on responses
- `FewShotExample`: Learning examples
- `AIProvider`: Provider configuration
- `RequestType`: price_optimization, demand_forecast, asset_creator, predictive_maintenance, general
- `ProviderType`: claude, gpt4o, gemini, mistral, ollama
- `PatternType`: 7 PII patterns (name, email, iban, phone, address, tax_id, customer_id)

**Domain Errors:**
- ErrAIRequestNotFound, ErrProviderNotAvailable, ErrNoProviders
- ErrInvalidRating, ErrAnonymizationFailed, ErrAIProviderError

### 2. Ports (Interfaces)
**Location:** `internal/ports/repositories.go`

Defines contracts for data access:
- `AIRequestRepository`
- `AIFeedbackRepository`
- `FewShotExampleRepository`
- `AIProviderRepository`

### 3. Application Layer
**Location:** `internal/application/`

**Main Service (ai_service.go):**
- Orchestrates all AI operations
- Manages requests, feedback, examples, and providers
- Coordinates between other services

**Provider Service (provider_service.go):**
- Multi-provider abstraction with interface `AIProvider`
- 5 implementations:
  - `ClaudeProvider`: Anthropic Claude
  - `GPT4oProvider`: OpenAI GPT-4o
  - `GeminiProvider`: Google Gemini
  - `MistralProvider`: Mistral AI
  - `OllamaProvider`: Local Ollama
- Provider routing and fallback

**Anonymization Service (anonymization_service.go):**
- 7 regex patterns for PII detection:
  1. Name: German name patterns → [NAME]
  2. Email: Standard regex → [EMAIL]
  3. IBAN: DE\d{20} → [IBAN]
  4. Phone: German formats → [PHONE]
  5. Address: Street patterns → [ADDRESS]
  6. Tax ID: Steuernummer patterns → [TAX_ID]
  7. Customer ID: K-\d{4,8} → [CUSTOMER_ID]
- Anonymization with reporting

**Prediction Service (prediction_service.go):**
- **PriceOptimization**: Equipment type, rental days, season → price recommendation
- **DemandForecast**: Category, period → demand prediction
- **SmartAssetCreator**: Description → asset metadata suggestions
- **PredictiveMaintenance**: Equipment usage → maintenance recommendations

**DTOs (dto.go):**
- AIRequestDTO, AIFeedbackDTO, FewShotExampleDTO, AIProviderDTO
- DashboardDTO with metrics
- Conversion functions: ToAIRequestDTO, ToAIFeedbackDTO, etc.

**Commands (commands.go):**
- CreateAIRequestCommand
- SubmitFeedbackCommand
- CreateFewShotExampleCommand
- CreateAIProviderCommand
- AnonymizeCommand
- PriceOptimizationCommand
- DemandForecastCommand
- SmartAssetCreatorCommand
- PredictiveMaintenanceCommand

### 4. Infrastructure Layer
**Location:** `internal/infrastructure/repositories/`

PostgreSQL implementations:
- `PostgresAIRequestRepository`: CRUD + List with pagination
- `PostgresAIFeedbackRepository`: Save, retrieve by request
- `PostgresFewShotExampleRepository`: List by type, increment usage count
- `PostgresAIProviderRepository`: Active provider filtering

### 5. Adapters Layer
**Location:** `internal/adapters/http/`

**Handlers (handlers.go):**
- 14 HTTP handlers covering all endpoints
- JSON request/response handling
- Tenant ID extraction from headers (X-Tenant-ID)

**Router (router.go):**
- Route setup for all 14 endpoints
- HTTP method and path patterns

---

## API Endpoints

### AI Requests
- `POST /api/v1/ai/complete` - Create and process AI request
- `GET /api/v1/ai/requests` - List requests (paginated)
- `GET /api/v1/ai/requests/{id}` - Get single request

### Feedback
- `POST /api/v1/ai/feedback` - Submit feedback (1-5 rating)

### Anonymization
- `POST /api/v1/ai/anonymize` - Anonymize text (7 PII patterns)

### Predictions
- `POST /api/v1/ai/price-optimize` - Price optimization recommendation
- `POST /api/v1/ai/demand-forecast` - Demand forecasting
- `POST /api/v1/ai/asset-create` - Smart asset creation
- `POST /api/v1/ai/predict-maintenance` - Predictive maintenance

### Few-Shot Learning
- `GET /api/v1/ai/few-shots` - List examples (paginated)
- `POST /api/v1/ai/few-shots` - Create example

### Providers
- `GET /api/v1/ai/providers` - List providers for tenant
- `POST /api/v1/ai/providers` - Register new provider

### Dashboard
- `GET /api/v1/ai/dashboard` - Statistics and metrics

### Health
- `GET /health` - Service health check
- `GET /ready` - Readiness check (database ping)

---

## Key Implementation Details

### Multi-Provider Architecture
Each provider implements the `AIProvider` interface:
```go
type AIProvider interface {
    Name() string
    Complete(ctx context.Context, req ProviderRequest) (*ProviderResponse, error)
    IsAvailable() bool
}
```

Stub implementations for all 5 providers. Production: Replace with actual API calls.

### Anonymization
7 regex patterns detect and replace PII before sending to AI providers:
- Supports German naming conventions
- IBAN, phone, address detection
- Tax ID and customer ID patterns
- Configurable replacement tokens

### Prediction Services
4 stub methods prepare detailed prompts for AI providers:
- **Price Optimization**: Equipment type + rental days + season → price range
- **Demand Forecast**: Category + period → demand level + inventory recommendations
- **Smart Asset Creator**: Description → category, pricing, maintenance
- **Predictive Maintenance**: Equipment usage → maintenance timeline + cost

### Hexagonal Architecture Features
- Clean separation of concerns (domain, application, infrastructure)
- Port-based dependency injection
- Repository pattern for data access
- HTTP adapters for external communication
- Domain-driven entities
- DTOs for API contracts

---

## Dependencies

```go
github.com/google/uuid v1.6.0        // UUID generation
github.com/lib/pq v1.12.0            // PostgreSQL driver
github.com/jeckersberger/rentflow/pkg/common v0.0.0  // Shared utilities (logger, config)
```

---

## Configuration

### Environment Variables
- `DATABASE_URL`: PostgreSQL connection string (from config.Load())
- `SERVICE_PORT`: Override default port 8013
- `LOG_LEVEL`: Logging level

### Logging
- Uses `github.com/rs/zerolog` via pkg/common/logger
- Structured logging throughout
- Service name: "ai-service"

---

## Building & Running

### Docker Build
```bash
docker build -t ai-service:latest .
```

### Local Build (requires Go 1.24)
```bash
GOWORK=off go build ./cmd/server
```

### Run
```bash
./server
# Service listens on port 8013
# Health check: curl http://localhost:8013/health
```

---

## Testing Endpoints

### Example: Create AI Request
```bash
curl -X POST http://localhost:8013/api/v1/ai/complete \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: tenant-123" \
  -d '{
    "provider_id": "provider-uuid",
    "request_type": "price_optimization",
    "input_text": "Equipment type: crane, Rental days: 7, Season: summer",
    "anonymize": true
  }'
```

### Example: Submit Feedback
```bash
curl -X POST http://localhost:8013/api/v1/ai/feedback \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "tenant-123",
    "request_id": "request-uuid",
    "rating": 5,
    "is_correct": true,
    "comment": "Excellent pricing recommendation"
  }'
```

### Example: Anonymize Text
```bash
curl -X POST http://localhost:8013/api/v1/ai/anonymize \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "tenant-123",
    "text": "Contact Max Müller at max.mueller@example.com, IBAN DE12345678901234567890"
  }'
```

---

## Production Considerations

### Provider Implementation
All 5 providers are stubs returning mock data. For production:
1. Implement actual HTTP calls to respective APIs
2. Add API key management
3. Implement rate limiting and retry logic
4. Add request/response logging
5. Handle provider-specific error codes

### Anonymization Enhancements
- Additional pattern types (credit cards, bank accounts)
- Machine learning-based NER for entity detection
- Audit logging of anonymization actions

### Analytics & Monitoring
- Dashboard queries for request/day, average latency, success rate
- Provider performance tracking
- Feedback rating trends
- Few-shot example effectiveness

### Error Handling
- Comprehensive error messages
- Provider fallback chain
- Graceful degradation
- Circuit breaker pattern for failing providers

---

## Compliance & Security

- GDPR-compliant anonymization before external API calls
- Tenant isolation (all queries filtered by tenant_id)
- Audit trail for feedback and requests
- No credentials stored in code (via pkg/common/config)
- Request/response logging for debugging

---

## Notes

- All code follows hexagonal architecture pattern (matching crew-service)
- Database schema uses TIMESTAMPTZ for UTC timestamps
- Repositories use context for cancellation support
- Services follow dependency injection pattern
- HTTP handlers use Go 1.22+ route patterns
- Connection pooling configured (25 max, 5 idle)
- Graceful shutdown with 30-second timeout
