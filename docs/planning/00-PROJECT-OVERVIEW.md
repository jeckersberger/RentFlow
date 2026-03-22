# MyRMS — Projektplan & Architektur-Übersicht

**Stand:** 20. März 2026
**Version:** 1.0
**Branch:** `claude/event-tech-inventory-system-H42yD`

---

## 1. Projektvision

MyRMS (My Rental Management System) ist eine selbst-gehostete, Open-Source-Lösung für Verleihunternehmen in der Veranstaltungstechnik. Sie ersetzt Excel, Word und WhatsApp durch ein integriertes System für Lagerverwaltung, Projektplanung, Rechnungsstellung und Partner-Zusammenarbeit.

### Zielgruppe
- Veranstaltungstechnik-Firmen mit 1-100 Mitarbeitern im DACH-Raum
- Unternehmen die Self-Hosting bevorzugen (eigene Daten, keine SaaS-Abhängigkeit)
- Firmen die regelmäßig Equipment untereinander verleihen (Federation)

### Abgrenzung zu Wettbewerbern

| Aspekt | Rentman | Easyjob | MyRMS |
|--------|---------|---------|-------|
| Hosting | Cloud-only | On-Premise (Windows) | Self-hosted Docker (jedes OS) |
| Mobile | Eingeschränkt | Sehr eingeschränkt | Full Web-App (responsive + PWA) |
| Sub-Rental | Nur Rentman↔Rentman | Manuell + Assistent | Federation API (offenes Protokoll) |
| Preis | $19-25/User/Monat | ~4.000€ + Support | Self-hosted, keine Lizenzkosten |
| Buchhaltung | QuickBooks/Xero | Eigenes Modul | DATEV/WISO-Export |
| Scanner | QR/Barcode/RFID | Barcode + App | USB, Android-Handheld, Kamera, RFID |
| Offline | Nein | Ja (Desktop) | PWA mit Offline-Sync |
| KI-Integration | Nein | Nein | Multi-Provider AI (Claude, OpenAI, Ollama, ...) |
| E-Check/DGUV V3 | Nein | Nein | Integriert mit Messgeräte-Import |

---

## 2. Architektur-Prinzipien

**Microservice Architecture (MSA) + Event Sourcing + CQRS**

- Jeder Bounded Context ist ein eigenständiger Go-Microservice
- **KurrentDB** (ehemals EventStoreDB) als Event Store — Source of Truth für alle Domain Events
- Services subscriben auf KurrentDB-Streams und projizieren Events in ihre **PostgreSQL-Schemas** (Read-Models)
- Synchrone Kommunikation nur wo nötig (API Gateway → Service)
- Database-per-Service: jeder Service besitzt sein eigenes PostgreSQL-Schema für Read-Projections
- Saga Pattern über Event-Choreografie (KurrentDB Subscriptions)
- CQRS: Write-Side = Events in KurrentDB, Read-Side = Projections in PostgreSQL
- KurrentDB Relational Sink für automatische Projektion nach PostgreSQL
- Kein separater Message Broker nötig — KurrentDB übernimmt Event-Bus + Persistenz

## 3. Tech-Stack

| Schicht | Technologie | Begründung |
|---------|-------------|------------|
| **Backend** | Go 1.22+ (pro Microservice) | Performance, Single-Binary, starke Concurrency |
| **API Gateway** | Traefik v3 | Service-Routing, Load Balancing, Rate Limiting, Docker-native |
| **Event Store** | KurrentDB (ehemals EventStoreDB) | Event-native DB, Event Sourcing, Subscriptions, Relational Sink, Go-Client |
| **Frontend** | TypeScript + React 18 + Sass/SCSS | Großes Ökosystem, TypeScript-First, PWA-fähig |
| **Datenbank** | PostgreSQL 16 (Schema-per-Service) | Read-Projections, JSON-Support, Row-Level Security |
| **Cache** | Redis 7 | Session Store, API-Caching, Rate Limiting |
| **Deployment** | Docker / Docker Compose | Ein `docker compose up` startet alles |
| **Service Discovery** | Docker DNS + Traefik Labels | Kein Consul nötig für Self-Hosting |
| **i18n** | DE/EN von Beginn an | react-i18next (Frontend), Go embed (Backend) |
| **Scanner** | USB-Barcode, Zebra Android-Handheld, Handy-Kamera, RFID | Alle gängigen Szenarien abgedeckt |
| **PDF-Erzeugung** | Go: go-pdf / chromedp | Rechnungen, Verträge, Prüfprotokolle, Labels |
| **E-Mail** | Go: net/smtp + gomail | Rechnungsversand, Mahnungen, Benachrichtigungen |
| **KI** | Multi-Provider (Claude, OpenAI, Gemini, Mistral, Ollama) | Flexibilität, kein Vendor Lock-in |
| **Echtzeit** | WebSocket (gorilla/websocket) | Live-Updates Dashboard, Scanner-Feedback, Notifications |
| **Kalender-Sync** | CalDAV / ICS-Export | Crew-Kalender kompatibel mit iPhone, Android, Outlook |
| **Backup** | pg_dump + AES-256 Verschlüsselung | Automatisch, Multi-Destination |
| **Observability** | Prometheus + Grafana + zerolog | Metriken, Dashboards, strukturiertes Logging |
| **Tracing** | Correlation-ID (alle Services) | Verteiltes Request-Tracing über Service-Grenzen |

---

## 4. System-Architektur (MSA + EDA)

```
┌────────────────────────────────────────────────────────────────┐
│                          Clients                                │
│  Desktop-Browser │ Tablet │ Handy (PWA) │ Zebra TC21 │ Portal   │
└──────────────────────────┬─────────────────────────────────────┘
                           │ HTTPS
┌──────────────────────────▼─────────────────────────────────────┐
│                   Traefik v3 (API Gateway)                      │
│          Routing │ Load Balancing │ Rate Limiting │ TLS          │
└──┬───┬───┬───┬───┬───┬───┬───┬───┬───┬───┬───┬───┬───┬───┬──┘
   │   │   │   │   │   │   │   │   │   │   │   │   │   │   │
   ▼   ▼   ▼   ▼   ▼   ▼   ▼   ▼   ▼   ▼   ▼   ▼   ▼   ▼   ▼
┌─────┐┌─────┐┌─────┐┌─────┐┌─────┐┌─────┐┌─────┐┌─────┐┌─────┐
│Auth ││Inven││Proj ││Scan ││Ware ││Invoic││Doc  ││Crew ││Feder│
│Svc  ││tory ││ect  ││ner  ││house││e Svc ││Svc  ││Svc  ││ation│
│:8001││:8002││:8003││:8004││:8005││:8006 ││:8007││:8008││:8009│
└──┬──┘└──┬──┘└──┬──┘└──┬──┘└──┬──┘└──┬──┘└──┬──┘└──┬──┘└──┬──┘
   │      │      │      │      │      │      │      │      │
   ▼      ▼      ▼      ▼      ▼      ▼      ▼      ▼      ▼
┌────────────────────────────────────────────────────────────────┐
│          KurrentDB (Event Store — Source of Truth)              │
│                                                                 │
│  Streams: equipment-, project-, invoice-, warehouse-, scan-,    │
│           maintenance-, transport-, insurance-, crew-, workflow-,│
│           document-, ai-, notification-                          │
│                                                                 │
│  Subscriptions → Services projizieren in PostgreSQL             │
│  Relational Sink → automatische Projektion nach PostgreSQL      │
└──┬──────────────────────────────────────────────────────────┬──┘
   │  Catch-up Subscriptions                                  │
   ▼                                                          ▼
┌─────┐┌─────┐┌─────┐┌─────┐┌─────┐┌─────┐┌─────┐┌────────────┐
│Maint││Trans││Insur││Work ││AI   ││Notif││Repor││Audit Svc   │
│enanc││port ││ance ││flow ││Svc  ││icat.││ting ││(GoBD)      │
│e Svc││Svc  ││Svc  ││Svc  ││:8014││Svc  ││Svc  ││:8017       │
│:8010││:8011││:8012││:8013││     ││:8015││:8016││(konsumiert  │
└──┬──┘└──┬──┘└──┬──┘└──┬──┘└──┬──┘└──┬──┘└──┬──┘│ALLE Events)│
   │      │      │      │      │      │      │   └────────────┘
   ▼      ▼      ▼      ▼      ▼      ▼      ▼
┌────────────────────────────────────────────────────────────────┐
│         PostgreSQL 16 (Read-Projections, Schema-per-Service)    │
│                                                                 │
│  auth_schema │ inventory_schema │ project_schema │ scanner_     │
│  warehouse_  │ invoice_schema   │ document_      │ crew_schema  │
│  federation_ │ maintenance_     │ transport_     │ insurance_   │
│  workflow_   │ ai_schema        │ notification_  │ reporting_   │
│  audit_schema                                                   │
└────────────────────────────────────────────────────────────────┘

            ┌──────────────────────────────────┐
            │           Redis 7                 │
            │  Sessions │ Cache │ Rate Limits   │
            └──────────────────────────────────┘

Firma A ◄──── Federation API (mTLS REST) ────► Firma B
```

### Microservice-Übersicht

| Service | Port | Module | Verantwortung |
|---------|------|--------|---------------|
| auth-service | :8001 | AUTH | JWT, RBAC, User, Tenants |
| inventory-service | :8002 | INV, A1, L1, L2, N5, N3 | Equipment, Kategorien, Preise, Labels, Flightcases |
| project-service | :8003 | PROJ, N6 | Projekte, Packlisten, Reservierungen, Timeline |
| scanner-service | :8004 | SCAN | QR/Barcode/RFID, Check-In/Out, Offline-Sync |
| warehouse-service | :8005 | L3, L4, L5 | Lagerplätze, Warenbewegung, Inventur, Optimierung, Reporting |
| invoice-service | :8006 | INV-FIN, BANK | Rechnungen, Mahnwesen, DATEV, Bank-Import |
| document-service | :8007 | DOC, J1, D2 | PDF, Templates, Verträge, Labels, I17-OCR |
| crew-service | :8008 | N1, I14 | Crew, Freelancer, Zeiterfassung, CalDAV |
| federation-service | :8009 | FED | mTLS P2P, Equipment-Sharing, Sub-Rental |
| maintenance-service | :8010 | J2, N2 | Wartung, E-Check/DGUV V3, Checklisten |
| transport-service | :8011 | J3, I11, I16 | Fahrzeuge, Touren, Routen, Transportkosten |
| insurance-service | :8012 | K2, K3, I12, I13 | Policen, Schäden, Risiko-Analyse |
| workflow-service | :8013 | K1 | No-Code Workflows, Event-Trigger |
| ai-service | :8014 | I1-I3, I6, I9, I10, I15 | Multi-Provider AI, Anonymisierung, Prognosen |
| notification-service | :8015 | N7, EMAIL | In-App, Push, E-Mail, Preferences |
| reporting-service | :8016 | DASH, L5-Nachh | CQRS Read-Side, KPIs, Dashboards, ESG |
| audit-service | :8017 | N8 | GoBD Audit-Trail, Checksummen |

### Event-Flow Beispiel: Equipment Check-Out

```
1. Lisa scannt QR-Code → scanner-service
2. scanner-service appended Event "scan.completed" → KurrentDB Stream "scan-events"
3. scanner-service ruft inventory-service API auf: POST /equipment/{id}/check-out
4. inventory-service ändert Status → appended Event "equipment.checked_out" → KurrentDB Stream "equipment-{id}"
5. Parallel reagieren:
   ├─ project-service:     aktualisiert Packlisten-Status
   ├─ warehouse-service:   bucht Warenbewegung (Auslagerung)
   ├─ notification-service: benachrichtigt Projektleiter
   ├─ reporting-service:   aktualisiert Auslastungs-KPI
   └─ audit-service:       loggt unveränderlich (GoBD)
```

### Event-Schema (Standard für alle Services)

```json
{
  "event_id": "550e8400-e29b-41d4-a716-446655440000",
  "event_type": "equipment.checked_out",
  "aggregate_type": "equipment",
  "aggregate_id": "uuid-des-equipments",
  "tenant_id": "uuid-der-firma",
  "correlation_id": "uuid-für-request-tracing",
  "timestamp": "2026-03-20T10:30:00Z",
  "version": 1,
  "data": {
    "equipment_id": "...",
    "project_id": "...",
    "checked_out_by": "user-uuid",
    "condition": "ok"
  },
  "metadata": {
    "user_id": "uuid",
    "source_service": "inventory-service"
  }
}
```

---

## 5. Docker Compose Stack (Produktion)

```yaml
services:
  # ─── Infrastruktur ───────────────────────────────────────
  traefik:
    image: traefik:v3.0
    command:
      - "--api.insecure=false"
      - "--providers.docker=true"
      - "--providers.docker.exposedbydefault=false"
      - "--entrypoints.web.address=:80"
      - "--entrypoints.websecure.address=:443"
      - "--certificatesresolvers.letsencrypt.acme.httpchallenge.entrypoint=web"
      - "--certificatesresolvers.letsencrypt.acme.email=${ADMIN_EMAIL}"
      - "--certificatesresolvers.letsencrypt.acme.storage=/letsencrypt/acme.json"
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - letsencrypt:/letsencrypt
    restart: unless-stopped

  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: myrms
      POSTGRES_USER: myrms
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - pgdata:/var/lib/postgresql/data
      - ./scripts/init-schemas.sql:/docker-entrypoint-initdb.d/01-schemas.sql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U myrms"]
      interval: 5s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  kurrentdb:
    image: ghcr.io/kurrent-io/kurrentdb:latest
    environment:
      KURRENTDB_INSECURE: "true"  # Für Dev; in Prod: TLS-Zertifikate
      KURRENTDB_ENABLE_ATOM_PUB_OVER_HTTP: "true"
      KURRENTDB_MEM_DB: "false"
    ports:
      - "2113:2113"   # HTTP API + Admin UI
    volumes:
      - kurrentdata:/var/lib/kurrentdb
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:2113/health/live"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    command: ["redis-server", "--appendonly", "yes"]
    volumes:
      - redisdata:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 3
    restart: unless-stopped

  # ─── Frontend ────────────────────────────────────────────
  frontend:
    image: myrms/frontend:latest
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.frontend.rule=Host(`${DOMAIN}`)"
      - "traefik.http.routers.frontend.tls.certresolver=letsencrypt"
    restart: unless-stopped

  # ─── Microservices ───────────────────────────────────────
  auth-service:
    image: myrms/auth-service:latest
    environment:
      DATABASE_URL: postgres://myrms:${DB_PASSWORD}@db:5432/myrms?sslmode=disable&search_path=auth_schema
      KURRENTDB_URL: esdb://kurrentdb:2113?tls=false
      REDIS_URL: redis://redis:6379
      JWT_SECRET: ${SECRET_KEY}
    depends_on: { db: { condition: service_healthy }, kurrentdb: { condition: service_healthy }, redis: { condition: service_healthy } }
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.auth.rule=Host(`${DOMAIN}`) && PathPrefix(`/api/auth`, `/api/users`, `/api/roles`)"
    restart: unless-stopped

  inventory-service:
    image: myrms/inventory-service:latest
    environment:
      DATABASE_URL: postgres://myrms:${DB_PASSWORD}@db:5432/myrms?sslmode=disable&search_path=inventory_schema
      KURRENTDB_URL: esdb://kurrentdb:2113?tls=false
      REDIS_URL: redis://redis:6379
      AUTH_SERVICE_URL: http://auth-service:8001
    depends_on: { db: { condition: service_healthy }, kurrentdb: { condition: service_healthy } }
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.inventory.rule=Host(`${DOMAIN}`) && PathPrefix(`/api/equipment`, `/api/categories`, `/api/labels`, `/api/price-rules`, `/api/bundles`)"
    restart: unless-stopped

  project-service:
    image: myrms/project-service:latest
    environment:
      DATABASE_URL: postgres://myrms:${DB_PASSWORD}@db:5432/myrms?sslmode=disable&search_path=project_schema
      KURRENTDB_URL: esdb://kurrentdb:2113?tls=false
      REDIS_URL: redis://redis:6379
    depends_on: { db: { condition: service_healthy }, kurrentdb: { condition: service_healthy } }
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.project.rule=Host(`${DOMAIN}`) && PathPrefix(`/api/projects`)"
    restart: unless-stopped

  scanner-service:
    image: myrms/scanner-service:latest
    environment:
      DATABASE_URL: postgres://myrms:${DB_PASSWORD}@db:5432/myrms?sslmode=disable&search_path=scanner_schema
      KURRENTDB_URL: esdb://kurrentdb:2113?tls=false
    depends_on: { db: { condition: service_healthy }, kurrentdb: { condition: service_healthy } }
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.scanner.rule=Host(`${DOMAIN}`) && PathPrefix(`/api/scan`, `/api/check-in`, `/api/check-out`)"
    restart: unless-stopped

  warehouse-service:
    image: myrms/warehouse-service:latest
    environment:
      DATABASE_URL: postgres://myrms:${DB_PASSWORD}@db:5432/myrms?sslmode=disable&search_path=warehouse_schema
      KURRENTDB_URL: esdb://kurrentdb:2113?tls=false
    depends_on: { db: { condition: service_healthy }, kurrentdb: { condition: service_healthy } }
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.warehouse.rule=Host(`${DOMAIN}`) && PathPrefix(`/api/warehouse`, `/api/inventory/optimization`)"
    restart: unless-stopped

  invoice-service:
    image: myrms/invoice-service:latest
    environment:
      DATABASE_URL: postgres://myrms:${DB_PASSWORD}@db:5432/myrms?sslmode=disable&search_path=invoice_schema
      KURRENTDB_URL: esdb://kurrentdb:2113?tls=false
      SMTP_HOST: ${SMTP_HOST:-}
      SMTP_PORT: ${SMTP_PORT:-587}
      SMTP_USER: ${SMTP_USER:-}
      SMTP_PASSWORD: ${SMTP_PASSWORD:-}
    depends_on: { db: { condition: service_healthy }, kurrentdb: { condition: service_healthy } }
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.invoice.rule=Host(`${DOMAIN}`) && PathPrefix(`/api/invoices`, `/api/credit-notes`, `/api/dunning`, `/api/datev`, `/api/bank`)"
    restart: unless-stopped

  document-service:
    image: myrms/document-service:latest
    environment:
      DATABASE_URL: postgres://myrms:${DB_PASSWORD}@db:5432/myrms?sslmode=disable&search_path=document_schema
      KURRENTDB_URL: esdb://kurrentdb:2113?tls=false
    depends_on: { db: { condition: service_healthy }, kurrentdb: { condition: service_healthy } }
    volumes:
      - uploads:/app/uploads
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.document.rule=Host(`${DOMAIN}`) && PathPrefix(`/api/documents`, `/api/templates`, `/api/contracts`)"
    restart: unless-stopped

  crew-service:
    image: myrms/crew-service:latest
    environment:
      DATABASE_URL: postgres://myrms:${DB_PASSWORD}@db:5432/myrms?sslmode=disable&search_path=crew_schema
      KURRENTDB_URL: esdb://kurrentdb:2113?tls=false
    depends_on: { db: { condition: service_healthy }, kurrentdb: { condition: service_healthy } }
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.crew.rule=Host(`${DOMAIN}`) && PathPrefix(`/api/crew`)"
    restart: unless-stopped

  federation-service:
    image: myrms/federation-service:latest
    environment:
      DATABASE_URL: postgres://myrms:${DB_PASSWORD}@db:5432/myrms?sslmode=disable&search_path=federation_schema
      KURRENTDB_URL: esdb://kurrentdb:2113?tls=false
    depends_on: { db: { condition: service_healthy }, kurrentdb: { condition: service_healthy } }
    volumes:
      - federation_certs:/app/certs
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.federation.rule=Host(`${DOMAIN}`) && PathPrefix(`/api/federation`)"
    restart: unless-stopped

  maintenance-service:
    image: myrms/maintenance-service:latest
    environment:
      DATABASE_URL: postgres://myrms:${DB_PASSWORD}@db:5432/myrms?sslmode=disable&search_path=maintenance_schema
      KURRENTDB_URL: esdb://kurrentdb:2113?tls=false
    depends_on: { db: { condition: service_healthy }, kurrentdb: { condition: service_healthy } }
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.maintenance.rule=Host(`${DOMAIN}`) && PathPrefix(`/api/maintenance`, `/api/echeck`)"
    restart: unless-stopped

  transport-service:
    image: myrms/transport-service:latest
    environment:
      DATABASE_URL: postgres://myrms:${DB_PASSWORD}@db:5432/myrms?sslmode=disable&search_path=transport_schema
      KURRENTDB_URL: esdb://kurrentdb:2113?tls=false
    depends_on: { db: { condition: service_healthy }, kurrentdb: { condition: service_healthy } }
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.transport.rule=Host(`${DOMAIN}`) && PathPrefix(`/api/vehicles`, `/api/tours`, `/api/ai/routes`, `/api/ai/transport`)"
    restart: unless-stopped

  insurance-service:
    image: myrms/insurance-service:latest
    environment:
      DATABASE_URL: postgres://myrms:${DB_PASSWORD}@db:5432/myrms?sslmode=disable&search_path=insurance_schema
      KURRENTDB_URL: esdb://kurrentdb:2113?tls=false
    depends_on: { db: { condition: service_healthy }, kurrentdb: { condition: service_healthy } }
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.insurance.rule=Host(`${DOMAIN}`) && PathPrefix(`/api/insurance`, `/api/damage-reports`, `/api/ai/insurance`, `/api/ai/damage`)"
    restart: unless-stopped

  workflow-service:
    image: myrms/workflow-service:latest
    environment:
      DATABASE_URL: postgres://myrms:${DB_PASSWORD}@db:5432/myrms?sslmode=disable&search_path=workflow_schema
      KURRENTDB_URL: esdb://kurrentdb:2113?tls=false
    depends_on: { db: { condition: service_healthy }, kurrentdb: { condition: service_healthy } }
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.workflow.rule=Host(`${DOMAIN}`) && PathPrefix(`/api/workflows`)"
    restart: unless-stopped

  ai-service:
    image: myrms/ai-service:latest
    environment:
      DATABASE_URL: postgres://myrms:${DB_PASSWORD}@db:5432/myrms?sslmode=disable&search_path=ai_schema
      KURRENTDB_URL: esdb://kurrentdb:2113?tls=false
      REDIS_URL: redis://redis:6379
    depends_on: { db: { condition: service_healthy }, kurrentdb: { condition: service_healthy } }
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.ai.rule=Host(`${DOMAIN}`) && PathPrefix(`/api/ai`)"
    restart: unless-stopped

  notification-service:
    image: myrms/notification-service:latest
    environment:
      DATABASE_URL: postgres://myrms:${DB_PASSWORD}@db:5432/myrms?sslmode=disable&search_path=notification_schema
      KURRENTDB_URL: esdb://kurrentdb:2113?tls=false
      SMTP_HOST: ${SMTP_HOST:-}
      SMTP_PORT: ${SMTP_PORT:-587}
      SMTP_USER: ${SMTP_USER:-}
      SMTP_PASSWORD: ${SMTP_PASSWORD:-}
    depends_on: { db: { condition: service_healthy }, kurrentdb: { condition: service_healthy } }
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.notification.rule=Host(`${DOMAIN}`) && PathPrefix(`/api/notifications`)"
    restart: unless-stopped

  reporting-service:
    image: myrms/reporting-service:latest
    environment:
      DATABASE_URL: postgres://myrms:${DB_PASSWORD}@db:5432/myrms?sslmode=disable&search_path=reporting_schema
      KURRENTDB_URL: esdb://kurrentdb:2113?tls=false
      REDIS_URL: redis://redis:6379
    depends_on: { db: { condition: service_healthy }, kurrentdb: { condition: service_healthy } }
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.reporting.rule=Host(`${DOMAIN}`) && PathPrefix(`/api/reports`, `/api/dashboard`, `/api/kpis`)"
    restart: unless-stopped

  audit-service:
    image: myrms/audit-service:latest
    environment:
      DATABASE_URL: postgres://myrms:${DB_PASSWORD}@db:5432/myrms?sslmode=disable&search_path=audit_schema
      KURRENTDB_URL: esdb://kurrentdb:2113?tls=false
    depends_on: { db: { condition: service_healthy }, kurrentdb: { condition: service_healthy } }
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.audit.rule=Host(`${DOMAIN}`) && PathPrefix(`/api/audit`)"
    restart: unless-stopped

volumes:
  pgdata:
  kurrentdata:
  redisdata:
  uploads:
  backups:
  letsencrypt:
  federation_certs:
```

**Minimal `.env`:**
```env
DOMAIN=myrms.example.com
DB_PASSWORD=sicheres-passwort-hier
ADMIN_EMAIL=marco@example.com
SECRET_KEY=generierter-jwt-secret
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=noreply@example.com
SMTP_PASSWORD=smtp-passwort
```

### Skalierung (bei Bedarf)

```yaml
# Einzelne Services horizontal skalieren:
services:
  scanner-service:
    deploy:
      replicas: 3  # Festival-Saison: viele parallele Scans
  ai-service:
    deploy:
      replicas: 2  # KI-Anfragen parallelisieren
      resources:
        limits:
          memory: 2G
```

---

## 5. Rollen & Berechtigungsmodell

| Rolle | Dashboard | Projekte | Lager/Scan | Preise | Rechnungen | Crew | Federation | Admin |
|-------|-----------|----------|------------|--------|------------|------|------------|-------|
| **Admin** | Voll | Voll | Voll | Ja | Voll | Voll | Voll | Ja |
| **Projektleiter** | Voll | Voll | Lesen | Ja | Freigabe | Planen | Anfragen | Nein |
| **Lager** | Lager-Widget | Packlisten | Voll | Nein | Nein | Nein | In/Out | Nein |
| **Buchhaltung** | Finanz-Widget | Lesen | Nein | Ja | Voll | Nein | Rechnungen | Nein |
| **Freelancer** | Eigene Jobs | Eigene | Scan | Nein | Nein | Eigene | Nein | Nein |
| **Kunden-Portal** | Nein | Eigene | Nein | Eigene | Download | Nein | Nein | Nein |
| **Benutzerdefiniert** | Konfigurierbar | Konfigurierbar | Konfigurierbar | Konfigurierbar | Konfigurierbar | Konfigurierbar | Konfigurierbar | Nein |

---

## 6. Modulübersicht

### Basis-Module (Phase 1 — MVP)

| ID | Modul | Beschreibung | Persona |
|----|-------|-------------|---------|
| AUTH | Authentifizierung & Rollen | JWT, RBAC, 6 Rollen, Einladungslinks | Admin |
| DASH | Dashboard | 30-Sekunden-Überblick, rollenbasiert | Marco |
| INV | Inventar & Assets | Equipment-Verwaltung, Kategorien, Seriennummern | Lisa, Marco |
| PROJ | Projekte & Packlisten | Projekt-Lifecycle, Equipment-Zuordnung, Doppelbuchungs-Check | Marco, Lisa |
| SCAN | Scanner & Check-In/Out | QR, Barcode, RFID, Kamera, Zebra TC21, Offline-fähig | Lisa |
| A1 | Artikelnummern & QR-Labels | Interne Nummern, ZPL-Labels, Metrawatt-Kompatibilität | Lisa, Admin |

### Business-Module (Phase 2)

| ID | Modul | Beschreibung | Persona |
|----|-------|-------------|---------|
| L1 | Flexible Lagerorte | Hierarchisch, Emoji-Icons, Bulk-Move | Lisa |
| L2 | Preiskalkulations-Engine | Staffelpreise, Mengenrabatte, Saisonzuschläge, Bundles | Marco |
| DOC | Dokument-Engine | Angebote, Rechnungen, Lieferscheine, Layout-Editor | Thomas, Marco |
| INV-FIN | Rechnungen & Buchhaltung | Angebot→Rechnung, Teilrechnungen, Gutschriften, Mahnwesen | Thomas |
| BANK | Bank-Integration | DATEV/WISO-Export, Kontoauszug-Import, OP-Abgleich | Thomas |
| EMAIL | E-Mail-System | Rechnungs-/Mahnversand, Templates, Anhänge | Thomas |
| FED | Federation | mTLS P2P, Equipment-Sharing, Sub-Rental, Partner-Preise | Stefan/SoundPro |
| N8 | Audit-Trail | GoBD-konform, unveränderliches Log, jede Entität | Thomas, Admin |

### Erweiterte Module (Phase 3)

| ID | Modul | Beschreibung | Persona |
|----|-------|-------------|---------|
| N1 | Crew & Personalplanung | Freelancer-DB, Schichtplanung, Zeiterfassung, CalDAV-Sync | Marco, Kevin |
| N5 | Flightcase-Management | 3-Tier-Validierung, RFID-Scan, Typ-basierte Zubehör-Zuordnung | Lisa |
| N2 | E-Check / DGUV V3 | Prüfdoku, Messgeräte-Import (IZYTRON.IQ), Gerätesperre | Lisa, Admin |
| N3 | Verbrauchsmaterial | Mengenbasiert, Mindestbestand, Nachbestellung, Lieferanten | Lisa |
| J1 | Vertragsmanagement | Templates, Platzhalter, Versionierung, digitale Signatur | Marco |
| J2 | Wartung & Predictive Maintenance | Wartungspläne, Checklisten, Erinnerungen, Foto-Doku | Lisa |
| J3 | Transport & Logistik | Fahrzeuge, Fahrer, Tourenplanung, Kapazitätsprüfung | Marco |
| K1 | Workflow-Engine | No-Code WENN→DANN, Event/CRON/Manuelle Trigger | Marco, Admin |
| K2 | Versicherungsmanagement | Policen, Deckungsprüfung, Ablauf-Warnungen | Marco |
| K3 | Schadenmanagement | 8-Schritte-Prozess, KVA, Fotos, Versicherungsmeldung | Lisa, Marco |
| N6 | Kalender & Timeline | Gantt-Projekte, Equipment-Verfügbarkeit, Crew-Kalender | Marco |
| N7 | Benachrichtigungs-Center | In-App Bell, Push, E-Mail, pro-User konfigurierbar | Alle |

### Portal & Erweiterungen (Phase 4)

| ID | Modul | Beschreibung | Persona |
|----|-------|-------------|---------|
| N4 | Kunden-Portal | Self-Service: Angebote, Rechnungen, Verfügbarkeit, Signatur | Kunden |
| L3 | Online-Buchungsportal | Öffentlicher Katalog, Warenkorb, Registrierung | Kunden |
| L4 | Backup & Disaster Recovery | Auto-Backup, Verschlüsselung, Restore-Test | Admin |
| L5 | Nachhaltigkeits-Reporting | CO2-Tracking, Energieverbrauch, ESG-Reports | Marco |
| D2 | Label-Designer | Zebra ZPL, visueller Editor, Artikelnummer+QR | Lisa |

### KI-Module (Phase 5)

| ID | Modul | Beschreibung |
|----|-------|-------------|
| I1-I3 | Multi-KI-Provider | 6 Adapter, Task-Routing, Fallback, Usage-Tracking |
| I6 | KI-Anonymisierung | Cloud-Pflicht-Anonymisierung, 7 Regex-Patterns |
| I9 | Smart Asset Creator | Hersteller→Auto-Fill, Foto-Erkennung, Bulk |
| I10 | KI-Lernsystem | Feedback, Few-Shot, Prompt-Versioning, A/B |
| I11 | KI-Transportkosten | Prognose, TCO, Budget-Forecast, Anomalien |
| I12 | KI-Versicherungsanalyse | Risiko-Bewertung, Deckungslücken, Claim-Assistent |
| I13 | KI-Bestandsoptimierung | Nachfrage-Prognose, Lagermengen, Nachbestellung |
| I14 | KI-Mahnwesen | Zahlungswahrscheinlichkeit, Mahnstufen-Optimierung |
| I15 | KI-Reporting | Anomalie-Erkennung, Trend-Vorhersage, NLP-Summaries |
| I16 | KI-Workflow-Optimierung | Vorschläge, Engpass-Erkennung, Trigger-Optimierung |
| I17 | KI-Packoptimierung | Packsequenz, Gewichtsverteilung, Fehlteile-Vorhersage |

---

## 7. Git-Workflow

```
main (stable releases)
 ├── develop (aktive Entwicklung)
 │    ├── feature/* (pro Modul/Feature)
 │    └── fix/* (Bugfixes)
 └── release/vX.Y.Z (Release-Vorbereitung)
```

### Commit-Konventionen
```
<type>(<scope>): <kurze Beschreibung>

Types: feat, fix, docs, refactor, test, chore, perf
Scopes: auth, inventory, scanner, invoicing, federation, crew, echeck, ...
```

### Versioning (SemVer)
- `MAJOR` = Breaking Changes / große neue Features
- `MINOR` = Neue Features, abwärtskompatibel
- `PATCH` = Bugfixes, kleine Verbesserungen

---

## 8. Verknüpfte Dokumente

| Datei | Inhalt |
|-------|--------|
| [01-DATABASE-SCHEMA.md](./01-DATABASE-SCHEMA.md) | Komplettes PostgreSQL-Schema aller Module |
| [02-BACKEND-SERVICES.md](./02-BACKEND-SERVICES.md) | Go-Packages, Services, API-Endpunkte |
| [03-FRONTEND-COMPONENTS.md](./03-FRONTEND-COMPONENTS.md) | React-Komponenten, Routing, State |
| [04-MODULE-SPECS.md](./04-MODULE-SPECS.md) | Detailspezifikationen aller Module |
| [05-IMPLEMENTATION-PHASES.md](./05-IMPLEMENTATION-PHASES.md) | Phasenplan, Prioritäten, Abhängigkeiten |
