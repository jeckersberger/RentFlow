# Inventory Service Implementation

This document describes the complete implementation of the inventory-service following the hexagonal architecture pattern.

## Architecture Overview

The service is structured in four main layers:

### 1. Domain Layer (`internal/domain/`)

The core business logic and domain objects:

- **equipment.go**: Equipment aggregate with status machine and condition tracking
  - Equipment entity with full lifecycle management
  - Status transitions: Available → Reserved → CheckedOut → Available/InMaintenance/Damaged → Retired
  - Condition tracking: New → Good → Fair → Poor → Defective
  - Image and tag management
  - Custom fields support

- **category.go**: Hierarchical equipment categories
  - Parent-child relationships for category hierarchy
  - Icon and color support for UI

- **flightcase.go**: Equipment containers/groupings
  - Logical grouping of equipment items
  - Item management with quantities
  - Weight and location tracking

- **events.go**: Domain events for event sourcing
  - EquipmentCreated, EquipmentUpdated, EquipmentStatusChanged
  - EquipmentLocationChanged, EquipmentConditionUpdated, EquipmentImageAdded
  - EquipmentRetired, CategoryCreated, CategoryUpdated, CategoryDeleted
  - FlightcaseCreated, FlightcaseItemAdded, FlightcaseItemRemoved

- **errors.go**: Domain-specific errors with structured error types

### 2. Application Layer (`internal/application/`)

Use cases and application services:

- **equipment_service.go**: Equipment business operations
  - Create, read, update, delete operations
  - Status and condition management
  - Image uploads and management
  - Barcode lookups
  - Search and filtering with pagination

- **category_service.go**: Category management
  - CRUD operations for categories
  - Hierarchical category support

- **flightcase_service.go**: Flightcase management
  - CRUD operations for flightcases
  - Item management (add/remove from flightcase)

- **dto.go**: Data Transfer Objects for API responses
  - EquipmentDTO, CategoryDTO, FlightcaseDTO
  - Conversion functions from domain objects

- **commands.go**: Command objects for all operations
  - CreateEquipmentCommand, UpdateEquipmentCommand, ChangeStatusCommand
  - CreateCategoryCommand, UpdateCategoryCommand
  - CreateFlightcaseCommand, AddFlightcaseItemCommand

- **queries.go**: Query objects for filtering and searching
  - ListEquipmentQuery, SearchEquipmentQuery
  - Pagination support

### 3. Ports Layer (`internal/ports/`)

Interface definitions for dependencies:

- **repository.go**: Repository interfaces
  - EquipmentRepository: CRUD + search + barcode lookup
  - CategoryRepository: CRUD for categories
  - FlightcaseRepository: CRUD + item management

These interfaces are implemented by the infrastructure layer, allowing for dependency injection and testability.

### 4. Infrastructure Layer (`internal/infrastructure/`)

Concrete implementations:

- **repositories/equipment_postgres.go**: PostgreSQL equipment storage
  - Full-text search on name, description, SKU, barcode, serial number
  - Filtering by status, category, location
  - Pagination support
  - Tenant isolation

- **repositories/category_postgres.go**: PostgreSQL category storage
  - Hierarchical category support
  - Sorted by sort_order

- **repositories/flightcase_postgres.go**: PostgreSQL flightcase storage
  - Item management with quantities
  - Cascade delete on item removal

- **repositories/helpers.go**: Helper functions
  - JSONB conversion utilities
  - Array type handling

### 5. Adapter Layer (`internal/adapters/http/`)

HTTP REST API implementation:

- **router.go**: HTTP route definitions
  - All endpoints mapped to handlers
  - Grouped by resource type

- **handlers.go**: HTTP request handlers
  - Request parsing and validation
  - Service invocation
  - Response serialization
  - Error handling with appropriate HTTP status codes

## Database Schema

### equipment table
```sql
- id, tenant_id, name, description
- category_id, sku, serial_number, barcode (unique per tenant)
- status (available, reserved, checked_out, in_maintenance, damaged, retired)
- condition (new, good, fair, poor, defective)
- purchase_date, purchase_price
- rental_price_day, rental_price_week
- weight, dimensions (length, width, height, unit)
- location_id
- image_refs (array), tags (array), custom_fields (jsonb)
- created_at, updated_at, created_by_user_id
```

Indexes:
- tenant_id (filtering)
- barcode (fast lookup)
- status, category_id, location_id (filtering)
- sku, serial_number (search)

### categories table
```sql
- id, tenant_id, name
- parent_id (for hierarchy)
- icon, color, sort_order
- created_at, created_by_user_id
```

Indexes:
- tenant_id, parent_id, sort_order

### flightcases table
```sql
- id, tenant_id, name, description
- barcode (unique per tenant)
- weight, location_id
- created_at, updated_at, created_by_user_id
```

### flightcase_items table
```sql
- flightcase_id (FK), equipment_id
- quantity, added_at
```

## API Endpoints

### Equipment

- `POST /api/v1/equipment` - Create equipment
- `GET /api/v1/equipment` - List equipment with filters (status, category_id, location_id, limit, offset)
- `GET /api/v1/equipment/{id}` - Get single equipment
- `GET /api/v1/equipment/search?q=term` - Search equipment
- `GET /api/v1/equipment/barcode/{barcode}` - Get by barcode
- `PUT /api/v1/equipment/{id}` - Update equipment
- `PATCH /api/v1/equipment/{id}/status` - Change status
- `PATCH /api/v1/equipment/{id}/condition` - Update condition
- `POST /api/v1/equipment/{id}/images` - Add image (multipart)
- `DELETE /api/v1/equipment/{id}` - Delete equipment

### Categories

- `POST /api/v1/categories` - Create category
- `GET /api/v1/categories` - List all categories
- `GET /api/v1/categories/{id}` - Get single category
- `PUT /api/v1/categories/{id}` - Update category
- `DELETE /api/v1/categories/{id}` - Delete category

### Flightcases

- `POST /api/v1/flightcases` - Create flightcase
- `GET /api/v1/flightcases` - List flightcases
- `GET /api/v1/flightcases/{id}` - Get single flightcase
- `PUT /api/v1/flightcases/{id}` - Update flightcase
- `POST /api/v1/flightcases/{id}/items` - Add item to flightcase
- `DELETE /api/v1/flightcases/{id}/items` - Remove item from flightcase
- `DELETE /api/v1/flightcases/{id}` - Delete flightcase

### Health

- `GET /health` - Health check
- `GET /ready` - Readiness check

## Tenant Isolation

All queries automatically filter by tenant_id from the X-Tenant-ID HTTP header. This ensures complete data isolation between tenants.

## Key Features

### 1. Status Management
Equipment follows a state machine with valid transitions:
- Available (ready for use)
- Reserved (booked for future job)
- CheckedOut (in use)
- InMaintenance (undergoing maintenance)
- Damaged (has damage)
- Retired (no longer in use)

### 2. Condition Tracking
Equipment condition is tracked separately:
- New, Good, Fair, Poor, Defective

### 3. Search & Filtering
- Full-text search across name, description, SKU, barcode, serial number
- Filter by status, category, location
- Pagination support (limit, offset)

### 4. Image Management
- Equipment can have multiple images
- Images can be stored via external storage adapter or referenced by filename
- Images are stored as array of refs

### 5. Hierarchical Categories
Categories support parent-child relationships for hierarchical organization (e.g., Audio > Microphones > Wireless)

### 6. Flightcases
Logical groupings of equipment for transport/storage. Supports:
- Multiple items per case
- Quantity tracking
- Location management
- Weight tracking

## Error Handling

The service uses structured domain errors with codes:

```go
ErrEquipmentNotFound      = "equipment not found"
ErrCategoryNotFound       = "category not found"
ErrFlightcaseNotFound     = "flightcase not found"
ErrBarcodeAlreadyExists   = "barcode already exists"
ErrInvalidEquipmentStatus = "invalid equipment status"
ErrInvalidEquipmentCondition = "invalid equipment condition"
ErrTenantIDRequired       = "tenant ID is required"
```

HTTP status codes are mapped appropriately:
- 404 Not Found
- 400 Bad Request (validation errors)
- 409 Conflict (duplicate barcode)
- 401 Unauthorized (missing tenant ID)
- 500 Internal Server Error

## Testing Considerations

The implementation follows SOLID principles making it testable:

- **Dependency Injection**: Services receive dependencies (repositories, logger)
- **Repository Interfaces**: Implementations can be mocked
- **Domain Objects**: Pure business logic without external dependencies
- **Application Services**: Orchestrate domain and infrastructure layers

## Future Enhancements

1. Event sourcing integration with Kurrent DB
2. Caching layer for frequently accessed categories/equipment
3. Bulk import/export operations
4. Image compression and optimization
5. Batch equipment condition updates
6. Equipment reservation system
7. Audit logging for all changes
8. Integration with warehouse management system
