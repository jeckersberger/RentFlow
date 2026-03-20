# RentFlow Docker Setup Guide

This guide covers the production-ready Docker Compose configuration for RentFlow, a comprehensive event and equipment rental management system.

## Overview

The Docker setup includes:
- **18 microservices** running in isolated containers
- **Traefik** API Gateway with automatic routing and HTTPS support
- **PostgreSQL** for relational data (18 databases, one per service)
- **KurrentDB** (EventStore) for event sourcing and CQRS
- **Redis** for caching and session management
- **Prometheus** for metrics collection
- **React 18 Frontend** with TypeScript and Vite

## Quick Start

### Prerequisites
- Docker Engine 20.10+
- Docker Compose 2.0+
- 8GB+ RAM
- 20GB+ disk space

### Setup

1. Clone and configure environment:
```bash
cd /path/to/rentflow
cp .env.example .env
```

2. Edit `.env` with your configuration:
```bash
# Database
POSTGRES_USER=rentflow
POSTGRES_PASSWORD=<secure-password>

# Redis
REDIS_PASSWORD=<secure-password>

# JWT
JWT_SECRET=<base64-encoded-secret-min-32-chars>

# Domain (production)
DOMAIN=rentflow.example.com
LETSENCRYPT_EMAIL=admin@rentflow.example.com
```

3. Start all services:
```bash
# Production (with Traefik TLS)
docker-compose up -d

# Development (without TLS, with hot reload)
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up
```

4. Verify services are running:
```bash
docker-compose ps
```

5. Access the application:
- **Frontend**: http://localhost:3000 (dev) or https://rentflow.example.com (prod)
- **Traefik Dashboard**: http://localhost:8080
- **Prometheus**: http://localhost:9090

## Architecture

### Service Layout

```
rentflow/
├── services/
│   ├── auth-service         (JWT, tenant isolation)
│   ├── inventory-service    (Equipment, categories)
│   ├── project-service      (Projects, packlists)
│   ├── scanner-service      (RFID/barcode scanning)
│   ├── warehouse-service    (Locations, movements)
│   ├── invoice-service      (PDF generation, DATEV)
│   ├── document-service     (OCR, template management)
│   ├── crew-service         (CalDAV integration)
│   ├── federation-service   (Multi-tenant federation)
│   ├── maintenance-service  (DGUV compliance)
│   ├── transport-service    (Vehicles, tours)
│   ├── insurance-service    (Claims, policies)
│   ├── workflow-service     (Orchestration)
│   ├── ai-service           (Ollama, LLM integration)
│   ├── notification-service (Email, webhooks)
│   ├── reporting-service    (Analytics, materialized views)
│   ├── audit-service        (GoBD compliance)
│   └── expense-service      (Cost tracking)
├── frontend/                (React 18 + Vite)
├── infra/
│   ├── traefik/             (API Gateway config)
│   ├── postgres/            (Database init)
│   └── prometheus/          (Metrics config)
├── docker-compose.yml       (Production config)
├── docker-compose.dev.yml   (Development overrides)
└── .env.example             (Environment template)
```

### Network Flow

```
Internet
   ↓
Traefik (Port 80/443)
   ↓
├─→ Frontend (React app)
│
├─→ Auth Service (8001)
│   ↓
│   PostgreSQL
│   KurrentDB
│   Redis
│
├─→ [17 other services] (8002-8018)
    ↓
    Shared PostgreSQL
    Shared KurrentDB (audit-only)
    Shared Redis
```

## Configuration Files

### Root Level

- **`docker-compose.yml`** - Main production configuration with all 18 services
- **`docker-compose.dev.yml`** - Development overrides (hot reload, debug logging)
- **`.env.example`** - Environment variable template
- **`.env`** - Runtime configuration (not in Git)

### Infrastructure (`infra/`)

#### `traefik/traefik.yml`
Static Traefik configuration:
- Entrypoints (HTTP/HTTPS)
- Certificate resolver (Let's Encrypt + self-signed)
- Docker provider for service discovery
- File provider for dynamic routing

#### `traefik/dynamic/config.yml`
Dynamic routing configuration:
- Middleware (CORS, rate limiting, security headers)
- Service routing rules (PathPrefix matching)
- Load balancing

#### `postgres/init.sql`
Database initialization:
- Creates 18 databases (one per service)
- Sets up user privileges
- Shared schema for cross-service queries (if needed)

#### `prometheus/prometheus.yml`
Metrics scraping:
- Prometheus self-monitoring
- All 18 services metrics endpoints (/metrics)
- Retention policy

## Service Ports

| Service | Port | Database | Route Prefix |
|---------|------|----------|--------------|
| Frontend | 3000 | — | / |
| Traefik | 8080 | — | /dashboard, /api/rawdata |
| Prometheus | 9090 | — | /prometheus |
| Auth | 8001 | auth_service | /api/v1/auth, /api/v1/users, /api/v1/tenants |
| Inventory | 8002 | inventory_service | /api/v1/equipment, /api/v1/categories |
| Project | 8003 | project_service | /api/v1/projects, /api/v1/packlists |
| Scanner | 8004 | scanner_service | /api/v1/scan |
| Warehouse | 8005 | warehouse_service | /api/v1/locations, /api/v1/movements, /api/v1/inventory |
| Invoice | 8006 | invoice_service | /api/v1/invoices |
| Document | 8007 | document_service | /api/v1/documents, /api/v1/templates, /api/v1/ocr |
| Crew | 8008 | crew_service | /api/v1/crew, /api/v1/time-entries |
| Federation | 8009 | federation_service | /api/v1/federation |
| Maintenance | 8010 | maintenance_service | /api/v1/maintenance, /api/v1/dguv |
| Transport | 8011 | transport_service | /api/v1/vehicles, /api/v1/tours |
| Insurance | 8012 | insurance_service | /api/v1/policies, /api/v1/claims |
| Workflow | 8013 | workflow_service | /api/v1/workflows |
| AI | 8014 | ai_service | /api/v1/ai |
| Notification | 8015 | notification_service | /api/v1/notifications |
| Reporting | 8016 | reporting_service | /api/v1/reports |
| Audit | 8017 | audit_service | /api/v1/audit |
| Expense | 8018 | expense_service | /api/v1/expenses |

## Development Workflow

### Running a Single Service with Hot Reload

Uncomment the service in `docker-compose.dev.yml` and use:

```bash
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up auth-service
```

This will:
- Run `go run ./cmd/server` instead of building
- Mount source code volumes for hot reload
- Enable debug logging
- Disable health checks for faster iteration

### Running All Infrastructure, Develop Services Locally

```bash
# Terminal 1: Start infrastructure only
docker-compose up postgres redis kurrentdb traefik prometheus

# Terminal 2-3: Run services locally with hot reload
cd services/auth-service && go run ./cmd/server
cd services/inventory-service && go run ./cmd/server
```

### Rebuilding Services

```bash
# Rebuild specific service
docker-compose build auth-service
docker-compose up -d auth-service

# Rebuild all services
docker-compose build
docker-compose up -d
```

## Environment Variables

### Database
```bash
POSTGRES_USER=rentflow          # PostgreSQL user
POSTGRES_PASSWORD=***           # PostgreSQL password
POSTGRES_HOST=postgres          # DB host (service name)
POSTGRES_PORT=5432             # DB port
POSTGRES_DB=rentflow           # Default database
```

### KurrentDB (EventStore)
```bash
KURRENTDB_HOST=kurrentdb       # EventStore host
KURRENTDB_PORT=2113            # EventStore port
KURRENTDB_SCHEME=http          # http or https
KURRENTDB_USER=admin           # EventStore user
KURRENTDB_PASSWORD=***         # EventStore password
```

### Redis
```bash
REDIS_HOST=redis               # Redis host
REDIS_PORT=6379               # Redis port
REDIS_PASSWORD=***            # Redis password
REDIS_DB=0                     # Redis database
REDIS_CACHE_TTL_SECONDS=300   # Cache TTL
```

### Security
```bash
JWT_SECRET=***                 # Min 32 chars, base64-encoded
JWT_EXPIRY_HOURS=24           # Token expiry
JWT_REFRESH_EXPIRY_DAYS=7     # Refresh token expiry
SESSION_SECRET=***             # Session secret
ALLOWED_ORIGINS=*              # CORS origins
```

### Traefik
```bash
LETSENCRYPT_EMAIL=***         # Let's Encrypt email
DASHBOARD_HTPASSWD=***        # Dashboard basic auth (htpasswd format)
METRICS_HTPASSWD=***          # Metrics basic auth
TRAEFIK_ENTRYPOINT=https      # Default entrypoint
```

### Logging
```bash
LOG_LEVEL=info                # info, debug, warn, error
DEBUG_MODE=false              # Enable debug output
```

### Optional Services
```bash
AI_PROVIDERS=ollama           # ollama, anthropic, openai, mistral, gemini
OLLAMA_ENDPOINT=http://localhost:11434
CLAUDE_API_KEY=***
OPENAI_API_KEY=***

SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=***
SMTP_PASSWORD=***
SMTP_FROM=noreply@rentflow.local
SMTP_TLS_ENABLED=true
```

## Monitoring & Logging

### Prometheus Metrics
All services expose metrics on `/metrics` endpoint. Configure monitoring:

```bash
# Access Prometheus dashboard
http://localhost:9090

# Query metrics
http_requests_total
http_request_duration_seconds
db_connection_pool_connections_used
cache_hits_total
```

### Service Logs
```bash
# View all logs
docker-compose logs -f

# View specific service
docker-compose logs -f auth-service

# View last 100 lines
docker-compose logs --tail=100 auth-service
```

### Health Checks
Each service exposes a health endpoint:
```bash
curl http://localhost:8001/health    # Auth service
curl http://localhost:8002/health    # Inventory service
# etc.
```

## Production Deployment

### SSL/TLS Configuration

Traefik automatically handles Let's Encrypt certificate generation and renewal:

1. Ensure `LETSENCRYPT_EMAIL` is set in `.env`
2. Point domain DNS to server IP
3. Traefik will:
   - Challenge DNS to verify domain ownership
   - Generate certificate automatically
   - Renew 30 days before expiry
   - Store certificates in `traefik_certs` volume

### Scaling Services

To run multiple instances of a service:

```bash
docker-compose up -d --scale auth-service=3
```

Traefik will automatically load-balance across instances.

### Backup Strategy

```bash
# Backup PostgreSQL
docker-compose exec postgres pg_dump -U rentflow rentflow > backup.sql

# Backup volumes
docker run --rm -v postgres_data:/data -v $(pwd):/backup \
  alpine tar czf /backup/postgres_data.tar.gz -C /data .

# Restore PostgreSQL
docker-compose exec -T postgres psql -U rentflow < backup.sql
```

### Resource Limits

Set resource limits in `docker-compose.yml`:

```yaml
auth-service:
  deploy:
    resources:
      limits:
        cpus: '0.5'
        memory: 512M
      reservations:
        cpus: '0.25'
        memory: 256M
```

## Troubleshooting

### Services won't start
```bash
# Check logs
docker-compose logs <service>

# Verify dependencies are healthy
docker-compose ps

# Rebuild from scratch
docker-compose down -v
docker-compose build --no-cache
docker-compose up -d
```

### Database connection errors
```bash
# Verify PostgreSQL is running
docker-compose ps postgres

# Test connection
docker-compose exec postgres psql -U rentflow -d rentflow -c "SELECT 1;"

# Check environment variables
docker-compose exec <service> env | grep DB_
```

### Traefik routing not working
```bash
# Check Traefik dashboard
http://localhost:8080

# View Traefik logs
docker-compose logs traefik

# Test direct service access
curl http://localhost:8001/health
```

### Port already in use
```bash
# Find process using port
lsof -i :8001

# Change port in docker-compose.yml or .env
ports:
  - "8001:8001"  # Change first number
```

## Advanced Configuration

### Custom Traefik Middleware

Edit `infra/traefik/dynamic/config.yml` to add middleware:

```yaml
http:
  middlewares:
    customHeader:
      headers:
        customRequestHeaders:
          X-Custom-Header: "value"
```

### Service-to-Service Communication

Services communicate via Docker network `rentflow`:

```bash
# From auth-service to inventory-service
curl http://inventory-service:8002/api/v1/equipment
```

### Database Replication

For high availability, use PostgreSQL streaming replication:

```yaml
postgres-replica:
  image: postgres:16-alpine
  environment:
    PGUSER: replicator
    REPLICATION_PASSWORD: ***
  command: |
    -c wal_level=replica
    -c max_wal_senders=3
    -c max_replication_slots=3
```

## See Also

- [Docker Compose Reference](https://docs.docker.com/compose/compose-file/)
- [Traefik Documentation](https://doc.traefik.io/)
- [PostgreSQL Best Practices](https://www.postgresql.org/docs/current/runtime.html)
- [KurrentDB Documentation](https://github.com/EventStore/EventStore)
