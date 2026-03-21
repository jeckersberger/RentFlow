# Inventory Service - File Index

## Quick Navigation

### Start Here
1. **[README.md](README.md)** - Quick start guide and feature overview
2. **[DELIVERY_SUMMARY.md](DELIVERY_SUMMARY.md)** - Complete delivery summary
3. **[COMPLETION_CHECKLIST.md](COMPLETION_CHECKLIST.md)** - Feature checklist

### Understanding the Service
1. **[ARCHITECTURE.md](ARCHITECTURE.md)** - Hexagonal architecture explanation
2. **[IMPLEMENTATION.md](IMPLEMENTATION.md)** - Feature details and business logic
3. **[API.md](API.md)** - Complete API reference with examples

### Source Code

#### Domain Layer (Core Business Logic)
- [internal/domain/equipment.go](internal/domain/equipment.go) - Equipment aggregate with status machine
- [internal/domain/category.go](internal/domain/category.go) - Equipment category with hierarchy
- [internal/domain/flightcase.go](internal/domain/flightcase.go) - Flightcase/container management
- [internal/domain/events.go](internal/domain/events.go) - Domain events for event sourcing
- [internal/domain/errors.go](internal/domain/errors.go) - Domain-specific errors

#### Application Layer (Use Cases)
- [internal/application/equipment_service.go](internal/application/equipment_service.go) - Equipment operations (11 operations)
- [internal/application/category_service.go](internal/application/category_service.go) - Category operations (5 operations)
- [internal/application/flightcase_service.go](internal/application/flightcase_service.go) - Flightcase operations (8 operations)
- [internal/application/dto.go](internal/application/dto.go) - Data Transfer Objects for API responses
- [internal/application/commands.go](internal/application/commands.go) - Command definitions for all operations
- [internal/application/queries.go](internal/application/queries.go) - Query definitions for filtering/searching

#### Ports Layer (Interface Definitions)
- [internal/ports/repository.go](internal/ports/repository.go) - Repository interfaces (EquipmentRepository, CategoryRepository, FlightcaseRepository)

#### Infrastructure Layer (Database)
- [internal/infrastructure/repositories/equipment_postgres.go](internal/infrastructure/repositories/equipment_postgres.go) - Equipment persistence
- [internal/infrastructure/repositories/category_postgres.go](internal/infrastructure/repositories/category_postgres.go) - Category persistence
- [internal/infrastructure/repositories/flightcase_postgres.go](internal/infrastructure/repositories/flightcase_postgres.go) - Flightcase persistence
- [internal/infrastructure/repositories/helpers.go](internal/infrastructure/repositories/helpers.go) - Database helper functions

#### Adapter Layer (HTTP REST API)
- [internal/adapters/http/router.go](internal/adapters/http/router.go) - Route definitions (25 endpoints)
- [internal/adapters/http/handlers.go](internal/adapters/http/handlers.go) - HTTP request handlers

#### Server
- [cmd/server/main.go](cmd/server/main.go) - Application entry point

### Database

#### Migrations
- [migrations/001_create_equipment.sql](migrations/001_create_equipment.sql) - Equipment table schema
- [migrations/002_create_categories.sql](migrations/002_create_categories.sql) - Categories table schema
- [migrations/003_create_flightcases.sql](migrations/003_create_flightcases.sql) - Flightcases table schema

### Configuration
- [go.mod](go.mod) - Go module definition with dependencies

## File Statistics

- **Total Files**: 28
- **Go Source Files**: 19
- **Database Migrations**: 3
- **Configuration Files**: 1
- **Documentation Files**: 5

**Lines of Code (approximate)**:
- Domain Layer: ~400 lines
- Application Layer: ~900 lines
- Infrastructure Layer: ~550 lines
- Adapter Layer: ~500 lines
- Main: ~80 lines
- Total: ~2,400 lines of production code

## Feature Mapping

### Equipment Management
**Files**: equipment_service.go, equipment.go, equipment_postgres.go, handlers.go (equipment methods)

Features:
- Create/Read/Update/Delete
- List with filtering (status, category, location)
- Search with full-text indexing
- Barcode lookup
- Status management with state machine
- Condition tracking
- Image management
- Custom fields support

### Category Management
**Files**: category_service.go, category.go, category_postgres.go, handlers.go (category methods)

Features:
- Create/Read/Update/Delete
- Hierarchical relationships (parent-child)
- Icon and color support
- Sorting

### Flightcase Management
**Files**: flightcase_service.go, flightcase.go, flightcase_postgres.go, handlers.go (flightcase methods)

Features:
- Create/Read/Update/Delete
- Item management (add/remove with quantities)
- Weight tracking
- Location management

### API Endpoints
**Files**: router.go, handlers.go

25 endpoints:
- 11 Equipment endpoints
- 5 Category endpoints
- 7 Flightcase endpoints
- 2 Health check endpoints

## Database Schema

**Equipment Table**
- File: migrations/001_create_equipment.sql
- Implementation: equipment_postgres.go
- 27 columns, 7 indexes, full-text search

**Categories Table**
- File: migrations/002_create_categories.sql
- Implementation: category_postgres.go
- Hierarchical support with parent_id FK

**Flightcases Table**
- File: migrations/003_create_flightcases.sql
- Implementation: flightcase_postgres.go
- Junction table for items with cascade delete

## Architecture Layers

### 1. Domain (internal/domain/)
- Equipment, Category, Flightcase aggregates
- Status machine and condition enums
- Domain events and errors
- Pure business logic, no external dependencies

### 2. Application (internal/application/)
- EquipmentService, CategoryService, FlightcaseService
- Command and Query objects
- Data Transfer Objects (DTOs)
- Use case coordination

### 3. Ports (internal/ports/)
- EquipmentRepository interface
- CategoryRepository interface
- FlightcaseRepository interface

### 4. Infrastructure (internal/infrastructure/)
- PostgreSQL repository implementations
- Database helper functions
- SQL query construction

### 5. Adapter (internal/adapters/http/)
- HTTP router with 25 endpoints
- Request handlers
- Response serialization
- Error mapping

## Getting Started

1. Read [README.md](README.md) for quick start
2. Review [API.md](API.md) for endpoint examples
3. Study [ARCHITECTURE.md](ARCHITECTURE.md) to understand design
4. Explore domain layer (internal/domain/) for business logic
5. Check application layer (internal/application/) for use cases
6. Review repository implementations (internal/infrastructure/) for persistence

## Common Tasks

### Add a new field to Equipment
1. Update equipment.go in domain layer
2. Update equipment_postgres.go for persistence
3. Update database migration
4. Update service and handlers

### Add a new endpoint
1. Add handler method in handlers.go
2. Register route in router.go
3. Document in API.md
4. Test with curl examples

### Modify database schema
1. Create new migration file
2. Update repository implementation
3. Update domain model if needed

### Change API response format
1. Modify DTOs in dto.go
2. Update handlers.go if needed
3. Update API.md documentation

## Dependencies

External:
- github.com/lib/pq - PostgreSQL driver

Internal:
- github.com/jeckersberger/rentflow/pkg/common - Shared utilities

Standard library only for HTTP, database, JSON, logging.

## Testing

The architecture supports:
- Unit tests: Domain logic (no database needed)
- Integration tests: Repositories (with test database)
- E2E tests: Full HTTP flow
- Mock repositories: Service testing without database

## Deployment

Files needed for deployment:
- Compiled binary (go build ./cmd/server)
- Database migrations (migrations/*.sql)
- Configuration (environment variables)
- Dockerfile (already exists)

## Monitoring Points

Health check: GET /health
Readiness check: GET /ready
Logging: Throughout application

## Documentation Links

- [README.md](README.md) - Overview
- [API.md](API.md) - Endpoint reference
- [ARCHITECTURE.md](ARCHITECTURE.md) - Design patterns
- [IMPLEMENTATION.md](IMPLEMENTATION.md) - Feature details
- [COMPLETION_CHECKLIST.md](COMPLETION_CHECKLIST.md) - Feature checklist
- [DELIVERY_SUMMARY.md](DELIVERY_SUMMARY.md) - Delivery summary
- [INDEX.md](INDEX.md) - This file

---

**Status**: Production Ready
**Last Updated**: March 21, 2026
**Version**: 1.0 Complete Implementation
