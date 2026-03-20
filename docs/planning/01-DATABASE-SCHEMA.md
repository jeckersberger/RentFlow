# MyRMS — Datenbankschema (PostgreSQL + KurrentDB + Redis)

**Stand:** 20. März 2026
**Version:** 1.0

---

## Überblick

MyRMS verwendet eine **Dual-Database-Architektur**:

- **KurrentDB** (ehemals EventStoreDB): Event Store, Source of Truth, alle Zustandsänderungen als immutable Events
- **PostgreSQL 16**: Read-Projections, Schema-per-Service, CQRS Read-Side
- **Redis 7**: Session Store, API-Cache, Rate Limiting, Pub/Sub

Jeder der 17 Microservices besitzt sein eigenes PostgreSQL-Schema. Cross-Schema-Joins sind **verboten** — Daten anderer Services werden nur über Events oder synchrone API-Calls bezogen.

---

## KurrentDB — Event Streams

### Stream-Konventionen

```
Aggregate-Streams:    <aggregate-type>-<uuid>
                      z.B. equipment-550e8400-e29b-41d4-a716-446655440000

Kategorie-Streams:    $ce-<kategorie>  (automatisch von KurrentDB)
                      z.B. $ce-equipment, $ce-project, $ce-invoice

System-Streams:       $all (alle Events aller Streams)
```

### Standard-Event-Envelope

```json
{
  "event_id":       "UUID v4 (idempotency key)",
  "event_type":     "domain.action_past_tense",
  "aggregate_type": "equipment | project | invoice | ...",
  "aggregate_id":   "UUID des Aggregats",
  "tenant_id":      "UUID der Firma (Mandantenfähigkeit)",
  "correlation_id": "UUID für verteiltes Request-Tracing",
  "causation_id":   "event_id des auslösenden Events",
  "timestamp":      "ISO-8601 UTC",
  "version":        "Integer (Aggregat-Version für Optimistic Concurrency)",
  "schema_version": "Integer (Event-Schema-Version für Migration)",
  "data":           { ... },
  "metadata": {
    "user_id":        "UUID des auslösenden Users",
    "source_service": "inventory-service",
    "ip_address":     "optional, für Audit"
  }
}
```

---

## Service 1: auth-service (Port :8001)

### PostgreSQL-Schema: `auth_schema`

```sql
CREATE SCHEMA auth_schema;
SET search_path = auth_schema;

-- Mandanten (Firmen)
CREATE TABLE tenants (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(255) NOT NULL,
    slug            VARCHAR(100) UNIQUE NOT NULL,       -- URL-freundlicher Bezeichner
    email           VARCHAR(255),
    phone           VARCHAR(50),
    address_street  VARCHAR(255),
    address_city    VARCHAR(100),
    address_zip     VARCHAR(20),
    address_country CHAR(2) DEFAULT 'DE',
    tax_number      VARCHAR(50),                        -- Steuernummer
    vat_id          VARCHAR(30),                        -- USt-IdNr.
    iban            VARCHAR(34),
    bic             VARCHAR(11),
    bank_name       VARCHAR(100),
    logo_url        VARCHAR(500),
    currency        CHAR(3) DEFAULT 'EUR',
    locale          VARCHAR(10) DEFAULT 'de-DE',
    timezone        VARCHAR(50) DEFAULT 'Europe/Berlin',
    invoice_prefix  VARCHAR(20) DEFAULT 'RE',
    invoice_counter INTEGER DEFAULT 1,
    settings        JSONB DEFAULT '{}',                 -- Mandanten-spezifische Einstellungen
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Benutzer
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email           VARCHAR(255) NOT NULL,
    password_hash   VARCHAR(255),                       -- bcrypt, NULL bei SSO
    first_name      VARCHAR(100) NOT NULL,
    last_name       VARCHAR(100) NOT NULL,
    phone           VARCHAR(50),
    avatar_url      VARCHAR(500),
    locale          VARCHAR(10) DEFAULT 'de-DE',
    timezone        VARCHAR(50) DEFAULT 'Europe/Berlin',
    is_active       BOOLEAN DEFAULT TRUE,
    is_verified     BOOLEAN DEFAULT FALSE,
    last_login_at   TIMESTAMPTZ,
    failed_logins   INTEGER DEFAULT 0,
    locked_until    TIMESTAMPTZ,
    mfa_secret      VARCHAR(100),                       -- TOTP-Secret (verschlüsselt)
    mfa_enabled     BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(tenant_id, email)
);

-- Rollen
CREATE TABLE roles (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name        VARCHAR(100) NOT NULL,
    slug        VARCHAR(100) NOT NULL,                  -- admin, project_manager, warehouse, ...
    description TEXT,
    is_system   BOOLEAN DEFAULT FALSE,                  -- Systemrollen sind nicht löschbar
    permissions JSONB DEFAULT '[]',                     -- Array von Permission-Strings
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(tenant_id, slug)
);

-- Benutzer-Rollen-Zuordnung
CREATE TABLE user_roles (
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id    UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    granted_by UUID REFERENCES users(id),
    granted_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

-- Einladungen
CREATE TABLE invitations (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email       VARCHAR(255) NOT NULL,
    role_id     UUID NOT NULL REFERENCES roles(id),
    token       VARCHAR(255) UNIQUE NOT NULL,
    invited_by  UUID REFERENCES users(id),
    expires_at  TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Passwort-Reset-Tokens
CREATE TABLE password_reset_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Refresh-Tokens (für JWT-Rotation)
CREATE TABLE refresh_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) UNIQUE NOT NULL,             -- SHA-256 des Tokens
    user_agent VARCHAR(500),
    ip_address INET,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indizes
CREATE INDEX idx_users_tenant ON users(tenant_id);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_roles_tenant ON roles(tenant_id);
CREATE INDEX idx_invitations_token ON invitations(token);
CREATE INDEX idx_refresh_tokens_hash ON refresh_tokens(token_hash);
```

### KurrentDB-Events (auth)

| Event-Typ | Beschreibung |
|-----------|-------------|
| `TenantCreated` | Neue Firma angelegt |
| `TenantUpdated` | Firmendaten geändert |
| `UserRegistered` | Neuer User via Einladung |
| `UserLoggedIn` | Login erfolgreich |
| `UserLoggedOut` | Logout |
| `UserPasswordChanged` | Passwort geändert |
| `UserRoleAssigned` | Rolle zugewiesen |
| `UserRoleRevoked` | Rolle entzogen |
| `UserDeactivated` | User deaktiviert |
| `InvitationCreated` | Einladungslink erstellt |
| `InvitationAccepted` | Einladung angenommen |
| `MFAEnabled` | Zwei-Faktor aktiviert |
| `LoginFailed` | Fehlgeschlagener Login |

---

## Service 2: inventory-service (Port :8002)

### PostgreSQL-Schema: `inventory_schema`

```sql
CREATE SCHEMA inventory_schema;
SET search_path = inventory_schema;

-- Kategorien (Hierarchisch)
CREATE TABLE categories (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    parent_id   UUID REFERENCES categories(id),
    name        VARCHAR(255) NOT NULL,
    slug        VARCHAR(255) NOT NULL,
    description TEXT,
    icon        VARCHAR(100),                           -- Emoji oder Icon-Name
    color       VARCHAR(7),                             -- Hex-Farbe
    sort_order  INTEGER DEFAULT 0,
    is_active   BOOLEAN DEFAULT TRUE,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(tenant_id, slug)
);

-- Equipment / Artikel
CREATE TABLE equipment (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID NOT NULL,
    category_id         UUID REFERENCES categories(id),
    internal_number     VARCHAR(100),                   -- Interne Artikelnummer (z.B. "LT-0042")
    serial_number       VARCHAR(255),                   -- Seriennummer des Herstellers
    name                VARCHAR(500) NOT NULL,
    manufacturer        VARCHAR(255),
    model               VARCHAR(255),
    description         TEXT,
    technical_specs     JSONB DEFAULT '{}',             -- Hersteller-Spezifikationen
    internal_notes      TEXT,                           -- Interne Hinweise (nur Staff)
    purchase_price      NUMERIC(12,2),                  -- Einkaufspreis
    replacement_value   NUMERIC(12,2),                  -- Wiederbeschaffungswert (für Versicherung)
    weight_kg           NUMERIC(8,3),
    dimensions_cm       JSONB,                          -- {"length": 60, "width": 40, "height": 20}
    power_consumption_w INTEGER,                        -- Leistungsaufnahme in Watt
    quantity_total      INTEGER NOT NULL DEFAULT 1,
    quantity_available  INTEGER NOT NULL DEFAULT 1,     -- Denormalisiert für schnelle Abfragen
    quantity_rented     INTEGER NOT NULL DEFAULT 0,
    quantity_defect     INTEGER NOT NULL DEFAULT 0,
    quantity_maintenance INTEGER NOT NULL DEFAULT 0,
    status              VARCHAR(50) DEFAULT 'available', -- available, rented, defect, maintenance, retired
    condition           VARCHAR(50) DEFAULT 'good',      -- new, good, used, defect
    location_id         UUID,                            -- Aktueller Lagerplatz (FK zu warehouse_schema)
    barcode             VARCHAR(255),                    -- Barcode-Wert
    qr_code             VARCHAR(255),                   -- QR-Code-Wert (häufig = barcode)
    rfid_tag            VARCHAR(255),                   -- RFID-Tag-ID
    label_template_id   UUID,                           -- FK zu document_schema.label_templates
    is_consumable       BOOLEAN DEFAULT FALSE,          -- Verbrauchsmaterial
    is_bundle           BOOLEAN DEFAULT FALSE,          -- Ist selbst Teil eines Bundles
    min_stock_quantity  INTEGER DEFAULT 0,              -- Mindestbestand-Alarm
    last_maintenance_at TIMESTAMPTZ,
    next_maintenance_at TIMESTAMPTZ,
    echeck_interval_days INTEGER,                       -- DGUV V3 Prüfintervall
    last_echeck_at      TIMESTAMPTZ,
    next_echeck_at      TIMESTAMPTZ,
    retired_at          TIMESTAMPTZ,
    images              JSONB DEFAULT '[]',             -- Array von Bild-URLs
    attachments         JSONB DEFAULT '[]',             -- Handbücher, Datenblätter
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- Preisregeln
CREATE TABLE price_rules (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    equipment_id    UUID REFERENCES equipment(id) ON DELETE CASCADE,
    category_id     UUID REFERENCES categories(id),     -- Preis für alle Artikel der Kategorie
    name            VARCHAR(255) NOT NULL,
    price_type      VARCHAR(50) NOT NULL,               -- daily, weekly, monthly, flat, hourly
    base_price      NUMERIC(12,2) NOT NULL,
    currency        CHAR(3) DEFAULT 'EUR',
    min_duration    INTEGER DEFAULT 1,                  -- Mindestmietdauer (Tage)
    -- Staffelpreise
    tier_2_from     INTEGER,                            -- Ab N Tagen Tier-2-Preis
    tier_2_price    NUMERIC(12,2),
    tier_3_from     INTEGER,
    tier_3_price    NUMERIC(12,2),
    -- Saisonzuschläge
    season_surcharge_pct  NUMERIC(5,2) DEFAULT 0,      -- Prozent-Aufschlag
    season_start    DATE,
    season_end      DATE,
    -- Mengenstaffel
    qty_discount_threshold INTEGER,                    -- Ab N Stück Rabatt
    qty_discount_pct       NUMERIC(5,2) DEFAULT 0,
    is_default      BOOLEAN DEFAULT FALSE,
    is_active       BOOLEAN DEFAULT TRUE,
    valid_from      DATE,
    valid_until     DATE,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Bundles (Equipment-Pakete)
CREATE TABLE bundles (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    bundle_price NUMERIC(12,2),                        -- Paketpreis (statt Einzelpreise)
    is_active   BOOLEAN DEFAULT TRUE,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE bundle_items (
    bundle_id    UUID NOT NULL REFERENCES bundles(id) ON DELETE CASCADE,
    equipment_id UUID NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
    quantity     INTEGER NOT NULL DEFAULT 1,
    PRIMARY KEY (bundle_id, equipment_id)
);

-- Flightcases
CREATE TABLE flightcases (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    barcode         VARCHAR(255) UNIQUE,
    name            VARCHAR(255) NOT NULL,
    case_type       VARCHAR(100),                       -- rack, transport, custom
    max_weight_kg   NUMERIC(8,2),
    dimensions_cm   JSONB,
    rfid_tag        VARCHAR(255),
    current_contents JSONB DEFAULT '[]',               -- Snapshot der aktuellen Inhalte
    notes           TEXT,
    images          JSONB DEFAULT '[]',
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Indizes
CREATE INDEX idx_equipment_tenant ON equipment(tenant_id);
CREATE INDEX idx_equipment_category ON equipment(category_id);
CREATE INDEX idx_equipment_barcode ON equipment(barcode);
CREATE INDEX idx_equipment_serial ON equipment(serial_number);
CREATE INDEX idx_equipment_status ON equipment(status);
CREATE INDEX idx_price_rules_equipment ON price_rules(equipment_id);
```

### KurrentDB-Events (inventory)

| Event-Typ | Beschreibung |
|-----------|-------------|
| `EquipmentCreated` | Neuer Artikel angelegt |
| `EquipmentUpdated` | Artikeldaten geändert |
| `EquipmentCheckedOut` | Artikel ausgegeben |
| `EquipmentCheckedIn` | Artikel zurückgenommen |
| `EquipmentReserved` | Artikel für Projekt reserviert |
| `EquipmentReleased` | Reservierung aufgehoben |
| `EquipmentConditionChanged` | Zustand geändert (OK → Defekt) |
| `EquipmentRetired` | Artikel ausgemustert |
| `StockAdjusted` | Bestandskorrektur (Inventur) |
| `PriceRuleCreated` | Neue Preisregel |
| `PriceRuleUpdated` | Preisregel geändert |
| `BundleCreated` | Neues Bundle |
| `MinStockAlert` | Mindestbestand unterschritten |

---

## Service 3: project-service (Port :8003)

### PostgreSQL-Schema: `project_schema`

```sql
CREATE SCHEMA project_schema;
SET search_path = project_schema;

-- Kunden
CREATE TABLE customers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    customer_number VARCHAR(50),
    type            VARCHAR(20) DEFAULT 'company',      -- company, individual
    company_name    VARCHAR(255),
    first_name      VARCHAR(100),
    last_name       VARCHAR(100),
    email           VARCHAR(255),
    phone           VARCHAR(50),
    mobile          VARCHAR(50),
    address_street  VARCHAR(255),
    address_city    VARCHAR(100),
    address_zip     VARCHAR(20),
    address_country CHAR(2) DEFAULT 'DE',
    vat_id          VARCHAR(30),
    payment_terms   INTEGER DEFAULT 14,                 -- Zahlungsziel in Tagen
    credit_limit    NUMERIC(12,2),
    notes           TEXT,
    tags            JSONB DEFAULT '[]',
    is_active       BOOLEAN DEFAULT TRUE,
    portal_access   BOOLEAN DEFAULT FALSE,
    portal_email    VARCHAR(255),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Projekte / Jobs
CREATE TABLE projects (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    customer_id     UUID REFERENCES customers(id),
    project_number  VARCHAR(50) UNIQUE,
    name            VARCHAR(500) NOT NULL,
    description     TEXT,
    status          VARCHAR(50) DEFAULT 'inquiry',
    -- Status: inquiry, offer_sent, confirmed, in_preparation, active, completed, cancelled, invoiced
    location_name   VARCHAR(255),
    location_address TEXT,
    location_lat    NUMERIC(10,7),
    location_lng    NUMERIC(10,7),
    setup_start     TIMESTAMPTZ,
    event_start     TIMESTAMPTZ NOT NULL,
    event_end       TIMESTAMPTZ NOT NULL,
    teardown_end    TIMESTAMPTZ,
    project_manager_id UUID,                            -- User-ID (aus auth_schema via API)
    crew_ids        JSONB DEFAULT '[]',                 -- Zugewiesene Crew-Mitglieder
    total_value     NUMERIC(12,2),                      -- Kalkulierter Projektwert
    notes           TEXT,
    internal_notes  TEXT,
    tags            JSONB DEFAULT '[]',
    attachments     JSONB DEFAULT '[]',
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Packlisten-Positionen
CREATE TABLE project_equipment (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    equipment_id    UUID NOT NULL,                      -- aus inventory_schema via API
    quantity        INTEGER NOT NULL DEFAULT 1,
    quantity_checked_out INTEGER DEFAULT 0,
    quantity_checked_in  INTEGER DEFAULT 0,
    status          VARCHAR(50) DEFAULT 'planned',      -- planned, reserved, checked_out, checked_in, partial
    unit_price      NUMERIC(12,2),                      -- Preis zum Zeitpunkt der Planung
    discount_pct    NUMERIC(5,2) DEFAULT 0,
    notes           TEXT,
    is_sub_rental   BOOLEAN DEFAULT FALSE,              -- Fremdmaterial
    sub_rental_from VARCHAR(255),                       -- Quellfirma bei Sub-Rental
    sort_order      INTEGER DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Packlisten-Abschnitte (Gruppierung)
CREATE TABLE packing_list_sections (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,                  -- "Ton", "Licht", "Video"
    sort_order  INTEGER DEFAULT 0
);

-- Verfügbarkeits-Konflikte (für Doppelbuchungs-Prüfung)
CREATE TABLE availability_conflicts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    equipment_id    UUID NOT NULL,
    project_id_1    UUID NOT NULL,
    project_id_2    UUID NOT NULL,
    conflict_start  TIMESTAMPTZ NOT NULL,
    conflict_end    TIMESTAMPTZ NOT NULL,
    resolved        BOOLEAN DEFAULT FALSE,
    resolved_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Indizes
CREATE INDEX idx_projects_tenant ON projects(tenant_id);
CREATE INDEX idx_projects_customer ON projects(customer_id);
CREATE INDEX idx_projects_status ON projects(status);
CREATE INDEX idx_projects_dates ON projects(event_start, event_end);
CREATE INDEX idx_project_equipment_project ON project_equipment(project_id);
CREATE INDEX idx_project_equipment_equipment ON project_equipment(equipment_id);
```

### KurrentDB-Events (project)

| Event-Typ | Beschreibung |
|-----------|-------------|
| `ProjectCreated` | Neues Projekt/Anfrage |
| `ProjectStatusChanged` | Statuswechsel |
| `ProjectEquipmentAdded` | Equipment zur Packliste hinzugefügt |
| `ProjectEquipmentRemoved` | Equipment aus Packliste entfernt |
| `ProjectEquipmentQuantityChanged` | Menge geändert |
| `ProjectCompleted` | Projekt abgeschlossen |
| `ProjectCancelled` | Projekt storniert |
| `AvailabilityConflictDetected` | Doppelbuchung erkannt |
| `CustomerCreated` | Neuer Kunde angelegt |
| `CustomerUpdated` | Kundendaten geändert |

---

## Service 4: scanner-service (Port :8004)

### PostgreSQL-Schema: `scanner_schema`

```sql
CREATE SCHEMA scanner_schema;
SET search_path = scanner_schema;

-- Scan-Ereignisse (vollständige Historie aller Scans)
CREATE TABLE scan_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    scan_type       VARCHAR(50) NOT NULL,               -- checkin, checkout, inventory, damage_report, locate
    equipment_id    UUID NOT NULL,
    project_id      UUID,                               -- Kontext: welcher Job
    user_id         UUID NOT NULL,
    device_type     VARCHAR(50),                        -- zebra_tc21, iphone, usb_scanner, rfid
    device_id       VARCHAR(100),                       -- Geräte-Identifier
    barcode_value   VARCHAR(255),
    scan_method     VARCHAR(50),                        -- barcode, qr_code, rfid, manual
    condition_before VARCHAR(50),
    condition_after  VARCHAR(50),
    notes           TEXT,
    photos          JSONB DEFAULT '[]',
    location_id     UUID,
    latitude        NUMERIC(10,7),
    longitude       NUMERIC(10,7),
    offline_sync    BOOLEAN DEFAULT FALSE,              -- War offline, später synchronisiert
    synced_at       TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Offline-Scan-Queue (für Offline-Modus)
CREATE TABLE offline_scan_queue (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    device_id       VARCHAR(100) NOT NULL,
    user_id         UUID NOT NULL,
    scan_data       JSONB NOT NULL,                     -- Vollständige Scan-Daten
    local_timestamp TIMESTAMPTZ NOT NULL,
    synced          BOOLEAN DEFAULT FALSE,
    synced_at       TIMESTAMPTZ,
    sync_error      TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Aktive Check-Outs (denormalisiert für schnelle Abfragen)
CREATE TABLE active_checkouts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    equipment_id    UUID NOT NULL,
    project_id      UUID,
    checked_out_by  UUID NOT NULL,
    checked_out_at  TIMESTAMPTZ NOT NULL,
    due_back_at     TIMESTAMPTZ,
    notes           TEXT,
    UNIQUE(tenant_id, equipment_id)                     -- Ein Artikel kann nur einmal ausgecheckt sein
);

-- Indizes
CREATE INDEX idx_scan_events_tenant ON scan_events(tenant_id);
CREATE INDEX idx_scan_events_equipment ON scan_events(equipment_id);
CREATE INDEX idx_scan_events_project ON scan_events(project_id);
CREATE INDEX idx_scan_events_created ON scan_events(created_at DESC);
CREATE INDEX idx_active_checkouts_equipment ON active_checkouts(equipment_id);
```

### KurrentDB-Events (scanner)

| Event-Typ | Beschreibung |
|-----------|-------------|
| `ScanCompleted` | Scan durchgeführt |
| `CheckoutCompleted` | Equipment ausgegeben |
| `CheckinCompleted` | Equipment zurückgenommen |
| `DamageReportedViaScan` | Schaden beim Scan erfasst |
| `OfflineScanSynced` | Offline-Scan synchronisiert |

---

## Service 5: warehouse-service (Port :8005)

### PostgreSQL-Schema: `warehouse_schema`

```sql
CREATE SCHEMA warehouse_schema;
SET search_path = warehouse_schema;

-- Lager
CREATE TABLE warehouses (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    address     TEXT,
    is_main     BOOLEAN DEFAULT FALSE,
    is_active   BOOLEAN DEFAULT TRUE,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Lagerzonen (z.B. "Halle A", "Büro", "Außenlager")
CREATE TABLE warehouse_zones (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    warehouse_id    UUID NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    sort_order      INTEGER DEFAULT 0
);

-- Lagerorte (Regale, Fächer, Ebenen)
CREATE TABLE storage_locations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    warehouse_id    UUID NOT NULL REFERENCES warehouses(id),
    zone_id         UUID REFERENCES warehouse_zones(id),
    parent_id       UUID REFERENCES storage_locations(id),  -- Hierarchisch
    name            VARCHAR(255) NOT NULL,
    barcode         VARCHAR(255),
    qr_code         VARCHAR(255),
    location_type   VARCHAR(50) DEFAULT 'shelf',        -- shelf, rack, floor, vehicle
    max_weight_kg   NUMERIC(8,2),
    description     TEXT,
    icon            VARCHAR(100),
    sort_order      INTEGER DEFAULT 0,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Equipment-Lagerplatz-Zuordnung
CREATE TABLE equipment_locations (
    equipment_id    UUID NOT NULL,
    location_id     UUID NOT NULL REFERENCES storage_locations(id),
    tenant_id       UUID NOT NULL,
    quantity        INTEGER NOT NULL DEFAULT 1,
    notes           TEXT,
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (equipment_id, location_id)
);

-- Lagerbewegungen (Warenbewegung)
CREATE TABLE stock_movements (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    equipment_id    UUID NOT NULL,
    movement_type   VARCHAR(50) NOT NULL,               -- inbound, outbound, transfer, adjustment
    quantity        INTEGER NOT NULL,
    from_location_id UUID REFERENCES storage_locations(id),
    to_location_id  UUID REFERENCES storage_locations(id),
    project_id      UUID,
    reference_id    UUID,                               -- Scan-Event-ID, Invoice-ID, etc.
    reference_type  VARCHAR(50),
    user_id         UUID NOT NULL,
    notes           TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Inventuren
CREATE TABLE inventory_counts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    name            VARCHAR(255) NOT NULL,
    status          VARCHAR(50) DEFAULT 'draft',        -- draft, in_progress, completed, cancelled
    count_type      VARCHAR(50) DEFAULT 'full',         -- full, partial, location, category
    scope           JSONB DEFAULT '{}',                 -- Filter für partielle Inventur
    started_by      UUID NOT NULL,
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    notes           TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE inventory_count_items (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    count_id        UUID NOT NULL REFERENCES inventory_counts(id) ON DELETE CASCADE,
    equipment_id    UUID NOT NULL,
    location_id     UUID REFERENCES storage_locations(id),
    expected_qty    INTEGER NOT NULL,
    counted_qty     INTEGER,
    difference      INTEGER GENERATED ALWAYS AS (COALESCE(counted_qty, 0) - expected_qty) STORED,
    counted_by      UUID,
    counted_at      TIMESTAMPTZ,
    notes           TEXT
);

-- Indizes
CREATE INDEX idx_storage_locations_tenant ON storage_locations(tenant_id);
CREATE INDEX idx_storage_locations_barcode ON storage_locations(barcode);
CREATE INDEX idx_equipment_locations_equipment ON equipment_locations(equipment_id);
CREATE INDEX idx_stock_movements_equipment ON stock_movements(equipment_id);
CREATE INDEX idx_stock_movements_created ON stock_movements(created_at DESC);
```

### KurrentDB-Events (warehouse)

| Event-Typ | Beschreibung |
|-----------|-------------|
| `StockMoved` | Lagerbewegung |
| `EquipmentLocated` | Lagerplatz zugewiesen |
| `InventoryCountStarted` | Inventur gestartet |
| `InventoryCountCompleted` | Inventur abgeschlossen |
| `StockDiscrepancyFound` | Differenz bei Inventur |
| `LocationCreated` | Neuer Lagerplatz |

---

## Service 6: invoice-service (Port :8006)

### PostgreSQL-Schema: `invoice_schema`

```sql
CREATE SCHEMA invoice_schema;
SET search_path = invoice_schema;

-- Angebote
CREATE TABLE quotes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    customer_id     UUID NOT NULL,
    project_id      UUID,
    quote_number    VARCHAR(50) UNIQUE NOT NULL,
    status          VARCHAR(50) DEFAULT 'draft',        -- draft, sent, accepted, rejected, expired
    subject         VARCHAR(500),
    intro_text      TEXT,
    outro_text      TEXT,
    subtotal        NUMERIC(12,2) DEFAULT 0,
    discount_amount NUMERIC(12,2) DEFAULT 0,
    discount_pct    NUMERIC(5,2) DEFAULT 0,
    tax_rate        NUMERIC(5,2) DEFAULT 19.00,
    tax_amount      NUMERIC(12,2) DEFAULT 0,
    total           NUMERIC(12,2) DEFAULT 0,
    currency        CHAR(3) DEFAULT 'EUR',
    valid_until     DATE,
    payment_terms   INTEGER DEFAULT 14,
    notes           TEXT,
    internal_notes  TEXT,
    created_by      UUID NOT NULL,
    sent_at         TIMESTAMPTZ,
    accepted_at     TIMESTAMPTZ,
    rejected_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Rechnungen
CREATE TABLE invoices (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    customer_id     UUID NOT NULL,
    project_id      UUID,
    quote_id        UUID REFERENCES quotes(id),
    parent_invoice_id UUID REFERENCES invoices(id),    -- Für Teilrechnungen
    invoice_number  VARCHAR(50) UNIQUE NOT NULL,
    invoice_type    VARCHAR(50) DEFAULT 'invoice',      -- invoice, partial, credit_note, dunning, proforma
    status          VARCHAR(50) DEFAULT 'draft',        -- draft, sent, paid, partial_paid, overdue, cancelled
    subject         VARCHAR(500),
    intro_text      TEXT,
    outro_text      TEXT,
    subtotal        NUMERIC(12,2) DEFAULT 0,
    discount_amount NUMERIC(12,2) DEFAULT 0,
    discount_pct    NUMERIC(5,2) DEFAULT 0,
    tax_rate        NUMERIC(5,2) DEFAULT 19.00,
    tax_amount      NUMERIC(12,2) DEFAULT 0,
    total           NUMERIC(12,2) DEFAULT 0,
    paid_amount     NUMERIC(12,2) DEFAULT 0,
    outstanding     NUMERIC(12,2) GENERATED ALWAYS AS (total - paid_amount) STORED,
    currency        CHAR(3) DEFAULT 'EUR',
    invoice_date    DATE NOT NULL,
    due_date        DATE NOT NULL,
    payment_terms   INTEGER DEFAULT 14,
    partial_pct     NUMERIC(5,2),                       -- Bei Teilrechnungen: % des Gesamtbetrags
    datev_account   VARCHAR(20),                        -- DATEV-Kontonummer
    datev_cost_center VARCHAR(20),
    notes           TEXT,
    internal_notes  TEXT,
    created_by      UUID NOT NULL,
    sent_at         TIMESTAMPTZ,
    paid_at         TIMESTAMPTZ,
    cancelled_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Rechnungspositionen
CREATE TABLE invoice_items (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id      UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    position        INTEGER NOT NULL,
    item_type       VARCHAR(50) DEFAULT 'equipment',    -- equipment, service, transport, misc, discount
    equipment_id    UUID,
    description     VARCHAR(1000) NOT NULL,
    quantity        NUMERIC(10,3) NOT NULL DEFAULT 1,
    unit            VARCHAR(50) DEFAULT 'Stk',
    unit_price      NUMERIC(12,2) NOT NULL,
    discount_pct    NUMERIC(5,2) DEFAULT 0,
    tax_rate        NUMERIC(5,2) DEFAULT 19.00,
    total           NUMERIC(12,2) NOT NULL,
    rental_days     INTEGER,
    rental_from     DATE,
    rental_to       DATE,
    datev_account   VARCHAR(20),
    sort_order      INTEGER DEFAULT 0
);

-- Zahlungseingänge
CREATE TABLE payments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    invoice_id      UUID NOT NULL REFERENCES invoices(id),
    amount          NUMERIC(12,2) NOT NULL,
    payment_date    DATE NOT NULL,
    payment_method  VARCHAR(50),                        -- bank_transfer, cash, card, direct_debit
    reference       VARCHAR(255),                       -- Verwendungszweck / Buchungsreferenz
    notes           TEXT,
    bank_transaction_id UUID,
    created_by      UUID NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Mahnungen
CREATE TABLE dunning_notices (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    invoice_id      UUID NOT NULL REFERENCES invoices(id),
    dunning_level   INTEGER NOT NULL,                   -- 1, 2, 3
    sent_at         TIMESTAMPTZ,
    due_date        DATE,
    fee_amount      NUMERIC(12,2) DEFAULT 0,
    notes           TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Bank-Transaktionen (Import)
CREATE TABLE bank_transactions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    transaction_date DATE NOT NULL,
    value_date      DATE,
    amount          NUMERIC(12,2) NOT NULL,
    currency        CHAR(3) DEFAULT 'EUR',
    counterparty    VARCHAR(255),
    iban            VARCHAR(34),
    reference       VARCHAR(500),
    bank_reference  VARCHAR(255) UNIQUE,
    matched_invoice_id UUID REFERENCES invoices(id),
    match_status    VARCHAR(50) DEFAULT 'unmatched',    -- unmatched, matched, partial, ignored
    imported_at     TIMESTAMPTZ DEFAULT NOW()
);

-- Indizes
CREATE INDEX idx_invoices_tenant ON invoices(tenant_id);
CREATE INDEX idx_invoices_customer ON invoices(customer_id);
CREATE INDEX idx_invoices_status ON invoices(status);
CREATE INDEX idx_invoices_due_date ON invoices(due_date);
CREATE INDEX idx_payments_invoice ON payments(invoice_id);
```

### KurrentDB-Events (invoice)

| Event-Typ | Beschreibung |
|-----------|-------------|
| `InvoiceCreated` | Rechnung erstellt |
| `InvoiceSent` | Rechnung versendet |
| `InvoicePaid` | Rechnung bezahlt |
| `InvoicePartiallyPaid` | Teilzahlung |
| `InvoiceOverdue` | Rechnung überfällig |
| `InvoiceCancelled` | Rechnung storniert |
| `CreditNoteCreated` | Gutschrift erstellt |
| `DunningSent` | Mahnung versendet |
| `PaymentReceived` | Zahlung erfasst |
| `BankTransactionImported` | Kontoauszug importiert |
| `BankTransactionMatched` | Zahlung zugeordnet |
| `QuoteCreated` | Angebot erstellt |
| `QuoteAccepted` | Angebot angenommen |

---

## Service 7: document-service (Port :8007)

### PostgreSQL-Schema: `document_schema`

```sql
CREATE SCHEMA document_schema;
SET search_path = document_schema;

-- Dokument-Templates
CREATE TABLE document_templates (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    name            VARCHAR(255) NOT NULL,
    doc_type        VARCHAR(50) NOT NULL,               -- quote, invoice, delivery_note, pack_list, contract, label
    content         TEXT NOT NULL,                      -- HTML-Template mit Platzhaltern
    css             TEXT,
    is_default      BOOLEAN DEFAULT FALSE,
    version         INTEGER DEFAULT 1,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Generierte Dokumente
CREATE TABLE documents (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    template_id     UUID REFERENCES document_templates(id),
    reference_id    UUID NOT NULL,                      -- Invoice-ID, Project-ID, etc.
    reference_type  VARCHAR(50) NOT NULL,               -- invoice, quote, project, delivery_note
    filename        VARCHAR(500) NOT NULL,
    file_path       VARCHAR(1000),                      -- Pfad auf dem Dateisystem
    file_size_bytes INTEGER,
    mime_type       VARCHAR(100) DEFAULT 'application/pdf',
    checksum        VARCHAR(64),                        -- SHA-256 für GoBD
    generated_at    TIMESTAMPTZ DEFAULT NOW(),
    generated_by    UUID
);

-- Label-Vorlagen
CREATE TABLE label_templates (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    name        VARCHAR(255) NOT NULL,
    format      VARCHAR(50) DEFAULT 'zpl',              -- zpl, pdf, png
    width_mm    NUMERIC(6,2),
    height_mm   NUMERIC(6,2),
    content     TEXT NOT NULL,                          -- ZPL-Template oder HTML
    is_default  BOOLEAN DEFAULT FALSE,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Indizes
CREATE INDEX idx_documents_tenant ON documents(tenant_id);
CREATE INDEX idx_documents_reference ON documents(reference_id, reference_type);
```

---

## Service 8: crew-service (Port :8008)

### PostgreSQL-Schema: `crew_schema`

```sql
CREATE SCHEMA crew_schema;
SET search_path = crew_schema;

-- Crew-Mitglieder (Festangestellte + Freelancer)
CREATE TABLE crew_members (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    user_id         UUID,                               -- Verknüpfung mit auth_schema (optional)
    first_name      VARCHAR(100) NOT NULL,
    last_name       VARCHAR(100) NOT NULL,
    email           VARCHAR(255),
    phone           VARCHAR(50),
    mobile          VARCHAR(50),
    employment_type VARCHAR(50) DEFAULT 'freelancer',   -- employee, freelancer
    address         JSONB,
    emergency_contact JSONB,
    qualifications  JSONB DEFAULT '[]',                 -- ["fachkraft_va", "rigger", "ton", "licht"]
    certifications  JSONB DEFAULT '[]',                 -- Zertifikate mit Ablaufdatum
    daily_rate      NUMERIC(10,2),
    hourly_rate     NUMERIC(10,2),
    tax_id          VARCHAR(50),                        -- Steuerident für Freelancer
    iban            VARCHAR(34),
    caldav_url      VARCHAR(500),                       -- CalDAV-Abo-URL
    notes           TEXT,
    is_active       BOOLEAN DEFAULT TRUE,
    available_from  DATE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Verfügbarkeiten
CREATE TABLE crew_availabilities (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    crew_member_id  UUID NOT NULL REFERENCES crew_members(id) ON DELETE CASCADE,
    date_from       TIMESTAMPTZ NOT NULL,
    date_to         TIMESTAMPTZ NOT NULL,
    status          VARCHAR(50) DEFAULT 'available',    -- available, unavailable, tentative
    reason          TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Projektzuweisungen
CREATE TABLE crew_assignments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    crew_member_id  UUID NOT NULL REFERENCES crew_members(id),
    project_id      UUID NOT NULL,
    role            VARCHAR(100),                       -- "FOH Techniker", "Licht-Operator"
    date_from       TIMESTAMPTZ NOT NULL,
    date_to         TIMESTAMPTZ NOT NULL,
    status          VARCHAR(50) DEFAULT 'planned',      -- planned, confirmed, completed, cancelled
    agreed_rate     NUMERIC(10,2),
    notes           TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Zeiterfassung
CREATE TABLE time_entries (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    crew_member_id  UUID NOT NULL REFERENCES crew_members(id),
    project_id      UUID,
    assignment_id   UUID REFERENCES crew_assignments(id),
    check_in        TIMESTAMPTZ NOT NULL,
    check_out       TIMESTAMPTZ,
    duration_minutes INTEGER GENERATED ALWAYS AS (
        EXTRACT(EPOCH FROM (check_out - check_in)) / 60
    ) STORED,
    break_minutes   INTEGER DEFAULT 0,
    notes           TEXT,
    approved_by     UUID,
    approved_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Indizes
CREATE INDEX idx_crew_members_tenant ON crew_members(tenant_id);
CREATE INDEX idx_crew_assignments_project ON crew_assignments(project_id);
CREATE INDEX idx_time_entries_crew ON time_entries(crew_member_id);
CREATE INDEX idx_time_entries_project ON time_entries(project_id);
```

### KurrentDB-Events (crew)

| Event-Typ | Beschreibung |
|-----------|-------------|
| `CrewMemberCreated` | Neues Crew-Mitglied |
| `CrewAssigned` | Crew-Mitglied zu Projekt zugewiesen |
| `CrewAssignmentConfirmed` | Zusage des Crew-Mitglieds |
| `TimeEntryStarted` | Check-in |
| `TimeEntryCompleted` | Check-out |
| `AvailabilityChanged` | Verfügbarkeit geändert |

---

## Service 9: federation-service (Port :8009)

### PostgreSQL-Schema: `federation_schema`

```sql
CREATE SCHEMA federation_schema;
SET search_path = federation_schema;

-- Partner-Instanzen
CREATE TABLE federation_partners (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    partner_name    VARCHAR(255) NOT NULL,
    partner_url     VARCHAR(500) NOT NULL,              -- HTTPS-Endpunkt der Partner-Instanz
    partner_tenant_id UUID,                             -- UUID der Partner-Instanz
    certificate_pem TEXT,                               -- mTLS-Zertifikat des Partners
    our_cert_pem    TEXT,                               -- Unser Zertifikat für diesen Partner
    our_cert_key    TEXT,                               -- Privater Schlüssel (verschlüsselt)
    status          VARCHAR(50) DEFAULT 'pending',      -- pending, active, suspended, revoked
    invitation_token VARCHAR(255),
    token_expires_at TIMESTAMPTZ,
    last_seen_at    TIMESTAMPTZ,
    shared_categories JSONB DEFAULT '[]',               -- Kategorien die wir teilen
    pricing_config  JSONB DEFAULT '{}',                 -- Partner-Konditionen
    notes           TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Sub-Rental-Anfragen
CREATE TABLE sub_rental_requests (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    partner_id      UUID NOT NULL REFERENCES federation_partners(id),
    direction       VARCHAR(20) NOT NULL,               -- outgoing (wir anfragen) / incoming (wir werden angefragt)
    equipment_external_id VARCHAR(255),
    equipment_name  VARCHAR(500),
    quantity        INTEGER NOT NULL,
    date_from       TIMESTAMPTZ NOT NULL,
    date_to         TIMESTAMPTZ NOT NULL,
    status          VARCHAR(50) DEFAULT 'requested',    -- requested, confirmed, declined, completed, cancelled
    agreed_price    NUMERIC(12,2),
    currency        CHAR(3) DEFAULT 'EUR',
    project_id      UUID,
    notes           TEXT,
    confirmed_at    TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Indizes
CREATE INDEX idx_federation_partners_tenant ON federation_partners(tenant_id);
CREATE INDEX idx_sub_rental_requests_tenant ON sub_rental_requests(tenant_id);
```

---

## Service 10: maintenance-service (Port :8010)

### PostgreSQL-Schema: `maintenance_schema`

```sql
CREATE SCHEMA maintenance_schema;
SET search_path = maintenance_schema;

-- Wartungspläne
CREATE TABLE maintenance_plans (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    equipment_id    UUID NOT NULL,
    plan_type       VARCHAR(50) NOT NULL,               -- scheduled, after_use, condition_based
    interval_type   VARCHAR(50),                        -- days, uses, months
    interval_value  INTEGER,
    description     TEXT,
    checklist       JSONB DEFAULT '[]',                 -- Array von Checklisten-Punkten
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Wartungsaufgaben
CREATE TABLE maintenance_tasks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    plan_id         UUID REFERENCES maintenance_plans(id),
    equipment_id    UUID NOT NULL,
    task_type       VARCHAR(50) DEFAULT 'maintenance',  -- maintenance, repair, echeck, inspection
    title           VARCHAR(500) NOT NULL,
    description     TEXT,
    status          VARCHAR(50) DEFAULT 'open',         -- open, in_progress, completed, cancelled
    priority        VARCHAR(20) DEFAULT 'normal',       -- low, normal, high, critical
    scheduled_for   DATE,
    due_date        DATE,
    assigned_to     UUID,
    checklist_results JSONB DEFAULT '{}',
    result          TEXT,
    cost            NUMERIC(12,2),
    photos          JSONB DEFAULT '[]',
    documents       JSONB DEFAULT '[]',
    completed_by    UUID,
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- E-Check / DGUV V3 Prüfungen
CREATE TABLE echeck_records (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    equipment_id    UUID NOT NULL,
    check_date      DATE NOT NULL,
    next_check_date DATE NOT NULL,
    result          VARCHAR(50) NOT NULL,               -- passed, failed, conditional
    performed_by    VARCHAR(255),                       -- Name des Prüfers
    measuring_device VARCHAR(255),                      -- Verwendetes Prüfgerät
    measuring_device_serial VARCHAR(100),
    measurements    JSONB DEFAULT '{}',                 -- Messwerte (Spannung, Strom, etc.)
    notes           TEXT,
    certificate_url VARCHAR(500),
    izytron_import_id VARCHAR(100),                     -- ID beim IZYTRON.IQ Import
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Indizes
CREATE INDEX idx_maintenance_tasks_equipment ON maintenance_tasks(equipment_id);
CREATE INDEX idx_maintenance_tasks_status ON maintenance_tasks(status);
CREATE INDEX idx_maintenance_tasks_due ON maintenance_tasks(due_date);
CREATE INDEX idx_echeck_records_equipment ON echeck_records(equipment_id);
CREATE INDEX idx_echeck_records_next ON echeck_records(next_check_date);
```

---

## Service 11: transport-service (Port :8011)

### PostgreSQL-Schema: `transport_schema`

```sql
CREATE SCHEMA transport_schema;
SET search_path = transport_schema;

-- Fahrzeuge
CREATE TABLE vehicles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    license_plate   VARCHAR(20) UNIQUE NOT NULL,
    vehicle_type    VARCHAR(50),                        -- sprinter, lkw_7_5t, van, pkw
    brand           VARCHAR(100),
    model           VARCHAR(100),
    year            INTEGER,
    payload_kg      NUMERIC(8,2),
    volume_m3       NUMERIC(6,2),
    fuel_type       VARCHAR(50),                        -- diesel, electric, hybrid
    fuel_consumption_per_100km NUMERIC(5,2),
    next_inspection DATE,
    next_service_date DATE,
    status          VARCHAR(50) DEFAULT 'available',    -- available, in_use, maintenance
    notes           TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Touren / Fahrten
CREATE TABLE tours (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    vehicle_id      UUID REFERENCES vehicles(id),
    driver_id       UUID,                               -- Crew-Mitglied
    project_id      UUID,
    tour_type       VARCHAR(50) DEFAULT 'delivery',     -- delivery, pickup, delivery_and_pickup
    status          VARCHAR(50) DEFAULT 'planned',      -- planned, in_progress, completed
    departure_location TEXT,
    destination     TEXT,
    planned_start   TIMESTAMPTZ,
    planned_end     TIMESTAMPTZ,
    actual_start    TIMESTAMPTZ,
    actual_end      TIMESTAMPTZ,
    distance_km     NUMERIC(8,2),
    fuel_cost       NUMERIC(10,2),
    route_data      JSONB,                              -- Optimierte Route (Waypoints)
    notes           TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Indizes
CREATE INDEX idx_vehicles_tenant ON vehicles(tenant_id);
CREATE INDEX idx_tours_vehicle ON tours(vehicle_id);
CREATE INDEX idx_tours_project ON tours(project_id);
```

---

## Service 12: insurance-service (Port :8012)

### PostgreSQL-Schema: `insurance_schema`

```sql
CREATE SCHEMA insurance_schema;
SET search_path = insurance_schema;

-- Versicherungspolicen
CREATE TABLE insurance_policies (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    policy_number   VARCHAR(100) NOT NULL,
    insurer_name    VARCHAR(255) NOT NULL,
    policy_type     VARCHAR(100),                       -- equipment, liability, transport, all_risk
    coverage_amount NUMERIC(14,2),
    premium_annual  NUMERIC(12,2),
    deductible      NUMERIC(12,2),
    valid_from      DATE NOT NULL,
    valid_until     DATE NOT NULL,
    auto_renew      BOOLEAN DEFAULT FALSE,
    document_url    VARCHAR(500),
    notes           TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Schadensfälle
CREATE TABLE damage_claims (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    policy_id       UUID REFERENCES insurance_policies(id),
    equipment_id    UUID NOT NULL,
    project_id      UUID,
    claim_number    VARCHAR(100),
    incident_date   DATE NOT NULL,
    reported_date   DATE NOT NULL,
    description     TEXT NOT NULL,
    damage_amount   NUMERIC(12,2),
    status          VARCHAR(50) DEFAULT 'reported',     -- reported, in_review, approved, rejected, settled
    photos          JSONB DEFAULT '[]',
    repair_quote_url VARCHAR(500),
    settlement_amount NUMERIC(12,2),
    settled_at      TIMESTAMPTZ,
    notes           TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Indizes
CREATE INDEX idx_insurance_policies_tenant ON insurance_policies(tenant_id);
CREATE INDEX idx_damage_claims_equipment ON damage_claims(equipment_id);
```

---

## Service 13: workflow-service (Port :8013)

### PostgreSQL-Schema: `workflow_schema`

```sql
CREATE SCHEMA workflow_schema;
SET search_path = workflow_schema;

-- Workflow-Definitionen (No-Code WENN-DANN)
CREATE TABLE workflows (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    trigger_type VARCHAR(50) NOT NULL,                  -- event, schedule, manual
    trigger_config JSONB NOT NULL,                      -- Event-Typ, Cron-Expression, etc.
    conditions  JSONB DEFAULT '[]',                     -- Array von Bedingungen
    actions     JSONB NOT NULL,                         -- Array von Aktionen
    is_active   BOOLEAN DEFAULT TRUE,
    last_run_at TIMESTAMPTZ,
    run_count   INTEGER DEFAULT 0,
    created_by  UUID NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Workflow-Ausführungen
CREATE TABLE workflow_executions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id     UUID NOT NULL REFERENCES workflows(id),
    trigger_event_id VARCHAR(255),
    status          VARCHAR(50) DEFAULT 'running',      -- running, completed, failed
    context         JSONB DEFAULT '{}',
    results         JSONB DEFAULT '[]',
    error           TEXT,
    duration_ms     INTEGER,
    started_at      TIMESTAMPTZ DEFAULT NOW(),
    completed_at    TIMESTAMPTZ
);

-- Indizes
CREATE INDEX idx_workflows_tenant ON workflows(tenant_id);
CREATE INDEX idx_workflow_executions_workflow ON workflow_executions(workflow_id);
```

---

## Service 14: ai-service (Port :8014)

### PostgreSQL-Schema: `ai_schema`

```sql
CREATE SCHEMA ai_schema;
SET search_path = ai_schema;

-- KI-Provider-Konfigurationen
CREATE TABLE ai_providers (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    provider    VARCHAR(50) NOT NULL,                   -- anthropic, openai, google, mistral, ollama
    api_key     TEXT,                                   -- Verschlüsselt gespeichert
    model_name  VARCHAR(100),
    endpoint_url VARCHAR(500),                          -- Für Ollama/Custom
    is_active   BOOLEAN DEFAULT TRUE,
    is_default  BOOLEAN DEFAULT FALSE,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- KI-Anfragen (für Usage-Tracking und Feedback)
CREATE TABLE ai_requests (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    provider_id     UUID REFERENCES ai_providers(id),
    task_type       VARCHAR(100) NOT NULL,              -- price_optimization, demand_forecast, asset_recognition
    prompt_tokens   INTEGER,
    completion_tokens INTEGER,
    total_tokens    INTEGER,
    cost_usd        NUMERIC(10,6),
    duration_ms     INTEGER,
    success         BOOLEAN DEFAULT TRUE,
    error           TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Preis-Optimierungsvorschläge
CREATE TABLE price_suggestions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    equipment_id    UUID NOT NULL,
    suggested_price NUMERIC(12,2),
    current_price   NUMERIC(12,2),
    confidence      NUMERIC(5,4),                       -- 0.0 - 1.0
    reasoning       TEXT,
    accepted        BOOLEAN,
    accepted_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Indizes
CREATE INDEX idx_ai_requests_tenant ON ai_requests(tenant_id);
CREATE INDEX idx_ai_requests_created ON ai_requests(created_at DESC);
```

---

## Service 15: notification-service (Port :8015)

### PostgreSQL-Schema: `notification_schema`

```sql
CREATE SCHEMA notification_schema;
SET search_path = notification_schema;

-- Benachrichtigungen
CREATE TABLE notifications (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    user_id         UUID NOT NULL,
    type            VARCHAR(100) NOT NULL,              -- equipment_overdue, invoice_paid, project_confirmed, ...
    title           VARCHAR(500) NOT NULL,
    body            TEXT,
    data            JSONB DEFAULT '{}',
    channel         VARCHAR(50) DEFAULT 'in_app',       -- in_app, email, push, sms
    is_read         BOOLEAN DEFAULT FALSE,
    read_at         TIMESTAMPTZ,
    sent_at         TIMESTAMPTZ,
    send_error      TEXT,
    reference_id    UUID,
    reference_type  VARCHAR(50),
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- User-Benachrichtigungs-Präferenzen
CREATE TABLE notification_preferences (
    user_id         UUID NOT NULL,
    tenant_id       UUID NOT NULL,
    notification_type VARCHAR(100) NOT NULL,
    in_app          BOOLEAN DEFAULT TRUE,
    email           BOOLEAN DEFAULT TRUE,
    push            BOOLEAN DEFAULT FALSE,
    PRIMARY KEY (user_id, notification_type)
);

-- Push-Tokens
CREATE TABLE push_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL,
    platform    VARCHAR(20) NOT NULL,                   -- ios, android, web
    token       TEXT NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Indizes
CREATE INDEX idx_notifications_user ON notifications(user_id, is_read);
CREATE INDEX idx_notifications_tenant ON notifications(tenant_id, created_at DESC);
```

---

## Service 16: reporting-service (Port :8016)

### PostgreSQL-Schema: `reporting_schema`

```sql
CREATE SCHEMA reporting_schema;
SET search_path = reporting_schema;

-- Materialized Views / KPI-Snapshots
CREATE TABLE kpi_snapshots (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    snapshot_date   DATE NOT NULL,
    kpi_type        VARCHAR(100) NOT NULL,
    value           NUMERIC(18,4),
    metadata        JSONB DEFAULT '{}',
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(tenant_id, snapshot_date, kpi_type)
);

-- Dashboard-Konfigurationen
CREATE TABLE dashboard_configs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    user_id     UUID,                                   -- NULL = Tenant-Standard
    name        VARCHAR(255) NOT NULL,
    config      JSONB NOT NULL,                         -- Widget-Layout und Einstellungen
    is_default  BOOLEAN DEFAULT FALSE,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Gespeicherte Reports
CREATE TABLE saved_reports (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    name        VARCHAR(255) NOT NULL,
    report_type VARCHAR(100) NOT NULL,
    parameters  JSONB DEFAULT '{}',
    schedule    VARCHAR(100),                           -- Cron für automatische Ausführung
    last_run_at TIMESTAMPTZ,
    created_by  UUID NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Indizes
CREATE INDEX idx_kpi_snapshots_tenant_date ON kpi_snapshots(tenant_id, snapshot_date DESC);
```

---

## Service 17: audit-service (Port :8017)

### PostgreSQL-Schema: `audit_schema`

```sql
CREATE SCHEMA audit_schema;
SET search_path = audit_schema;

-- GoBD-konformer Audit-Trail
-- Alle Einträge sind UNVERÄNDERLICH (keine UPDATE/DELETE-Rechte für Services)
CREATE TABLE audit_log (
    id              BIGSERIAL PRIMARY KEY,              -- Sequential für GoBD-Reihenfolge
    log_uuid        UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    event_id        UUID NOT NULL,                      -- KurrentDB event_id
    event_type      VARCHAR(200) NOT NULL,
    aggregate_type  VARCHAR(100) NOT NULL,
    aggregate_id    UUID NOT NULL,
    user_id         UUID,
    user_email      VARCHAR(255),
    ip_address      INET,
    source_service  VARCHAR(100) NOT NULL,
    correlation_id  UUID,
    payload         JSONB NOT NULL,                     -- Vollständige Event-Daten
    -- GoBD-Felder
    checksum        VARCHAR(64) NOT NULL,               -- SHA-256(event_id + payload + prev_checksum)
    prev_checksum   VARCHAR(64),                        -- Kette: jeder Eintrag verweist auf den vorherigen
    sequence_number BIGINT NOT NULL,
    recorded_at     TIMESTAMPTZ DEFAULT NOW() NOT NULL,

    CONSTRAINT audit_log_immutable CHECK (TRUE)         -- Constraint als Erinnerung für Reviews
);

-- Checksummen-Verifikation
CREATE TABLE audit_chain_verifications (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    verified_from   BIGINT NOT NULL,
    verified_to     BIGINT NOT NULL,
    is_valid        BOOLEAN NOT NULL,
    verified_at     TIMESTAMPTZ DEFAULT NOW(),
    error_at_sequence BIGINT
);

-- Row Level Security — audit_service-User darf NUR LESEN und EINFÜGEN
ALTER TABLE audit_log ENABLE ROW LEVEL SECURITY;

-- Indizes (auf Spalten die GoBD-Auswertungen benötigen)
CREATE INDEX idx_audit_log_tenant ON audit_log(tenant_id, recorded_at DESC);
CREATE INDEX idx_audit_log_aggregate ON audit_log(aggregate_type, aggregate_id);
CREATE INDEX idx_audit_log_user ON audit_log(user_id);
CREATE INDEX idx_audit_log_event_type ON audit_log(event_type);
CREATE INDEX idx_audit_log_sequence ON audit_log(sequence_number);
```

---

## Redis 7 — Datenstrukturen

### Session Store (auth-service)

```
KEY: session:{session_id}
TYPE: Hash
TTL: 3600s (1 Stunde, renewbar)
FIELDS:
  user_id     → UUID
  tenant_id   → UUID
  email       → String
  roles       → JSON-Array
  ip          → IP-Adresse
  user_agent  → String
  created_at  → Unix-Timestamp
```

### Rate Limiting (Traefik / auth-service)

```
KEY: rate:{tenant_id}:{user_id}:{endpoint}
TYPE: String (Counter via INCR)
TTL: 60s (gleitendes Fenster via Redis EXPIRE)
VALUE: Integer (Anzahl der Requests im Fenster)

KEY: rate:ip:{ip_address}
TYPE: String
TTL: 60s
VALUE: Integer (für Login-Rate-Limiting)
```

### Equipment-Verfügbarkeits-Cache (inventory-service)

```
KEY: avail:{tenant_id}:{equipment_id}:{date_from}:{date_to}
TYPE: String (JSON)
TTL: 300s (5 Minuten)
VALUE: {"available": true, "quantity": 3, "conflicts": [...]}
```

### API-Response-Cache (verschiedene Services)

```
KEY: cache:{service}:{tenant_id}:{hash_of_query}
TYPE: String (JSON)
TTL: 60-300s (je nach Endpoint)
VALUE: Serialisierte API-Antwort
```

### WebSocket-Verbindungen (notification-service)

```
KEY: ws:connections:{tenant_id}
TYPE: Set
TTL: keins (persistent, Members werden bei Disconnect entfernt)
MEMBERS: {user_id}:{connection_id}

KEY: ws:user:{user_id}
TYPE: Set
MEMBERS: {connection_id}:{server_id}
```

### Pub/Sub-Kanäle (für Real-time Updates)

```
CHANNEL: events:{tenant_id}              → Alle Events eines Mandanten
CHANNEL: notifications:{user_id}         → User-spezifische Notifications
CHANNEL: equipment:{tenant_id}           → Equipment-Status-Updates
CHANNEL: project:{tenant_id}:{project_id} → Projekt-Updates
```

### Workflow-Locks (workflow-service)

```
KEY: lock:workflow:{workflow_id}:{execution_id}
TYPE: String
TTL: 300s (Schutz vor Doppelausführung)
VALUE: "locked"
```

### KI-Anfragen-Cache (ai-service)

```
KEY: ai:cache:{hash_of_prompt}
TYPE: String (JSON)
TTL: 3600s (1 Stunde, für identische Anfragen)
VALUE: KI-Antwort
```

---

## Migrations-Strategie

Jeder Service verwaltet seine eigenen Datenbankmigrationen unabhängig:

```
services/inventory-service/
└── migrations/
    ├── 000001_create_categories.up.sql
    ├── 000001_create_categories.down.sql
    ├── 000002_create_equipment.up.sql
    └── ...
```

- **Tool:** `golang-migrate/migrate`
- **Ausführung:** Automatisch beim Service-Start (configurable via `MIGRATE_ON_START=true`)
- **Transaktions-sicher:** Jede Migration in einer Transaktion
- **Reihenfolge:** Numerisches Präfix sichert korrekte Reihenfolge
- **Rollback:** Down-Migrations für alle Up-Migrations vorhanden

---

## Zusammenfassung Schema-Übersicht

| Service | PostgreSQL-Schema | Haupttabellen | KurrentDB-Streams |
|---------|------------------|---------------|-------------------|
| auth | auth_schema | tenants, users, roles, refresh_tokens | user-{id}, tenant-{id} |
| inventory | inventory_schema | equipment, categories, price_rules, bundles | equipment-{id} |
| project | project_schema | projects, customers, project_equipment | project-{id}, customer-{id} |
| scanner | scanner_schema | scan_events, active_checkouts | scan-{id} |
| warehouse | warehouse_schema | storage_locations, stock_movements, inventory_counts | warehouse-{id} |
| invoice | invoice_schema | invoices, quotes, payments, bank_transactions | invoice-{id}, quote-{id} |
| document | document_schema | document_templates, documents, label_templates | document-{id} |
| crew | crew_schema | crew_members, crew_assignments, time_entries | crew-{id} |
| federation | federation_schema | federation_partners, sub_rental_requests | federation-{id} |
| maintenance | maintenance_schema | maintenance_tasks, echeck_records | maintenance-{id} |
| transport | transport_schema | vehicles, tours | transport-{id} |
| insurance | insurance_schema | insurance_policies, damage_claims | insurance-{id} |
| workflow | workflow_schema | workflows, workflow_executions | workflow-{id} |
| ai | ai_schema | ai_providers, ai_requests, price_suggestions | ai-{id} |
| notification | notification_schema | notifications, notification_preferences, push_tokens | notification-{id} |
| reporting | reporting_schema | kpi_snapshots, dashboard_configs, saved_reports | — |
| audit | audit_schema | audit_log (UNVERÄNDERLICH) | — (subscribt auf $all) |
