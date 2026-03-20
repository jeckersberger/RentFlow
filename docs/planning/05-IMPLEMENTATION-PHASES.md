# 05 - Implementierungsphasen: MyRMS Entwicklungs-Roadmap

## Übersicht

Dieses Dokument beschreibt den vollständigen Implementierungsplan für **MyRMS** – das selbst-gehostete Rental Management System für Veranstaltungstechnik-Unternehmen im DACH-Raum. Die Entwicklung ist in **5 Phasen** unterteilt (Phase 0–4), die sequenziell aufbauen und jeweils deploybare Zwischenstände liefern.

**Gesamtdauer:** 12–18 Monate (Solo-Entwickler + gelegentliche Freelancer-Unterstützung)
**Deployment-Strategie:** Self-hosted via Docker Compose, kein Cloud-Zwang
**Zielgruppe Phase 1 MVP:** Unternehmen wie Marco Bergers Firma (2–5 Mitarbeiter, 800k€ Equipment-Wert)

---

## Phasen-Übersicht

| Phase | Name | Dauer | Hauptziel |
|-------|------|-------|-----------|
| Phase 0 | Infrastruktur & Setup | 4–6 Wochen | Docker-Stack läuft, CI/CD aktiv |
| Phase 1 | Core MVP | 10–14 Wochen | Inventar + Projekte + Rechnung + Auth |
| Phase 2 | Operations | 8–12 Wochen | Lager + Scanner + Transport + Wartung |
| Phase 3 | Business | 8–10 Wochen | Crew + Dokumente + Versicherung + Reporting |
| Phase 4 | Advanced Features | 10–14 Wochen | KI + Workflow + Federation + Audit |

---

## Phase 0: Infrastruktur & Setup

**Dauer:** 4–6 Wochen
**Voraussetzung:** Keine
**Lieferbar:** Vollständig konfigurierter Docker-Stack, leere Datenbanken, CI/CD-Pipeline

### Ziele

- Lokale Entwicklungsumgebung für alle 17 Microservices
- Docker Compose Production-Setup
- KurrentDB + PostgreSQL + Redis laufen stabil
- Traefik als API Gateway konfiguriert
- CI/CD-Pipeline mit GitHub Actions
- Shared Go-Library (pkg/common) implementiert
- Frontend-Scaffold mit Vite + React + TanStack

### Meilensteine

#### M0.1: Repository & Projektstruktur (Woche 1)

```
myRMS/
├── services/
│   ├── auth/           # Port 8001
│   ├── inventory/      # Port 8002
│   ├── project/        # Port 8003
│   ├── scanner/        # Port 8004
│   ├── warehouse/      # Port 8005
│   ├── invoice/        # Port 8006
│   ├── document/       # Port 8007
│   ├── crew/           # Port 8008
│   ├── federation/     # Port 8009
│   ├── maintenance/    # Port 8010
│   ├── transport/      # Port 8011
│   ├── insurance/      # Port 8012
│   ├── workflow/       # Port 8013
│   ├── ai/             # Port 8014
│   ├── notification/   # Port 8015
│   ├── reporting/      # Port 8016
│   └── audit/          # Port 8017
├── pkg/
│   └── common/         # Shared Go-Library
├── frontend/           # React 18 + Vite
├── infra/
│   ├── docker/
│   │   ├── docker-compose.yml
│   │   ├── docker-compose.dev.yml
│   │   └── docker-compose.prod.yml
│   ├── traefik/
│   │   ├── traefik.yml
│   │   └── dynamic/
│   ├── postgres/
│   │   └── init-schemas.sql
│   └── kurrentdb/
│       └── kurrentdb.conf
├── docs/
│   ├── planning/
│   └── architecture/
├── .github/
│   └── workflows/
│       ├── ci.yml
│       ├── build.yml
│       └── deploy.yml
├── Makefile
└── README.md
```

**Aufgaben:**
- [ ] Git-Repository initialisieren
- [ ] Verzeichnisstruktur anlegen
- [ ] `.gitignore` für Go + Node + Docker konfigurieren
- [ ] `go.work` für Multi-Module Workspace einrichten
- [ ] Makefile mit Standard-Targets erstellen

#### M0.2: Docker Compose Stack (Woche 1–2)

**`infra/docker/docker-compose.yml`:**

```yaml
version: "3.9"

networks:
  myrms-internal:
    driver: bridge
  myrms-external:
    driver: bridge

volumes:
  postgres_data:
  kurrentdb_data:
  redis_data:
  traefik_certs:

services:
  traefik:
    image: traefik:v3.0
    container_name: myrms-traefik
    restart: unless-stopped
    command:
      - "--api.insecure=false"
      - "--providers.docker=true"
      - "--providers.docker.exposedbydefault=false"
      - "--entrypoints.web.address=:80"
      - "--entrypoints.websecure.address=:443"
      - "--certificatesresolvers.letsencrypt.acme.httpchallenge=true"
      - "--certificatesresolvers.letsencrypt.acme.email=${ACME_EMAIL}"
      - "--certificatesresolvers.letsencrypt.acme.storage=/certs/acme.json"
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - traefik_certs:/certs
    networks:
      - myrms-internal
      - myrms-external

  postgres:
    image: postgres:16-alpine
    container_name: myrms-postgres
    restart: unless-stopped
    environment:
      POSTGRES_USER: ${POSTGRES_USER}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      POSTGRES_DB: myrms
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./infra/postgres/init-schemas.sql:/docker-entrypoint-initdb.d/init.sql
    networks:
      - myrms-internal
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER}"]
      interval: 10s
      timeout: 5s
      retries: 5

  kurrentdb:
    image: docker.eventstore.com/eventstore/eventstoredb:23.10.0-jammy
    container_name: myrms-kurrentdb
    restart: unless-stopped
    environment:
      EVENTSTORE_CLUSTER_SIZE: 1
      EVENTSTORE_RUN_PROJECTIONS: All
      EVENTSTORE_START_STANDARD_PROJECTIONS: true
      EVENTSTORE_INSECURE: false
      EVENTSTORE_CERTIFICATE_FILE: /etc/eventstore/certs/node.crt
      EVENTSTORE_CERTIFICATE_PRIVATE_KEY_FILE: /etc/eventstore/certs/node.key
    volumes:
      - kurrentdb_data:/var/lib/eventstore
    networks:
      - myrms-internal
    healthcheck:
      test: ["CMD-SHELL", "curl -sf http://localhost:2113/health/live || exit 1"]
      interval: 15s
      timeout: 10s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: myrms-redis
    restart: unless-stopped
    command: redis-server --requirepass ${REDIS_PASSWORD} --appendonly yes
    volumes:
      - redis_data:/data
    networks:
      - myrms-internal
    healthcheck:
      test: ["CMD", "redis-cli", "--no-auth-warning", "-a", "${REDIS_PASSWORD}", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
```

**Aufgaben:**
- [ ] `docker-compose.yml` für Infrastruktur-Services
- [ ] `docker-compose.dev.yml` mit Hot-Reload für alle Services
- [ ] `docker-compose.prod.yml` mit Ressource-Limits
- [ ] `.env.example` Datei mit allen Umgebungsvariablen
- [ ] PostgreSQL Init-Skript für alle 17 Schemas
- [ ] KurrentDB TLS-Zertifikat-Generierung
- [ ] Redis-Passwort und Persistence konfigurieren
- [ ] Traefik TLS + ACME konfigurieren
- [ ] Health-Check-Skript: `scripts/health-check.sh`

#### M0.3: PostgreSQL Schema-Initialisierung (Woche 2)

**`infra/postgres/init-schemas.sql`:**

```sql
-- Alle 17 Schemas anlegen
CREATE SCHEMA IF NOT EXISTS auth_schema;
CREATE SCHEMA IF NOT EXISTS inventory_schema;
CREATE SCHEMA IF NOT EXISTS project_schema;
CREATE SCHEMA IF NOT EXISTS scanner_schema;
CREATE SCHEMA IF NOT EXISTS warehouse_schema;
CREATE SCHEMA IF NOT EXISTS invoice_schema;
CREATE SCHEMA IF NOT EXISTS document_schema;
CREATE SCHEMA IF NOT EXISTS crew_schema;
CREATE SCHEMA IF NOT EXISTS federation_schema;
CREATE SCHEMA IF NOT EXISTS maintenance_schema;
CREATE SCHEMA IF NOT EXISTS transport_schema;
CREATE SCHEMA IF NOT EXISTS insurance_schema;
CREATE SCHEMA IF NOT EXISTS workflow_schema;
CREATE SCHEMA IF NOT EXISTS ai_schema;
CREATE SCHEMA IF NOT EXISTS notification_schema;
CREATE SCHEMA IF NOT EXISTS reporting_schema;
CREATE SCHEMA IF NOT EXISTS audit_schema;

-- Service-User mit Schema-Rechten anlegen
CREATE USER auth_user WITH PASSWORD '${AUTH_DB_PASSWORD}';
GRANT USAGE, CREATE ON SCHEMA auth_schema TO auth_user;
-- [... je User für alle 17 Services ...]

-- Extensions aktivieren
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";  -- Volltext-Suche
CREATE EXTENSION IF NOT EXISTS "btree_gin"; -- GIN-Indizes
```

**Aufgaben:**
- [ ] Init-SQL für alle 17 Schemas
- [ ] Service-spezifische Datenbanknutzer
- [ ] Extensions (uuid-ossp, pg_trgm, btree_gin)
- [ ] Migration-Tool `golang-migrate/migrate` einrichten
- [ ] `migrations/` Verzeichnis pro Service anlegen
- [ ] Makefile-Target `make migrate-up` / `make migrate-down`

#### M0.4: Shared Go-Library `pkg/common` (Woche 2–3)

**Zu implementierende Pakete:**

```
pkg/common/
├── eventstore/
│   ├── client.go        # KurrentDB-Client-Wrapper
│   ├── aggregate.go     # Base-Aggregate mit AppendToStream/LoadFromStream
│   ├── envelope.go      # Event-Envelope-Struct
│   └── projection.go    # Subscription-Helper
├── database/
│   ├── postgres.go      # pgxpool.Pool-Factory mit Retry
│   └── redis.go         # go-redis/v9-Client-Factory
├── middleware/
│   ├── auth.go          # JWT-Validierung (chi-Middleware)
│   ├── cors.go          # CORS-Middleware
│   ├── logging.go       # Request-Logging mit zerolog
│   ├── ratelimit.go     # Redis-basiertes Rate-Limiting
│   └── recovery.go      # Panic-Recovery
├── logger/
│   └── logger.go        # Zerolog-Setup mit Correlation-ID
├── health/
│   └── handler.go       # /health/live + /health/ready Handler
├── errors/
│   └── errors.go        # Domänen-Fehlertypen
├── pagination/
│   └── pagination.go    # Cursor-basiertes Paginierungsmodell
└── grpc/
    ├── server.go         # gRPC-Server-Bootstrap
    └── client.go         # gRPC-Client mit mTLS
```

**Aufgaben:**
- [ ] KurrentDB Go-Client-Wrapper mit exponential backoff
- [ ] PostgreSQL Pool-Factory (pgxpool.Pool) mit Connection-Retry
- [ ] Redis-Client-Factory
- [ ] JWT-Middleware für chi-Router
- [ ] Zerolog-Setup mit Correlation-ID, Request-ID
- [ ] Health-Check-Handler (liveness + readiness)
- [ ] Domänen-Fehlertypen (NotFound, Conflict, Unauthorized, etc.)
- [ ] Cursor-Paginierung (keyset pagination)
- [ ] gRPC-Server-Bootstrap mit TLS
- [ ] Unit-Tests für alle gemeinsamen Pakete

#### M0.5: Service-Template (Woche 3)

Für jeden der 17 Microservices ein einheitliches Hexagonal-Architecture-Template:

```
services/[name]/
├── cmd/
│   └── server/
│       └── main.go           # Entrypoint: Config, DI, Start
├── internal/
│   ├── domain/
│   │   ├── aggregate.go      # Aggregate-Root
│   │   ├── events.go         # Domain-Events
│   │   └── commands.go       # Command-Structs
│   ├── application/
│   │   ├── command_handler.go
│   │   └── query_handler.go
│   ├── infrastructure/
│   │   ├── eventstore/
│   │   │   └── repository.go # KurrentDB Write-Side
│   │   ├── postgres/
│   │   │   └── repository.go # PostgreSQL Read-Side
│   │   └── grpc/
│   │       └── client.go     # Ausgeh. gRPC-Calls
│   └── ports/
│       ├── http/
│       │   ├── router.go     # chi-Router-Registrierung
│       │   └── handlers.go   # HTTP-Handler
│       └── grpc/
│           └── server.go     # Eingehende gRPC-Handler
├── migrations/               # SQL-Migrationsdateien
├── Dockerfile
├── Makefile
└── go.mod
```

**Aufgaben:**
- [ ] Service-Template-Generator (Bash/Go-Skript)
- [ ] Template für alle 17 Services anwenden
- [ ] Dockerfile (Multi-Stage Build, distroless/scratch)
- [ ] Makefile pro Service (build, test, run, lint)
- [ ] `go.work` alle Services einschließen

#### M0.6: Frontend-Scaffold (Woche 3–4)

```bash
npm create vite@latest frontend -- --template react-ts
cd frontend
npm install \
  @tanstack/react-router \
  @tanstack/react-query \
  @tanstack/react-query-devtools \
  zustand \
  @radix-ui/themes \
  @radix-ui/react-icons \
  react-i18next \
  i18next \
  axios \
  date-fns \
  vite-plugin-pwa \
  sass
```

**Aufgaben:**
- [ ] Vite-Projekt mit React + TypeScript
- [ ] TanStack Router mit Route-Tree
- [ ] TanStack Query mit Axios-Client
- [ ] Zustand-Stores (auth, ui, scanner, notification)
- [ ] Radix UI Themes + Farbpalette
- [ ] SCSS Design-Tokens (`src/styles/tokens.scss`)
- [ ] i18n-Setup (DE/EN, Namespace-Splitting)
- [ ] PWA-Config (`vite.config.ts` + `vite-plugin-pwa`)
- [ ] Service-Worker für Offline-Funktionalität
- [ ] ESLint + Prettier-Konfiguration
- [ ] Vitest + Testing-Library Setup

#### M0.7: CI/CD-Pipeline (Woche 4–5)

**`.github/workflows/ci.yml`:**

```yaml
name: CI Pipeline
on: [push, pull_request]

jobs:
  lint-and-test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        service: [auth, inventory, project, scanner, warehouse, invoice,
                  document, crew, federation, maintenance, transport, insurance,
                  workflow, ai, notification, reporting, audit]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: "1.22" }
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v4
        with:
          working-directory: services/${{ matrix.service }}
      - name: Tests
        run: |
          cd services/${{ matrix.service }}
          go test ./... -race -coverprofile=coverage.out
      - name: Coverage-Report
        uses: codecov/codecov-action@v4

  build-images:
    needs: lint-and-test
    runs-on: ubuntu-latest
    strategy:
      matrix:
        service: [auth, inventory, project, scanner, warehouse, invoice,
                  document, crew, federation, maintenance, transport, insurance,
                  workflow, ai, notification, reporting, audit]
    steps:
      - uses: actions/checkout@v4
      - uses: docker/setup-buildx-action@v3
      - uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - uses: docker/build-push-action@v5
        with:
          context: services/${{ matrix.service }}
          push: ${{ github.ref == 'refs/heads/main' }}
          tags: ghcr.io/${{ github.repository }}/${{ matrix.service }}:latest
          cache-from: type=gha
          cache-to: type=gha,mode=max

  frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with: { node-version: "20" }
      - run: cd frontend && npm ci
      - run: cd frontend && npm run lint
      - run: cd frontend && npm run test
      - run: cd frontend && npm run build
```

**Aufgaben:**
- [ ] GitHub Actions CI für alle Services
- [ ] Docker-Image-Build und Push nach GHCR
- [ ] Frontend-Build und -Tests in CI
- [ ] Deployment-Skript für Self-Hosted (`scripts/deploy.sh`)
- [ ] Watchtower für automatisches Image-Update konfigurieren
- [ ] Rollback-Skript (`scripts/rollback.sh`)
- [ ] Secrets-Management mit `.env`-Datei und Docker Secrets

#### M0.8: Monitoring & Logging (Woche 5–6)

**Aufgaben:**
- [ ] Prometheus Scraping-Config für alle Services
- [ ] Grafana Dashboard-Templates (Latenz, Error-Rate, Event-Throughput)
- [ ] Loki für zentrales Log-Aggregation
- [ ] Alertmanager-Regeln (Service Down, hohe Fehlerrate)
- [ ] `docker-compose.monitoring.yml` (Prometheus + Grafana + Loki)
- [ ] Health-Check-Endpoint für alle Services (`GET /health/live`, `GET /health/ready`)

### Phase-0-Abnahmekriterien

- [ ] `docker compose up -d` startet alle Infrastruktur-Services ohne Fehler
- [ ] KurrentDB Admin-UI erreichbar unter `http://localhost:2113`
- [ ] PostgreSQL-Verbindung mit allen 17 Schemas erfolgreich
- [ ] Redis `PING` liefert `PONG`
- [ ] Traefik-Dashboard erreichbar
- [ ] CI-Pipeline läuft durch (grün)
- [ ] Frontend lädt unter `http://localhost:5173`
- [ ] Alle Services kompilieren fehlerfrei

---

## Phase 1: Core MVP

**Dauer:** 10–14 Wochen
**Voraussetzung:** Phase 0 abgeschlossen
**Lieferbar:** Vollständig nutzbare Grundfunktionen: Auth, Inventar, Projekte, Rechnungen

**Zielgruppe Phase 1:**
- Marco Berger (GF, 3 Mitarbeiter) kann Tagesgeschäft komplett in MyRMS abwickeln
- Lisa Huber kann Equipment per Barcode scannen und verwalten
- Thomas Berger kann Rechnungen erstellen und DATEV exportieren

### Meilensteine

#### M1.1: Auth-Service (Woche 1–2)

**Port:** 8001 | **Schema:** `auth_schema`

**Domain-Events (KurrentDB Stream: `user-{uuid}`):**
- `UserRegistered` — Erstregistrierung
- `UserLoggedIn` — Erfolgreiche Anmeldung
- `UserRoleChanged` — Rollenänderung durch Admin
- `PasswordChanged` — Passwortänderung
- `UserDeactivated` — Deaktivierung

**REST-Endpoints:**
```
POST   /api/v1/auth/login          # JWT-Token ausstellen
POST   /api/v1/auth/refresh        # Token erneuern
POST   /api/v1/auth/logout         # Token invalidieren (Redis)
GET    /api/v1/users               # Alle Nutzer (Admin)
POST   /api/v1/users               # Nutzer anlegen
PUT    /api/v1/users/{id}/role     # Rolle ändern
DELETE /api/v1/users/{id}          # Nutzer deaktivieren
GET    /api/v1/users/me            # Eigenes Profil
PUT    /api/v1/users/me/password   # Passwort ändern
```

**JWT-Payload:**
```json
{
  "sub": "user-uuid",
  "email": "max@firma.de",
  "role": "admin",
  "tenant": "firma-uuid",
  "exp": 1700000000,
  "iat": 1699996400,
  "jti": "token-uuid"
}
```

**Aufgaben:**
- [ ] PostgreSQL-Schema: `users`, `sessions`, `refresh_tokens`
- [ ] Argon2id-Passwort-Hashing
- [ ] JWT-Ausstellung (RS256, 1h Lebensdauer)
- [ ] Refresh-Token in Redis (7 Tage)
- [ ] RBAC-Rollen: `superadmin`, `admin`, `manager`, `warehouse`, `driver`, `crew`, `freelancer`, `readonly`
- [ ] Permissions-Matrix implementieren
- [ ] Brute-Force-Schutz (5 Fehlversuche → 15 Minuten Sperre, Redis)
- [ ] Rate-Limiting: 10 Anfragen/Minute pro IP
- [ ] Initialer Superadmin-Nutzer per Umgebungsvariable

#### M1.2: Inventory-Service (Woche 2–5)

**Port:** 8002 | **Schema:** `inventory_schema`

**Domain-Events (Stream: `equipment-{uuid}`):**
- `EquipmentCreated`
- `EquipmentUpdated`
- `EquipmentStatusChanged` (available/rented/maintenance/retired)
- `EquipmentAssignedToProject`
- `EquipmentReturnedFromProject`
- `EquipmentSetQuantity`
- `EquipmentImageUploaded`
- `EquipmentQRCodeGenerated`

**REST-Endpoints:**
```
GET    /api/v1/equipment                    # Liste mit Filtern + Paginierung
POST   /api/v1/equipment                    # Neue Position anlegen
GET    /api/v1/equipment/{id}               # Detail-View
PUT    /api/v1/equipment/{id}              # Aktualisieren
DELETE /api/v1/equipment/{id}              # Archivieren (soft delete)
POST   /api/v1/equipment/{id}/images        # Bild hochladen (S3/MinIO)
GET    /api/v1/equipment/{id}/history       # Event-Historie
GET    /api/v1/equipment/{id}/availability  # Verfügbarkeit für Zeitraum
POST   /api/v1/equipment/{id}/qr-code       # QR-Code / Label generieren (ZPL)
GET    /api/v1/categories                   # Kategoriebaum
POST   /api/v1/categories                   # Kategorie anlegen
PUT    /api/v1/categories/{id}             # Kategorie umbenennen
GET    /api/v1/equipment/search             # Volltext-Suche (pg_trgm)
GET    /api/v1/equipment/availability-check # Batch-Verfügbarkeitscheck
```

**Kern-Algorithmus: Verfügbarkeitscheck**

```go
type AvailabilityChecker struct {
    db *pgxpool.Pool
}

func (a *AvailabilityChecker) CheckAvailability(
    ctx context.Context,
    equipmentID uuid.UUID,
    startDate, endDate time.Time,
    requestedQty int,
) (AvailabilityResult, error) {
    // 1. Gesamtbestand aus inventory_schema.equipment_stock
    totalQty, err := a.getTotalQuantity(ctx, equipmentID)
    if err != nil { return AvailabilityResult{}, err }

    // 2. Reservierungen im Zeitraum aus project_schema.project_equipment
    reservedQty, err := a.getReservedQuantity(ctx, equipmentID, startDate, endDate)
    if err != nil { return AvailabilityResult{}, err }

    // 3. Geräte in Wartung aus maintenance_schema.maintenance_plans
    inMaintenanceQty, err := a.getMaintenanceQuantity(ctx, equipmentID, startDate, endDate)
    if err != nil { return AvailabilityResult{}, err }

    available := totalQty - reservedQty - inMaintenanceQty
    return AvailabilityResult{
        EquipmentID: equipmentID,
        TotalQty:    totalQty,
        Available:   available,
        Reserved:    reservedQty,
        Requested:   requestedQty,
        Feasible:    available >= requestedQty,
    }, nil
}
```

**Preiskalkulation:**

```go
type PriceEngine struct{}

func (p *PriceEngine) Calculate(e Equipment, days int, discount float64) PriceResult {
    basePrice := e.DailyRate * float64(days)

    // Mengenstaffel
    var volumeDiscount float64
    switch {
    case days >= 30: volumeDiscount = 0.20
    case days >= 14: volumeDiscount = 0.15
    case days >= 7:  volumeDiscount = 0.10
    case days >= 3:  volumeDiscount = 0.05
    }

    subtotal := basePrice * (1 - volumeDiscount) * (1 - discount)
    vat := subtotal * 0.19  // Deutsche MwSt

    return PriceResult{
        BasePrice:      basePrice,
        VolumeDiscount: volumeDiscount,
        CustomDiscount: discount,
        Subtotal:       subtotal,
        VAT:            vat,
        Total:          subtotal + vat,
    }
}
```

**Aufgaben:**
- [ ] PostgreSQL-Migration: `equipment`, `equipment_stock`, `categories`, `equipment_images`
- [ ] KurrentDB-Aggregate `EquipmentAggregate`
- [ ] Volltext-Suche mit `pg_trgm`
- [ ] Batch-Verfügbarkeitscheck (bis 500 Items gleichzeitig)
- [ ] Preiskalkulations-Engine mit Staffelrabatten
- [ ] ZPL-Label-Generierung für Zebra-Drucker
- [ ] MinIO/S3-Integration für Equipment-Bilder
- [ ] QR-Code-Generierung (`skip2/go-qrcode`)
- [ ] Flightcase-Logik: Child-Equipment-Zuordnung
- [ ] CSV-Import aus Bestandsdaten alter Systeme

#### M1.3: Project-Service (Woche 4–7)

**Port:** 8003 | **Schema:** `project_schema`

**Domain-Events (Stream: `project-{uuid}`):**
- `ProjectCreated`
- `ProjectUpdated`
- `ProjectStatusChanged` (anfrage/angebot/bestätigt/vorbereitung/aufbau/aktiv/abbau/abgeschlossen/archiviert)
- `EquipmentReservedForProject`
- `EquipmentRemovedFromProject`
- `PackingListGenerated`
- `CustomerApproved`

**REST-Endpoints:**
```
GET    /api/v1/projects                        # Liste mit Filtern
POST   /api/v1/projects                        # Neues Projekt
GET    /api/v1/projects/{id}                   # Detail
PUT    /api/v1/projects/{id}                   # Aktualisieren
GET    /api/v1/projects/{id}/packing-list       # Packliste generieren
GET    /api/v1/projects/{id}/packing-list/pdf   # Packliste als PDF
POST   /api/v1/projects/{id}/equipment          # Equipment hinzufügen
DELETE /api/v1/projects/{id}/equipment/{eqId}  # Equipment entfernen
GET    /api/v1/projects/{id}/availability       # Alle Items prüfen
GET    /api/v1/projects/calendar                # Kalenderansicht (JSON)
GET    /api/v1/projects/conflicts               # Doppelbuchungen
```

**Aufgaben:**
- [ ] PostgreSQL-Migration: `projects`, `project_equipment`, `customers`
- [ ] Status-Zustandsmaschine (anfrage → archiviert)
- [ ] Doppelbuchungs-Erkennung über alle Projekte
- [ ] Packlisten-Generierung sortiert nach Lagerort
- [ ] Kalenderansicht für Projektzeitraum-Übersicht
- [ ] E-Mail-Benachrichtigung bei Status-Änderung
- [ ] Kunden-Verwaltung (DSGVO-konforme Pflichtfelder)
- [ ] Projekt-Kopieren-Funktion

#### M1.4: Invoice-Service (Woche 6–9)

**Port:** 8006 | **Schema:** `invoice_schema`

**Domain-Events (Stream: `invoice-{uuid}`):**
- `InvoiceCreated`
- `InvoiceSent`
- `PaymentRecorded`
- `InvoiceCancelled`
- `CreditNoteCreated`
- `ReminderSent`
- `InvoiceExportedToDatev`

**REST-Endpoints:**
```
GET    /api/v1/invoices                    # Liste
POST   /api/v1/invoices                    # Neue Rechnung / Angebot
GET    /api/v1/invoices/{id}               # Detail
POST   /api/v1/invoices/{id}/send          # Per E-Mail senden
POST   /api/v1/invoices/{id}/payment       # Zahlung erfassen
POST   /api/v1/invoices/{id}/cancel        # GoBD-konform stornieren
POST   /api/v1/invoices/{id}/credit-note   # Gutschrift erstellen
GET    /api/v1/invoices/{id}/pdf           # PDF herunterladen
POST   /api/v1/invoices/from-project/{id}  # Aus Projekt generieren
GET    /api/v1/invoices/datev-export        # DATEV CSV-Export
GET    /api/v1/invoices/open               # Offene Forderungen
```

**Aufgaben:**
- [ ] Fortlaufende GoBD-konforme Rechnungsnummern (pro Mandant)
- [ ] PDF-Generierung mit `chromedp` (HTML-Template → PDF)
- [ ] Angebot → Auftragsbestätigung → Rechnung Konvertierung
- [ ] Mahnwesen (3-stufig: Zahlungserinnerung / Mahnung 1 / Mahnung 2)
- [ ] DATEV-Export SKR03 (CSV-Format)
- [ ] Teilrechnung mit Restbetrag-Tracking
- [ ] Stornierung nach GoBD (keine Löschung, nur Gegenbuchung)
- [ ] Gutschriften mit Pflicht-Verweis auf Originalrechnung
- [ ] E-Mail-Versand der Rechnung als PDF-Anhang

#### M1.5: Frontend Core-Pages (Woche 5–10)

**Zu implementierende Pages:**

1. **Login-Page** (`/login`)
   - Formular (E-Mail + Passwort)
   - JWT speichern in Zustand-Store + Cookie
   - Redirect nach Login

2. **Dashboard** (`/dashboard`)
   - KPIs: Aktive Projekte, Equipment-Verfügbarkeit, Offene Rechnungen
   - Heutige Ausgaben/Rücknahmen
   - Alarme: Wartung fällig, Überfällige Rechnungen
   - Schnell-Navigation

3. **Equipment-Liste** (`/equipment`)
   - DataGrid mit Filtern (Kategorie, Status, Suche)
   - Inline-Verfügbarkeitsanzeige
   - Bulk-Actions (Etiketten drucken)
   - Import-Button

4. **Equipment-Detail** (`/equipment/:id`)
   - Alle Felder editierbar
   - Bild-Upload
   - Verfügbarkeits-Kalender
   - Event-Historie (Audit-Trail)
   - QR-Code-Druck

5. **Projekte-Liste** (`/projects`)
   - Kalender-Ansicht + Listen-Ansicht
   - Status-Filter
   - Schnell-Anlegen

6. **Projekt-Detail** (`/projects/:id`)
   - Status-Änderung mit Tabs
   - Equipment-Auswahl mit Verfügbarkeitscheck
   - Packliste (sortiert nach Lager)
   - Rechnungsvorschau

7. **Rechnungen-Liste** (`/invoices`)
   - Tabelle: Nummer, Kunde, Betrag, Status, Fälligkeit
   - Farbliche Ampel für Fälligkeitsstatus
   - DATEV-Export-Button

8. **Rechnung-Detail** (`/invoices/:id`)
   - WYSIWYG-ähnliche Bearbeitungsansicht
   - Live-PDF-Vorschau
   - Zahlungserfassung
   - Mahnung senden

**Aufgaben:**
- [ ] TanStack Router Route-Tree aufbauen
- [ ] Alle 8 Core-Pages implementieren
- [ ] TanStack Query für alle API-Calls
- [ ] Optimistic Updates für häufige Operationen
- [ ] Fehlerbehandlung + Toast-Benachrichtigungen
- [ ] Responsive Design (Desktop + Tablet Prio)
- [ ] Dark-Mode-Toggle
- [ ] Command Palette (⌘K) für Navigation

#### M1.6: Scanner-Integration (Woche 9–11)

Einfache Scan-Funktion für Phase 1 (QR-Code-Kamera-Scan):

**Aufgaben:**
- [ ] `ScannerPage` (`/scan`) mit Kamera-QR-Scan
- [ ] Equipment-Lookup nach QR-Code-UUID
- [ ] Check-Out: Equipment einem Projekt zuweisen
- [ ] Check-In: Equipment vom Projekt zurückgeben
- [ ] Scan-Feedback (visuell + vibration)
- [ ] Offline-Unterstützung für Service-Worker

### Phase-1-Abnahmekriterien

- [ ] Marco kann ein neues Projekt anlegen und Equipment zuweisen
- [ ] Lisa kann Equipment per Smartphone-Kamera scannen und check-out durchführen
- [ ] Thomas kann eine Rechnung erstellen, als PDF herunterladen und per E-Mail senden
- [ ] Thomas kann DATEV-Export für einen Monat herunterladen
- [ ] Alle CRUD-Operationen für Equipment, Projekte, Rechnungen funktionieren
- [ ] JWT-Auth schützt alle Endpunkte
- [ ] Keine SQL-Injections, XSS, CSRF

---

## Phase 2: Operations

**Dauer:** 8–12 Wochen
**Voraussetzung:** Phase 1 abgeschlossen
**Lieferbar:** Lager-Management, Zebra-Scanner, Transport-Logistik, Wartungs-Tracking

### Meilensteine

#### M2.1: Warehouse-Service (Woche 1–3)

**Port:** 8005 | **Schema:** `warehouse_schema`

**Aufgaben:**
- [ ] PostgreSQL-Migration: `warehouses`, `zones`, `racks`, `bays`, `stock_locations`, `movements`
- [ ] Hierarchische Lagerstruktur (Lager → Zone → Regal → Fach)
- [ ] Standort-Zuweisung für Equipment-Bestände
- [ ] QR-Code-Label für jeden Lagerplatz (ZPL-Format für Zebra)
- [ ] Bewegungs-Tracking: jede physische Bewegung als Event
- [ ] Inventur-Workflow: Voll-Inventur, Zyklus-Inventur, Spot-Kontrolle
- [ ] Lagerplatz-Kapazitäts-Tracking
- [ ] KI-Lageroptimierung (Vorbereitung für Phase 4)

#### M2.2: Scanner-Service mit Zebra TC21 (Woche 2–5)

**Port:** 8004 | **Schema:** `scanner_schema`

**DataWedge-Integration:**

```javascript
// In React Native oder PWA (Android)
document.addEventListener('message', (event) => {
    const data = JSON.parse(event.data);
    if (data.source === 'datawedge') {
        const barcode = data.data;
        const symbology = data.labelType; // QR_CODE, CODE_128, etc.
        handleScan({ barcode, symbology, timestamp: new Date() });
    }
});
```

**DataWedge-Profil für MyRMS:**
```json
{
  "PROFILE_NAME": "MyRMS",
  "PROFILE_ENABLED": "true",
  "CONFIG_MODE": "CREATE_IF_NOT_EXIST",
  "PLUGIN_CONFIG": {
    "PLUGIN_NAME": "INTENT",
    "PARAM_LIST": {
      "intent_output_enabled": "true",
      "intent_action": "de.myrms.SCAN",
      "intent_delivery": "2"
    }
  },
  "ASSOCIATED_APP_LIST": [
    { "PACKAGE_NAME": "de.myrms.app", "ACTIVITY_LIST": ["*"] }
  ]
}
```

**Aufgaben:**
- [ ] Zebra DataWedge-Intent-Empfang in PWA
- [ ] Scan-Kontext-System (check-out, check-in, lager-einräumen, inventur)
- [ ] Multi-Scan für Packlisten (alle Items eines Projekts scannen)
- [ ] Audio + visuelles Feedback pro Scan-Ergebnis
- [ ] Offline-Scan-Queue (IndexedDB, bis 500 Scans)
- [ ] Background-Sync wenn Verbindung wiederhergestellt
- [ ] Scan-Session-Protokoll in `scanner_schema`
- [ ] RFID-Vorbereitung (Zebra FX9600 Interface)
- [ ] USB-Scanner via WebHID-API

#### M2.3: Transport-Service (Woche 3–6)

**Port:** 8011 | **Schema:** `transport_schema`

**Aufgaben:**
- [ ] PostgreSQL-Migration: `vehicles`, `tours`, `tour_equipment`, `driver_logs`
- [ ] Fahrzeug-Verwaltung (Kennzeichen, Kapazität kg + m³, DGUV-Prüfung)
- [ ] Tour-Planung: welches Equipment in welches Fahrzeug
- [ ] Kapazitäts-Check: Gewicht + Volumen nicht überschreiten
- [ ] Fahrer-Zuweisung aus Crew-Service (gRPC)
- [ ] Lieferschein-Generierung (Anknüpfung an Document-Service Phase 3)
- [ ] Kilometer + Kosten-Tracking pro Tour
- [ ] Tachograph-kompatible Fahrerlisten

#### M2.4: Maintenance-Service (Woche 5–8)

**Port:** 8010 | **Schema:** `maintenance_schema`

**Aufgaben:**
- [ ] PostgreSQL-Migration: `maintenance_plans`, `maintenance_tasks`, `checklists`, `electrical_tests`
- [ ] Wartungsplan-Typen: Intervall (täglich/wöchentlich/monatlich/jährlich), Nach-Einsatz, Stunden-basiert
- [ ] Aufgaben-Zustandsmaschine: geplant → in_bearbeitung → abgeschlossen
- [ ] Checklisten mit JSON-Format
- [ ] E-Check / DGUV Vorschrift 3 Workflow:
  - Erstellt E-Check-Auftrag pro Gerät
  - Fälligkeits-Erinnerung (E-Mail + In-App)
  - VDE-0701-0702-Messwerte speichern
  - Prüfplaketten-Druck (ZPL-Label)
- [ ] IZYTRON.IQ XML-Import (Prüfergebnis-Import aus externer Prüf-Software)
- [ ] Equipment automatisch in Wartungs-Status setzen (sperrt für Buchung)
- [ ] Wartungs-Dashboard: fällige und überfällige Wartungen

#### M2.5: Frontend Operations-Pages (Woche 4–10)

**Neue Pages:**

1. **Lager-Übersicht** (`/warehouse`)
   - Lager-Karte mit Zonen + Regalanzeige
   - Equipment-Verteilung pro Lagerplatz

2. **Mobiler Scanner** (`/scan`) — erweitert
   - Zebra DataWedge-Intent-Empfang
   - Scan-Kontext-Auswahl (Check-Out / Check-In / Inventur)
   - Offline-Indikator + Sync-Status
   - Große Buttons (Lisa Persona: max. 2 Taps)

3. **Transport** (`/transport`)
   - Touren-Liste und -Planung
   - Fahrzeug-Verfügbarkeitsanzeige
   - Packlist-zu-Tour-Zuordnung

4. **Wartung** (`/maintenance`)
   - Fälligkeits-Kalender
   - E-Check-Dashboard
   - Aufgaben-Zuweisung

### Phase-2-Abnahmekriterien

- [ ] Lisa kann Equipment mit Zebra TC21 scannen und in ein Projekt auschecken
- [ ] Offline-Scans werden bei Reconnect automatisch synchronisiert
- [ ] Wartungsfällige Geräte erscheinen in Dashboard-Warnung
- [ ] DGUV/E-Check-Status pro Gerät ist sichtbar
- [ ] Touren können Fahrzeugen zugewiesen werden mit Kapazitätsprüfung

---

## Phase 3: Business

**Dauer:** 8–10 Wochen
**Voraussetzung:** Phase 2 abgeschlossen
**Lieferbar:** Crew-Management, Dokument-Management, Versicherung, Reporting

### Meilensteine

#### M3.1: Crew-Service (Woche 1–3)

**Port:** 8008 | **Schema:** `crew_schema`

**Aufgaben:**
- [ ] PostgreSQL-Migration: `crew_members`, `qualifications`, `crew_assignments`, `time_records`
- [ ] Qualifikations-System (IPAF, Rigger, Tonmeister, etc.)
- [ ] Verfügbarkeits-Matrix für Crew-Mitglieder
- [ ] Konflikt-Erkennung: Doppel-Buchung eines Crew-Mitglieds
- [ ] CalDAV-Feed für Crew-Kalender (Thunderbird/iOS)
- [ ] Freelancer-Portal: Kevin kann eigene Einsätze und Aufgaben sehen
- [ ] Zeiterfassung pro Einsatz (Start/Stop-Timer)
- [ ] Fahrerliste: Wer fährt wann mit welchem Fahrzeug

#### M3.2: Document-Service (Woche 2–5)

**Port:** 8007 | **Schema:** `document_schema`

**Aufgaben:**
- [ ] PostgreSQL-Migration: `documents`, `document_versions`, `signatures`
- [ ] Dokument-Typen: Angebot, Auftragsbestätigung, Lieferschein, Abholschein, Rechnung, Mietvertrag, Schadensbericht
- [ ] Go-Template-basierte PDF-Generierung (chromedp)
- [ ] Lieferschein-Generierung aus Packliste
- [ ] Digitale Unterschriften (Canvas-Signature-Pad im Browser)
- [ ] GoBD-konformes Archiv: SHA-256-Checksummen-Kette
- [ ] Versionierung: `v1`, `v2` ... bei Änderungen
- [ ] Scan-to-Document: Scan-Upload + OCR (tesseract)
- [ ] Kunden-Portal: Dokument zum Unterschreiben per Link schicken

#### M3.3: Insurance-Service (Woche 3–5)

**Port:** 8012 | **Schema:** `insurance_schema`

**Aufgaben:**
- [ ] PostgreSQL-Migration: `policies`, `claims`, `claim_items`
- [ ] Versicherungspolice-Verwaltung (Betriebshaftpflicht, Kaskoversicherung, Transportversicherung, Mieterversicherung)
- [ ] Equipment automatisch versicherungsrelevant markieren (ab Wiederbeschaffungswert)
- [ ] Schadensfall-Workflow: Erfassung → Dokumentation → Einreichung → Abschluss
- [ ] Schadensformular mit Foto-Upload
- [ ] Forderungs-Tracking: Wieviel erstattet, wieviel offen
- [ ] Jahres-Auswertung: Schäden pro Jahr, Prämienentwicklung

#### M3.4: Reporting-Service (Woche 4–7)

**Port:** 8016 | **Schema:** `reporting_schema`

**Aufgaben:**
- [ ] PostgreSQL-Migration: `report_definitions`, `report_runs`, `kpi_snapshots`
- [ ] Rollen-spezifische KPI-Dashboards:
  - GF (Marco): Umsatz, Marge, Auslastung, Top-Equipment
  - Lager (Lisa): Equipment-Status, Standorte, Bewegungen
  - Buchaltung (Thomas): Offene Forderungen, DATEV-Export, Mahnstatistiken
- [ ] Report-Typen: Umsatzbericht, Auslastung, Mahnstatistik, Equipment-Inventur
- [ ] Zeitraumauswahl: Woche, Monat, Quartal, Jahr, Custom
- [ ] CSV + PDF-Export für jeden Report
- [ ] Geplante E-Mail-Reports (wöchentlicher Umsatz-Report an Marco)
- [ ] Recharts/Victory-basierte Diagramme im Frontend

#### M3.5: Frontend Business-Pages (Woche 3–9)

**Neue Pages:**

1. **Crew-Management** (`/crew`)
   - Crew-Mitglieder-Liste
   - Qualifikations-Badges
   - Verfügbarkeitskalender
   - Einsatz-Zuordnung

2. **Freelancer-Ansicht** (`/my-assignments`) — Kevin Persona
   - Eigene Einsätze
   - Packliste für Einsatz
   - Zeiterfassung (Start/Stop)
   - Auf Mobilgerät optimiert

3. **Dokumente** (`/documents`)
   - Dokument-Archiv mit Vorschau
   - Lieferschein aus Projekt generieren
   - Unterschriften-Status

4. **Versicherung** (`/insurance`)
   - Policen-Übersicht
   - Schadensmeldung (Formular)
   - Aktive Schadensfälle

5. **Reports** (`/reports`)
   - KPI-Dashboard rollenspezifisch
   - Diagramme (Umsatz, Auslastung)
   - Export-Buttons

### Phase-3-Abnahmekriterien

- [ ] Kevin kann seine Einsätze auf dem iPhone sehen und Zeiten erfassen
- [ ] Lieferschein wird aus Packliste automatisch generiert und kann digital unterschrieben werden
- [ ] Thomas kann monatlichen Umsatzbericht als PDF/CSV exportieren
- [ ] Schadensfall kann dokumentiert und mit Fotos eingereicht werden
- [ ] GoBD-Archiv speichert Dokumente mit Checksummen-Kette

---

## Phase 4: Advanced Features

**Dauer:** 10–14 Wochen
**Voraussetzung:** Phase 3 abgeschlossen
**Lieferbar:** KI-Funktionen, Workflow-Engine, Federation (Multi-Company), Audit/GoBD

### Meilensteine

#### M4.1: AI-Service (Woche 1–4)

**Port:** 8014 | **Schema:** `ai_schema`

**Aufgaben:**
- [ ] PostgreSQL-Migration: `ai_requests`, `ai_feedback`, `few_shot_examples`, `ai_providers`
- [ ] Multi-Provider-Architektur: Anthropic Claude / OpenAI GPT-4o / Google Gemini / Mistral / Ollama
- [ ] Provider-Config per YAML (Priorität, Fallback, Kostenbudget)
- [ ] Anonymisierungs-Service (7 Muster: Name, E-Mail, IBAN, Tel., Adresse, Steuernummer, Kundennummer)
- [ ] Deanonymisierungs-Map in Redis (15 Minuten TTL)
- [ ] KI-Funktionen:
  - Preisoptimierung: Saisonale + nachfragebasierte Empfehlungen
  - Nachfrageprognose: Prophet-ähnliche Trendanalyse auf historischen Daten
  - Smart Asset Creator: GPT-4o Vision für automatisches Befüllen von Equipment-Daten aus Foto
  - Prädiktive Wartung: Anomalieerkennung in Nutzungsmustern
  - Dunning-Empfehlung: KI entscheidet über Mahnstufe
  - Equipment-Texte: Automatisch Beschreibungen und Wartungshinweise generieren
- [ ] Feedback-Mechanismus: Nutzer bewertet KI-Vorschläge
- [ ] Few-Shot-Learning: Bewertungen verbessern zukünftige Prompts

#### M4.2: Workflow-Service (Woche 2–5)

**Port:** 8013 | **Schema:** `workflow_schema`

**Aufgaben:**
- [ ] PostgreSQL-Migration: `workflow_definitions`, `workflow_instances`, `workflow_steps`
- [ ] Trigger-Typen: Event-basiert, Zeitbasiert (Cron), Manuell
- [ ] Aktions-Typen: E-Mail senden, Webhook, Service-Call (gRPC), Status-Änderung
- [ ] Bedingungen: IF/THEN/ELSE Logik
- [ ] Vordefinierte Templates:
  - Onboarding neuer Kunde
  - Projekt-Erinnerung 7 Tage vorher
  - Equipment-Wartung fällig
  - Überfällige Rechnung
  - Rücknahme-Erinnerung
  - Willkommens-E-Mail neuer Nutzer
  - Lager-Bestandswarnung
- [ ] Visueller Workflow-Editor (React Flow)
- [ ] Workflow-Log mit Ausführungshistorie

#### M4.3: Federation-Service (Woche 4–7)

**Port:** 8009 | **Schema:** `federation_schema`

**Aufgaben:**
- [ ] mTLS-Zertifikat-Management (Cert-auf-Cert-Austausch)
- [ ] Partner-Einrichtung: Stefan (SoundPro GmbH) kann sich verbinden
- [ ] PostgreSQL-Migration: `federation_partners`, `federation_requests`, `shared_equipment`
- [ ] Equipment-Verfügbarkeit im Partner-Netzwerk abfragen
- [ ] Sub-Vermietungs-Anfrage senden + empfangen
- [ ] Handover-Dokumentation bei Übergabe
- [ ] Automatische Eingangsrechnung bei Rückgabe
- [ ] Data-Sovereignty: Nur explizit freigegebene Equipment-Daten werden geteilt
- [ ] Verschlüsselte Kommunikation ausschließlich P2P (kein zentraler Server)

#### M4.4: Notification-Service (Woche 3–6)

**Port:** 8015 | **Schema:** `notification_schema`

**Aufgaben:**
- [ ] PostgreSQL-Migration: `notifications`, `notification_channels`, `user_preferences`
- [ ] E-Mail via SMTP (Go `gomail/gomail`)
- [ ] Web Push (VAPID, Service Worker)
- [ ] In-App-Benachrichtigungen (WebSocket + Redis Pub/Sub)
- [ ] Nutzer-Präferenzen: Welcher Kanal für welchen Ereignistyp
- [ ] Notification-Center (Glocken-Icon im Header)
- [ ] Tages-Digest-E-Mail (jeden Morgen 7 Uhr)
- [ ] Stille Stunden (keine Push-Benachrichtigungen 22–7 Uhr)
- [ ] Batch-Benachrichtigungen (sammeln statt sofort senden)

#### M4.5: Audit-Service (Woche 5–8)

**Port:** 8017 | **Schema:** `audit_schema`

**Aufgaben:**
- [ ] PostgreSQL-Migration: `audit_log`, `audit_exports`
- [ ] Immutables Audit-Log: SHA-256-Checksummen-Kette für GoBD-Konformität
- [ ] Alle schreibenden Operationen in Audit-Log schreiben
- [ ] Audit-Log-Abfrage-UI: Filter nach Zeitraum, Nutzer, Ressource, Aktion
- [ ] GoBD-Prüfung: Prüf-Funktion für Checksummen-Integrität
- [ ] DSGVO-Pseudonymisierung: Nach Nutzerlöschung Daten pseudonymisieren
- [ ] Betriebsprüfungs-Export: ZIP mit allen Audit-Logs + KurrentDB-Events für ein Geschäftsjahr
- [ ] Aufbewahrungsfristen: 10 Jahre für steuerrelevante Daten (GoBD), 3 Jahre für sonstige

#### M4.6: Frontend Advanced-Pages (Woche 4–11)

**Neue Pages:**

1. **KI-Assistent** (`/ai`)
   - Chat-Interface für freie KI-Anfragen
   - Preisoptimierungs-Widget
   - Smart Asset Creator (Foto-Upload → Equipment-Daten)
   - Provider-Auswahl

2. **Workflow-Editor** (`/workflows`)
   - React Flow Canvas
   - Drag & Drop Trigger/Aktionen
   - Template-Bibliothek
   - Ausführungs-Log

3. **Federation** (`/federation`)
   - Partner-Netzwerk-Übersicht
   - Verfügbarkeitsanfrage an Partner
   - Handover-Protokoll

4. **Audit-Log** (`/audit`)
   - Volltext-Suchbare Log-Tabelle
   - GoBD-Prüfbericht
   - Betriebsprüfungs-Export

5. **Admin** (`/admin`)
   - System-Einstellungen
   - Nutzer-Verwaltung
   - Service-Health-Status
   - KurrentDB-Stream-Browser

### Phase-4-Abnahmekriterien

- [ ] KI-Preisoptimierung liefert nachvollziehbare Empfehlungen für 5 Test-Equipment-Positionen
- [ ] Stefan (SoundPro) kann sich als Federation-Partner verbinden und Equipment-Verfügbarkeit abfragen
- [ ] Audit-Log zeigt lückenlose Checksummen-Kette (GoBD-Prüfung bestanden)
- [ ] Workflow-Engine sendet automatisch E-Mail 7 Tage vor Projektstart
- [ ] Web Push funktioniert auf Chrome Desktop + Android + iOS Safari

---

## Querschnittsthemen (alle Phasen)

### Sicherheit

**Zu implementieren (kontinuierlich):**
- [ ] OWASP Top 10 Maßnahmen (XSS, CSRF, SQL-Injection, etc.)
- [ ] Dependency-Scanning (Dependabot, `govulncheck`)
- [ ] Container-Scanning (Trivy in CI)
- [ ] Secrets niemals im Code (`.env` + Docker Secrets)
- [ ] Rate-Limiting auf allen Public-Endpunkten
- [ ] Input-Validierung in allen Services (Go: `go-playground/validator`)
- [ ] HTTPS-only (Traefik HTTP→HTTPS Redirect)
- [ ] Security-Header (CSP, HSTS, X-Frame-Options via Traefik Middleware)
- [ ] Audit-Log für alle sensitiven Operationen
- [ ] Penetrations-Test nach Phase 2 (eigenes Testing)

### Performance

**Ziele:**
- API-Antwortzeit: < 200ms für 95% der Requests (p95)
- Frontend First Contentful Paint: < 1,5s
- Offline-Verfügbarkeit: Scanner-Grundfunktionen auch ohne Internet

**Maßnahmen:**
- [ ] PostgreSQL-Indizes auf alle Foreign Keys und häufige Filter-Spalten
- [ ] Redis-Cache für Equipment-Verfügbarkeits-Ergebnisse (60s TTL)
- [ ] React-Abfragen mit TanStack Query (staleTime 30s für statische Daten)
- [ ] Frontend Code-Splitting (lazy loading pro Route)
- [ ] Virtualisierung für lange Listen (TanStack Virtual)
- [ ] Database Connection Pooling (pgxpool, max 20 connections pro Service)
- [ ] gRPC-Kommunikation zwischen Services (statt HTTP)
- [ ] PostgreSQL EXPLAIN ANALYZE für langsame Queries

### Datenschutz (DSGVO)

**Maßnahmen:**
- [ ] Datenschutzerklärung im Frontend
- [ ] Cookie-Banner für optionale Tracking-Cookies
- [ ] Recht auf Auskunft: `GET /api/v1/users/{id}/data-export`
- [ ] Recht auf Löschung: `DELETE /api/v1/users/{id}` (Pseudonymisierung statt Löschung)
- [ ] Daten-Minimierung: Nur notwendige Felder speichern
- [ ] Auftragsverarbeitungsvertrag (AVV) für selbst-gehostete Systeme nicht notwendig
- [ ] Verschlüsselung at-rest: PostgreSQL PGDATA auf verschlüsseltem Dateisystem
- [ ] Verschlüsselung in-transit: TLS für alle Verbindungen
- [ ] Löschfristen: Technische Implementierung der GoBD-Aufbewahrungsfristen

---

## Technische Schulden & Refactoring-Planung

### Nach Phase 1 (vor Phase 2)

- [ ] Code-Review aller Phase-1-Services durch zweite Person (Freelancer)
- [ ] Performance-Profiling der häufigsten API-Pfade
- [ ] Fehler-Handling vereinheitlichen (domänenspezifische Fehler-Typen)
- [ ] Integrations-Tests für alle kritischen Pfade (Buchung, Rechnung)
- [ ] API-Dokumentation mit Swagger/OpenAPI

### Nach Phase 2 (vor Phase 3)

- [ ] Scanner-Offline-Queue Stress-Test (500 Scans gleichzeitig)
- [ ] KurrentDB-Subscription-Recovery testen (Verbindungsabbruch)
- [ ] Load-Test mit k6 (100 gleichzeitige Nutzer)
- [ ] Datenbank-Migrations-Strategie für Produktionsupdates
- [ ] Backup-Strategie verifizieren (WAL-G für PostgreSQL, KurrentDB-Backup)

### Nach Phase 3 (vor Phase 4)

- [ ] Microservice-zu-Microservice-Kommunikation auditieren
- [ ] Shared `pkg/common` Refactoring nach Erfahrungen aus 3 Phasen
- [ ] Frontend Performance-Budget einhalten (< 300kB JS initial)
- [ ] Accessibility-Audit (WCAG 2.1 AA)
- [ ] Mobile-UX-Test mit echten Zebra TC21-Geräten

---

## Deployment-Strategie

### Erstinstallation (Zero-to-Running)

```bash
# 1. Repo klonen
git clone https://github.com/myrms/myrms.git /opt/myrms
cd /opt/myrms

# 2. Konfiguration
cp .env.example .env
# .env editieren: Passwörter, Domain, E-Mail-Server, etc.

# 3. TLS-Zertifikate (selbst-signiert für lokale Nutzung)
./scripts/generate-certs.sh

# 4. Starten
docker compose up -d

# 5. Initialen Superadmin anlegen
./scripts/create-admin.sh admin@firma.de

# 6. Browser öffnen
echo "MyRMS läuft unter: https://$(hostname -f)"
```

**Geschätzter Zeitaufwand für Erstinstallation:** 30 Minuten (Linux/Mac), 60 Minuten (Windows mit WSL2)

### Updates

```bash
# Automatisch via Watchtower (täglich 3 Uhr)
# Oder manuell:
./scripts/update.sh

# Rollback im Notfall:
./scripts/rollback.sh v1.2.3
```

### Backup

```bash
# Tägliches automatisches Backup
# Via cron (crontab -e):
0 2 * * * /opt/myrms/scripts/backup.sh >> /var/log/myrms-backup.log 2>&1

# Backup-Inhalt:
# - PostgreSQL pg_dump (alle Schemas)
# - KurrentDB chunk-backup
# - Redis BGSAVE + RDB-Datei-Kopie
# - .env (verschlüsselt)
# - /uploads (Bilder, Dokumente)
```

---

## Risiken und Gegenmaßnahmen

| Risiko | Wahrscheinlichkeit | Schwere | Gegenmaßnahme |
|--------|-------------------|---------|---------------|
| KurrentDB-Instabilität | Niedrig | Hoch | KurrentDB 23.10 LTS, automated health checks, Restart-Policy |
| PostgreSQL-Datenverlust | Sehr Niedrig | Kritisch | WAL-G Backup, tägliche pg_dump, Offsite-Backup |
| Zebra TC21 DataWedge inkompatibel | Mittel | Mittel | Frühzeitiger Hardware-Test in Phase 2, Fallback auf Kamera-Scan |
| Lange AI-Antwortzeiten (>30s) | Mittel | Niedrig | Async Jobs mit SSE/WebSocket, Ollama lokal als Fallback |
| DATEV-Export-Format-Änderung | Niedrig | Mittel | DATEV-Format versioniert, regelmäßige Prüfung |
| GoBD-Anforderungsänderung | Sehr Niedrig | Hoch | Modulare Audit-Implementierung, Fachanwalt konsultieren |
| Entwickler-Burnout (Solo) | Mittel | Hoch | Klarer Phasenplan, realistische Zeitschätzungen, Puffer einplanen |
| Fehlende Nutzer-Akzeptanz | Mittel | Hoch | Frühes User-Testing mit Persona-Vorbildern, UX-Fokus |

---

## Zeitplan-Übersicht (Gantt-Darstellung)

```
Monat:      1    2    3    4    5    6    7    8    9   10   11   12   13   14   15   16   17   18
           |----|----|----|----|----|----|----|----|----|----|----|----|----|----|----|----|----|----|
Phase 0:   [====]
Phase 1:        [===========]
Phase 2:                    [=========]
Phase 3:                              [========]
Phase 4:                                        [===========]
Refactor:       [  ]        [  ]      [  ]      [  ]
Testing:                         [  ]      [  ]      [  ]      [  ]
Docs:      [=================================================]
```

**Legende:** `[====]` = Hauptarbeit, `[  ]` = Refactoring/Testing-Sprints

---

## Ressourcen und Abhängigkeiten

### Externe Dienste (optional, selbst-hostbar)

| Dienst | Zweck | Self-Hosted-Alternative |
|--------|-------|------------------------|
| SMTP-Server | E-Mail-Versand | Postfix + Dovecot |
| S3-kompatibler Speicher | Equipment-Bilder | MinIO |
| Ollama | Lokale KI-Modelle | Pflicht wenn kein Cloud-KI |
| IZYTRON.IQ | E-Check-Prüfsoftware | Externes Tool (Dritthersteller) |
| DATEV | Buchhaltungssoftware | Nur Export, kein Self-Host |

### Go-Abhängigkeiten (kritische Third-Party-Libs)

| Library | Version | Zweck |
|---------|---------|-------|
| `go.chi/chi` | v5.x | HTTP-Router |
| `jackc/pgx` | v5.x | PostgreSQL-Treiber |
| `EventStore/EventStore-Client-Go` | latest | KurrentDB-Client |
| `go-redis/redis` | v9.x | Redis-Client |
| `golang-jwt/jwt` | v5.x | JWT |
| `golang-migrate/migrate` | v4.x | DB-Migrationen |
| `rs/zerolog` | latest | Structured Logging |
| `go-playground/validator` | v10.x | Input-Validierung |
| `google/uuid` | latest | UUID-Generierung |
| `chromedp/chromedp` | latest | PDF-Generierung |
| `prometheus/client_golang` | latest | Metriken |

### NPM-Abhängigkeiten (kritische Frontend-Libs)

| Package | Version | Zweck |
|---------|---------|-------|
| `react` | 18.x | UI-Framework |
| `@tanstack/react-router` | latest | Routing |
| `@tanstack/react-query` | v5.x | Server State |
| `zustand` | v4.x | Client State |
| `@radix-ui/themes` | v3.x | UI-Komponenten |
| `vite` | v5.x | Build-Tool |
| `vite-plugin-pwa` | latest | PWA |
| `react-i18next` | latest | Internationalisierung |
| `date-fns` | v3.x | Datum-Hilfsfunktionen |
| `recharts` | v2.x | Diagramme |
| `@tanstack/react-virtual` | latest | Listenvirtualisierung |

---

## Qualitätssicherung

### Test-Strategie

**Unit-Tests (Go):**
- Domain-Aggregate (Event-Sourcing-Logik)
- Preis-Kalkulation
- Verfügbarkeits-Algorithmus
- Checksummen-Kette (Audit)
- Mindest-Coverage: 80%

**Integrations-Tests (Go + Testcontainers):**
- PostgreSQL-Repositories gegen echte Test-DB
- KurrentDB-Aggregate gegen Test-KurrentDB
- gRPC-Kommunikation zwischen Services

**End-to-End-Tests (Playwright):**
- Kritischer Pfad: Projekt anlegen → Equipment zuweisen → Rechnung erstellen → DATEV-Export
- Scanner-Workflow: Scan → Check-Out → Check-In
- Krit. Pfad für Freelancer Kevin: Login → Einsatz sehen → Zeiterfassung

**Performance-Tests (k6):**
- 100 gleichzeitige Nutzer auf `/api/v1/equipment` (Verfügbarkeitscheck)
- Scanner-Batch: 500 Scans in 10 Sekunden
- Ziel: p95 < 200ms

### Definition of Done (DoD)

Für jede Feature-Implementation gilt ein Ticket als "Done" wenn:
- [ ] Code kompiliert ohne Warnungen
- [ ] Unit-Tests geschrieben und grün (min. 80% Abdeckung für neue Funktionen)
- [ ] Integrations-Test für kritische Datenbankoperationen
- [ ] Kein `golangci-lint` Fehler
- [ ] API-Endpunkt in Swagger/OpenAPI dokumentiert
- [ ] CHANGELOG.md aktualisiert
- [ ] Docker-Build läuft durch
- [ ] CI-Pipeline grün
- [ ] Peer-Review (wenn Entwicklerteam > 1 Person)

---

## Changelog-Strategie

**Versionierung:** Semantic Versioning (`MAJOR.MINOR.PATCH`)
- `MAJOR`: Breaking API-Änderungen
- `MINOR`: Neue Features (rückwärtskompatibel)
- `PATCH`: Bugfixes

**Release-Zyklus:**
- Patch-Releases: Bei kritischen Bugs (sofort)
- Minor-Releases: Am Ende jeder Phase
- Major-Releases: Bei Breaking Changes (selten)

**Git-Flow:**
```
main         ─────────────────────────────────────►  Produktionsstand
              ↑           ↑          ↑
develop      ─────────────────────────────────────►  Entwicklungsstand
              ↑           ↑          ↑
feature/*    ─────►      ─────►     ─────►           Feature-Branches
```

---

## Erste Schritte für neue Entwickler

```bash
# 1. Prerequisites installieren
# - Go 1.22+
# - Node.js 20+
# - Docker Desktop (oder Docker Engine + Compose)
# - VS Code + Go-Extension + ESLint-Extension

# 2. Repo klonen
git clone https://github.com/myrms/myrms.git
cd myrms

# 3. Infrastruktur starten (Datenbanken)
docker compose -f docker-compose.dev.yml up -d postgres kurrentdb redis

# 4. Alle Services im Dev-Modus starten
make dev

# 5. Datenbank-Migrationen ausführen
make migrate-up

# 6. Frontend starten
cd frontend && npm install && npm run dev

# 7. Browser öffnen
# Frontend: http://localhost:5173
# API: http://localhost:8001 (auth-service)
# KurrentDB: http://localhost:2113
# Traefik Dashboard: http://localhost:8080

# 8. Test-Daten einspielen
make seed-dev-data
```

---

*Dokument erstellt: März 2026 | Zielplattform: MyRMS v1.0 | Sprache: Deutsch*
*Alle Zeitschätzungen basieren auf Solo-Entwicklung; bei Teamgröße ≥ 2 entsprechend reduzieren.*
