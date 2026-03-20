# RentFlow Monorepo Structure - Phase 0 Foundation

This document describes the complete monorepo structure for the RentFlow project with Event Sourcing and CQRS architecture.

## Project Overview

RentFlow is built as a microservices architecture with:
- **18 Microservices** for different business domains
- **Event Sourcing** via KurrentDB for immutable event logs
- **CQRS** pattern for separation of concerns
- **Hexagonal Architecture** for each service
- **Shared Common Library** for cross-cutting concerns

## Directory Structure

```
rentflow/
├── services/                    # 18 microservices
│   ├── auth-service/           # Port 8001 - User authentication & authorization
│   ├── inventory-service/      # Port 8002 - Equipment & item management
│   ├── project-service/        # Port 8003 - Project lifecycle management
│   ├── scanner-service/        # Port 8004 - QR/barcode scanning
│   ├── warehouse-service/      # Port 8005 - Warehouse operations
│   ├── invoice-service/        # Port 8006 - Invoice generation & management
│   ├── document-service/       # Port 8007 - Document storage & retrieval
│   ├── crew-service/           # Port 8008 - Crew/team management
│   ├── federation-service/     # Port 8009 - Multi-tenant federation
│   ├── maintenance-service/    # Port 8010 - Equipment maintenance
│   ├── transport-service/      # Port 8011 - Transportation & logistics
│   ├── insurance-service/      # Port 8012 - Insurance & coverage
│   ├── workflow-service/       # Port 8013 - Workflow orchestration
│   ├── ai-service/             # Port 8014 - AI/ML features
│   ├── notification-service/   # Port 8015 - Email, SMS, push notifications
│   ├── reporting-service/      # Port 8016 - Analytics & reports
│   ├── audit-service/          # Port 8017 - Audit logging
│   └── expense-service/        # Port 8018 - Expense tracking
│
├── pkg/                         # Shared packages
│   └── common/                  # Common library for all services
│       ├── config/              # Configuration management
│       ├── errors/              # Error handling utilities
│       ├── events/              # Event definitions & interfaces
│       ├── health/              # Health check implementation
│       ├── logger/              # Logging utilities
│       ├── middleware/          # HTTP middleware
│       ├── storage/             # Storage abstractions (placeholder)
│       └── testing/             # Testing utilities (placeholder)
│
├── infra/                       # Infrastructure & DevOps
│   ├── docker/                  # Docker Compose configuration
│   │   └── docker-compose.yml  # Multi-service orchestration
│   ├── traefik/                 # API Gateway configuration
│   │   └── dynamic/
│   │       └── routers.yml      # Service routing rules
│   ├── postgres/                # PostgreSQL initialization
│   │   └── init.sql             # Database schema creation
│   ├── kurrentdb/               # Event store configuration (placeholder)
│   └── prometheus/              # Metrics collection
│       └── prometheus.yml       # Scrape configs for all services
│
├── .github/
│   └── workflows/               # CI/CD pipelines (placeholder)
│
├── frontend/                    # React/Vue frontend (existing)
├── docs/                        # Documentation (existing)
├── config/                      # Global configuration (existing)
│
├── go.work                      # Go workspace file
├── go.mod                       # Root module (placeholder)
├── Makefile                     # Build & development commands
├── .gitignore                   # Git ignore rules
└── MONOREPO_STRUCTURE.md        # This file
```

## Hexagonal Architecture (Per Service)

Each microservice follows the hexagonal (ports & adapters) architecture:

```
{service}/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── domain/                  # Business logic & domain models
│   │   └── entities.go
│   ├── application/             # Use cases & orchestration
│   │   └── {use_case}.go
│   ├── infrastructure/          # External integrations
│   │   ├── events/              # Event store integration
│   │   │   └── event_store.go
│   │   └── repositories/        # Data persistence
│   │       └── {aggregate}_repository.go
│   ├── ports/                   # Interface definitions
│   │   └── {interface}.go
│   └── adapters/                # External interface implementations
│       └── http/
│           └── handlers.go      # HTTP handlers
├── config/
│   └── .env.example             # Configuration template
└── migrations/                  # Database schema & seed data
    └── V001__init_schema.sql
```

## Key Files

### Root Level

- **go.work**: Go workspace configuration enabling all modules to work together
- **Makefile**: Development targets (build, test, lint, docker, etc.)
- **MONOREPO_STRUCTURE.md**: This file

### pkg/common

Core libraries used by all services:

- **config/config.go**: Environment-based configuration management
- **errors/errors.go**: Structured error handling with codes
- **logger/logger.go**: Logging abstraction
- **health/health.go**: Health check patterns
- **events/event.go**: Event definitions & store interfaces
- **middleware/middleware.go**: HTTP middleware utilities

### Infrastructure

- **docker-compose.yml**: Orchestrates PostgreSQL, KurrentDB, Traefik, Prometheus
- **routers.yml**: Traefik routing rules for all 18 services
- **init.sql**: PostgreSQL database initialization (creates 18 service databases)
- **prometheus.yml**: Prometheus scrape configurations

## Getting Started

### Prerequisites

- Go 1.22+
- Docker & Docker Compose (optional, for containerized dev)
- Make

### Installation

```bash
# Verify Go workspace
go work sync

# Build all services
make build

# Run tests
make test

# Start services locally
make dev

# With Docker
make docker-up

# View all available commands
make help
```

## Development Workflow

### Building

```bash
# Build all services
make build

# Binary output goes to ./bin/ directory
ls -la bin/
```

### Testing

```bash
# Unit tests
make test

# Integration tests
make test-integration

# End-to-end tests
make test-e2e
```

### Code Quality

```bash
# Run linter
make lint
```

### Database Operations

```bash
# Run migrations
make migrate-up

# Rollback
make migrate-down
```

### Docker Operations

```bash
# Build Docker images
make docker-build

# Start all services
make docker-up

# Stop all services
make docker-down
```

## Go Modules

### Workspace Setup (go.work)

All modules are managed through a Go workspace at the root level. This allows:

- Unified dependency management
- Easy navigation between packages
- Shared `go.mod` and `go.sum`

### Service Modules

Each service has its own `go.mod`:
- Declares service-specific dependencies
- Requires `github.com/jeckersberger/rentflow/pkg/common`
- Uses `replace` to reference local common package

### Common Package

Shared library module at `pkg/common/go.mod`:
- No external dependencies initially
- Foundation for cross-cutting concerns
- Imports across services must use full import path

## Port Assignments

| Service | Port | Purpose |
|---------|------|---------|
| auth-service | 8001 | Authentication & Authorization |
| inventory-service | 8002 | Equipment & Inventory |
| project-service | 8003 | Project Management |
| scanner-service | 8004 | QR/Barcode Scanning |
| warehouse-service | 8005 | Warehouse Operations |
| invoice-service | 8006 | Invoicing & Billing |
| document-service | 8007 | Document Management |
| crew-service | 8008 | Team & Crew Management |
| federation-service | 8009 | Multi-Tenant Federation |
| maintenance-service | 8010 | Equipment Maintenance |
| transport-service | 8011 | Logistics & Transport |
| insurance-service | 8012 | Insurance Management |
| workflow-service | 8013 | Workflow Orchestration |
| ai-service | 8014 | AI/ML Features |
| notification-service | 8015 | Notifications (Email, SMS) |
| reporting-service | 8016 | Analytics & Reporting |
| audit-service | 8017 | Audit Logging |
| expense-service | 8018 | Expense Management |

## Database Strategy

### PostgreSQL (per-service databases)

Each microservice has its own PostgreSQL database:
- Ensures loose coupling
- Initialized via `infra/postgres/init.sql`
- Connection string: `postgresql://postgres:password@localhost:5432/{service_name}`

### KurrentDB (Event Store)

Centralized event store for all services:
- Single source of truth for events
- Immutable event log
- Supports event replays
- Initialized in `docker-compose.yml`
- URL: `http://localhost:2113`

## API Gateway (Traefik)

Routes all service requests through a single entry point:

- **Dashboard**: http://localhost:8080
- **Web Access**: http://localhost
- **WebSecure**: https://localhost (443)

Routes are defined in `infra/traefik/dynamic/routers.yml`:
- `/auth/*` → auth-service
- `/inventory/*` → inventory-service
- etc.

## Monitoring (Prometheus)

Metrics collection via Prometheus:

- **Scrape Interval**: 15 seconds
- **Dashboard**: http://localhost:9090
- **All services**: Configured to expose `/metrics` endpoint

## Next Steps

1. **Add Dependencies**: Update `pkg/common/go.mod` with required libraries
2. **Implement Services**: Fill in domain logic, repositories, and handlers for each service
3. **Add Database Migrations**: Create schema files in `{service}/migrations/`
4. **Add Tests**: Create `_test.go` files alongside implementation
5. **CI/CD**: Set up GitHub Actions workflows in `.github/workflows/`
6. **Protocol Buffers**: Add protobuf schemas for inter-service communication
7. **Event Definitions**: Define domain events in each service
8. **Documentation**: Add API documentation and architecture decision records

## File Locations Summary

**Key paths for reference:**

- All services: `/sessions/beautiful-adoring-keller/repo/services/`
- Shared library: `/sessions/beautiful-adoring-keller/repo/pkg/common/`
- Infrastructure: `/sessions/beautiful-adoring-keller/repo/infra/`
- Build output: `/sessions/beautiful-adoring-keller/repo/bin/`
- Go workspace: `/sessions/beautiful-adoring-keller/repo/go.work`
- Build commands: `/sessions/beautiful-adoring-keller/repo/Makefile`
