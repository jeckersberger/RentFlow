# Architektur-Übersicht — Microservice + Event Sourcing + CQRS

**Stand:** 20. März 2026

## Architektur-Prinzipien

1. **Microservice Architecture (MSA)**: Jeder Bounded Context ist ein eigenständiger Go-Service
2. **Event Sourcing**: KurrentDB (ehemals EventStoreDB) als Source of Truth — alle Zustandsänderungen als immutable Events
3. **CQRS**: Write-Side = Events in KurrentDB, Read-Side = Projections in PostgreSQL
4. **Database-per-Service**: Jeder Service besitzt sein eigenes PostgreSQL-Schema für Read-Projections
5. **API Gateway Pattern**: Traefik routet alle Client-Requests zum richtigen Service
6. **Saga Pattern**: Verteilte Transaktionen über Event-Choreografie (KurrentDB Subscriptions)
7. **Relational Sink**: KurrentDB v26 projiziert Events automatisch nach PostgreSQL

## Tech-Stack

| Komponente | Technologie | Begründung |
|-----------|-------------|------------|
| Backend | Go 1.22+ (pro Service) | Performance, Single-Binary, Concurrency |
| API Gateway | Traefik v3 | Docker-native, automatisches Service-Discovery |
| Event Store | KurrentDB (ehemals EventStoreDB) | Event-native DB, Subscriptions, Relational Sink, Go-Client |
| Frontend | TypeScript + React 18 + Sass/SCSS | PWA-fähig, großes Ökosystem |
| Read-DB | PostgreSQL 16 | Schema-per-Service Projections, JSONB, Row-Level Security |
| Cache | Redis 7 | Sessions, API-Cache, Rate Limiting |
| Deployment | Docker Compose | Self-Hosting, ein Befehl startet alles |
| Scanner | USB-Barcode, Zebra TC21, Handy-Kamera, RFID | Alle VT-Szenarien |
| Observability | Prometheus + Grafana + zerolog | Metriken, Dashboards, Logging |

## System-Architektur

```
┌────────────────────────────────────────────────────────────┐
│                        Clients                              │
│  Browser (Desktop/Tablet/Handy) │ PWA │ Scanner │ Portal    │
└──────────────────────┬─────────────────────────────────────┘
                       │ HTTPS
┌──────────────────────▼─────────────────────────────────────┐
│               Traefik v3 (API Gateway)                      │
│       TLS Termination │ Routing │ Load Balancing             │
│       Rate Limiting │ Circuit Breaker │ Health Checks        │
└──┬───┬───┬───┬───┬───┬───┬───┬───┬───┬───┬───┬───┬───┬──┘
   │   │   │   │   │   │   │   │   │   │   │   │   │   │
   ▼   ▼   ▼   ▼   ▼   ▼   ▼   ▼   ▼   ▼   ▼   ▼   ▼   ▼
 ┌───┐┌───┐┌───┐┌───┐┌───┐┌───┐┌───┐┌───┐┌───┐┌───┐┌───┐
 │AUT││INV││PRJ││SCN││WHS││INV││DOC││CRW││FED││MNT││...│
 │H  ││ENT││ECT││NER││E  ││OIC││   ││   ││   ││   ││   │
 └─┬─┘└─┬─┘└─┬─┘└─┬─┘└─┬─┘└─┬─┘└─┬─┘└─┬─┘└─┬─┘└─┬─┘└─┬─┘
   └────┴────┴────┴────┴────┴────┴────┴────┴────┴────┴───┘
                              │
              ┌───────────────▼───────────────┐
              │   KurrentDB (Event Store)      │
              │   Source of Truth │ Immutable   │
              │   Subscriptions │ Replay        │
              │   Relational Sink → PostgreSQL  │
              └───────────────┬───────────────┘
                              │ Projections
              ┌───────────────▼───────────────┐
              │   PostgreSQL 16 (Read-Models)  │
              │   Schema-per-Service           │
              │   (eine Instanz, N Schemas)    │
              └───────────────────────────────┘
```

## Microservices (17 Services)

| # | Service | Port | Verantwortung |
|---|---------|------|---------------|
| 1 | auth-service | :8001 | JWT, RBAC, Users, Tenants, Sessions |
| 2 | inventory-service | :8002 | Equipment, Kategorien, Preise, Labels, Flightcases |
| 3 | project-service | :8003 | Projekte, Packlisten, Reservierungen |
| 4 | scanner-service | :8004 | QR/Barcode/RFID, Check-In/Out, Offline-Sync |
| 5 | warehouse-service | :8005 | Lagerplätze, Warenbewegung, Inventur, Optimierung |
| 6 | invoice-service | :8006 | Rechnungen, Mahnwesen, DATEV, Bank-Integration |
| 7 | document-service | :8007 | PDF, Templates, Verträge, OCR |
| 8 | crew-service | :8008 | Personal, Freelancer, Zeiterfassung, CalDAV |
| 9 | federation-service | :8009 | mTLS P2P, Equipment-Sharing, Sub-Rental |
| 10 | maintenance-service | :8010 | Wartung, E-Check/DGUV V3 |
| 11 | transport-service | :8011 | Fahrzeuge, Touren, Routen, Transportkosten |
| 12 | insurance-service | :8012 | Policen, Schäden, Risiko-Analyse |
| 13 | workflow-service | :8013 | No-Code Workflows, Event-Trigger |
| 14 | ai-service | :8014 | Multi-Provider AI, Anonymisierung, Prognosen |
| 15 | notification-service | :8015 | In-App, Push, E-Mail |
| 16 | reporting-service | :8016 | CQRS Read-Side, KPIs, Dashboards |
| 17 | audit-service | :8017 | GoBD Audit-Trail, immutables Log |

## Event Sourcing mit KurrentDB

### Stream-Konventionen

KurrentDB organisiert Events in Streams. Wir nutzen zwei Arten:

**Aggregate-Streams** (pro Entität): `equipment-{uuid}`, `project-{uuid}`, `invoice-{uuid}`
- Enthalten alle Events einer einzelnen Entität
- Für optimistic concurrency (expected version)

**Kategorie-Streams** (automatisch): `$ce-equipment`, `$ce-project`, `$ce-invoice`
- KurrentDB gruppiert automatisch alle Streams einer Kategorie
- Services subscriben auf Kategorie-Streams

### Event-Typen pro Domain

| Kategorie | Event-Typen | Producer |
|-----------|-------------|----------|
| equipment | EquipmentCreated, EquipmentUpdated, EquipmentCheckedOut, EquipmentCheckedIn, EquipmentReserved, EquipmentReleased | inventory-service |
| project | ProjectCreated, ProjectUpdated, ProjectStatusChanged, ProjectCompleted | project-service |
| scan | ScanCompleted, CheckinCompleted, CheckoutCompleted | scanner-service |
| warehouse | StockUpdated, StockBelowMinimum, MovementRecorded, ReorderSuggested | warehouse-service |
| invoice | InvoiceCreated, InvoiceSent, InvoicePaid, InvoiceOverdue, PaymentReceived, DunningSent | invoice-service |
| document | DocumentGenerated, DocumentUploaded, DocumentExtracted | document-service |
| crew | CrewAssigned, AvailabilityChanged, TimeRecorded | crew-service |
| federation | PartnerConnected, EquipmentShared, RentalRequested | federation-service |
| maintenance | MaintenanceScheduled, MaintenanceCompleted, MaintenanceOverdue, EcheckCompleted | maintenance-service |
| transport | TourCreated, TourOptimized, TourCompleted, VehicleAssigned | transport-service |
| insurance | PolicyCreated, ClaimFiled, DamageReported, RiskScoreUpdated | insurance-service |
| workflow | WorkflowExecuted, WorkflowFailed | workflow-service |
| ai | AnalysisCompleted, PredictionGenerated | ai-service |

### Subscription-Patterns

**Catch-up Subscriptions** (persistent, at-least-once):
- **audit-service**: Subscribt auf `$all` → immutable logging (GoBD)
- **reporting-service**: Subscribt auf `$all` → baut Materialized Views in PostgreSQL
- **notification-service**: Subscribt auf relevante Kategorien → transformiert zu User-Notifications
- **workflow-service**: Subscribt auf `$all` → evaluiert Trigger-Bedingungen

**Relational Sink** (KurrentDB v26):
- Automatische Projektion von Events nach PostgreSQL-Tabellen
- Konfigurierbar pro Service: welche Events → welche Tabellen
- Reduziert Boilerplate-Code für Read-Model-Updates

### Write-Side vs. Read-Side (CQRS)

```
WRITE (Command):                    READ (Query):
  Client → API Gateway               Client → API Gateway
    → Service                           → Service
      → Validate                          → PostgreSQL (Read-Model)
      → Append Event → KurrentDB          → Return Data
      → Return ACK
                    ↓
          KurrentDB Subscription
                    ↓
          Service projiziert Event
                    ↓
          PostgreSQL (Read-Model updated)
```

## Database-per-Service (Schema Isolation)

```sql
-- Eine PostgreSQL-Instanz, ein Schema pro Service
CREATE SCHEMA auth_schema;
CREATE SCHEMA inventory_schema;
CREATE SCHEMA project_schema;
CREATE SCHEMA scanner_schema;
CREATE SCHEMA warehouse_schema;
CREATE SCHEMA invoice_schema;
CREATE SCHEMA document_schema;
CREATE SCHEMA crew_schema;
CREATE SCHEMA federation_schema;
CREATE SCHEMA maintenance_schema;
CREATE SCHEMA transport_schema;
CREATE SCHEMA insurance_schema;
CREATE SCHEMA workflow_schema;
CREATE SCHEMA ai_schema;
CREATE SCHEMA notification_schema;
CREATE SCHEMA reporting_schema;
CREATE SCHEMA audit_schema;
-- Kein event_store_schema nötig — Events leben in KurrentDB
```

**Regeln:**
- Kein Cross-Schema-Join erlaubt
- Daten anderer Services nur über Events oder synchrone API-Calls
- Jeder Service verwaltet seine eigenen Migrationen

## Shared Go Library (pkg/common)

```
pkg/common/
├── auth/          # JWT-Validation Middleware, RBAC Helpers
├── events/        # KurrentDB Client Wrapper, Event-Schemas (Go Structs)
├── errors/        # Standard Error Types, HTTP Error Responses
├── logging/       # Structured Logging (zerolog)
├── middleware/     # Tenant Extraction, Request-ID, Correlation-ID
├── health/        # Health Check + Readiness Endpoints
├── config/        # Environment-basierte Konfiguration
└── database/      # PostgreSQL Connection Helper, Migration Runner
```

## Multi-Firma Federation (P2P)

```
Firma A (MyRMS-Instanz)          Firma B (MyRMS-Instanz)
┌──────────────────────┐         ┌──────────────────────┐
│  federation-service  │◄─mTLS──►│  federation-service  │
│  (eigene DB, Events) │  REST   │  (eigene DB, Events) │
└──────────────────────┘         └──────────────────────┘
```

- Peer-to-Peer über mTLS (gegenseitige Zertifikat-Authentifizierung)
- Jede Firma kontrolliert welche Daten geteilt werden
- Equipment-Verfügbarkeit, Sub-Rental, Partner-Preise

## Entscheidungen

| Entscheidung | Begründung |
|-------------|-----------|
| Go pro Service | Performance, Single-Binary pro Service, einfaches Docker-Image |
| KurrentDB statt Kafka/NATS | Event-native DB (nicht nachgerüstet), Event Sourcing built-in, Relational Sink nach PostgreSQL, Go-Client, Admin-UI |
| Traefik statt Kong/Nginx | Docker-native Service-Discovery, Labels statt Config-Files |
| Schema-per-Service statt DB-per-Service | Pragmatisch für Self-Hosting (eine DB-Instanz reicht), logisch trotzdem isoliert |
| Redis für Cache/Sessions | Einfach, schnell, weit verbreitet, minimaler Overhead |
| PostgreSQL statt MySQL | JSONB, Row-Level Security, Schema-Support, bessere Erweiterbarkeit |
| React statt Vue | Größeres Ökosystem, TypeScript-First, PWA-Libraries |
| REST statt gRPC (extern) | Einfacher zu debuggen, Firewall-freundlicher |
| mTLS für Federation | Keine zentrale Authority nötig, P2P-kompatibel |
