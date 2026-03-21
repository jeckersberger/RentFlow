# Inventory Service - RentFlow

Complete implementation of the inventory and warehouse management service for RentFlow, a comprehensive equipment rental management system.

## Quick Start

### Prerequisites
- Go 1.22+
- PostgreSQL 14+
- Environment variables configured

### Build & Run

```bash
# Build
go build -o inventory-service ./cmd/server

# Run
./inventory-service

# Server starts on :8002 by default
```

### Run Tests
```bash
go test ./...
```

## What's Implemented

### Core Features
✓ Equipment management with full lifecycle (create, update, delete)
✓ Status machine for equipment (Available → Reserved → CheckedOut → ... → Retired)
✓ Condition tracking (New, Good, Fair, Poor, Defective)
✓ Category hierarchy with parent-child relationships
✓ Flightcase/container management for equipment grouping
✓ Image management for equipment
✓ Full-text search across equipment
✓ Barcode/QR code lookup
✓ Pagination and filtering
✓ Tenant isolation (multi-tenant ready)

### API
25 REST endpoints:
- 11 Equipment endpoints (CRUD + status + condition + images + search)
- 5 Category endpoints (CRUD)
- 7 Flightcase endpoints (CRUD + item management)
- 2 Health endpoints (health + readiness)

See [API.md](API.md) for complete API reference with examples.

### Database
3 migration files with PostgreSQL schema:
- Equipment table with full-text search indexes
- Categories table with hierarchy support
- Flightcases table with items tracking

## Architecture

Hexagonal (clean) architecture with clear separation:

```
HTTP Handlers (Adapter)
    ↓
Services & Commands (Application)
    ↓
Domain Objects (Domain)
    ↓
Repository Interfaces (Ports)
    ↓
PostgreSQL Implementations (Infrastructure)
```

See [ARCHITECTURE.md](ARCHITECTURE.md) for detailed explanation.

## Project Structure

```
inventory-service/
├── cmd/server/              # Application entry point
├── internal/
│   ├── domain/              # Business logic (Equipment, Category, Flightcase)
│   ├── application/         # Use cases (Services, Commands, DTOs)
│   ├── ports/               # Interface definitions (Repository)
│   ├── infrastructure/       # Database implementations (PostgreSQL)
│   └── adapters/http/       # REST API (Router, Handlers)
├── migrations/              # Database schema migrations
├── API.md                   # API documentation
├── ARCHITECTURE.md          # Architecture explanation
├── IMPLEMENTATION.md        # Implementation details
└── README.md               # This file
```

## Key Components

### Domain Model
- **Equipment**: Core rental item with status machine, conditions, images, tags, custom fields
- **Category**: Hierarchical classification for equipment
- **Flightcase**: Container for grouping equipment items

### Services
- **EquipmentService**: Equipment lifecycle management
- **CategoryService**: Category management with hierarchy
- **FlightcaseService**: Flightcase and item management

### Repositories
- **EquipmentRepository**: Equipment persistence with search and filtering
- **CategoryRepository**: Category persistence with hierarchy support
- **FlightcaseRepository**: Flightcase persistence with item tracking

## Features

### Equipment Status Machine
```
Available → Reserved → CheckedOut → (Available | InMaintenance | Damaged) → Retired
```

Valid transitions are enforced at the domain level.

### Full-Text Search
Equipment can be searched by:
- Name
- Description
- SKU
- Barcode
- Serial number

### Filtering
Equipment list can be filtered by:
- Status (available, reserved, checked_out, etc.)
- Category ID
- Location ID

### Pagination
All list endpoints support pagination with `limit` and `offset` parameters.

### Tenant Isolation
All operations are automatically isolated by tenant ID from the X-Tenant-ID header. This ensures complete data isolation in multi-tenant deployments.

## API Examples

### Create Equipment
```bash
curl -X POST http://localhost:8002/api/v1/equipment \
  -H "X-Tenant-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Shure SM58 Microphone",
    "category_id": "cat_123",
    "barcode": "BARCODE123",
    "sku": "SM58",
    "purchase_price": 99.99,
    "rental_price_day": 25.00,
    "rental_price_week": 150.00
  }'
```

### List Equipment with Filters
```bash
curl "http://localhost:8002/api/v1/equipment?status=available&category_id=cat_123&limit=20&offset=0" \
  -H "X-Tenant-ID: 550e8400-e29b-41d4-a716-446655440000"
```

### Search Equipment
```bash
curl "http://localhost:8002/api/v1/equipment/search?q=microphone" \
  -H "X-Tenant-ID: 550e8400-e29b-41d4-a716-446655440000"
```

### Change Equipment Status
```bash
curl -X PATCH http://localhost:8002/api/v1/equipment/equip_123/status \
  -H "X-Tenant-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "checked_out",
    "reason": "rented_to_customer"
  }'
```

For more examples, see [API.md](API.md).

## Configuration

Configure via environment variables:

```bash
# Service
SERVICE_NAME=inventory-service
SERVICE_PORT=8002
LOG_LEVEL=info

# Database
DATABASE_URL=postgres://user:password@localhost:5432/rentflow?sslmode=disable

# Environment
ENVIRONMENT=development
```

## Database Setup

1. Create database:
```sql
CREATE DATABASE rentflow;
```

2. Run migrations:
```bash
# Use your migration tool to run migrations/
# Example with PostgreSQL:
psql -U user -d rentflow -f migrations/001_create_equipment.sql
psql -U user -d rentflow -f migrations/002_create_categories.sql
psql -U user -d rentflow -f migrations/003_create_flightcases.sql
```

## Error Handling

The API returns appropriate HTTP status codes:
- **200**: Success
- **201**: Created
- **204**: No Content (success with no response body)
- **400**: Bad Request (validation error)
- **401**: Unauthorized (missing tenant ID)
- **404**: Not Found
- **409**: Conflict (e.g., duplicate barcode)
- **500**: Internal Server Error

Error responses include a descriptive message:
```json
{
  "error": "equipment not found"
}
```

## Testing

The hexagonal architecture enables comprehensive testing:

### Unit Tests (Domain)
Test domain objects and business rules directly without external dependencies.

### Integration Tests (Infrastructure)
Test database operations with test database.

### E2E Tests
Test full HTTP request/response cycle.

Mock repositories can be injected for service testing.

## Future Extensions

The architecture supports adding:
- Event sourcing with Kurrent DB
- Additional adapters (gRPC, GraphQL)
- Caching layer (Redis)
- Search integration (Elasticsearch)
- Message publishing for events
- Bulk import/export

All can be added without modifying domain logic.

## Production Readiness

The implementation is production-ready:
✓ Proper error handling throughout
✓ Input validation on all endpoints
✓ Tenant isolation enforced
✓ Database migrations provided
✓ Graceful shutdown handling
✓ Health checks implemented
✓ Logging at important points
✓ No hardcoded values
✓ Configuration from environment
✓ Clean, testable architecture

## Documentation

- [API.md](API.md) - Complete REST API reference with examples
- [ARCHITECTURE.md](ARCHITECTURE.md) - Hexagonal architecture details
- [IMPLEMENTATION.md](IMPLEMENTATION.md) - Feature and implementation details
- [COMPLETION_CHECKLIST.md](COMPLETION_CHECKLIST.md) - Implementation completion status

## License

Part of RentFlow - Equipment Rental Management System

## Contact

For questions or issues, please refer to the main RentFlow documentation.
