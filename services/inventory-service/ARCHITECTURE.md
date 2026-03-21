# Inventory Service - Hexagonal Architecture

## Project Structure

```
inventory-service/
├── cmd/
│   └── server/
│       └── main.go                  # Application entry point
├── internal/
│   ├── adapters/
│   │   └── http/
│   │       ├── handlers.go         # HTTP request handlers
│   │       └── router.go           # Route definitions
│   ├── application/
│   │   ├── equipment_service.go    # Equipment use cases
│   │   ├── category_service.go     # Category use cases
│   │   ├── flightcase_service.go   # Flightcase use cases
│   │   ├── dto.go                  # Data Transfer Objects
│   │   ├── commands.go             # Command definitions
│   │   └── queries.go              # Query definitions
│   ├── domain/
│   │   ├── equipment.go            # Equipment aggregate
│   │   ├── category.go             # Category entity
│   │   ├── flightcase.go           # Flightcase aggregate
│   │   ├── events.go               # Domain events
│   │   └── errors.go               # Domain errors
│   ├── infrastructure/
│   │   └── repositories/
│   │       ├── equipment_postgres.go   # Equipment persistence
│   │       ├── category_postgres.go    # Category persistence
│   │       ├── flightcase_postgres.go  # Flightcase persistence
│   │       └── helpers.go              # Helper functions
│   └── ports/
│       └── repository.go           # Repository interfaces
├── migrations/
│   ├── 001_create_equipment.sql
│   ├── 002_create_categories.sql
│   └── 003_create_flightcases.sql
├── go.mod                          # Go module definition
├── Dockerfile
├── IMPLEMENTATION.md               # Implementation details
└── ARCHITECTURE.md                 # This file
```

## Hexagonal Architecture Layers

### 1. Domain Layer
**Location**: `internal/domain/`

Pure business logic with no external dependencies. Contains:
- **Aggregates**: Equipment, Flightcase
- **Value Objects**: Dimensions, EquipmentStatus, EquipmentCondition
- **Domain Events**: All business events
- **Domain Errors**: Structured error definitions

Key characteristics:
- No database, HTTP, or external dependencies
- Business rules and validation
- State management and transitions

### 2. Application Layer
**Location**: `internal/application/`

Use case orchestration and coordination. Contains:
- **Services**: EquipmentService, CategoryService, FlightcaseService
- **Commands**: Input objects for operations
- **Queries**: Input objects for retrieving data
- **DTOs**: Output objects for API responses

Key characteristics:
- Coordinates domain and infrastructure layers
- Implements use cases
- Transforms domain objects to DTOs for API responses

### 3. Ports Layer
**Location**: `internal/ports/`

Interface definitions for outbound dependencies. Contains:
- **Repository Interfaces**: EquipmentRepository, CategoryRepository, FlightcaseRepository

Key characteristics:
- Defines contracts for infrastructure
- Enables dependency injection
- Allows for multiple implementations (test doubles, different DBs)

### 4. Infrastructure Layer
**Location**: `internal/infrastructure/`

Concrete implementations of ports. Contains:
- **PostgreSQL Repositories**: EquipmentPostgres, CategoryPostgres, FlightcasePostgres
- **Helper Functions**: JSONB conversion, array handling

Key characteristics:
- Database-specific implementation
- SQL queries and persistence logic
- Can be replaced without affecting domain or application

### 5. Adapter Layer
**Location**: `internal/adapters/http/`

External interface implementation. Contains:
- **HTTP Router**: Route definitions
- **HTTP Handlers**: Request parsing, response serialization
- **Error Mapping**: HTTP status code mapping

Key characteristics:
- Converts HTTP requests to commands/queries
- Converts responses to JSON
- Can be replaced with different transport layers (gRPC, etc.)

## Data Flow

### Request Flow
```
HTTP Request
    ↓
HTTP Handler (Adapter)
    ↓
Command/Query (Application)
    ↓
Service (Application)
    ↓
Domain Objects (Domain)
    ↓
Repository Interface (Ports)
    ↓
PostgreSQL Repository (Infrastructure)
    ↓
Database
```

### Response Flow
```
Database Query Result
    ↓
Domain Objects (Infrastructure)
    ↓
DTO Conversion (Application)
    ↓
JSON Response (Adapter)
    ↓
HTTP Response
```

## Dependency Direction

All dependencies point INWARD toward the domain:

```
HTTP Handlers (Adapter)
    ↓
Services & Commands (Application)
    ↓
Domain Objects (Domain)
    ↓
Repository Interfaces (Ports) ←← Infrastructure Implementations
```

The domain has NO dependencies on other layers. This allows:
- Easy testing (mock repositories)
- Multiple implementations (different databases, caches)
- Clear separation of concerns
- Future event sourcing integration

## Key Design Patterns

### 1. Dependency Injection
All services receive their dependencies through constructors:
```go
func NewEquipmentService(
    equipRepo EquipmentRepository,
    catRepo CategoryRepository,
    storage StorageAdapter,
    logger *logger.Logger,
) *EquipmentService
```

### 2. Repository Pattern
All data access through repository interfaces:
```go
type EquipmentRepository interface {
    Create(ctx context.Context, eq *Equipment) error
    GetByID(ctx context.Context, tenantID, id string) (*Equipment, error)
    // ... other methods
}
```

### 3. Value Objects
Immutable domain values (Status, Condition, Dimensions)

### 4. Aggregates
Equipment and Flightcase are aggregates with invariants:
- Equipment status transitions are validated
- Flightcase items maintain consistency

### 5. Domain Events
Events capture what happened:
- EquipmentCreated, EquipmentStatusChanged, etc.
- Prepared for event sourcing integration

## Tenant Isolation

All operations are tenant-isolated:
- X-Tenant-ID header required for all API calls
- All queries filter by tenant_id
- Database constraints ensure isolation
- No way to accidentally access another tenant's data

## Error Handling

Structured error types with codes:
```go
type DomainError struct {
    Code    string
    Message string
    Err     error
}
```

HTTP handlers map domain errors to appropriate status codes:
- 404 Not Found
- 400 Bad Request (validation)
- 409 Conflict (duplicate)
- 500 Internal Server Error

## Testing Strategy

The hexagonal architecture enables comprehensive testing:

### Unit Tests (Domain)
Test domain objects and business rules directly

### Integration Tests (Infrastructure)
Test database operations with test database

### Contract Tests (Repositories)
Test repository implementations against interfaces

### E2E Tests (Adapter)
Test full HTTP request/response cycle

Mock repositories can be injected for service testing without database.

## Future Extensions

The architecture supports adding:

### 1. Event Sourcing
- Domain events already defined
- Can replace write model with event store
- Maintains read model with projections

### 2. Additional Adapters
- gRPC adapter alongside HTTP
- Message queue adapter for async operations
- GraphQL adapter for more flexible queries

### 3. Caching Layer
- Cache repositories behind Cache interface
- Redis or other cache implementations

### 4. Search Integration
- Elasticsearch adapter for full-text search
- Behind search repository interface

### 5. Message Publishing
- Events published to message queue
- Other services subscribe to events
- Maintains eventual consistency

All these can be added without modifying domain or core business logic.
