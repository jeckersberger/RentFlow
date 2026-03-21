# RentFlow Microservices - Complete Implementation

This document describes the six microservices implemented for the RentFlow rental equipment management platform.

## Quick Start

### Prerequisites
- Go 1.22+
- PostgreSQL 13+
- Docker (optional)

### Service Ports
- **Federation-Service**: 8009
- **Maintenance-Service**: 8010
- **Transport-Service**: 8011
- **Notification-Service**: 8015
- **Reporting-Service**: 8016
- **Audit-Service**: 8017

### Running a Service

```bash
# Set environment
export DATABASE_URL="postgresql://user:password@localhost:5432/rentflow"
export LOG_LEVEL="debug"

# Run service
cd services/federation-service
go run cmd/server/main.go
```

### Health Check
```bash
curl http://localhost:8009/health
curl http://localhost:8009/ready
```

---

## 1. Federation-Service (Port 8009)

### Purpose
Manages inter-tenant equipment sharing and federation between different RentFlow instances.

### Key Entities
- **Peer**: Other RentFlow instances in the federation (active, pending, blocked)
- **SharedEquipment**: Equipment shared to other peers with availability and pricing
- **ShareRequest**: Cross-peer equipment sharing requests with approval workflow

### API Endpoints

#### Peer Management
```
POST   /api/v1/federation/peers              Create peer
GET    /api/v1/federation/peers              List all peers
GET    /api/v1/federation/peers/{id}        Get peer details
PUT    /api/v1/federation/peers/{id}        Update peer status
DELETE /api/v1/federation/peers/{id}        Remove peer
```

#### Equipment Sharing
```
POST   /api/v1/federation/share             Share equipment to peer
GET    /api/v1/federation/catalog           View federated equipment catalog
```

#### Share Requests
```
POST   /api/v1/federation/requests          Create share request
GET    /api/v1/federation/requests          List pending requests
PUT    /api/v1/federation/requests/{id}     Approve/reject request (action: approve|reject)
```

#### Network Operations
```
POST   /api/v1/federation/handshake         Record peer connection
```

### Example Usage

```bash
# Create peer
curl -X POST http://localhost:8009/api/v1/federation/peers \
  -H "X-Tenant-ID: tenant123" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Partner Rental Co",
    "url": "https://partner.rentflow.io",
    "public_key": "public_key_base64"
  }'

# Request equipment sharing
curl -X POST http://localhost:8009/api/v1/federation/requests \
  -H "X-Tenant-ID: tenant123" \
  -H "Content-Type: application/json" \
  -d '{
    "from_peer_id": "peer_123",
    "to_peer_id": "peer_456",
    "equipment_id": "eq_789",
    "start_date": "2026-04-01T00:00:00Z",
    "end_date": "2026-04-07T00:00:00Z"
  }'

# Approve request
curl -X PUT http://localhost:8009/api/v1/federation/requests/req_999 \
  -H "X-Tenant-ID: tenant123" \
  -H "Content-Type: application/json" \
  -d '{"action": "approve"}'
```

---

## 2. Maintenance-Service (Port 8010)

### Purpose
Manages equipment maintenance scheduling, execution, and DGUV V3 certification compliance.

### Key Entities
- **MaintenanceRecord**: Individual maintenance tasks (scheduled, in-progress, completed)
  - Types: scheduled, unscheduled, dguv_v3, repair
- **MaintenanceSchedule**: Recurring maintenance intervals per equipment

### API Endpoints

#### Maintenance Records
```
POST   /api/v1/maintenance                  Create maintenance record
GET    /api/v1/maintenance                  List all records
GET    /api/v1/maintenance/{id}            Get record details
PUT    /api/v1/maintenance/{id}            Update record
POST   /api/v1/maintenance/{id}/complete   Mark completed
```

#### Schedules
```
GET    /api/v1/maintenance/schedule         List all schedules
POST   /api/v1/maintenance/schedule         Create schedule
GET    /api/v1/maintenance/equipment/{id}   Get equipment schedule
GET    /api/v1/maintenance/overdue          List overdue tasks
```

#### DGUV Compliance
```
POST   /api/v1/dguv-import                  Import DGUV V3 certifications
GET    /api/v1/dguv/equipment/{id}         View DGUV status for equipment
```

### Example Usage

```bash
# Create maintenance record
curl -X POST http://localhost:8010/api/v1/maintenance \
  -H "X-Tenant-ID: tenant123" \
  -H "Content-Type: application/json" \
  -d '{
    "equipment_id": "eq_789",
    "type": "scheduled",
    "scheduled_date": "2026-03-25T09:00:00Z",
    "technician": "John Smith",
    "cost": 250.00,
    "notes": "Annual inspection and safety check"
  }'

# Create maintenance schedule
curl -X POST http://localhost:8010/api/v1/maintenance/schedule \
  -H "X-Tenant-ID: tenant123" \
  -H "Content-Type: application/json" \
  -d '{
    "equipment_id": "eq_789",
    "interval_days": 365
  }'

# Complete maintenance
curl -X POST http://localhost:8010/api/v1/maintenance/maint_123/complete \
  -H "X-Tenant-ID: tenant123" \
  -H "Content-Type: application/json" \
  -d '{
    "completed_date": "2026-03-25T14:30:00Z",
    "notes": "All systems operational",
    "certificate_ref": "CERT-2026-001"
  }'

# Check overdue maintenance
curl http://localhost:8010/api/v1/maintenance/overdue \
  -H "X-Tenant-ID: tenant123"
```

---

## 3. Transport-Service (Port 8011)

### Purpose
Manages vehicle fleet and transportation tour planning/execution.

### Key Entities
- **Vehicle**: Fleet vehicles (vans, trucks, trailers) with capacity tracking
- **Tour**: Route planning with multiple stops and status tracking
- **TourStop**: Individual pickup/delivery locations with timestamps

### API Endpoints

#### Vehicle Management
```
POST   /api/v1/vehicles                     Create vehicle
GET    /api/v1/vehicles                     List all vehicles
GET    /api/v1/vehicles/{id}               Get vehicle details
PUT    /api/v1/vehicles/{id}               Update vehicle status
```

#### Tour Management
```
POST   /api/v1/tours                        Create tour
GET    /api/v1/tours                        List all tours
GET    /api/v1/tours/{id}                  Get tour details
PUT    /api/v1/tours/{id}                  Update tour
PATCH  /api/v1/tours/{id}/status           Update status
GET    /api/v1/tours/date/{date}           List tours by date
```

### Example Usage

```bash
# Create vehicle
curl -X POST http://localhost:8011/api/v1/vehicles \
  -H "X-Tenant-ID: tenant123" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Van 001",
    "license_plate": "ABC-123",
    "type": "van",
    "capacity": 5000.00
  }'

# Create tour with multiple stops
curl -X POST http://localhost:8011/api/v1/tours \
  -H "X-Tenant-ID: tenant123" \
  -H "Content-Type: application/json" \
  -d '{
    "project_id": "proj_456",
    "vehicle_id": "veh_789",
    "driver_id": "driver_123",
    "date": "2026-03-25T08:00:00Z",
    "stops": [
      {
        "address": "123 Main St",
        "arrival_time": "2026-03-25T08:00:00Z",
        "departure_time": "2026-03-25T08:30:00Z",
        "type": "pickup"
      },
      {
        "address": "456 Oak Ave",
        "arrival_time": "2026-03-25T09:00:00Z",
        "departure_time": "2026-03-25T09:30:00Z",
        "type": "delivery"
      }
    ]
  }'

# Update tour status
curl -X PATCH http://localhost:8011/api/v1/tours/tour_999/status \
  -H "X-Tenant-ID: tenant123" \
  -H "Content-Type: application/json" \
  -d '{"status": "in_progress"}'

# List tours by date
curl http://localhost:8011/api/v1/tours/date/2026-03-25 \
  -H "X-Tenant-ID: tenant123"
```

---

## 4. Notification-Service (Port 8015)

### Purpose
Handles multi-channel notifications (in-app, email, push) with user preferences.

### Key Entities
- **Notification**: Delivery of message to user (in-app, email, or push)
- **NotificationPreference**: User settings per channel and event type

### API Endpoints

#### Notifications
```
GET    /api/v1/notifications                List user notifications
GET    /api/v1/notifications/unread-count   Get unread count
PUT    /api/v1/notifications/{id}/read      Mark as read
POST   /api/v1/notifications/mark-all-read  Mark all as read
POST   /api/v1/notifications/send           Send notification
```

#### Preferences
```
GET    /api/v1/notifications/preferences    Get user preferences
PUT    /api/v1/notifications/preferences    Update preferences
```

### Example Usage

```bash
# Send notification
curl -X POST http://localhost:8015/api/v1/notifications/send \
  -H "X-Tenant-ID: tenant123" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user_123",
    "type": "success",
    "channel": "in_app",
    "title": "Equipment Checked In",
    "message": "Equipment EQ-789 successfully checked in",
    "link": "/equipment/eq_789"
  }'

# Get unread notifications
curl http://localhost:8015/api/v1/notifications \
  -H "X-Tenant-ID: tenant123" \
  -H "X-User-ID: user_123"

# Mark notification as read
curl -X PUT http://localhost:8015/api/v1/notifications/notif_456/read \
  -H "X-Tenant-ID: tenant123"

# Mark all as read
curl -X POST http://localhost:8015/api/v1/notifications/mark-all-read \
  -H "X-Tenant-ID: tenant123" \
  -H "X-User-ID: user_123"
```

---

## 5. Reporting-Service (Port 8016)

### Purpose
Generates business intelligence reports and dashboard KPIs.

### Key Entities
- **Report**: Generated report with async status tracking
- **KPI**: Key performance indicators (revenue, utilization, inventory value, projects)

### API Endpoints

#### Reports
```
POST   /api/v1/reports/generate             Create report (async)
GET    /api/v1/reports/{id}                Get report status/results
GET    /api/v1/reports/revenue             List revenue reports
GET    /api/v1/reports/utilization         List utilization reports
GET    /api/v1/reports/inventory-value     List inventory value reports
GET    /api/v1/reports/equipment/{id}/history  Equipment history
```

#### Dashboard
```
GET    /api/v1/reports/kpi                 Get KPI dashboard
```

### Example Usage

```bash
# Get KPI dashboard
curl http://localhost:8016/api/v1/reports/kpi \
  -H "X-Tenant-ID: tenant123"

# Response example:
# {
#   "kpis": [
#     {"name": "Total Revenue", "value": 150000, "previous_value": 140000, "change_percent": 7.14, "period": "MTD"},
#     {"name": "Equipment Utilization", "value": 82.5, "previous_value": 78.0, "change_percent": 5.77, "period": "MTD"},
#     {"name": "Inventory Value", "value": 450000, "previous_value": 455000, "change_percent": -1.10, "period": "Current"},
#     {"name": "Active Projects", "value": 12, "previous_value": 10, "change_percent": 20.0, "period": "Active"}
#   ]
# }

# Generate report (async)
curl -X POST http://localhost:8016/api/v1/reports/generate \
  -H "X-Tenant-ID: tenant123" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "revenue",
    "parameters": {
      "start_date": "2026-03-01",
      "end_date": "2026-03-31",
      "group_by": "project"
    }
  }'

# Check report status
curl http://localhost:8016/api/v1/reports/report_123 \
  -H "X-Tenant-ID: tenant123"
```

---

## 6. Audit-Service (Port 8017)

### Purpose
GoBD-compliant audit logging with hash chain integrity verification.

### Key Features
- SHA-256 hash chain for immutable audit trail
- Complete state tracking (previous and current)
- User and IP address logging
- Integrity verification endpoint
- Time-range based log export

### Key Entities
- **AuditEntry**: Individual action with hash chain
- **IntegrityCheck**: Verification result

### API Endpoints

#### Audit Logs
```
POST   /api/v1/audit/logs                   Create audit entry
GET    /api/v1/audit/logs                   List logs (with time filtering)
GET    /api/v1/audit/entity/{type}/{id}    Get entity change history
GET    /api/v1/audit/user/{id}             Get user activity
```

#### Integrity
```
POST   /api/v1/audit/verify                 Verify hash chain integrity
GET    /api/v1/audit/verify/status          Get last verification status
```

#### Export
```
GET    /api/v1/audit/export?from=X&to=Y    Export logs as JSON
```

### Example Usage

```bash
# Log audit entry
curl -X POST http://localhost:8017/api/v1/audit/logs \
  -H "X-Tenant-ID: tenant123" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user_456",
    "action": "equipment_checked_out",
    "entity_type": "Equipment",
    "entity_id": "eq_789",
    "new_state": {
      "status": "checked_out",
      "renter": "customer_123",
      "checkout_date": "2026-03-21T10:00:00Z"
    }
  }'

# List audit logs with time range
curl "http://localhost:8017/api/v1/audit/logs?from=2026-03-01T00:00:00Z&to=2026-03-31T23:59:59Z" \
  -H "X-Tenant-ID: tenant123"

# Get entity history
curl http://localhost:8017/api/v1/audit/entity/Equipment/eq_789 \
  -H "X-Tenant-ID: tenant123"

# Verify integrity
curl -X POST http://localhost:8017/api/v1/audit/verify \
  -H "X-Tenant-ID: tenant123"

# Export logs
curl "http://localhost:8017/api/v1/audit/export?from=2026-03-01T00:00:00Z&to=2026-03-31T23:59:59Z" \
  -H "X-Tenant-ID: tenant123" \
  > audit_logs.json
```

---

## Common Patterns

### All Services Support

#### Headers
```
X-Tenant-ID: required on all endpoints (multi-tenant isolation)
X-User-ID: optional, used by notification service
User-Agent: auto-captured by audit service
```

#### Response Format
```json
// Success response
{
  "id": "resource_id",
  "field1": "value1",
  "created_at": "2026-03-21T10:00:00Z"
}

// List response
{
  "data": [
    { /* resource */ },
    { /* resource */ }
  ]
}

// Error response
{
  "error": "error message"
}
```

#### Status Codes
- `201 Created` - Resource created
- `200 OK` - Success
- `204 No Content` - Success with no body
- `400 Bad Request` - Invalid input
- `401 Unauthorized` - Missing tenant ID
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

---

## Database Schema

All services use PostgreSQL. Each service includes migration files:
- `/migrations/001_*_service.sql`

Run migrations before starting services:
```bash
psql postgresql://user:password@localhost:5432/rentflow < migrations/001_federation_service.sql
# Repeat for other services
```

---

## Testing

Each service can be tested independently:

```bash
# Unit test a service
cd services/federation-service
go test ./...

# Run with race detector
go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## Monitoring

### Health Checks
```bash
curl http://localhost:8009/health   # Service health
curl http://localhost:8009/ready    # Readiness
```

### Logging
All services log to stdout with structured logging:
```
time=2026-03-21T10:00:00Z level=info service=federation-service msg="Starting service" port=8009
time=2026-03-21T10:00:01Z level=info service=federation-service msg="Connected to database"
time=2026-03-21T10:00:02Z level=info service=federation-service msg="Listening" addr=:8009
```

---

## Deployment

### Docker
Each service includes a Dockerfile template for containerization.

### Kubernetes
Services are stateless and can be horizontally scaled:
- Use `X-Tenant-ID` header for multi-tenancy
- Database pooling handled by pgx
- Graceful shutdown on SIGINT/SIGTERM

### Environment Variables
```bash
DATABASE_URL=postgresql://user:password@host:5432/db
LOG_LEVEL=info|debug|warn|error
```

---

## Architecture Diagram

```
┌─────────────────────────────────────────┐
│         API Gateway / Router            │
└─────────────────────────────────────────┘
           │           │           │
    ┌──────┴──┐  ┌─────┴──┐  ┌────┴─────┐
    │          │  │        │  │          │
    ▼          ▼  ▼        ▼  ▼          ▼
Federation  Maintenance Transport Notification Reporting  Audit
 Service    Service       Service     Service      Service   Service
    │          │          │          │           │           │
    └──────────┴──────────┴──────────┴───────────┴───────────┘
                          │
                          ▼
                    PostgreSQL
              (Multi-tenant Database)
```

---

## Support

For issues or questions:
1. Check service logs: `docker logs container_name`
2. Verify database connectivity: `psql $DATABASE_URL`
3. Check health endpoint: `curl http://localhost:PORT/health`
4. Review API endpoint documentation in this file

---

**Version**: 1.0.0
**Last Updated**: March 21, 2026
**Services**: 6 complete microservices (5550+ lines of code)
