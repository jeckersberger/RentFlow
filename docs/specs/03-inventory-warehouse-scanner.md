# RentFlow — Inventory & Warehouse Scanner Detailspezifikation

**Stand:** 21. März 2026
**Version:** 1.0
**Status:** Implementation Ready
**Autoren:** Senior Inventory Domain Expert, Warehouse/Logistics Specialist, IoT/Scanner Engineer

---

## Executive Summary

Diese Spezifikation definiert die komplette Implementierung der Inventory-, Warehouse- und Scanner-Domain für RentFlow. Sie umfasst:

1. **Inventory Domain Model** — Equipment-Typen, Kategorien, Kombinationen, Status-Maschine, Verfügbarkeitsalgorithmus, Preisengine
2. **Warehouse Domain Model** — Hierarchische Lagerstruktur, Warenbewegungen, Inventur-Workflows, Lagerplatz-Optimierung
3. **Scanner Service** — Zebra TC21, Handy-Kamera, USB-Scanner, RFID-Konzept
4. **Scan-Workflows** — Checkout, Check-in, Lagerplatz-Suche, Quick-Info
5. **Offline-Modus** — IndexedDB, Sync-Protokoll, Service Worker, Konflikt-Handling
6. **Label/QR-Code Generierung** — Content, ZPL-Template, öffentliche Info-Page

Zielgruppe: **Lisa Huber** (Lagerverwalterin) — Schnell, zuverlässig, offline-fähig, keine 10-Klicks-Workflows.

---

## 1. Inventory Domain Model

### 1.1 Equipment-Typen vs. Individuelle Items

**Equipment-Typ** (Type):
- Eindeutige Identifikation: `equipment_type_id` (UUID)
- Beispiel: `"Shure SM58 Microphone"`, `"Claypaky Sharpy B-Eye Pro"`, `"12m XLR Kabel 3-Pin"`
- Attribute:
  - `name`: String (max 255)
  - `category_id`: Foreign Key zu EquipmentCategory
  - `manufacturer`: String (optional)
  - `model_number`: String (optional)
  - `description`: Text (optional, technische Spezifikationen)
  - `unit_of_measure`: Enum(PIECE, METER, SET)
  - `photo_url`: URL zum Produkt-Foto
  - `default_condition`: Enum(NEW, GOOD, FAIR)
  - `sku`: String (eindeutig pro Mandant)
  - `created_at`: Timestamp
  - `created_by_user_id`: UUID
  - `is_active`: Boolean

**Individuelle Items** (Equipment Instance):
- Eindeutige Identifikation: `equipment_item_id` (UUID)
- Wenn `unit_of_measure = PIECE` → **IMMER einzelne Items mit Seriennummern tracken**
- Beispiele:
  - `"Yamaha CL5 Mischpult SN:ABC123DEF"` (High-Value, eindeutig)
  - `"Claypaky Sharpy SN:2024-0847"` (Moving Head, individuell)
  - `"XLR 15m Kabel ID:KBL-20240115-007"` (Bulk-Tracking mit Chargennummer)
- Attribute:
  - `equipment_item_id`: UUID (primary key)
  - `equipment_type_id`: Foreign Key
  - `serial_number`: String (optional, aber IMMER bei teurem Equipment oder Einzelteilen)
  - `batch_id`: String (für Bulk-Items wie Kabel: Chargennummer)
  - `purchase_date`: Date
  - `purchase_price_cents`: Integer (in Cent)
  - `current_condition`: Enum(NEW, GOOD, FAIR, DAMAGED, DEFECTIVE)
  - `last_condition_check_at`: Timestamp
  - `location_id`: Foreign Key zu WarehouseLocation (aktueller Lagerplatz)
  - `is_active`: Boolean (Retirement-Flag)
  - `retired_at`: Timestamp (optional)
  - `retire_reason`: String (optional, z.B. "Verkauft", "Verschlissen", "Totalschaden")
  - `barcode`: String (eindeutig, generiert bei Erstellung)
  - `qr_code_data`: String (URL zur öffentlichen Info-Page)

```sql
-- inventory_schema

CREATE TABLE equipment_types (
  equipment_type_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(tenant_id),
  category_id UUID NOT NULL REFERENCES equipment_categories(category_id),
  name VARCHAR(255) NOT NULL,
  sku VARCHAR(50) NOT NULL,
  manufacturer VARCHAR(255),
  model_number VARCHAR(255),
  description TEXT,
  unit_of_measure VARCHAR(20) DEFAULT 'PIECE', -- PIECE, METER, SET
  photo_url VARCHAR(2048),
  default_condition VARCHAR(20) DEFAULT 'NEW',
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  created_by_user_id UUID NOT NULL,
  is_active BOOLEAN DEFAULT TRUE,
  UNIQUE(tenant_id, sku)
);

CREATE TABLE equipment_items (
  equipment_item_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(tenant_id),
  equipment_type_id UUID NOT NULL REFERENCES equipment_types(equipment_type_id),
  serial_number VARCHAR(255),
  batch_id VARCHAR(255),
  purchase_date DATE,
  purchase_price_cents BIGINT,
  current_condition VARCHAR(20) DEFAULT 'GOOD',
  last_condition_check_at TIMESTAMP WITH TIME ZONE,
  location_id UUID REFERENCES warehouse_locations(location_id),
  barcode VARCHAR(255) UNIQUE NOT NULL,
  qr_code_data VARCHAR(2048),
  is_active BOOLEAN DEFAULT TRUE,
  retired_at TIMESTAMP WITH TIME ZONE,
  retire_reason VARCHAR(255),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  created_by_user_id UUID NOT NULL,
  UNIQUE(tenant_id, serial_number), -- nur nicht-null Seriennummern eindeutig
  UNIQUE(tenant_id, batch_id) -- nur bei Bulk-Items
);

CREATE INDEX idx_equipment_items_location ON equipment_items(location_id);
CREATE INDEX idx_equipment_items_barcode ON equipment_items(barcode);
CREATE INDEX idx_equipment_items_type ON equipment_items(equipment_type_id);
CREATE INDEX idx_equipment_items_condition ON equipment_items(current_condition);
```

### 1.2 Equipment-Kategorien (Hierarchie)

**Flache Hierarchie mit Parent-Kind-Beziehung:**

```
Ton
├── PA-Systeme
│   ├── Line Arrays
│   │   └── L-Acoustics K2
│   └── Standboxen
├── Mischpulte
│   ├── Kompakt
│   └── Full-Size (CL5, PM5D)
├── Mikrofone & Zubehör
│   ├── Draht-Mikrofone
│   └── Funk-Mikrofone
└── Kabel & Stecker
    ├── XLR Kabel
    ├── Multicore
    └── Speakerboxen-Kabel

Licht
├── LED Scheinwerfer
├── Moving Heads
│   ├── Spots
│   ├── Washes
│   └── Effekte (Claypaky Sharpy, Chauvet etc.)
├── Traggalgen
└── Traversen

Mechanik
├── Traversen & Riggingelemente
├── Stative & Bodenständer
└── Cases & Transport

Video
├── Projektoren
├── Leinwände
└── Videowall-Module
```

```sql
CREATE TABLE equipment_categories (
  category_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(tenant_id),
  name VARCHAR(255) NOT NULL,
  icon_name VARCHAR(50), -- 'speaker', 'lightbulb', 'camera', etc.
  parent_category_id UUID REFERENCES equipment_categories(category_id),
  hierarchy_level INT GENERATED ALWAYS AS (
    CASE
      WHEN parent_category_id IS NULL THEN 0
      ELSE 1 + (SELECT hierarchy_level FROM equipment_categories ec WHERE ec.category_id = equipment_categories.parent_category_id)
    END
  ) STORED,
  sort_order INT DEFAULT 0,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(tenant_id, name, parent_category_id)
);

-- View für Kategorie-Pfade (z.B. "Licht > Moving Heads > Spots")
CREATE VIEW category_paths AS
WITH RECURSIVE cat_tree AS (
  SELECT
    category_id,
    name,
    parent_category_id,
    name AS full_path,
    0 AS depth
  FROM equipment_categories
  WHERE parent_category_id IS NULL

  UNION ALL

  SELECT
    ec.category_id,
    ec.name,
    ec.parent_category_id,
    ct.full_path || ' > ' || ec.name,
    ct.depth + 1
  FROM equipment_categories ec
  JOIN cat_tree ct ON ec.parent_category_id = ct.category_id
)
SELECT * FROM cat_tree;
```

### 1.3 Equipment-Kombinationen (Flightcases)

Ein Flightcase ist eine logische Gruppe von Equipment-Items, die als eine Einheit transportiert/eingelagert werden. Beispiel: Ein Case mit 4x Moving Heads, 2x Traggalgen, 8x XLR-Kabel.

```sql
CREATE TABLE flight_cases (
  flight_case_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(tenant_id),
  name VARCHAR(255) NOT NULL, -- z.B. "Case: 4x Claypaky Sharpy + Galgen"
  description TEXT,
  barcode VARCHAR(255) UNIQUE NOT NULL,
  location_id UUID REFERENCES warehouse_locations(location_id),
  weight_kg DECIMAL(7,2),
  volume_liters DECIMAL(8,2),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  created_by_user_id UUID NOT NULL,
  is_active BOOLEAN DEFAULT TRUE
);

-- Mapping: Welche Items sind in welchem Case?
CREATE TABLE flight_case_items (
  flight_case_id UUID NOT NULL REFERENCES flight_cases(flight_case_id) ON DELETE CASCADE,
  equipment_item_id UUID NOT NULL REFERENCES equipment_items(equipment_item_id),
  quantity INT DEFAULT 1,
  position_in_case INT, -- Für visuelle Anordnung (Foto-Guide)
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (flight_case_id, equipment_item_id)
);
```

### 1.4 Equipment-Status-Maschine

```
                    ┌─────────────────────────────────────────┐
                    │                                         │
                    ↓                                         │
    ┌─────────────────────────────────────────────────────────────┐
    │                     Available                               │
    │            (Item ist einsatzbereit im Lager)                │
    └───┬──────────────┬──────────────┬──────────────┬──────────┬─┘
        │              │              │              │          │
        │              │              │              │          └─→ Retired
        │              │              │              │         (Ausgemustert)
        ↓              ↓              ↓              ↓
    Reserved    CheckedOut       Maintenance    Defect
    (für Job    (auf Baustelle   (in Wartung)   (Problem
     geplant)    im Einsatz)                     erkannt)
        │              │              │              │
        │              │              │              ↓
        │              └──→ CheckedIn ←─────────────┘
        │                   (von
        │                    Baustelle)
        │                     │
        └─────────────────────┤
                              ↓
                          Available
```

**State Transitions:**
- `Available → Reserved` — Equipment wird für einen zukünftigen Job reserviert
- `Reserved → CheckedOut` — Checkout-Prozess initiiert
- `CheckedOut → InUse` — Bei Scan-Bestätigung auf Baustelle
- `InUse → CheckedIn` — Return-Scan startet
- `CheckedIn → Available` — Nach Zustandsprüfung OK
- `Available → Maintenance` — Wartung geplant
- `Maintenance → Available` — Nach erfolgreicher Reparatur
- `Available → Defect` — Defekt erkannt (ReportDamage-Event)
- `Defect → Maintenance` — Reparatur angenommen
- `Available/Defect/Maintenance → Retired` — Item wird nicht mehr verwendet

```go
// Go Structs für Events

type EquipmentStatusChangedEvent struct {
  EventID       string                   `json:"event_id"`
  AggregateID   string                   `json:"aggregate_id"` // equipment_item_id
  AggregateType string                   `json:"aggregate_type"` // "equipment"
  TenantID      string                   `json:"tenant_id"`
  CorrelationID string                   `json:"correlation_id"`
  CausationID   string                   `json:"causation_id"`
  Timestamp     time.Time                `json:"timestamp"`
  Version       int64                    `json:"version"`
  SchemaVersion int                      `json:"schema_version"`
  Data          EquipmentStatusChanged   `json:"data"`
  Metadata      EventMetadata            `json:"metadata"`
}

type EquipmentStatusChanged struct {
  EquipmentItemID  string    `json:"equipment_item_id"`
  EquipmentTypeID  string    `json:"equipment_type_id"`
  PreviousStatus   string    `json:"previous_status"` // "Available", "Reserved", "CheckedOut", etc.
  NewStatus        string    `json:"new_status"`
  Reason           string    `json:"reason"`          // z.B. "checkout_requested", "damage_reported"
  ChangedByUserID  string    `json:"changed_by_user_id"`
  ChangedAt        time.Time `json:"changed_at"`
}

type DamageReportedEvent struct {
  EventID       string           `json:"event_id"`
  AggregateID   string           `json:"aggregate_id"`
  AggregateType string           `json:"aggregate_type"`
  TenantID      string           `json:"tenant_id"`
  CorrelationID string           `json:"correlation_id"`
  CausationID   string           `json:"causation_id"`
  Timestamp     time.Time        `json:"timestamp"`
  Version       int64            `json:"version"`
  SchemaVersion int              `json:"schema_version"`
  Data          DamageReported   `json:"data"`
  Metadata      EventMetadata    `json:"metadata"`
}

type DamageReported struct {
  EquipmentItemID  string    `json:"equipment_item_id"`
  SerialNumber     string    `json:"serial_number"`
  DamageSeverity   string    `json:"damage_severity"` // "MINOR", "MEDIUM", "SEVERE"
  Description      string    `json:"description"`      // Freitextnotiz
  PhotoURLs        []string  `json:"photo_urls"`       // Scans von Fotos
  ReportedByUserID string    `json:"reported_by_user_id"`
  ReportedAt       time.Time `json:"reported_at"`
  JobID            string    `json:"job_id,omitempty"` // Wenn während Job erkannt
  ProjectID        string    `json:"project_id,omitempty"`
}

type EquipmentRetiredEvent struct {
  EventID       string        `json:"event_id"`
  AggregateID   string        `json:"aggregate_id"`
  AggregateType string        `json:"aggregate_type"`
  TenantID      string        `json:"tenant_id"`
  CorrelationID string        `json:"correlation_id"`
  CausationID   string        `json:"causation_id"`
  Timestamp     time.Time     `json:"timestamp"`
  Version       int64         `json:"version"`
  SchemaVersion int           `json:"schema_version"`
  Data          EquipmentRetired `json:"data"`
  Metadata      EventMetadata `json:"metadata"`
}

type EquipmentRetired struct {
  EquipmentItemID  string    `json:"equipment_item_id"`
  RetiredAt        time.Time `json:"retired_at"`
  Reason           string    `json:"reason"`           // "SOLD", "DAMAGED", "OBSOLETE", "LOST"
  Notes            string    `json:"notes,omitempty"`
  RetiredByUserID  string    `json:"retired_by_user_id"`
}

// Generische Event-Struktur
type EventMetadata struct {
  UserID        string `json:"user_id"`
  SourceService string `json:"source_service"` // "scanner-service", "inventory-service"
  IPAddress     string `json:"ip_address,omitempty"`
}
```

**SQL für Status-Tracking:**

```sql
-- Aktuelle Zustände (materialisierte View, gepflegt über Events)
CREATE TABLE equipment_item_states (
  equipment_item_id UUID PRIMARY KEY REFERENCES equipment_items(equipment_item_id),
  tenant_id UUID NOT NULL REFERENCES tenants(tenant_id),
  current_status VARCHAR(50) NOT NULL, -- 'Available', 'Reserved', 'CheckedOut', etc.
  status_changed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  changed_by_user_id UUID,
  -- für CheckedOut Status:
  project_id UUID REFERENCES projects(project_id),
  -- für Maintenance Status:
  maintenance_reason VARCHAR(255),
  maintenance_started_at TIMESTAMP WITH TIME ZONE,
  -- für Defect Status:
  defect_description TEXT,
  defect_severity VARCHAR(20), -- 'MINOR', 'MEDIUM', 'SEVERE'
  defect_reported_at TIMESTAMP WITH TIME ZONE,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Event Log (für Audit & Replay)
CREATE TABLE equipment_status_events (
  event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  equipment_item_id UUID NOT NULL REFERENCES equipment_items(equipment_item_id),
  tenant_id UUID NOT NULL,
  previous_status VARCHAR(50),
  new_status VARCHAR(50) NOT NULL,
  transition_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  changed_by_user_id UUID NOT NULL,
  reason VARCHAR(255),
  related_project_id UUID,
  event_data JSONB,
  INDEX idx_equipment_status_events (equipment_item_id, transition_at DESC)
);
```

### 1.5 Verfügbarkeitsalgorithmus

**Input:**
- `equipment_type_id`: Welcher Gerätetyp?
- `quantity`: Wie viele Exemplare?
- `from_date`: Startdatum (Buchung)
- `to_date`: Enddatum
- `exclude_equipment_item_ids`: Liste von Items, die ignoriert werden (z.B. aktuell auf dieser Packliste)

**Output:**
- `available_quantity`: Verfügbare Menge in diesem Zeitraum
- `conflicts`: Array von Konflikten (welche Items sind blockiert, warum)
- `can_proceed`: Boolean (true wenn `available_quantity >= quantity`)

**Pseudocode:**

```
FUNCTION CheckAvailability(
  equipmentTypeID,
  quantity,
  fromDate,
  toDate,
  excludeItemIds = []
):

  // Schritt 1: Alle aktiven Items des Typs holen (außer Retired)
  items = DB.query("""
    SELECT ei.equipment_item_id, ei.current_condition, eis.current_status, p.id as project_id
    FROM equipment_items ei
    LEFT JOIN equipment_item_states eis ON ei.equipment_item_id = eis.equipment_item_id
    LEFT JOIN checkouts co ON eis.project_id = co.project_id
    LEFT JOIN projects p ON co.project_id = p.id
    WHERE ei.equipment_type_id = equipmentTypeID
      AND ei.is_active = true
      AND ei.current_condition != 'DEFECTIVE'
      AND ei.equipment_item_id NOT IN excludeItemIds
  """)

  // Schritt 2: Für jedes Item prüfen: Ist es im Zeitraum [fromDate, toDate] verfügbar?
  available = []
  conflicts = []

  FOREACH item IN items:

    // Case 1: Status = "Available"
    IF item.current_status == "Available":
      available.push(item)

    // Case 2: Status = "Reserved"
    ELSE IF item.current_status == "Reserved":
      reservation = DB.query("""
        SELECT from_date, to_date FROM reservations
        WHERE equipment_item_id = item.id
        ORDER BY created_at DESC LIMIT 1
      """)
      // Ist das Item in unserem Zeitraum reserviert?
      IF reservation.from_date <= toDate AND reservation.to_date >= fromDate:
        conflicts.push({
          equipment_item_id: item.id,
          reason: "RESERVED",
          blocked_until: reservation.to_date,
          reserved_for_project: reservation.project_id
        })
      ELSE:
        available.push(item)

    // Case 3: Status = "CheckedOut"
    ELSE IF item.current_status == "CheckedOut":
      checkout = DB.query("""
        SELECT planned_return_date FROM checkouts
        WHERE equipment_item_id = item.id
        AND status != 'COMPLETED'
        ORDER BY created_at DESC LIMIT 1
      """)
      // Ist das Item in unserem Zeitraum im Einsatz?
      IF checkout.planned_return_date >= fromDate:
        conflicts.push({
          equipment_item_id: item.id,
          reason: "IN_USE",
          blocked_until: checkout.planned_return_date,
          on_project: checkout.project_id
        })
      ELSE:
        available.push(item)

    // Case 4: Status = "Maintenance"
    ELSE IF item.current_status == "Maintenance":
      maintenance = DB.query("""
        SELECT expected_completion_date FROM maintenance_orders
        WHERE equipment_item_id = item.id
        AND status IN ('SCHEDULED', 'IN_PROGRESS')
        ORDER BY created_at DESC LIMIT 1
      """)
      IF maintenance.expected_completion_date >= fromDate:
        conflicts.push({
          equipment_item_id: item.id,
          reason: "UNDER_MAINTENANCE",
          blocked_until: maintenance.expected_completion_date
        })
      ELSE:
        available.push(item)

    // Case 5: Status = "Defect"
    // => Niemals verfügbar
    ELSE IF item.current_status == "Defect":
      conflicts.push({
        equipment_item_id: item.id,
        reason: "DEFECTIVE",
        defect_description: item.defect_description
      })

  END FOREACH

  // Schritt 3: Sub-Rental prüfen (von anderen Firmen ausleihen)
  IF available.length < quantity:
    // Prüfe: Können wir diesen Typ von einem Partner ausleihen?
    available_rental = DB.query("""
      SELECT sr.partner_id, sr.available_quantity
      FROM sub_rental_inventory sr
      WHERE sr.equipment_type_id = equipmentTypeID
        AND sr.from_date <= fromDate
        AND sr.to_date >= toDate
        AND sr.available_quantity > 0
    """)
    // Sub-Rental wird als Konflikt-Warnung angezeigt, nicht als "verfügbar"
    // Die Preisengine muss berücksichtigen dass diese teurer sind

  // Schritt 4: Doppelbuchung-Check
  pending_checkouts = DB.query("""
    SELECT pli.equipment_item_id, pli.quantity
    FROM packing_list_items pli
    JOIN packing_lists pl ON pli.packing_list_id = pl.id
    WHERE pli.equipment_type_id = equipmentTypeID
      AND pl.planned_checkout_date >= TODAY()
      AND pl.status IN ('DRAFT', 'PENDING_APPROVAL', 'APPROVED')
      AND pl.id != CURRENT_PACKING_LIST_ID -- nicht die aktuelle Liste
  """)
  // Warnung: Diese Items sind schon geplant, aber noch nicht gebucht

  RETURN {
    available_count: available.length,
    conflicts: conflicts,
    can_proceed: available.length >= quantity,
    warnings: {
      sub_rental_needed: available.length < quantity,
      pending_checkouts: pending_checkouts.length
    }
  }

END FUNCTION
```

**Go Implementation Skeleton:**

```go
type AvailabilityRequest struct {
  EquipmentTypeID    string    `json:"equipment_type_id"`
  Quantity           int       `json:"quantity"`
  FromDate           time.Time `json:"from_date"`
  ToDate             time.Time `json:"to_date"`
  ExcludeItemIDs     []string  `json:"exclude_item_ids,omitempty"`
  CheckSubRental     bool      `json:"check_sub_rental,omitempty"`
}

type AvailabilityResponse struct {
  EquipmentTypeID  string                 `json:"equipment_type_id"`
  EquipmentName    string                 `json:"equipment_name"`
  AvailableCount   int                    `json:"available_count"`
  RequestedCount   int                    `json:"requested_count"`
  CanProceed       bool                   `json:"can_proceed"`
  Conflicts        []AvailabilityConflict `json:"conflicts"`
  Warnings         AvailabilityWarnings   `json:"warnings"`
}

type AvailabilityConflict struct {
  EquipmentItemID  string    `json:"equipment_item_id"`
  Reason           string    `json:"reason"` // "RESERVED", "IN_USE", "MAINTENANCE", "DEFECTIVE"
  BlockedUntil     *time.Time `json:"blocked_until,omitempty"`
  RelatedProjectID *string    `json:"related_project_id,omitempty"`
  Description      string    `json:"description,omitempty"`
}

type AvailabilityWarnings struct {
  SubRentalNeeded     bool `json:"sub_rental_needed"`
  PendingCheckouts    int  `json:"pending_checkouts"`
  MaintenanceRequired bool `json:"maintenance_required"`
}

// Service Method
func (s *InventoryService) CheckAvailability(
  ctx context.Context,
  tenantID string,
  req AvailabilityRequest,
) (AvailabilityResponse, error) {
  // Implementierung folgt Pseudocode oben
}
```

### 1.6 Preisengine

**Miet-Preismodell:**

```sql
CREATE TABLE rental_prices (
  rental_price_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(tenant_id),
  equipment_type_id UUID NOT NULL REFERENCES equipment_types(equipment_type_id),

  -- Zeitbasierte Preise (in Cent)
  daily_price_cents BIGINT,           -- z.B. 50€ = 5000 Cent
  weekly_price_cents BIGINT,          -- 7 Tage (oft Rabatt vs. 7x daily)
  monthly_price_cents BIGINT,         -- 30 Tage

  -- Staffelung
  quantity_from INT DEFAULT 1,        -- ab dieser Menge
  quantity_discount_percent DECIMAL(5,2) DEFAULT 0, -- z.B. 10% bei 5+ Stück

  -- Versicherung
  insurance_percent DECIMAL(5,2) DEFAULT 0, -- z.B. 3% Versicherungszuschlag

  -- Gültig ab/bis
  valid_from DATE NOT NULL DEFAULT CURRENT_DATE,
  valid_to DATE,

  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  created_by_user_id UUID NOT NULL,

  UNIQUE(tenant_id, equipment_type_id, quantity_from, valid_from)
);

-- Partner-spezifische Konditionen
CREATE TABLE partner_conditions (
  condition_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(tenant_id),
  partner_id UUID NOT NULL REFERENCES partners(partner_id), -- Sub-Rental Partner
  equipment_type_id UUID NOT NULL REFERENCES equipment_types(equipment_type_id),

  -- Prozentuale Rabatte/Aufschläge
  discount_percent DECIMAL(5,2), -- negativ = Aufschlag

  -- oder absolute Preisüberride
  override_daily_price_cents BIGINT,
  override_weekly_price_cents BIGINT,
  override_monthly_price_cents BIGINT,

  valid_from DATE NOT NULL DEFAULT CURRENT_DATE,
  valid_to DATE,

  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(tenant_id, partner_id, equipment_type_id, valid_from)
);
```

**Preisberechnung für Angebote/Rechnungen:**

```go
type PricingCalculationRequest struct {
  EquipmentTypeID string    `json:"equipment_type_id"`
  Quantity        int       `json:"quantity"`
  RentalDays      int       `json:"rental_days"`     // Gesamtmietdauer
  FromDate        time.Time `json:"from_date"`
  ToDate          time.Time `json:"to_date"`
  CustomerID      *string   `json:"customer_id,omitempty"`
  PartnerID       *string   `json:"partner_id,omitempty"` // Falls Sub-Rental
  IncludeInsurance bool    `json:"include_insurance"`
  Notes           string    `json:"notes,omitempty"` // z.B. "longterm customer"
}

type PricingCalculationResponse struct {
  EquipmentTypeID      string        `json:"equipment_type_id"`
  EquipmentName        string        `json:"equipment_name"`
  Quantity             int           `json:"quantity"`
  RentalDays           int           `json:"rental_days"`

  // Preis pro Einheit (nicht Stücklistenpreis)
  UnitDailyPrice       int           `json:"unit_daily_price_cents"`

  // Berechnungen
  CalculationMethod    string        `json:"calculation_method"` // "DAILY", "WEEKLY", "MONTHLY", "HYBRID"
  BasePrice            int           `json:"base_price_cents"`
  QuantityDiscount     int           `json:"quantity_discount_cents"`
  QuantityDiscountPct  float64       `json:"quantity_discount_pct"`
  InsuranceCharge      int           `json:"insurance_charge_cents"`
  PartnerDiscount      int           `json:"partner_discount_cents"`
  PartnerDiscountPct   float64       `json:"partner_discount_pct"`

  TotalPrice           int           `json:"total_price_cents"`

  // Für Transparenz
  Breakdown            string        `json:"breakdown"` // z.B. "Woche: 7x50€ = 350€, -10% Mengenrabatt = -35€, +3% Versicherung = +9.45€ = 324.45€"
}

// Pseudocode Berechnung:

FUNCTION CalculatePrice(req PricingCalculationRequest):

  // 1. Basis-Preis laden
  basePrice = DB.query("""
    SELECT daily_price_cents, weekly_price_cents, monthly_price_cents
    FROM rental_prices
    WHERE equipment_type_id = req.equipmentTypeID
      AND quantity_from <= req.quantity
      AND (valid_to IS NULL OR valid_to >= TODAY())
    ORDER BY quantity_from DESC
    LIMIT 1
  """)

  // 2. Optimale Zeiteinheit wählen
  // Beispiel: 15 Tage
  //   Variante A: 15x daily = 15 * 50€ = 750€
  //   Variante B: 2 * weekly + 1 * daily = (2 * 300€) + 50€ = 650€ ← besser
  //   Variante C: 1 * monthly = 500€ ← bestes
  // => Rückgabe: Variante C mit Explanation

  method = OptimalPricingMethod(req.rentalDays, basePrice)

  IF method == "MONTHLY":
    months = ceil(req.rentalDays / 30)
    unitPrice = basePrice.monthly_price_cents * months

  ELSE IF method == "WEEKLY":
    weeks = ceil(req.rentalDays / 7)
    unitPrice = basePrice.weekly_price_cents * weeks

  ELSE: // DAILY
    unitPrice = basePrice.daily_price_cents * req.rentalDays

  // 3. Mit Menge multiplizieren
  baseTotal = unitPrice * req.quantity

  // 4. Mengenrabatt
  quantityDiscount = 0
  IF req.quantity >= basePrice.quantity_from:
    quantityDiscount = baseTotal * (basePrice.quantity_discount_percent / 100)

  // 5. Partner-Konditionen (falls Sub-Rental)
  partnerDiscount = 0
  IF req.partnerID != null:
    conditions = DB.query("""
      SELECT discount_percent, override_*_price_cents
      FROM partner_conditions
      WHERE partner_id = req.partnerID
        AND equipment_type_id = req.equipmentTypeID
    """)
    IF conditions.override_daily_price_cents != null:
      // Neu berechnen mit Partner-Preis
      unitPrice = conditions.override_daily_price_cents * req.rentalDays
    ELSE IF conditions.discount_percent != null:
      partnerDiscount = (baseTotal - quantityDiscount) * (conditions.discount_percent / 100)

  // 6. Versicherung
  insuranceCharge = 0
  IF req.includeInsurance:
    insuranceCharge = (baseTotal - quantityDiscount - partnerDiscount) *
                      (basePrice.insurance_percent / 100)

  // 7. Gesamt
  totalPrice = baseTotal - quantityDiscount - partnerDiscount + insuranceCharge

  RETURN {
    calculation_method: method,
    base_price: baseTotal,
    quantity_discount: quantityDiscount,
    partner_discount: partnerDiscount,
    insurance_charge: insuranceCharge,
    total_price: totalPrice,
    breakdown: HUMAN_READABLE_BREAKDOWN(...)
  }

END FUNCTION
```

**Go Implementation:**

```go
type RentalPrice struct {
  RentalPriceID           string    `db:"rental_price_id"`
  EquipmentTypeID         string    `db:"equipment_type_id"`
  DailyPriceCents         int64     `db:"daily_price_cents"`
  WeeklyPriceCents        int64     `db:"weekly_price_cents"`
  MonthlyPriceCents       int64     `db:"monthly_price_cents"`
  QuantityFrom            int       `db:"quantity_from"`
  QuantityDiscountPercent float64   `db:"quantity_discount_percent"`
  InsurancePercent        float64   `db:"insurance_percent"`
  ValidFrom               time.Time `db:"valid_from"`
  ValidTo                 *time.Time `db:"valid_to"`
}

type PartnerCondition struct {
  ConditionID              string    `db:"condition_id"`
  PartnerID                string    `db:"partner_id"`
  EquipmentTypeID          string    `db:"equipment_type_id"`
  DiscountPercent          *float64  `db:"discount_percent"`
  OverrideDailyPriceCents  *int64    `db:"override_daily_price_cents"`
  OverrideWeeklyPriceCents *int64    `db:"override_weekly_price_cents"`
  OverrideMonthlyPriceCents *int64   `db:"override_monthly_price_cents"`
  ValidFrom                time.Time `db:"valid_from"`
  ValidTo                  *time.Time `db:"valid_to"`
}

func (s *InventoryService) CalculateRentalPrice(
  ctx context.Context,
  tenantID string,
  req PricingCalculationRequest,
) (*PricingCalculationResponse, error) {
  // Implementation
}
```

---

## 2. Warehouse Domain Model

### 2.1 Hierarchische Lagerstruktur

```
Standort (Site)
├── Gebäude (Building)
│   ├── Raum (Room/Zone)
│   │   ├── Regal (Shelf/Aisle)
│   │   │   └── Fach (Bin/Compartment)
│   │   └── Regal ...
│   ├── Raum ...
├── Gebäude ...
```

**Beispiel:**
```
"MB Veranstaltungstechnik — Hauptlager"
├── "Lager-Halle 1"
│   ├── "Licht-Bereich"
│   │   ├── "Regal A (Moving Heads)"
│   │   │   ├── "Fach A1 (oben)"
│   │   │   ├── "Fach A2 (mitte)"
│   │   │   └── "Fach A3 (unten)"
│   │   └── "Regal B (PA)"
│   └── "Ton-Bereich"
│       └── "Regal C (Mixer, Stecker)"
└── "Lager-Halle 2"
    └── "Aussenlager (große Cases)"
```

```sql
CREATE TABLE warehouse_sites (
  site_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(tenant_id),
  name VARCHAR(255) NOT NULL,
  address_street VARCHAR(255),
  address_city VARCHAR(255),
  address_postal_code VARCHAR(20),
  address_country VARCHAR(255),
  phone VARCHAR(20),
  email VARCHAR(255),
  is_primary BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  created_by_user_id UUID NOT NULL
);

CREATE TABLE warehouse_buildings (
  building_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  site_id UUID NOT NULL REFERENCES warehouse_sites(site_id),
  tenant_id UUID NOT NULL REFERENCES tenants(tenant_id),
  name VARCHAR(255) NOT NULL, -- z.B. "Halle 1", "Lager Ost"
  floor_count INT DEFAULT 1,
  total_area_sqm DECIMAL(8,2),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE warehouse_rooms (
  room_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  building_id UUID NOT NULL REFERENCES warehouse_buildings(building_id),
  tenant_id UUID NOT NULL REFERENCES tenants(tenant_id),
  name VARCHAR(255) NOT NULL, -- z.B. "Licht-Bereich", "Ton-Zone"
  zone_type VARCHAR(50), -- "STORAGE", "RECEIVING", "DISPATCH", "MAINTENANCE", "CLIMATE_CONTROLLED"
  temperature_min_celsius DECIMAL(4,1),
  temperature_max_celsius DECIMAL(4,1),
  humidity_min_percent INT,
  humidity_max_percent INT,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE warehouse_shelves (
  shelf_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  room_id UUID NOT NULL REFERENCES warehouse_rooms(room_id),
  tenant_id UUID NOT NULL REFERENCES tenants(tenant_id),
  name VARCHAR(255) NOT NULL, -- z.B. "Regal A", "Regal B"
  shelf_number VARCHAR(50), -- für Sortierung: A, B, C oder 1, 2, 3
  barcode VARCHAR(255) UNIQUE NOT NULL,
  qr_code_data VARCHAR(2048),
  capacity_items INT,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE warehouse_locations (
  location_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  shelf_id UUID NOT NULL REFERENCES warehouse_shelves(shelf_id),
  tenant_id UUID NOT NULL REFERENCES tenants(tenant_id),
  name VARCHAR(100) NOT NULL, -- z.B. "Fach 1", "Fach 2", oder "Oben", "Mitte", "Unten"
  position_order INT,
  barcode VARCHAR(255) UNIQUE NOT NULL,
  qr_code_data VARCHAR(2048),
  is_active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(shelf_id, position_order)
);

-- Schnelle Wege-Berechnung (Standort-Koordinaten für Routenoptimierung)
CREATE TABLE location_coordinates (
  location_id UUID PRIMARY KEY REFERENCES warehouse_locations(location_id),
  -- Vereinfachtes 3D-Koordinaten-System:
  -- x: left-right (in Metern oder Fächer-Nummern)
  -- y: front-back (Reihenfolge der Regale)
  -- z: up-down (Höhe in Fächern, 0=oben, 1=mitte, 2=unten)
  x DECIMAL(6,2),
  y DECIMAL(6,2),
  z INT,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- View: Vollständiger Lagerplatz-Pfad
CREATE VIEW location_full_paths AS
SELECT
  wl.location_id,
  ws.name AS site_name,
  wb.name AS building_name,
  wr.name AS room_name,
  wsh.name AS shelf_name,
  wl.name AS location_name,
  CONCAT(ws.name, ' > ', wb.name, ' > ', wr.name, ' > ', wsh.name, ' > ', wl.name) AS full_path,
  lc.x, lc.y, lc.z
FROM warehouse_locations wl
JOIN warehouse_shelves wsh ON wl.shelf_id = wsh.shelf_id
JOIN warehouse_rooms wr ON wsh.room_id = wr.room_id
JOIN warehouse_buildings wb ON wr.building_id = wb.building_id
JOIN warehouse_sites ws ON wb.site_id = ws.site_id
LEFT JOIN location_coordinates lc ON wl.location_id = lc.location_id;
```

### 2.2 Warenbewegungen (Stock Movements)

**Movement-Typen:**

```
INBOUND:
  - PURCHASE: Neukauf (Warehouse empfängt Equipment von Lieferant)
  - RETURN_FROM_JOB: Rückgabe nach Job (Equipment kommt von Baustelle zurück)
  - RETURN_FROM_SUB_RENTAL: Sub-Rental Return (Equipment kommt von Partner zurück)
  - INTERNAL_TRANSFER: Umlagern zwischen Lagerplätzen (intern)

OUTBOUND:
  - CHECKOUT_FOR_JOB: Checkout für Job (Equipment verlässt Lager zur Baustelle)
  - SUB_RENTAL_OUT: Sub-Rental Out (Equipment an Partner verliehen)
  - TRANSFER_TO_PARTNER: Transfer zu anderem Standort
  - DISPOSAL: Ausschuss (Equipment wird entfernt, z.B. Verkauf, Recycling)

INTERNAL:
  - RELOCATION: Umlagern zwischen Lagerplätzen
  - RESTOCK_FROM_RECEIVING: Von Empfangszone zu Lagerfach
```

```sql
CREATE TABLE stock_movements (
  movement_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(tenant_id),
  equipment_item_id UUID NOT NULL REFERENCES equipment_items(equipment_item_id),
  movement_type VARCHAR(50) NOT NULL, -- 'PURCHASE', 'CHECKOUT_FOR_JOB', 'RETURN_FROM_JOB', etc.

  -- Quelle & Ziel
  from_location_id UUID REFERENCES warehouse_locations(location_id),
  to_location_id UUID REFERENCES warehouse_locations(location_id),

  -- Kontext
  project_id UUID REFERENCES projects(project_id),
  packing_list_id UUID REFERENCES packing_lists(packing_list_id),
  sub_rental_agreement_id UUID REFERENCES sub_rental_agreements(sub_rental_agreement_id),

  -- Zeitstempel
  movement_requested_at TIMESTAMP WITH TIME ZONE,
  movement_confirmed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  executed_by_user_id UUID NOT NULL,

  -- Notizen
  notes TEXT,

  -- Für Audit
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_equipment_item_movements (equipment_item_id),
  INDEX idx_project_movements (project_id),
  INDEX idx_movement_type (movement_type, created_at DESC),
  INDEX idx_location_movements (from_location_id, to_location_id)
);

-- Events für Event Sourcing
CREATE TABLE movement_events (
  event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  movement_id UUID NOT NULL REFERENCES stock_movements(movement_id),
  tenant_id UUID NOT NULL,
  event_type VARCHAR(100) NOT NULL, -- 'StockMovementInitiated', 'StockMovementScanned', 'StockMovementCompleted'
  equipment_item_id UUID NOT NULL,
  from_location_id UUID,
  to_location_id UUID,
  event_data JSONB NOT NULL,
  triggered_by_user_id UUID NOT NULL,
  triggered_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_movement_events (movement_id)
);
```

### 2.3 Inventur-Workflows

**Typ 1: Vollständige Inventur**
- Alle Items im System zählen
- Soll/Ist-Vergleich
- Abweichungen dokumentieren

**Typ 2: Stichproben-Inventur**
- Einzelne Regale/Zonen auswählen
- Zählen und vergleichen
- Routinen auf größere Abweichungen aufschalem

**Typ 3: Scan-basierte Inventur**
- Jedes Item scannen (Barcode/QR)
- In Echtzeit mit System abgleichen
- Sofort Unterschiede erkennen

```sql
CREATE TABLE inventory_cycles (
  inventory_cycle_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(tenant_id),
  cycle_type VARCHAR(50) NOT NULL, -- 'FULL', 'SAMPLING', 'SCAN_BASED'
  status VARCHAR(50) DEFAULT 'PLANNED', -- 'PLANNED', 'IN_PROGRESS', 'COMPLETED'

  -- Scope
  planned_scope VARCHAR(255), -- z.B. "Regal A+B", "alle Räume", "nur Licht-Zone"
  scope_location_ids UUID[] DEFAULT ARRAY[]::uuid[], -- IDs der Räume/Regale

  -- Zeiten
  planned_start_date DATE NOT NULL,
  actual_start_at TIMESTAMP WITH TIME ZONE,
  expected_completion_date DATE NOT NULL,
  actual_completion_at TIMESTAMP WITH TIME ZONE,

  -- Verantwortung
  initiated_by_user_id UUID NOT NULL,
  assigned_to_user_ids UUID[] DEFAULT ARRAY[]::uuid[],

  -- Ergebnisse
  items_counted INT DEFAULT 0,
  items_with_variance INT DEFAULT 0,

  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Inventur-Ergebnisse
CREATE TABLE inventory_results (
  result_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  inventory_cycle_id UUID NOT NULL REFERENCES inventory_cycles(inventory_cycle_id),
  tenant_id UUID NOT NULL,
  location_id UUID NOT NULL REFERENCES warehouse_locations(location_id),
  equipment_item_id UUID NOT NULL REFERENCES equipment_items(equipment_item_id),

  -- Soll/Ist
  expected_quantity INT,
  actual_quantity INT,
  variance INT GENERATED ALWAYS AS (actual_quantity - expected_quantity) STORED,
  variance_percent DECIMAL(5,2) GENERATED ALWAYS AS (
    CASE
      WHEN expected_quantity = 0 THEN NULL
      ELSE (actual_quantity - expected_quantity)::DECIMAL / expected_quantity * 100
    END
  ) STORED,

  -- Erklärung
  notes TEXT,
  photo_url VARCHAR(2048),

  recorded_by_user_id UUID NOT NULL,
  recorded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

  INDEX idx_inventory_variance (variance, inventory_cycle_id)
);

-- View: Abweichungs-Summary
CREATE VIEW inventory_variance_summary AS
SELECT
  ic.inventory_cycle_id,
  ic.cycle_type,
  COUNT(*) as total_items_checked,
  SUM(CASE WHEN ir.variance = 0 THEN 1 ELSE 0 END) as items_correct,
  SUM(CASE WHEN ir.variance != 0 THEN 1 ELSE 0 END) as items_with_variance,
  ROUND(SUM(CASE WHEN ir.variance != 0 THEN 1 ELSE 0 END)::numeric / COUNT(*) * 100, 2) as variance_rate_percent,
  SUM(ABS(ir.variance)) as total_unit_difference,
  STRING_AGG(DISTINCT CASE WHEN ir.variance > 0 THEN 'missing' WHEN ir.variance < 0 THEN 'surplus' END, ', ') as variance_types
FROM inventory_cycles ic
LEFT JOIN inventory_results ir ON ic.inventory_cycle_id = ir.inventory_cycle_id
GROUP BY ic.inventory_cycle_id, ic.cycle_type;
```

### 2.4 Lagerplatz-Optimierung

**Routenoptimierte Packlisten:**

Bei Kommissionierung für einen Job wird die Packliste nach Lagerplätzen sortiert, nicht nach Equipment-Namen. Ziel: Lisa läuft **eine Runde durchs Lager**, nicht zig-zag.

```go
// Nearest-Neighbor Algorithmus für Routenoptimierung

type LocationCoordinate struct {
  LocationID string
  X, Y, Z    float64 // 3D-Koordinaten
}

type PackingListItem struct {
  EquipmentItemID string
  EquipmentName   string
  LocationID      string
  Coordinates     LocationCoordinate
  Quantity        int
  SequenceOrder   int // Sortierung für Picking
}

// Pseudocode:
FUNCTION OptimizePackingListRoute(items []PackingListItem):

  // 1. Startpunkt: Receiving/Dispatch Area (z.B. 0,0,0)
  currentLocation = LocationCoordinate{X: 0, Y: 0, Z: 0}

  remaining = items.copy()
  optimizedSequence = []

  // 2. Greedy Nearest-Neighbor
  WHILE remaining.length > 0:

    nextItem = remaining[0]
    minDistance = INFINITY

    FOREACH item IN remaining:
      distance = EuclideanDistance(currentLocation, item.Coordinates)
      IF distance < minDistance:
        nextItem = item
        minDistance = distance

    optimizedSequence.push(nextItem)
    remaining.remove(nextItem)
    currentLocation = nextItem.Coordinates

  // 3. Sortierung aktualisieren
  FOR i, item IN optimizedSequence:
    item.SequenceOrder = i + 1

  RETURN optimizedSequence

END FUNCTION

// Alternative: Traveling Salesman Problem (TSP) Heuristic für kleine Listen
// TSP ist NP-hard, aber für 50-100 Items ist 2-Opt oder Lin-Kernighan praktikabel
```

**Heatmap — Equipment-Kombinations-Häufigkeit:**

Welche Equipment-Kombinationen werden am häufigsten zusammen gebucht? Diese sollten räumlich nah beieinander stehen.

```sql
-- Häufigkeiten von Equipment-Paaren in Packlisten
CREATE MATERIALIZED VIEW equipment_pairing_heatmap AS
WITH packing_combos AS (
  SELECT
    pli1.equipment_type_id as eq1_type_id,
    pli2.equipment_type_id as eq2_type_id,
    COUNT(*) as frequency,
    COUNT(*) FILTER (WHERE ABS(EXTRACT(EPOCH FROM (pli2.created_at - pli1.created_at))) < 86400) as same_day_frequency
  FROM packing_list_items pli1
  JOIN packing_list_items pli2 ON pli1.packing_list_id = pli2.packing_list_id
    AND pli1.equipment_type_id < pli2.equipment_type_id -- Avoid duplicates
  WHERE pli1.created_at > CURRENT_DATE - INTERVAL '90 days'
  GROUP BY pli1.equipment_type_id, pli2.equipment_type_id
)
SELECT
  eq1_type_id,
  eq2_type_id,
  frequency,
  same_day_frequency,
  ROUND(frequency::numeric / (SELECT MAX(frequency) FROM packing_combos) * 100, 0) as popularity_percent
FROM packing_combos
WHERE frequency > 5 -- Nur signifikante Paare
ORDER BY frequency DESC;
```

**Umlagerungs-Vorschläge:**

```go
type RelocationSuggestion struct {
  EquipmentTypeID      string
  EquipmentName        string
  CurrentLocationID    string
  CurrentLocationPath  string
  SuggestedLocationID  string
  SuggestedLocationPath string
  Reason               string // "HOT_ITEM", "CONSOLIDATION", "FREQUENCY_PAIRING"
  ImpactScore          float64 // 0-100: Wie viel Weg wird gespart?
}

// Logik:
// 1. "HOT_ITEM": Equipment wird oft gebucht, sollte nah an Dispatch-Area sein
// 2. "CONSOLIDATION": Items des gleichen Typs sind an mehreren Plätzen, zusammenführen
// 3. "FREQUENCY_PAIRING": Equipment wird oft mit anderem Equipment gebucht, räumlich nah stellen
```

---

## 3. Scanner Service

### 3.1 Zebra TC21 Integration

**DataWedge Profil-Konfiguration:**

Zebra-Geräte verwenden **DataWedge** (proprietary Scanning-Engine). Konfiguration ist XML-basiert und kann per MDA (Mobile Device Management) deployed werden, oder lokal auf dem Gerät.

**Profil: "RentFlow-Scanner"**

```xml
<?xml version="1.0" encoding="UTF-8"?>
<Profile>
  <Name>RentFlow-Scanner</Name>
  <ActivityList>
    <!-- Jede App, die DataWedge-Events empfangen soll -->
    <ActivityItem>
      <AppName>com.rentflow.scanner</AppName>
      <ActivityName>com.rentflow.scanner.ui.MainActivity</ActivityName>
    </ActivityItem>
  </ActivityList>

  <!-- Scanner-Parameter -->
  <PluginConfig>
    <PluginName>Barcode</PluginName>
    <OutputPlugin>
      <!-- Zwei Output-Modi: Intent und Keystroke -->
      <Type>Intent</Type>
      <IntentOutput>
        <IntentAction>com.rentflow.scanner.SCAN_RESULT</IntentAction>
        <IntentDelivery>Broadcast</IntentDelivery>
      </IntentOutput>
    </OutputPlugin>

    <PluginProps>
      <!-- Scanner-Hardware aktivieren -->
      <Property>
        <Key>scanner_input_enabled</Key>
        <Value>true</Value>
      </Property>

      <!-- Barcode-Symbologien -->
      <Property>
        <Key>decoder_code128</Key>
        <Value>true</Value>
      </Property>
      <Property>
        <Key>decoder_code39</Key>
        <Value>true</Value>
      </Property>
      <Property>
        <Key>decoder_qrcode</Key>
        <Value>true</Value>
      </Property>
      <Property>
        <Key>decoder_datamatrix</Key>
        <Value>true</Value>
      </Property>

      <!-- Performance -->
      <Property>
        <Key>scan_timeout</Key>
        <Value>30</Value> <!-- Sekunden, dann Timeout -->
      </Property>
      <Property>
        <Key>minimum_good_decode</Key>
        <Value>1</Value> <!-- Nur 1x erfolgreich scannen, dann Ergebnis -->
      </Property>

      <!-- Audio/Visuelles Feedback vom Scanner -->
      <Property>
        <Key>notification_good_scan_ogg</Key>
        <Value>/system/media/audio/ui/good_scan.ogg</Value>
      </Property>
      <Property>
        <Key>notification_bad_scan_ogg</Key>
        <Value>/system/media/audio/ui/bad_scan.ogg</Value>
      </Property>
    </PluginProps>
  </PluginConfig>

  <!-- Klawiaturtaste Mapping: Scanner-Button → Scannen starten -->
  <KeyMap>
    <Key>SCAN</Key> <!-- Physischer Scan-Button auf Zebra TC21 -->
    <Action>StartScanning</Action>
  </KeyMap>
</Profile>
```

**Wie die Web-App Scans empfängt (Android WebView):**

```typescript
// TypeScript in der Browser-App (WebView auf Zebra TC21)

interface ScanIntent {
  barcode: string;
  symbology: string; // "CODE128", "QR_CODE", etc.
  timestamp: number;
}

// 1. Intent Receiver in Android-Wrapper
window.addEventListener('datawedge-scan', (event: CustomEvent<ScanIntent>) => {
  handleScanData(event.detail);
});

// Alternative: Hidden Input-Feld mit onchange Listener
// (wenn DataWedge im Keystroke Mode konfiguriert)
const hiddenScanInput = document.getElementById('hidden-scan-input') as HTMLInputElement;
hiddenScanInput.addEventListener('change', (e) => {
  const barcode = (e.target as HTMLInputElement).value;
  const symbology = detectSymbology(barcode);

  handleScanData({
    barcode,
    symbology,
    timestamp: Date.now()
  });

  // Input zurücksetzen für nächsten Scan
  (e.target as HTMLInputElement).value = '';
});

async function handleScanData(scan: ScanIntent) {
  try {
    // Scan an Backend senden
    const response = await fetch('/scanner/process-scan', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(scan)
    });

    const result = await response.json();

    // UI-Feedback: Grün (OK) oder Rot (Fehler)
    if (result.status === 'success') {
      showGreenFeedback(result.message); // z.B. "Yamaha CL5 — OK"
      playSound('success.mp3');
    } else {
      showRedFeedback(result.error); // z.B. "Equipment nicht in dieser Packliste"
      playSound('error.mp3');
    }
  } catch (err) {
    showRedFeedback('Scanner-Fehler');
  }
}
```

**Scanner-Typen & Symbologien:**

| Typ | Format | Chars | Barcode | QR | DataMatrix |
|-----|--------|-------|---------|----|----|
| 1D Linear | Code128, Code39 | 0-9, A-Z, Sonderzeichen | ✓ | ✗ | ✗ |
| 2D QR | QR-Code | bis 4296 Zeichen | ✗ | ✓ | ✗ |
| 2D Matrix | DataMatrix | bis 2335 Zeichen | ✗ | ✗ | ✓ |

**Performance-Ziel:** Scan → API-Verarbeitung → Feedback in **<200ms** (Lisa's Erwartung: maximal 1 Sekunde).

### 3.2 Handy-Kamera (iOS/Android)

Auf der Baustelle hat Lisa ihr iPhone, keinen Scanner. Die App muss QR-Codes lesen können.

```typescript
// TypeScript/React für Camera-QR-Scanning

import jsQR from 'jsqr';

interface CameraQRScannerProps {
  onScan: (data: string) => void;
  isActive: boolean;
}

export function CameraQRScanner({ onScan, isActive }: CameraQRScannerProps) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [hasFlash, setHasFlash] = useState(false);

  useEffect(() => {
    if (!isActive) return;

    const startCamera = async () => {
      try {
        const stream = await navigator.mediaDevices.getUserMedia({
          video: {
            facingMode: 'environment', // Rückkamera
            width: { ideal: 1280 },
            height: { ideal: 720 }
          }
        });

        if (videoRef.current) {
          videoRef.current.srcObject = stream;
          videoRef.current.play();

          // Flash-Verfügbarkeit prüfen
          const track = stream.getVideoTracks()[0];
          const capabilities = (track.getCapabilities?.() || {}) as any;
          setHasFlash(capabilities.torch === true);
        }
      } catch (err) {
        console.error('Camera access denied:', err);
      }
    };

    startCamera();

    return () => {
      if (videoRef.current?.srcObject) {
        const tracks = (videoRef.current.srcObject as MediaStream).getTracks();
        tracks.forEach(track => track.stop());
      }
    };
  }, [isActive]);

  // Continuous Scanning Loop
  useEffect(() => {
    if (!isActive || !videoRef.current || !canvasRef.current) return;

    const scanInterval = setInterval(() => {
      const canvas = canvasRef.current!;
      const ctx = canvas.getContext('2d');
      const video = videoRef.current!;

      if (video.readyState === video.HAVE_ENOUGH_DATA) {
        canvas.width = video.videoWidth;
        canvas.height = video.videoHeight;

        ctx!.drawImage(video, 0, 0, canvas.width, canvas.height);
        const imageData = ctx!.getImageData(0, 0, canvas.width, canvas.height);

        // jsQR: QR-Code erkennen
        const code = jsQR(imageData.data, canvas.width, canvas.height);

        if (code && code.data) {
          onScan(code.data);
          // Optional: Vibration + Sound
          if (navigator.vibrate) navigator.vibrate(100);
          playSound('success.mp3');
        }
      }
    }, 100); // 100ms = 10 FPS ausreichend für QR-Scanning

    return () => clearInterval(scanInterval);
  }, [isActive, onScan]);

  const toggleFlash = async () => {
    if (videoRef.current?.srcObject) {
      const track = (videoRef.current.srcObject as MediaStream).getVideoTracks()[0];
      await track.applyConstraints({
        advanced: [{ torch: !hasFlash }]
      } as any);
      setHasFlash(!hasFlash);
    }
  };

  return (
    <div className="camera-scanner">
      <video ref={videoRef} style={{ width: '100%' }} playsInline={true} />
      <canvas ref={canvasRef} style={{ display: 'none' }} />
      <button onClick={toggleFlash} className="flash-button">
        {hasFlash ? '💡 Aus' : '💡 An'}
      </button>
    </div>
  );
}
```

**Library-Alternativen:**
- **jsQR** — Pure JavaScript, kleiner (1.5KB), aber nur QR
- **html5-qrcode** — Größer (40KB), unterstützt auch Code128, UPC etc.
- **ZXing.js** — Vollständig, aber sehr groß (500KB+)

**Empfehlung für RentFlow:** jsQR für QR-Codes auf dem Handy, Zebra Hardware für Lagerverwaltung.

### 3.3 USB-Barcode-Scanner (PC)

Im Lager-Büro steht ein Standard USB-Barcode-Scanner (z.B. Honeywell, Symbol). Diese arbeiten im **Keyboard Wedge Mode** — eingescannte Daten landen wie Tastatureingaben im fokussierten Input-Feld.

```typescript
// PC Web-App (z.B. Inventur-Interface auf Desktop-Browser)

interface USBScannerConfig {
  focusedInputSelector: string; // z.B. '#barcode-input'
  prefixRegex?: RegExp; // Optional: Scanner sendet Prefix
  suffixChar?: string; // Optional: Scanner sendet Suffix (meist Enter)
}

export function USBBarcodeScanner(config: USBScannerConfig) {
  const [lastScan, setLastScan] = useState<string>('');
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Prüfe ob Focus im Scanner-Input ist
      if (document.activeElement !== inputRef.current) return;

      // Wenn Scanner Enter nach Barcode sendet:
      if (e.key === 'Enter' && inputRef.current?.value) {
        const barcode = inputRef.current.value.trim();

        // Prefix/Suffix entfernen falls konfiguriert
        let cleanBarcode = barcode;
        if (config.prefixRegex) {
          cleanBarcode = barcode.replace(config.prefixRegex, '');
        }
        if (config.suffixChar) {
          cleanBarcode = cleanBarcode.replace(
            new RegExp(`${config.suffixChar}$`),
            ''
          );
        }

        setLastScan(cleanBarcode);
        inputRef.current.value = '';
        e.preventDefault();

        // Verarbeiten
        handleScan(cleanBarcode);
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [config]);

  async function handleScan(barcode: string) {
    try {
      const response = await fetch('/scanner/verify', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ barcode })
      });

      const result = await response.json();
      if (result.valid) {
        // Feedback: Equipment-Info anzeigen
      }
    } catch (err) {
      console.error('Scan verification failed');
    }
  }

  return (
    <div>
      <input
        ref={inputRef}
        type="text"
        placeholder="Barcode einlesen..."
        autoFocus
      />
      <p>Letzter Scan: {lastScan}</p>
    </div>
  );
}
```

**Typische Scanner-Konfiguration (auf dem Gerät selbst via Menü):**

```
Prefix:        [leer]
Barcode Type:  Code128 + Code39 (unterstützen)
Suffix:        Enter (0x0D)
Timing:        10ms zwischen Zeichen
```

### 3.4 RFID — Zukunft (nicht in MVP, aber Konzept)

**UHF RFID (Ultra-High Frequency):**
- Reichweite: bis 10 Meter (je nach Tag-Leistung)
- Geschwindigkeit: 50 Items in 3 Sekunden scannen (vs. 1 Item pro Scan mit Barcode)
- Reader: z.B. Zebra FX9600 (ca. 2500€)
- Tags: Passive UHF-Sticker (ca. 0,20€ pro Tag bei 10k Stückzahl)

**Use Case:** Eingang-Prüfung (Pallet mit 50 Items auspacken, alle auf einmal scannen).

```go
// RFID Reader Integration Konzept (Skeleton)

type RFIDReaderConfig struct {
  ReaderIPAddress string
  ReaderPort      int
  AntennaCount    int // 1-4 Antennen
  ReadPowerDBm    int // -10 bis +32 dBm
  Protocol        string // "LLRP" (Low Level Reader Protocol)
}

type RFIDTag struct {
  EPC            string    // Electronic Product Code (96-bit)
  RSSI           int       // Signalstärke
  Antenna        int       // Welche Antenne hat erkannt?
  FirstSeenAt    time.Time
  LastSeenAt     time.Time
  ReadCount      int
}

type RFIDScanResult struct {
  Tags           []RFIDTag
  ScanDuration   time.Duration
  SuccessRate    float64 // % erfolgreich gelesen
}

// Mapping: RFID EPC → Equipment
type RFIDEquipmentMapping struct {
  EPC             string
  EquipmentItemID string
  AssignedAt      time.Time
  AssignedByUser  string
}

func (s *ScannerService) BulkScanRFID(ctx context.Context, duration time.Duration) (*RFIDScanResult, error) {
  // 1. RFID Reader konfigurieren
  // 2. Scan starten für <duration>
  // 3. Alle Tags einsammeln
  // 4. EPC → Equipment-ID mappen
  // 5. Duplikate filtern
  return nil, nil
}
```

---

## 4. Scan-Workflows (Step-by-Step)

### 4.1 Checkout (Equipment rausgeht)

**Szenario:** Lisa packt Job "Firmengala Mittwoch" — 45 Items aus Packliste.

```
Schritt 1: Packliste öffnen
┌─────────────────────────────────┐
│ Checkout: Firmengala Mittwoch   │
│ Job-Datum: Mi, 22. Mai 2026     │
│ Fahrer: Max Müller              │
│ Status: READY_FOR_SCANNING      │
├─────────────────────────────────┤
│ ☐ 12x PAR-64 (Regal A2)         │
│ ☐ 8x Moving Head (Regal A3)     │
│ ☐ 2x CL5 Mischpult (Regal C1)   │
│ ☐ ... (38 weitere Items)        │
├─────────────────────────────────┤
│ Fortschritt: 0/45               │
│ [SCAN] Button fokussieren       │
└─────────────────────────────────┘

Schritt 2: Scan Equipment-Barcode
→ Barcode: PAR_001 (oder PAR_002, PAR_003, ... PAR_012)
→ System: Dieses Item gehört zu dieser Packliste? JA ✓
→ Zustand: Gerät OK? JA ✓
→ Feedback:
  ✓ GRÜN Screen
  🔊 Beep (ca. 500Hz, 100ms)
  ✓ "PAR-64 #1 von 12 — OK"
  ✓ Item als "gescannt" markiert
  Fortschritt: 1/45 → 2/45 → 3/45 ...

Schritt 3: Fehlerfall — Gerät nicht in Packliste
→ Barcode: CL5_SN001 (nicht auf dieser Liste)
→ Feedback:
  ✗ ROTES Screen (3 Sekunden)
  🔊 Warnton (ca. 800Hz, 200ms, 2x)
  ✗ "Yamaha CL5 — NICHT in dieser Packliste!"
  → Mitarbeiter kann trotzdem scannen ("Force-Scan"), oder zurück ins Regal
  Fortschritt: 3/45 (unverändert)

Schritt 4: Warnung — Gerät ist defekt markiert
→ Barcode: SHARPY_007
→ System: Gerät ist als "defekt" gekennzeichnet
→ Feedback:
  ⚠ GELBES Screen (2 Sekunden)
  🔊 Ping-Ton (ca. 700Hz, 150ms)
  ⚠ "Claypaky Sharpy — DEFEKT MARKIERT. Einpacken?"
  → Mit "OK" bestätigen, oder "Verweigern" → zurück ins Regal

Schritt 5: Nach letztem Item — Zusammenfassung
→ Alle 45 Items gescannt ✓
→ Feedback:
  ✓ GRÜNER Screen, 2 Sekunden
  🔊 Erfolgs-Melodie (3-Noten-Beep)
  → Automatisch zur Zusammenfassung:

┌─────────────────────────────────┐
│ Checkout abgeschlossen          │
├─────────────────────────────────┤
│ Firmengala Mittwoch             │
│ 45 Items gescannt, 0 fehlend    │
│                                 │
│ [📸 Foto] [✍️ Unterschrift]     │
│ [❌ Abbrechen] [✅ Bestätigen]   │
└─────────────────────────────────┘

Schritt 6: Digitale Unterschrift
→ Lisa unterschreibt mit Finger auf Touch-Display
→ System speichert Bild + Zeitstempel + Benutzer
→ Events triggered:
  - CheckoutStarted (erste Scan)
  - ItemScanned (je Item, 45x)
  - CheckoutCompleted (mit Unterschrift)

Schritt 7: Lieferschein drucken
→ Papierausdruck mit QR-Code (enthält Job-ID)
→ Barcode (enthält alle 45 Item-IDs)
→ Signature + Timestamp
→ Max nimmt die Palette, nimmt Lieferschein mit
```

**Go Events:**

```go
type CheckoutStartedEvent struct {
  EventID          string    `json:"event_id"`
  AggregateID      string    `json:"aggregate_id"` // packing_list_id
  AggregateType    string    `json:"aggregate_type"` // "checkout"
  TenantID         string    `json:"tenant_id"`
  CorrelationID    string    `json:"correlation_id"`
  CausationID      string    `json:"causation_id"`
  Timestamp        time.Time `json:"timestamp"`
  Version          int64     `json:"version"`
  SchemaVersion    int       `json:"schema_version"`
  Data             struct {
    CheckoutID     string    `json:"checkout_id"`
    PackingListID  string    `json:"packing_list_id"`
    ProjectID      string    `json:"project_id"`
    CheckoutByUser string    `json:"checkout_by_user"`
    DriverName     string    `json:"driver_name,omitempty"`
    PlannedItems   int       `json:"planned_items"`
    StartedAt      time.Time `json:"started_at"`
  } `json:"data"`
  Metadata EventMetadata `json:"metadata"`
}

type ItemScannedEvent struct {
  EventID           string    `json:"event_id"`
  AggregateID       string    `json:"aggregate_id"` // checkout_id
  AggregateType     string    `json:"aggregate_type"`
  TenantID          string    `json:"tenant_id"`
  CorrelationID     string    `json:"correlation_id"`
  CausationID       string    `json:"causation_id"`
  Timestamp         time.Time `json:"timestamp"`
  Version           int64     `json:"version"`
  SchemaVersion     int       `json:"schema_version"`
  Data              struct {
    EquipmentItemID string    `json:"equipment_item_id"`
    Barcode         string    `json:"barcode"`
    Equipment Name  string    `json:"equipment_name"`
    SerialNumber    string    `json:"serial_number,omitempty"`
    Condition       string    `json:"condition"` // "NEW", "GOOD", "FAIR"
    IsExpected      bool      `json:"is_expected"` // War auf Packliste?
    SequenceNumber  int       `json:"sequence_number"` // 1., 2., 3. Item gescannt?
    ScannedAt       time.Time `json:"scanned_at"`
  } `json:"data"`
  Metadata EventMetadata `json:"metadata"`
}

type CheckoutCompletedEvent struct {
  EventID              string    `json:"event_id"`
  AggregateID          string    `json:"aggregate_id"`
  AggregateType        string    `json:"aggregate_type"`
  TenantID             string    `json:"tenant_id"`
  CorrelationID        string    `json:"correlation_id"`
  CausationID          string    `json:"causation_id"`
  Timestamp            time.Time `json:"timestamp"`
  Version              int64     `json:"version"`
  SchemaVersion        int       `json:"schema_version"`
  Data                 struct {
    CheckoutID         string    `json:"checkout_id"`
    ProjectID          string    `json:"project_id"`
    ItemsScanned       int       `json:"items_scanned"`
    ItemsMissing       int       `json:"items_missing"`
    ItemsWarning       int       `json:"items_warning"`
    CompletedAt        time.Time `json:"completed_at"`
    CompletedByUser    string    `json:"completed_by_user"`
    SignatureImageURL  string    `json:"signature_image_url,omitempty"`
    DocumentURL        string    `json:"document_url"` // PDF Lieferschein
  } `json:"data"`
  Metadata EventMetadata `json:"metadata"`
}
```

**SQL für Checkout-Tracking:**

```sql
CREATE TABLE checkouts (
  checkout_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  packing_list_id UUID NOT NULL REFERENCES packing_lists(packing_list_id),
  project_id UUID NOT NULL REFERENCES projects(project_id),
  status VARCHAR(50) DEFAULT 'IN_PROGRESS', -- 'IN_PROGRESS', 'PENDING_SIGNATURE', 'COMPLETED'

  -- Timeline
  started_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  completed_at TIMESTAMP WITH TIME ZONE,

  -- Benutzer
  initiated_by_user_id UUID NOT NULL,
  completed_by_user_id UUID,

  -- Ergebnis
  items_scanned INT DEFAULT 0,
  items_expected INT NOT NULL,
  items_missing INT DEFAULT 0,
  items_with_warning INT DEFAULT 0,

  -- Unterschrift
  signature_image_url VARCHAR(2048),

  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE checkout_item_scans (
  scan_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  checkout_id UUID NOT NULL REFERENCES checkouts(checkout_id),
  equipment_item_id UUID NOT NULL REFERENCES equipment_items(equipment_item_id),
  packing_list_item_id UUID NOT NULL,

  barcode_scanned VARCHAR(255) NOT NULL,
  is_expected BOOLEAN DEFAULT TRUE,
  condition_at_scan VARCHAR(50),

  sequence_number INT,
  scanned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

  INDEX idx_checkout_items (checkout_id)
);
```

### 4.2 Check-in (Equipment kommt zurück)

```
Szenario: Job ist vorbei. LKW kommt zurück mit Equipment.

Schritt 1: Ersten Item scannen (beliebig)
→ Barcode: PAR_001
→ System: Dieses Item kommt von welchem Job?
→ Suche in der Datenbank:
  SELECT co.project_id FROM checkouts co
  WHERE co.equipment_item_id = PAR_001
    AND co.status = 'COMPLETED'
  ORDER BY completed_at DESC LIMIT 1
→ Ergebnis: Firmengala Mittwoch (Project-ID: proj_123)

Feedback:
✓ GRÜN Screen
✓ "Firmengala Mittwoch — Return erkannt. Weitere Items?"
Fortschritt: Checkin für "Firmengala Mittwoch" initiiert

Schritt 2: Zustandsprüfung des ersten Items
→ System fragt: "PAR-64 #1 — Zustand?"
→ Optionen: [✓ OK] [⚠️ Beschädigt] [✗ Defekt]
→ Lisa prüft das Gerät physisch
→ Sie wählt "OK" → Scan-Eingabefokus
→ Oder "Beschädigt" → Foto-Mode (Kamera starten)

Wenn "Beschädigt":
┌─────────────────────────────────┐
│ Beschädigung dokumentieren      │
├─────────────────────────────────┤
│ [📷 Foto 1] [+]                 │
│ [📷 Foto 2] [+]                 │
│ [📷 Foto 3] [+]                 │
│                                 │
│ Notiz:                          │
│ [Kratzer an Seite, Lüfter...]   │
│                                 │
│ Schweregrad:                    │
│ [○ Leicht] [●● Mittel] [○ Schwer]
│                                 │
│ [Zurück] [Speichern & Weiter]   │
└─────────────────────────────────┘
→ Lisa macht Fotos, beschreibt Schaden
→ Speichern → Zurück zum Scan

Schritt 3: Vergleich: Was sollte zurückkommen vs. was kommt zurück
→ System hat Packliste für Firmengala:
  12x PAR-64, 8x Moving Head, 2x CL5, ...

→ Bei jedem Scan:
  Soll: 12x PAR-64
  Ist: 1x gescannt, dann 2x, dann ... bis 12x
  Fortschritt: 12/12 ✓

→ Wenn fehlend (z.B. nur 10x):
  Soll: 12x PAR-64
  Ist: 10x gescannt
  Feedback: ✗ "2x PAR-64 FEHLEN!"
  → Liste mit fehlenden Teilen anzeigen

Schritt 4: Sub-Rental Rückgabe
→ Wenn Equipment von Partner (SoundPro) war:
→ System erkennt an Barcode dass es leihweise war
→ Zusätzliche Frage: "Condition OK?"
→ Lisa dokumentiert mit Fotos (Kratzer vorher vs. nachher)
→ Rückgabe-Bestätigung signieren

Schritt 5: Zusammenfassung vor Abschluss
┌─────────────────────────────────┐
│ Firmengala Mittwoch — Return    │
├─────────────────────────────────┤
│ Items OK: 43/45                 │
│ Items beschädigt: 2              │
│ Items fehlend: 0                │
│                                 │
│ Beschädigte Items:              │
│ • Multicore 16/4 Ch.7 — Wackler │
│ • XLR 15m #005 — gebrochener... │
│                                 │
│ ⚠️ 2 Items zur Reparatur        │
│                                 │
│ [✍️ Unterschrift] [Fertig]      │
└─────────────────────────────────┘
```

**Go Events:**

```go
type CheckinStartedEvent struct {
  EventID           string    `json:"event_id"`
  AggregateID       string    `json:"aggregate_id"` // checkin_id
  AggregateType     string    `json:"aggregate_type"`
  TenantID          string    `json:"tenant_id"`
  CorrelationID     string    `json:"correlation_id"`
  CausationID       string    `json:"causation_id"`
  Timestamp         time.Time `json:"timestamp"`
  Version           int64     `json:"version"`
  SchemaVersion     int       `json:"schema_version"`
  Data              struct {
    CheckinID       string    `json:"checkin_id"`
    AutoDetectedProject string `json:"auto_detected_project"` // Erkannt vom ersten Scan
    StartedAt       time.Time `json:"started_at"`
    InitiatedByUser string    `json:"initiated_by_user"`
  } `json:"data"`
  Metadata EventMetadata `json:"metadata"`
}

type ItemReturnedEvent struct {
  EventID           string    `json:"event_id"`
  AggregateID       string    `json:"aggregate_id"`
  AggregateType     string    `json:"aggregate_type"`
  TenantID          string    `json:"tenant_id"`
  CorrelationID     string    `json:"correlation_id"`
  CausationID       string    `json:"causation_id"`
  Timestamp         time.Time `json:"timestamp"`
  Version           int64     `json:"version"`
  SchemaVersion     int       `json:"schema_version"`
  Data              struct {
    EquipmentItemID string    `json:"equipment_item_id"`
    Barcode         string    `json:"barcode"`
    EquipmentName   string    `json:"equipment_name"`
    ProjectID       string    `json:"project_id"`
    ConditionAtReturn string  `json:"condition_at_return"` // "OK", "DAMAGED", "DEFECTIVE"
    DamagePhotos    []string  `json:"damage_photos,omitempty"`
    DamageNotes     string    `json:"damage_notes,omitempty"`
    ReturnedAt      time.Time `json:"returned_at"`
  } `json:"data"`
  Metadata EventMetadata `json:"metadata"`
}

type DamageReportedEvent struct {
  // (schon oben im Status-Model)
}

type CheckinCompletedEvent struct {
  EventID           string    `json:"event_id"`
  AggregateID       string    `json:"aggregate_id"`
  AggregateType     string    `json:"aggregate_type"`
  TenantID          string    `json:"tenant_id"`
  CorrelationID     string    `json:"correlation_id"`
  CausationID       string    `json:"causation_id"`
  Timestamp         time.Time `json:"timestamp"`
  Version           int64     `json:"version"`
  SchemaVersion     int       `json:"schema_version"`
  Data              struct {
    CheckinID         string    `json:"checkin_id"`
    ProjectID         string    `json:"project_id"`
    ItemsReturned     int       `json:"items_returned"`
    ItemsExpected     int       `json:"items_expected"`
    ItemsOK           int       `json:"items_ok"`
    ItemsDamaged      int       `json:"items_damaged"`
    ItemsMissing      int       `json:"items_missing"`
    CompletedAt       time.Time `json:"completed_at"`
    CompletedByUser   string    `json:"completed_by_user"`
    SignatureImageURL string    `json:"signature_image_url,omitempty"`
  } `json:"data"`
  Metadata EventMetadata `json:"metadata"`
}
```

### 4.3 Lagerplatz-Scan

```
Szenario: "Wo stehen die Claypaky Sharpys?"

Schritt 1: Regal-Barcode scannen
→ Barcode: SHELF_A3
→ System: Regal A3 ausgewählt

Feedback:
✓ GRÜN Screen
✓ "Regal A3 — 8 Items"
✓ Liste anzeigen:
  • Claypaky Sharpy SN:2024-0847
  • Claypaky Sharpy SN:2024-0848
  • Claypaky Sharpy SN:2024-0849
  • Claypaky Sharpy SN:2024-0850
  • Claypaky B-Eye SN:2024-0851
  • Claypaky B-Eye SN:2024-0852
  • ... (2 weitere)

Schritt 2: Einzelnes Item scannen
→ Barcode: SHARPY_2024_0847
→ System: Zeige Details

┌─────────────────────────────────┐
│ Claypaky Sharpy                 │
│ SN: 2024-0847                   │
├─────────────────────────────────┤
│ Zustand: GOOD                   │
│ Lagerplatz: Regal A3, Fach 2    │
│ Letzter Einsatz: 15. Mai (Job)  │
│ Nächster Einsatz: 22. Mai       │
│ Letzter E-Check: 01. März       │
│ Foto: [BILD]                    │
│ Notizen: "Neu, Lens kalibriert" │
└─────────────────────────────────┘

Schritt 3: Route anzeigen
→ "Regal A3, Fach 2"
→ Mit Pfeilen zum Lagerplatz navigieren (optional AR)
→ Oder: "Oben, rechts, 3. Fach von links"
```

### 4.4 Quick-Info-Scan

```
Szenario: Lisa scannt beliebiges Equipment
→ Barcode: XLR_ADAPTER_KOFFER

Sofort-Info:
┌─────────────────────────────────┐
│ XLR-Adapter-Koffer              │
│ SKU: KAB-ADI-2401               │
├─────────────────────────────────┤
│ Kategorie: Kabel & Adapter > DI │
│ Zustand: GOOD                   │
│ Lagerplatz: Regal C, Fach 3     │
│ Seriennummer/Batch: BATCH#2024- │
│                                 │
│ Letzter Einsatz:                │
│   Projekt: Festival 20.-22. April│
│   Rückgabe: 23. April           │
│                                 │
│ Nächster geplanter Einsatz:     │
│   Firmengala, 22. Mai           │
│                                 │
│ Wartung:                        │
│   Letzter E-Check: 15. Jan      │
│   Nächster Termin: 15. April ✓  │
│                                 │
│ [Foto] [Daten] [Historie]       │
└─────────────────────────────────┘
```

---

## 5. Offline-Modus (Scanner)

### 5.1 IndexedDB Schema

```typescript
// Offline Database Schema für PWA Scanner-App

// IndexedDB:
const DB_NAME = 'rentflow_scanner_cache';
const DB_VERSION = 1;

const STORES = {
  // Stammdaten (read-only, regelmäßig synced)
  'equipment_types': {
    keyPath: 'equipment_type_id',
    indexes: [
      { name: 'category_id', unique: false },
      { name: 'sku', unique: true }
    ]
  },

  'equipment_items': {
    keyPath: 'equipment_item_id',
    indexes: [
      { name: 'equipment_type_id', unique: false },
      { name: 'barcode', unique: true },
      { name: 'location_id', unique: false }
    ]
  },

  'warehouse_locations': {
    keyPath: 'location_id',
    indexes: [
      { name: 'shelf_id', unique: false },
      { name: 'full_path', unique: true }
    ]
  },

  'packing_lists': {
    keyPath: 'packing_list_id',
    indexes: [
      { name: 'project_id', unique: false },
      { name: 'planned_checkout_date', unique: false }
    ]
  },

  'packing_list_items': {
    keyPath: 'packing_list_item_id',
    indexes: [
      { name: 'packing_list_id', unique: false },
      { name: 'equipment_item_id', unique: false }
    ]
  },

  // Offline Queue (transient)
  'pending_scans': {
    keyPath: 'scan_id',
    indexes: [
      { name: 'status', unique: false },
      { name: 'created_at', unique: false }
    ]
  },

  // Sync Metadata
  'sync_metadata': {
    keyPath: 'table_name'
  }
};

// Struktur:
interface EquipmentItemOffline {
  equipment_item_id: string;
  equipment_type_id: string;
  barcode: string;
  serial_number?: string;
  current_condition: string; // LOCAL cache
  location_id?: string;
  last_sync_at: number; // Timestamp wann von Server geholt
}

interface PendingScan {
  scan_id: string;
  scan_type: 'CHECKOUT_ITEM' | 'CHECKIN_ITEM' | 'LOCATION_LOOKUP';
  barcode: string;
  event_data: any; // Volle Event-Struktur
  status: 'PENDING' | 'RETRYING' | 'FAILED';
  retry_count: number;
  last_retry_at: number;
  created_at: number;
  should_retry_after: number; // Exponential Backoff
}

interface SyncMetadata {
  table_name: string;
  last_synced_at: number; // ISO timestamp
  local_version: number; // Für Optimistic Update Detection
  is_syncing: boolean;
}
```

### 5.2 Sync-Protokoll

```typescript
// Sync-Engine für Offline-Events

interface SyncEngine {
  queuePendingScan(scan: PendingScan): Promise<void>;
  syncToServer(): Promise<SyncResult>;
  getConflicts(): Promise<SyncConflict[]>;
  resolvConflict(conflict: SyncConflict): Promise<void>;
}

type SyncResult = {
  success: boolean;
  synced_count: number;
  failed_count: number;
  conflicts: SyncConflict[];
  server_updates: ServerUpdate[]; // Neue Daten vom Server
};

type SyncConflict = {
  scan_id: string;
  conflict_type: 'EQUIPMENT_CHANGED' | 'PACKING_LIST_UPDATED' | 'NETWORK_RETRY';
  local_version: any;
  server_version: any;
  resolution: 'ACCEPT_LOCAL' | 'ACCEPT_SERVER' | 'MANUAL';
};

// Pseudocode:
FUNCTION SyncToServer(session: OfflineSession):

  offlineDrafts = DB.pending_scans.where(status in ['PENDING', 'RETRYING']).sort_by(created_at)

  synced = 0
  failed = 0
  conflicts = []

  FOR EACH draft IN offlineDrafts:

    TRY:
      // Chronologische Reihenfolge beibehalten!
      response = await POST /scanner/sync {
        scan_event: draft.event_data,
        idempotency_key: draft.scan_id, // Verhindert Duplikate
        expected_local_version: draft.local_version
      }

      IF response.status == 200:
        // Erfolgreich
        draft.status = 'SYNCED'
        synced++

        // Lokale Daten mit Server-Response aktualisieren
        applyServerUpdates(response.server_updates)

      ELSE IF response.status == 409:
        // Conflict: Equipment oder Packliste hat sich geändert
        conflicts.push({
          scan_id: draft.scan_id,
          conflict_type: response.conflict_type,
          local_version: draft.local_version,
          server_version: response.server_version
        })

      ELSE:
        throw new NetworkError()

    CATCH NetworkError:
      draft.retry_count++
      draft.status = 'RETRYING'

      // Exponential Backoff: 1s, 2s, 4s, 8s, 16s, 32s, 60s (max)
      wait_seconds = MIN(2 ^ retry_count, 60)
      draft.should_retry_after = NOW() + wait_seconds

      failed++

  END FOR

  RETURN {
    success: failed == 0,
    synced_count: synced,
    failed_count: failed,
    conflicts: conflicts
  }

END FUNCTION

// Konflikt-Auflösung Beispiel:
// "Equipment wurde offline gescannt, aber auf dem Server inzwischen zu Maintenance markiert"
FUNCTION ResolveConflict(conflict: SyncConflict, resolution: string):

  scan = DB.pending_scans.get(conflict.scan_id)

  IF resolution == 'ACCEPT_LOCAL':
    // Nehme meinen lokalen Scan als gültig an
    // Force-Push zum Server mit Override-Flag
    await POST /scanner/sync {
      scan_event: scan.event_data,
      force_override: true,
      idempotency_key: scan.scan_id
    }

  ELSE IF resolution == 'ACCEPT_SERVER':
    // Verwerfe meinen lokalen Scan, nimm Server-Version
    DB.pending_scans.delete(scan.scan_id)

    // Aktualisiere lokale Daten vom Server
    applyServerUpdates(conflict.server_version)

  ELSE IF resolution == 'MANUAL':
    // User muss entscheiden (UI zeigt beide Versionen)
    // ...

  END IF

END FUNCTION
```

### 5.3 Service Worker — Caching-Strategie

```typescript
// Service Worker für PWA

const CACHE_VERSION = 'v1';
const CACHE_NAME = `rentflow-scanner-${CACHE_VERSION}`;

const STATIC_ASSETS = [
  '/index.html',
  '/styles.css',
  '/app.js',
  '/manifest.json',
  '/icons/logo-192.png',
  '/icons/logo-512.png'
];

const API_ENDPOINTS = [
  /^\/api\/equipment\//,
  /^\/api\/scanner\//,
  /^\/api\/packing-lists\//
];

// Installation
self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => {
      return cache.addAll(STATIC_ASSETS);
    })
  );
});

// Fetch-Handler
self.addEventListener('fetch', (event) => {
  const url = new URL(event.request.url);

  // 1. Statische Assets: Cache-First
  if (STATIC_ASSETS.some(asset => url.pathname.endsWith(asset))) {
    event.respondWith(
      caches.match(event.request).then((cached) => {
        return cached || fetch(event.request);
      })
    );
  }

  // 2. API-Calls: Network-First (aber mit Offline-Fallback)
  else if (API_ENDPOINTS.some(pattern => pattern.test(url.pathname))) {
    event.respondWith(
      fetch(event.request).then((response) => {
        // Erfolgreiche Response cachen
        if (response.status === 200) {
          const cached = response.clone();
          caches.open(CACHE_NAME).then((cache) => {
            cache.put(event.request, cached);
          });
        }
        return response;
      }).catch(() => {
        // Offline: Cache-Fallback für GET-Requests
        if (event.request.method === 'GET') {
          return caches.match(event.request).then((cached) => {
            return cached || new Response(
              JSON.stringify({ error: 'offline' }),
              { status: 503, headers: { 'Content-Type': 'application/json' } }
            );
          });
        }
        // POST/PUT offline: Queue für später
        return new Response(
          JSON.stringify({ error: 'offline_queued' }),
          { status: 202, headers: { 'Content-Type': 'application/json' } }
        );
      })
    );
  }

  // 3. Andere (z.B. CSS): Stale-While-Revalidate
  else {
    event.respondWith(
      caches.match(event.request).then((cached) => {
        const fetching = fetch(event.request).then((response) => {
          if (response.status === 200) {
            caches.open(CACHE_NAME).then((cache) => {
              cache.put(event.request, response.clone());
            });
          }
          return response;
        });

        return cached || fetching;
      })
    );
  }
});
```

### 5.4 Offline-Indikator

```typescript
// UI-Komponente für Offline-Status

export function OfflineIndicator() {
  const [isOnline, setIsOnline] = useState(navigator.onLine);
  const [syncStatus, setSyncStatus] = useState<'idle' | 'syncing' | 'error'>('idle');
  const [pendingCount, setPendingCount] = useState(0);

  useEffect(() => {
    const handleOnline = () => {
      setIsOnline(true);
      syncToServer(); // Auto-sync wenn online
    };

    const handleOffline = () => setIsOnline(false);

    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);

    return () => {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
    };
  }, []);

  if (isOnline && syncStatus === 'idle' && pendingCount === 0) {
    return null; // Nichts anzeigen
  }

  return (
    <div className={`offline-banner ${isOnline ? 'online' : 'offline'}`}>
      {!isOnline && (
        <>
          <span className="icon">📡</span>
          <span className="text">Offline-Modus — Daten werden automatisch synchronisiert</span>
        </>
      )}

      {isOnline && syncStatus === 'syncing' && (
        <>
          <span className="spinner">⏳</span>
          <span className="text">Synchronisiere {pendingCount} Scans...</span>
        </>
      )}

      {syncStatus === 'error' && (
        <>
          <span className="icon">⚠️</span>
          <span className="text">{pendingCount} Scans ausstehend — Synchronisierung fehlgeschlagen</span>
          <button onClick={syncToServer}>Erneut versuchen</button>
        </>
      )}
    </div>
  );
}

// CSS
.offline-banner {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 12px 16px;
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 14px;
  z-index: 1000;
  box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.1);
}

.offline-banner.offline {
  background-color: #ffd700; /* Gelb */
  color: #000;
}

.offline-banner.online.syncing {
  background-color: #87ceeb; /* Hellblau */
  color: #000;
}

.offline-banner.error {
  background-color: #ff6b6b; /* Rot */
  color: #fff;
}
```

---

## 6. Label/QR-Code Generierung

### 6.1 QR-Code Content

```
QR-Code Zielformat: URL zur öffentlichen Info-Page

Beispiel:
  https://rentflow.example.com/public/equipment/550e8400-e29b-41d4-a716-446655440000

Content:
  - 36 Zeichen (UUID)
  - Max 154 Byte für QR-Code Standard (Level M, bis zu 2953 numerisch)
  - QR-Version: V3-V5 ausreichend
  - Error Correction: M (7% Fehler-Toleranz)
```

### 6.2 ZPL-Label Template (Zebra Drucker)

```zebra
^XA
^MMT
^PW812
^LL1218

~JSN
^LT
^MNW
^MTT
^PON
^PMN
^LH0,0
^JMA

^PRC
^PAI

~SD15

^FO50,100
^A0N,70,70
^FD$EQUIPMENT_NAME^FS

^FO50,200
^BY4,3,100
^BQN,0,6
^FDHTTPS://rentflow.example.com/PUBLIC/$EQUIPMENT_ID^FS

^FO50,350
^A0N,30,30
^FDSKU: $SKU^FS

^FO50,400
^A0N,30,30
^FDSN: $SERIAL_NUMBER^FS

^FO50,450
^A0N,30,30
^FDKat: $CATEGORY^FS

^XZ
```

**Variablen:**
- `$EQUIPMENT_NAME` — z.B. "Claypaky Sharpy"
- `$EQUIPMENT_ID` — UUID (für öffentliche Info-Page)
- `$SKU` — z.B. "CLY-SHA-2024"
- `$SERIAL_NUMBER` — z.B. "2024-0847" (falls vorhanden)
- `$CATEGORY` — z.B. "Licht > Moving Heads"

**Label-Größe:** 4"x6" (Standard, passt auf Zebra ZD420T Drucker)

### 6.3 Öffentliche Info-Page

```
GET /public/equipment/{equipment_id}

Antwort (HTML/JSON):

{
  "equipment": {
    "equipment_item_id": "550e8400-e29b-41d4-a716-446655440000",
    "equipment_type": "Claypaky Sharpy",
    "serial_number": "2024-0847",
    "category": "Licht > Moving Heads",
    "photo_url": "https://cdn.example.com/equipment/sharpy-0847.jpg",
    "current_condition": "GOOD",
    "warehouse_location": "Halle 1 > Licht-Bereich > Regal A > Fach 3",
    "last_condition_check": "2026-03-15T14:00:00Z",
    "technical_specs": {
      "power_consumption": "320W",
      "output_lumens": "32000",
      "beam_angle": "2.5-15 degrees"
    },
    "rental_price": {
      "daily_cents": 50000,
      "weekly_cents": 300000,
      "monthly_cents": 500000
    },
    "last_projects": [
      { "project": "Festival 20.-22. April", "returned": "2026-04-23" },
      { "project": "Firmengala 15. April", "returned": "2026-04-16" }
    ]
  },
  "can_be_rented": true,
  "next_available_date": "2026-03-25"
}
```

**Öffentliche HTML-Seite (Mobile-optimiert):**

```html
<!DOCTYPE html>
<html lang="de">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Claypaky Sharpy - RentFlow</title>
  <style>
    body { font-family: sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; }
    .header { background: #222; color: #fff; padding: 20px; border-radius: 8px; }
    .photo { width: 100%; max-width: 400px; margin: 20px 0; border-radius: 8px; }
    .specs { background: #f5f5f5; padding: 15px; margin: 10px 0; border-radius: 8px; }
    .availability { color: green; font-weight: bold; }
    .unavailable { color: red; }
  </style>
</head>
<body>
  <div class="header">
    <h1>Claypaky Sharpy</h1>
    <p>SN: 2024-0847 | Status: ✓ Einsatzbereit</p>
  </div>

  <img src="..." alt="Claypaky Sharpy" class="photo">

  <div class="specs">
    <h3>Spezifikationen</h3>
    <ul>
      <li>Power: 320W</li>
      <li>Lumen: 32000</li>
      <li>Beam Angle: 2.5-15°</li>
    </ul>
  </div>

  <div class="specs">
    <h3>Verfügbarkeit</h3>
    <p class="availability">Ab sofort verfügbar</p>
  </div>

  <div class="specs">
    <h3>Zuletzt eingesetzt</h3>
    <ul>
      <li>Festival 20.-22. April (Rückgabe: 23.4.)</li>
      <li>Firmengala 15. April (Rückgabe: 16.4.)</li>
    </ul>
  </div>

  <p style="text-align: center; color: #999; font-size: 12px;">
    Kontakt: info@mb-veranstaltungstechnik.de | RentFlow v1.0
  </p>
</body>
</html>
```

---

## 7. Backend Service Endpoints (Scanner-Service)

```yaml
# scanner-service API

POST /scanner/process-scan
  Input:
    barcode: string
    symbology: "CODE128" | "QR_CODE" | "DATAMATRIX"
    timestamp: number (ms)
    context: "CHECKOUT" | "CHECKIN" | "INVENTORY" | "LOCATION_LOOKUP" | "QUICK_INFO"
    packing_list_id?: string (wenn Checkout/Checkin)

  Output (on success):
    status: "ok"
    equipment_item_id: string
    equipment_name: string
    equipment_type_id: string
    serial_number?: string
    current_condition: string
    is_expected: boolean (falls context=CHECKOUT)
    location_full_path?: string
    feedback_type: "success" | "warning" | "error"
    feedback_message: string
    can_proceed: boolean

  Output (on error):
    status: "error"
    error_code: string
    message: string
    suggestions: string[] (Hilfetext)

POST /scanner/checkout/start
  Input:
    packing_list_id: string
    tenant_id: string
    initiated_by_user_id: string

  Output:
    checkout_id: string
    items_expected: int
    packing_list_items: {
      packing_list_item_id: string
      equipment_item_id: string
      equipment_name: string
      barcode: string
      quantity: int
    }[]

POST /scanner/checkout/scan-item
  Input:
    checkout_id: string
    barcode: string
    condition: "GOOD" | "FAIR" | "DAMAGED"

  Output:
    item_id: string
    is_expected: boolean
    sequence: number (Wievieltes Item?)
    items_scanned: int
    items_expected: int
    progress: float (0-100%)

POST /scanner/checkout/complete
  Input:
    checkout_id: string
    signature_image_url: string
    completed_by_user_id: string

  Output:
    status: "completed"
    summary: {
      items_scanned: int
      items_missing: int
      items_with_warning: int
      document_url: string (PDF Lieferschein)
    }

POST /scanner/checkin/start
  Input:
    first_scan_barcode: string
    tenant_id: string
    initiated_by_user_id: string

  Output:
    checkin_id: string
    auto_detected_project_id: string
    project_name: string
    items_expected: int

POST /scanner/checkin/scan-item
  Input:
    checkin_id: string
    barcode: string
    condition: "OK" | "DAMAGED" | "DEFECTIVE"
    damage_photos?: string[] (URLs)
    damage_notes?: string

  Output:
    item_id: string
    returned_successfully: boolean
    items_returned: int
    items_expected: int

GET /scanner/equipment/{barcode}
  Input: barcode (string)

  Output:
    equipment_item_id: string
    equipment_name: string
    serial_number?: string
    category: string
    current_condition: string
    current_location: {
      location_id: string
      full_path: string
    }
    last_checkout?: {
      project_id: string
      project_name: string
      checkout_date: date
      planned_return: date
    }
    last_condition_check: date
    technical_notes: string

GET /public/equipment/{equipment_id}
  Output: (siehe 6.3)

POST /scanner/sync
  Input (für Offline-Modus):
    scan_event: object (Event-Struktur)
    idempotency_key: string
    expected_local_version: int

  Output:
    status: 200 | 409 (Conflict)
    server_updates?: object (neue Daten vom Server)
    conflict_type?: string
```

---

## 8. Fehlerbehandlung & Validierung

```go
// Error Codes

const (
  ErrBarcodeNotFound           = "BARCODE_NOT_FOUND"
  ErrBarcodeAmbiguous          = "BARCODE_AMBIGUOUS" // Mehrere Items mit gleicher Nummer?
  ErrEquipmentNotInPackingList = "NOT_IN_PACKING_LIST"
  ErrEquipmentDefective        = "EQUIPMENT_DEFECTIVE"
  ErrEquipmentInMaintenance    = "EQUIPMENT_IN_MAINTENANCE"
  ErrWrongProjectCheckin       = "WRONG_PROJECT"
  ErrNetworkError              = "NETWORK_ERROR"
  ErrDatabaseError             = "DATABASE_ERROR"
  ErrUnauthorized              = "UNAUTHORIZED"
)

// Validation Rules

func ValidateScan(barcode string) error {
  // 1. Format
  if len(barcode) == 0 || len(barcode) > 255 {
    return errors.New(ErrBarcodeNotFound)
  }

  // 2. Datenbankabfrage (mit Caching)
  item := db.FindEquipmentByBarcode(barcode)
  if item == nil {
    return errors.New(ErrBarcodeNotFound)
  }

  // 3. Status-Checks (je nach Context)
  if item.CurrentCondition == "DEFECTIVE" {
    return errors.New(ErrEquipmentDefective)
  }

  // 4. Duplikate prüfen (sollte nicht passieren, aber sichern)
  if db.CountEquipmentByBarcode(barcode) > 1 {
    return errors.New(ErrBarcodeAmbiguous)
  }

  return nil
}
```

---

## 9. Zusammenfassung: Implementierungs-Roadmap

1. **Phase 1 (MVP):**
   - [ ] Inventory Domain Model (Equipment-Typen, Items, Status-Maschine)
   - [ ] Warehouse Hierarchie (Standort → Gebäude → Raum → Regal → Fach)
   - [ ] Zebra TC21 Scan-Integration (DataWedge Intent Output)
   - [ ] Checkout-Workflow (Scan → Bestätigung → Unterschrift)
   - [ ] Check-in-Workflow (Auto-Job-Detection, Zustandsprüfung)
   - [ ] IndexedDB Offline-Modus (für Lager-WLAN)
   - [ ] QR-Code Label-Generierung

2. **Phase 2 (Post-MVP):**
   - [ ] Verfügbarkeitsalgorithmus (mit Doppelbuchungs-Check)
   - [ ] Preisengine (Miet-Preise, Partner-Konditionen)
   - [ ] Lagerplatz-Optimierung (Nearest-Neighbor, Heatmap)
   - [ ] Handy-Kamera QR-Scanning (jsQR library)
   - [ ] Inventur-Workflows (Vollständig, Stichprobe, Scan-basiert)

3. **Phase 3 (Erweitert):**
   - [ ] RFID-Reader Integration
   - [ ] Mobile App für Baustelle (iOS/Android)
   - [ ] Sub-Rental Federation (mTLS P2P mit Partnern)
   - [ ] Erweiterte Reporting/Analytics

---

**Stand:** 21. März 2026
**Nächste Aktion:** Implementation Scanner Service Backend (Go) + Frontend (React TypeScript)

