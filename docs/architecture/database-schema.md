# Database Schema - Veranstaltungstechnik Lagerverwaltung & Rechnungssoftware

## 1. Übersicht und Design-Principles

### Design-Entscheidungen
- **UUID Primary Keys**: Alle Tabellen nutzen UUID (v4) als Primärschlüssel für bessere Verteilbarkeit und Security
- **Audit Trail**: `created_at`, `updated_at`, `deleted_at` in allen Tabellen für Nachverfolgbarkeit
- **Soft Deletes**: `deleted_at` Spalte für DSGVO-Compliance und Daten-Recovery
- **DECIMAL für Geldbeträge**: `DECIMAL(19,2)` überall statt FLOAT (Vermeidung von Floating-Point-Fehlern)
- **JSONB für Metadaten**: Flexible Datenstruktur für Equipment-Spezifikationen, Custom-Felder
- **Normalisierung**: 3NF Minimum, denormalisiert wo nötig für Performance (z.B. Packlisten)
- **Availability Calculation**: Through Views/Queries bei Bedarf, nicht in der DB gespeichert

### Enum-Typen
```sql
CREATE TYPE equipment_condition AS ENUM ('operational', 'damaged', 'maintenance', 'decommissioned');
CREATE TYPE project_status AS ENUM ('draft', 'active', 'archived', 'completed', 'cancelled');
CREATE TYPE offer_status AS ENUM ('draft', 'sent', 'accepted', 'rejected', 'invoiced');
CREATE TYPE invoice_status AS ENUM ('draft', 'sent', 'paid', 'partially_paid', 'overdue', 'cancelled');
CREATE TYPE invoice_type AS ENUM ('invoice', 'credit_note', 'reversal');
CREATE TYPE transaction_type AS ENUM ('income', 'expense', 'sub_rental_in', 'sub_rental_out', 'payment');
CREATE TYPE user_role AS ENUM ('admin', 'project_manager', 'warehouse', 'accounting', 'freelancer', 'custom');
CREATE TYPE sub_rental_status AS ENUM ('requested', 'quoted', 'confirmed', 'active', 'returned', 'cancelled');
CREATE TYPE storage_location_type AS ENUM ('site', 'room', 'shelf', 'compartment');
CREATE TYPE equipment_type AS ENUM ('single', 'bulk', 'combination');
CREATE TYPE audit_action AS ENUM ('create', 'update', 'delete', 'restore', 'export', 'merge');
```

---

## 2. Kern-Tabellen

### 2.1 Users & Roles

```sql
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email VARCHAR(255) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  first_name VARCHAR(100),
  last_name VARCHAR(100),
  role user_role NOT NULL DEFAULT 'custom',
  is_active BOOLEAN NOT NULL DEFAULT true,
  last_login TIMESTAMP WITH TIME ZONE,
  phone VARCHAR(20),
  metadata JSONB DEFAULT '{}', -- Custom Felder wie Adresse, Bankdaten für Freelancer
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_is_active ON users(is_active);
CREATE INDEX idx_users_role ON users(role);

CREATE TABLE roles (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(100) NOT NULL UNIQUE,
  description TEXT,
  is_system_role BOOLEAN NOT NULL DEFAULT false,
  permissions JSONB NOT NULL DEFAULT '{}', -- {module: [read, write, delete], ...}
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE user_roles (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(user_id, role_id)
);

CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
```

---

### 2.2 Lagerplätze (Hierarchisch)

```sql
CREATE TABLE storage_locations (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  parent_id UUID REFERENCES storage_locations(id) ON DELETE CASCADE,
  name VARCHAR(255) NOT NULL,
  type storage_location_type NOT NULL, -- site, room, shelf, compartment
  description TEXT,
  barcode VARCHAR(100) UNIQUE,
  qr_code VARCHAR(500), -- Data für QR-Code
  capacity_unit VARCHAR(50), -- "pcs", "m³", "kg", NULL wenn unbegrenzt
  capacity_value DECIMAL(10,2),
  is_active BOOLEAN NOT NULL DEFAULT true,
  metadata JSONB DEFAULT '{}', -- Koordinaten, Temperatur-Anforderungen, etc.
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_storage_locations_parent_id ON storage_locations(parent_id);
CREATE INDEX idx_storage_locations_type ON storage_locations(type);
CREATE INDEX idx_storage_locations_barcode ON storage_locations(barcode) WHERE deleted_at IS NULL;
CREATE INDEX idx_storage_locations_is_active ON storage_locations(is_active);

-- Materialized View für Performance: Gesamter Pfad vom Root zur Location
CREATE TABLE storage_location_hierarchy (
  location_id UUID PRIMARY KEY REFERENCES storage_locations(id) ON DELETE CASCADE,
  path TEXT[] NOT NULL, -- Array der IDs vom Root zu dieser Location
  depth INTEGER NOT NULL,
  root_id UUID NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_storage_hierarchy_root_id ON storage_location_hierarchy(root_id);
CREATE INDEX idx_storage_hierarchy_depth ON storage_location_hierarchy(depth);
```

---

### 2.3 Equipment-Katalog & Inventar

```sql
CREATE TABLE equipment_types_catalog (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(255) NOT NULL UNIQUE,
  category VARCHAR(100) NOT NULL, -- "lighting", "sound", "video", "structural", "cable", etc.
  description TEXT,
  manufacturer VARCHAR(255),
  model VARCHAR(255),
  specifications JSONB DEFAULT '{}', -- {weight, dimensions, power_consumption, connectors, etc.}
  icon_path VARCHAR(500),
  replacement_value DECIMAL(19,2), -- Für Versicherung/Schadensersatz
  depreciation_years INTEGER,
  is_consumable BOOLEAN NOT NULL DEFAULT false, -- Verschleißteil
  metadata JSONB DEFAULT '{}',
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_equipment_types_category ON equipment_types_catalog(category);
CREATE INDEX idx_equipment_types_name ON equipment_types_catalog(name);

CREATE TABLE equipment (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  equipment_type_id UUID NOT NULL REFERENCES equipment_types_catalog(id),
  type equipment_type NOT NULL, -- single, bulk, combination
  name VARCHAR(255) NOT NULL,
  barcode VARCHAR(100) UNIQUE,
  qr_code VARCHAR(500),
  serial_number VARCHAR(100), -- Nur für teure Geräte relevant
  serial_tracked BOOLEAN NOT NULL DEFAULT false, -- true wenn Seriennummer wichtig ist
  quantity_unit VARCHAR(50), -- "pcs", "m", "kg", NULL wenn single-item
  quantity_available DECIMAL(10,2) NOT NULL DEFAULT 1, -- Verfügbare Menge (bulk)
  quantity_total DECIMAL(10,2) NOT NULL DEFAULT 1, -- Gesamte Menge im Lager
  condition equipment_condition NOT NULL DEFAULT 'operational',
  primary_storage_location_id UUID REFERENCES storage_locations(id) ON DELETE SET NULL,
  purchase_date DATE,
  purchase_cost DECIMAL(19,2),
  current_value DECIMAL(19,2), -- Nach Depreciation
  operating_hours INTEGER, -- Für Wartung/Tracking
  last_maintenance_date DATE,
  next_maintenance_date DATE,
  is_active BOOLEAN NOT NULL DEFAULT true,
  metadata JSONB DEFAULT '{}', -- Farbe, Besonderheiten, Notizen
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_equipment_barcode ON equipment(barcode) WHERE deleted_at IS NULL;
CREATE INDEX idx_equipment_serial_number ON equipment(serial_number) WHERE serial_tracked = true;
CREATE INDEX idx_equipment_type_id ON equipment(equipment_type_id);
CREATE INDEX idx_equipment_primary_storage ON equipment(primary_storage_location_id);
CREATE INDEX idx_equipment_condition ON equipment(condition) WHERE condition != 'operational';
CREATE INDEX idx_equipment_is_active ON equipment(is_active);

-- Equipment-Kombinationen: z.B. Flightcase mit Inhalt
CREATE TABLE equipment_combinations (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  parent_equipment_id UUID NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
  child_equipment_id UUID NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
  quantity DECIMAL(10,2) NOT NULL DEFAULT 1,
  sequence_order INTEGER, -- Reihenfolge im Flightcase
  description TEXT, -- "3x Scheinwerfer im Case"
  is_optional BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(parent_equipment_id, child_equipment_id)
);

CREATE INDEX idx_equipment_combinations_parent ON equipment_combinations(parent_equipment_id);
CREATE INDEX idx_equipment_combinations_child ON equipment_combinations(child_equipment_id);

-- Defekt- und Zustandsdokumentation
CREATE TABLE equipment_damage_reports (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  equipment_id UUID NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
  reported_by_user_id UUID NOT NULL REFERENCES users(id),
  damage_date DATE NOT NULL,
  description TEXT NOT NULL,
  severity_level INTEGER CHECK (severity_level >= 1 AND severity_level <= 5), -- 1-5 Skala
  estimated_repair_cost DECIMAL(19,2),
  actual_repair_cost DECIMAL(19,2),
  repair_status VARCHAR(50), -- "reported", "in_repair", "repaired", "damaged_beyond_repair"
  photo_paths VARCHAR(500)[] DEFAULT '{}', -- Array von Pfaden
  resolved_date DATE,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_damage_reports_equipment_id ON equipment_damage_reports(equipment_id);
CREATE INDEX idx_damage_reports_status ON equipment_damage_reports(repair_status);
```

---

### 2.4 Projekte

```sql
CREATE TABLE projects (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(255) NOT NULL,
  description TEXT,
  client_name VARCHAR(255),
  project_number VARCHAR(100) UNIQUE,
  status project_status NOT NULL DEFAULT 'draft',
  start_date DATE NOT NULL,
  end_date DATE NOT NULL,
  location TEXT,
  project_manager_id UUID REFERENCES users(id) ON DELETE SET NULL,
  notes TEXT,
  metadata JSONB DEFAULT '{}', -- Event-Details, Besonderheiten
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_projects_status ON projects(status);
CREATE INDEX idx_projects_project_manager_id ON projects(project_manager_id);
CREATE INDEX idx_projects_start_date ON projects(start_date);
CREATE INDEX idx_projects_end_date ON projects(end_date);
CREATE INDEX idx_projects_project_number ON projects(project_number) WHERE deleted_at IS NULL;

-- Equipment-Zuweisung zu Projekten (Reservierung)
CREATE TABLE project_equipment (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  equipment_id UUID NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
  quantity DECIMAL(10,2) NOT NULL DEFAULT 1,
  allocated_from_date DATE NOT NULL,
  allocated_until_date DATE NOT NULL,
  notes TEXT,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(project_id, equipment_id)
);

CREATE INDEX idx_project_equipment_project_id ON project_equipment(project_id);
CREATE INDEX idx_project_equipment_equipment_id ON project_equipment(equipment_id);
CREATE INDEX idx_project_equipment_dates ON project_equipment(allocated_from_date, allocated_until_date);

-- Freelancer auf Projekten
CREATE TABLE project_freelancers (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  freelancer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role VARCHAR(100), -- "sound_tech", "lighting_tech", "setup_crew"
  hourly_rate DECIMAL(19,2),
  assigned_from_date DATE NOT NULL,
  assigned_until_date DATE NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(project_id, freelancer_id)
);

CREATE INDEX idx_project_freelancers_project_id ON project_freelancers(project_id);
CREATE INDEX idx_project_freelancers_freelancer_id ON project_freelancers(freelancer_id);

-- Packlisten
CREATE TABLE packing_lists (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  is_master_list BOOLEAN NOT NULL DEFAULT true, -- Gibt es mehrere Packlisten pro Projekt?
  created_by_user_id UUID NOT NULL REFERENCES users(id),
  sequence_order INTEGER,
  metadata JSONB DEFAULT '{}', -- Routing-Info, Fahrzeug-Zuweisung
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_packing_lists_project_id ON packing_lists(project_id);

-- Packlisten-Positionen
CREATE TABLE packing_list_items (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  packing_list_id UUID NOT NULL REFERENCES packing_lists(id) ON DELETE CASCADE,
  equipment_id UUID NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
  quantity DECIMAL(10,2) NOT NULL,
  picked_quantity DECIMAL(10,2) DEFAULT 0, -- Tracking beim Abhaken
  is_picked BOOLEAN NOT NULL DEFAULT false,
  picked_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  picked_at TIMESTAMP WITH TIME ZONE,
  sequence_order INTEGER,
  notes TEXT,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_packing_list_items_packing_list ON packing_list_items(packing_list_id);
CREATE INDEX idx_packing_list_items_equipment_id ON packing_list_items(equipment_id);
CREATE INDEX idx_packing_list_items_is_picked ON packing_list_items(is_picked);

-- Zeiterfassung für Freelancer
CREATE TABLE time_entries (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  freelancer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  date_worked DATE NOT NULL,
  hours_worked DECIMAL(10,2) NOT NULL,
  hourly_rate DECIMAL(19,2),
  description TEXT,
  verified_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  is_verified BOOLEAN NOT NULL DEFAULT false,
  verified_at TIMESTAMP WITH TIME ZONE,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_time_entries_project_id ON time_entries(project_id);
CREATE INDEX idx_time_entries_freelancer_id ON time_entries(freelancer_id);
CREATE INDEX idx_time_entries_date_worked ON time_entries(date_worked);
CREATE INDEX idx_time_entries_is_verified ON time_entries(is_verified);
```

---

### 2.5 Angebote & Rechnungen

```sql
CREATE TABLE offers (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
  offer_number VARCHAR(100) NOT NULL UNIQUE,
  client_name VARCHAR(255) NOT NULL,
  client_email VARCHAR(255),
  client_address TEXT,
  status offer_status NOT NULL DEFAULT 'draft',
  creation_date DATE NOT NULL DEFAULT CURRENT_DATE,
  valid_until DATE,
  total_net DECIMAL(19,2) NOT NULL DEFAULT 0,
  vat_rate DECIMAL(5,2) NOT NULL DEFAULT 19.00, -- Prozentsatz
  total_vat DECIMAL(19,2) NOT NULL DEFAULT 0,
  total_gross DECIMAL(19,2) NOT NULL DEFAULT 0,
  currency VARCHAR(3) NOT NULL DEFAULT 'EUR',
  notes TEXT,
  created_by_user_id UUID NOT NULL REFERENCES users(id),
  sent_date DATE,
  acceptance_date DATE,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_offers_project_id ON offers(project_id);
CREATE INDEX idx_offers_offer_number ON offers(offer_number);
CREATE INDEX idx_offers_status ON offers(status);
CREATE INDEX idx_offers_creation_date ON offers(creation_date);

CREATE TABLE offer_items (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  offer_id UUID NOT NULL REFERENCES offers(id) ON DELETE CASCADE,
  description VARCHAR(500) NOT NULL,
  quantity DECIMAL(10,2) NOT NULL,
  unit_price DECIMAL(19,2) NOT NULL,
  position INTEGER NOT NULL,
  equipment_id UUID REFERENCES equipment(id) ON DELETE SET NULL, -- Optional, kann auch service sein
  metadata JSONB DEFAULT '{}',
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_offer_items_offer_id ON offer_items(offer_id);

CREATE TABLE invoices (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  invoice_number VARCHAR(100) NOT NULL UNIQUE,
  offer_id UUID REFERENCES offers(id) ON DELETE SET NULL, -- Aus Angebot erzeugt
  project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
  type invoice_type NOT NULL DEFAULT 'invoice', -- invoice, credit_note, reversal
  related_invoice_id UUID REFERENCES invoices(id) ON DELETE SET NULL, -- Für Gutschriften/Stornos
  client_name VARCHAR(255) NOT NULL,
  client_email VARCHAR(255),
  client_address TEXT,
  status invoice_status NOT NULL DEFAULT 'draft',
  invoice_date DATE NOT NULL DEFAULT CURRENT_DATE,
  due_date DATE NOT NULL,
  payment_date DATE,
  total_net DECIMAL(19,2) NOT NULL DEFAULT 0,
  vat_rate DECIMAL(5,2) NOT NULL DEFAULT 19.00,
  total_vat DECIMAL(19,2) NOT NULL DEFAULT 0,
  total_gross DECIMAL(19,2) NOT NULL DEFAULT 0,
  amount_paid DECIMAL(19,2) NOT NULL DEFAULT 0,
  currency VARCHAR(3) NOT NULL DEFAULT 'EUR',
  payment_method VARCHAR(50), -- "bank_transfer", "cash", "check", "credit_card"
  payment_reference VARCHAR(255),
  notes TEXT,
  reminder_count INTEGER NOT NULL DEFAULT 0,
  last_reminder_date DATE,
  created_by_user_id UUID NOT NULL REFERENCES users(id),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_invoices_invoice_number ON invoices(invoice_number);
CREATE INDEX idx_invoices_offer_id ON invoices(offer_id);
CREATE INDEX idx_invoices_project_id ON invoices(project_id);
CREATE INDEX idx_invoices_status ON invoices(status);
CREATE INDEX idx_invoices_invoice_date ON invoices(invoice_date);
CREATE INDEX idx_invoices_due_date ON invoices(due_date);
CREATE INDEX idx_invoices_client_name ON invoices(client_name);

CREATE TABLE invoice_items (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
  description VARCHAR(500) NOT NULL,
  quantity DECIMAL(10,2) NOT NULL,
  unit_price DECIMAL(19,2) NOT NULL,
  position INTEGER NOT NULL,
  equipment_id UUID REFERENCES equipment(id) ON DELETE SET NULL,
  project_equipment_id UUID REFERENCES project_equipment(id) ON DELETE SET NULL,
  time_entry_id UUID REFERENCES time_entries(id) ON DELETE SET NULL,
  metadata JSONB DEFAULT '{}',
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_invoice_items_invoice_id ON invoice_items(invoice_id);
CREATE INDEX idx_invoice_items_equipment_id ON invoice_items(equipment_id);

-- Teilrechnungen (z.B. 30% Anzahlung, 50% bei Lieferung, 20% am Ende)
CREATE TABLE invoice_installments (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
  installment_number INTEGER NOT NULL,
  percentage DECIMAL(5,2) NOT NULL, -- z.B. 30.00 für 30%
  amount_net DECIMAL(19,2) NOT NULL,
  amount_vat DECIMAL(19,2) NOT NULL,
  amount_gross DECIMAL(19,2) NOT NULL,
  due_date DATE NOT NULL,
  payment_date DATE,
  status invoice_status NOT NULL DEFAULT 'sent',
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(invoice_id, installment_number)
);

CREATE INDEX idx_invoice_installments_invoice_id ON invoice_installments(invoice_id);
CREATE INDEX idx_invoice_installments_due_date ON invoice_installments(due_date);

-- Mahnungen
CREATE TABLE reminders (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
  reminder_level INTEGER NOT NULL DEFAULT 1, -- 1st, 2nd, 3rd reminder
  reminder_date DATE NOT NULL DEFAULT CURRENT_DATE,
  next_reminder_date DATE,
  sent_date DATE,
  reminder_amount DECIMAL(19,2),
  reminder_fee DECIMAL(19,2),
  status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, sent, payment_received
  created_by_user_id UUID NOT NULL REFERENCES users(id),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_reminders_invoice_id ON reminders(invoice_id);
CREATE INDEX idx_reminders_reminder_date ON reminders(reminder_date);
CREATE INDEX idx_reminders_status ON reminders(status);
```

---

### 2.6 Sub-Rental (Equipment-Vermietung von/an Partner)

```sql
CREATE TABLE sub_rental_requests (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
  requesting_from_partner_id UUID REFERENCES federation_partners(id) ON DELETE SET NULL,
  request_date DATE NOT NULL DEFAULT CURRENT_DATE,
  needed_from_date DATE NOT NULL,
  needed_until_date DATE NOT NULL,
  status sub_rental_status NOT NULL DEFAULT 'requested',
  equipment_specifications JSONB NOT NULL, -- Was wird benötigt?
  estimated_value DECIMAL(19,2),
  confirmed_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  notes TEXT,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_sub_rental_requests_project_id ON sub_rental_requests(project_id);
CREATE INDEX idx_sub_rental_requests_partner_id ON sub_rental_requests(requesting_from_partner_id);
CREATE INDEX idx_sub_rental_requests_status ON sub_rental_requests(status);

CREATE TABLE sub_rental_equipments (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  request_id UUID NOT NULL REFERENCES sub_rental_requests(id) ON DELETE CASCADE,
  equipment_id UUID NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
  equipment_name VARCHAR(255), -- Fallback wenn Equipment nicht lokal bekannt
  partner_equipment_id VARCHAR(255), -- ID im Partner-System
  quantity DECIMAL(10,2) NOT NULL,
  daily_rental_rate DECIMAL(19,2),
  notes TEXT,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sub_rental_equipments_request_id ON sub_rental_equipments(request_id);

CREATE TABLE sub_rental_conditions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  partner_id UUID NOT NULL REFERENCES federation_partners(id) ON DELETE CASCADE,
  rental_rate_per_day DECIMAL(19,2),
  rental_rate_per_week DECIMAL(19,2),
  rental_rate_per_month DECIMAL(19,2),
  damage_deposit_percentage DECIMAL(5,2), -- z.B. 10% vom Wert
  cancellation_fee_percentage DECIMAL(5,2),
  minimum_rental_period INTEGER, -- In Tagen
  equipment_categories VARCHAR(100)[], -- Welche Equipment-Kategorien?
  notes TEXT,
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_sub_rental_conditions_partner_id ON sub_rental_conditions(partner_id);

-- Zustandsdokumentation bei Übergabe
CREATE TABLE sub_rental_handover_reports (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  request_id UUID NOT NULL REFERENCES sub_rental_requests(id) ON DELETE CASCADE,
  report_type VARCHAR(50) NOT NULL, -- "pickup", "return"
  report_date DATE NOT NULL,
  reported_by_user_id UUID NOT NULL REFERENCES users(id),
  partner_contact_name VARCHAR(255),
  equipment_condition JSONB NOT NULL, -- {equipment_id: {condition, photos, notes}, ...}
  overall_notes TEXT,
  photo_paths VARCHAR(500)[] DEFAULT '{}',
  signature_paths VARCHAR(500)[] DEFAULT '{}',
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_handover_reports_request_id ON sub_rental_handover_reports(request_id);
```

---

### 2.7 Federation (Peer-to-Peer Partner-Verbindungen)

```sql
CREATE TABLE federation_partners (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(255) NOT NULL UNIQUE,
  company_name VARCHAR(255),
  website VARCHAR(500),
  contact_email VARCHAR(255),
  contact_phone VARCHAR(20),
  contact_person VARCHAR(255),
  api_key VARCHAR(500), -- Für API-Zugriff
  api_key_hash VARCHAR(255), -- Gehashed für DB-Sicherheit
  public_key TEXT, -- RSA Public Key für API-Requests
  is_active BOOLEAN NOT NULL DEFAULT true,
  is_trusted BOOLEAN NOT NULL DEFAULT false, -- Automatisches Quoting?
  shared_equipment_categories VARCHAR(100)[], -- Welche Kategorien teilen?
  notes TEXT,
  last_sync_date TIMESTAMP WITH TIME ZONE,
  metadata JSONB DEFAULT '{}',
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_federation_partners_is_active ON federation_partners(is_active);
CREATE INDEX idx_federation_partners_name ON federation_partners(name);
CREATE INDEX idx_federation_partners_is_trusted ON federation_partners(is_trusted);

-- Synchronisierte Partner-Equipment (Read-Only lokale Kopie)
CREATE TABLE federated_equipment (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  partner_id UUID NOT NULL REFERENCES federation_partners(id) ON DELETE CASCADE,
  partner_equipment_id VARCHAR(255) NOT NULL,
  name VARCHAR(255) NOT NULL,
  type equipment_type NOT NULL,
  category VARCHAR(100),
  description TEXT,
  quantity_available DECIMAL(10,2),
  quantity_total DECIMAL(10,2),
  daily_rental_rate DECIMAL(19,2),
  specifications JSONB DEFAULT '{}',
  last_synced TIMESTAMP WITH TIME ZONE,
  is_available BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(partner_id, partner_equipment_id)
);

CREATE INDEX idx_federated_equipment_partner_id ON federated_equipment(partner_id);
CREATE INDEX idx_federated_equipment_category ON federated_equipment(category);
CREATE INDEX idx_federated_equipment_is_available ON federated_equipment(is_available);

-- Federation API-Anfragen-Log (für Debugging und Audit)
CREATE TABLE federation_api_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  partner_id UUID REFERENCES federation_partners(id) ON DELETE SET NULL,
  endpoint VARCHAR(255) NOT NULL,
  method VARCHAR(10) NOT NULL, -- GET, POST, etc.
  request_body JSONB,
  response_code INTEGER,
  response_body JSONB,
  error_message TEXT,
  ip_address INET,
  user_agent VARCHAR(500),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_federation_api_logs_partner_id ON federation_api_logs(partner_id);
CREATE INDEX idx_federation_api_logs_created_at ON federation_api_logs(created_at);
```

---

### 2.8 DATEV Export & Finanzen

```sql
CREATE TABLE datev_exports (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  export_date DATE NOT NULL DEFAULT CURRENT_DATE,
  export_period_from DATE NOT NULL,
  export_period_until DATE NOT NULL,
  invoice_count INTEGER NOT NULL DEFAULT 0,
  total_amount DECIMAL(19,2) NOT NULL DEFAULT 0,
  export_file_path VARCHAR(500),
  export_status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, exported, imported_in_datev, failed
  error_message TEXT,
  created_by_user_id UUID NOT NULL REFERENCES users(id),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_datev_exports_export_date ON datev_exports(export_date);
CREATE INDEX idx_datev_exports_status ON datev_exports(export_status);

-- Eingangsrechnungen (von Lieferanten)
CREATE TABLE incoming_invoices (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  invoice_number VARCHAR(100) NOT NULL,
  vendor_name VARCHAR(255) NOT NULL,
  vendor_contact VARCHAR(255),
  invoice_date DATE NOT NULL,
  due_date DATE NOT NULL,
  receipt_date DATE,
  total_net DECIMAL(19,2) NOT NULL,
  vat_rate DECIMAL(5,2) NOT NULL DEFAULT 19.00,
  total_vat DECIMAL(19,2) NOT NULL,
  total_gross DECIMAL(19,2) NOT NULL,
  payment_date DATE,
  payment_method VARCHAR(50),
  payment_reference VARCHAR(255),
  document_path VARCHAR(500), -- Scan/PDF
  status VARCHAR(50) NOT NULL DEFAULT 'received', -- received, reviewed, booked, paid
  notes TEXT,
  created_by_user_id UUID NOT NULL REFERENCES users(id),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_incoming_invoices_vendor_name ON incoming_invoices(vendor_name);
CREATE INDEX idx_incoming_invoices_invoice_date ON incoming_invoices(invoice_date);
CREATE INDEX idx_incoming_invoices_status ON incoming_invoices(status);

-- Kostenstellen (optional, für detaillierte Kostenrechnung)
CREATE TABLE cost_centers (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code VARCHAR(20) NOT NULL UNIQUE,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_cost_centers_code ON cost_centers(code);
```

---

### 2.9 Audit-Log & Sicherheit

```sql
CREATE TABLE audit_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  action audit_action NOT NULL,
  entity_type VARCHAR(100) NOT NULL, -- "invoice", "equipment", "project", etc.
  entity_id UUID NOT NULL,
  old_values JSONB, -- Nur bei Update/Delete
  new_values JSONB, -- Bei Create/Update
  change_description TEXT,
  ip_address INET,
  user_agent VARCHAR(500),
  status VARCHAR(50) NOT NULL DEFAULT 'success', -- success, failure
  error_message TEXT,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_entity_type ON audit_logs(entity_type);
CREATE INDEX idx_audit_logs_entity_id ON audit_logs(entity_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);

-- Session-Management
CREATE TABLE sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash VARCHAR(255) NOT NULL UNIQUE,
  ip_address INET,
  user_agent VARCHAR(500),
  last_activity TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX idx_sessions_is_active ON sessions(is_active);

-- API-Keys für externe Integrations
CREATE TABLE api_keys (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  key_hash VARCHAR(255) NOT NULL UNIQUE,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  last_used_at TIMESTAMP WITH TIME ZONE,
  expires_at TIMESTAMP WITH TIME ZONE,
  is_active BOOLEAN NOT NULL DEFAULT true,
  permissions JSONB NOT NULL DEFAULT '{}', -- {module: [read, write], ...}
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX idx_api_keys_is_active ON api_keys(is_active);
```

---

### 2.10 Scanner-Service Verknüpfung

```sql
CREATE TABLE scanner_sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  device_id VARCHAR(255), -- Eindeutige Device-ID
  packing_list_id UUID REFERENCES packing_lists(id) ON DELETE SET NULL,
  session_type VARCHAR(50) NOT NULL, -- "packing", "return", "inventory"
  status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, closed, paused
  started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  ended_at TIMESTAMP WITH TIME ZONE,
  items_scanned INTEGER NOT NULL DEFAULT 0,
  metadata JSONB DEFAULT '{}',
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_scanner_sessions_user_id ON scanner_sessions(user_id);
CREATE INDEX idx_scanner_sessions_packing_list_id ON scanner_sessions(packing_list_id);
CREATE INDEX idx_scanner_sessions_status ON scanner_sessions(status);

CREATE TABLE scanner_scans (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id UUID NOT NULL REFERENCES scanner_sessions(id) ON DELETE CASCADE,
  barcode VARCHAR(100) NOT NULL,
  equipment_id UUID REFERENCES equipment(id) ON DELETE SET NULL,
  storage_location_id UUID REFERENCES storage_locations(id) ON DELETE SET NULL,
  quantity DECIMAL(10,2),
  is_valid BOOLEAN NOT NULL DEFAULT true,
  error_message TEXT,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_scanner_scans_session_id ON scanner_scans(session_id);
CREATE INDEX idx_scanner_scans_equipment_id ON scanner_scans(equipment_id);
CREATE INDEX idx_scanner_scans_created_at ON scanner_scans(created_at);
```

---

### 2.11 System & Konfiguration

```sql
CREATE TABLE system_config (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  config_key VARCHAR(255) NOT NULL UNIQUE,
  config_value TEXT,
  value_type VARCHAR(50) NOT NULL, -- string, integer, boolean, json
  description TEXT,
  is_editable BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_system_config_config_key ON system_config(config_key);

-- Backups & Maintenance
CREATE TABLE backups (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  backup_type VARCHAR(50) NOT NULL, -- full, incremental, export
  backup_path VARCHAR(500),
  file_size_bytes BIGINT,
  backup_status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, completed, failed
  error_message TEXT,
  backup_duration_seconds INTEGER,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_backups_created_at ON backups(created_at);
CREATE INDEX idx_backups_backup_status ON backups(backup_status);
```

---

## 3. Verfügbarkeitsverwaltung (Views & Berechnungen)

```sql
-- Materialisierte View: Equipment-Verfügbarkeit pro Tag
CREATE MATERIALIZED VIEW equipment_availability AS
SELECT
  e.id AS equipment_id,
  e.name,
  e.quantity_available,
  COALESCE(pe.projected_allocated_quantity, 0) AS allocated_quantity,
  (e.quantity_available - COALESCE(pe.projected_allocated_quantity, 0)) AS available_quantity,
  CURRENT_DATE AS calculation_date
FROM equipment e
LEFT JOIN (
  SELECT
    equipment_id,
    SUM(quantity) AS projected_allocated_quantity
  FROM project_equipment
  WHERE
    allocated_from_date <= CURRENT_DATE
    AND allocated_until_date >= CURRENT_DATE
  GROUP BY equipment_id
) pe ON e.id = pe.equipment_id
WHERE e.deleted_at IS NULL AND e.is_active = true;

CREATE INDEX idx_equipment_availability_equipment_id ON equipment_availability(equipment_id);

-- View: Offene Posten (Rechnungen)
CREATE VIEW open_invoices AS
SELECT
  i.id,
  i.invoice_number,
  i.client_name,
  i.total_gross,
  i.amount_paid,
  (i.total_gross - i.amount_paid) AS amount_open,
  i.due_date,
  i.status,
  CASE
    WHEN i.due_date < CURRENT_DATE AND i.status NOT IN ('paid', 'cancelled') THEN 'overdue'
    WHEN (CURRENT_DATE - i.due_date) > 30 AND i.status NOT IN ('paid', 'cancelled') THEN 'very_overdue'
    ELSE 'normal'
  END AS aging_status
FROM invoices i
WHERE
  i.deleted_at IS NULL
  AND (i.total_gross - i.amount_paid) > 0;

-- View: Projekt-Equipment-Zusammenfassung
CREATE VIEW project_equipment_summary AS
SELECT
  p.id AS project_id,
  p.name AS project_name,
  COUNT(DISTINCT pe.equipment_id) AS equipment_count,
  SUM(pe.quantity * e.current_value) AS total_equipment_value,
  MIN(p.start_date) AS earliest_date,
  MAX(p.end_date) AS latest_date
FROM projects p
LEFT JOIN project_equipment pe ON p.id = pe.project_id
LEFT JOIN equipment e ON pe.equipment_id = e.id
WHERE p.deleted_at IS NULL
GROUP BY p.id, p.name;
```

---

## 4. Indizes - Zusammenfassung und Begründung

| Tabelle | Index | Spalten | Grund |
|---------|-------|---------|-------|
| users | idx_users_email | email | Login, Unique Constraint |
| users | idx_users_is_active | is_active | Filtering aktive Benutzer |
| storage_locations | idx_storage_locations_barcode | barcode | Scanner-Lookup |
| storage_locations | idx_storage_locations_type | type | UI-Filter (Raum/Regal/etc.) |
| equipment | idx_equipment_barcode | barcode | Scanner-Lookup |
| equipment | idx_equipment_serial_number | serial_number | Seriennummern-Suche |
| equipment | idx_equipment_type_id | equipment_type_id | Katalog-Joins |
| equipment | idx_equipment_condition | condition | Defekte filtern |
| project_equipment | idx_project_equipment_dates | allocated_from_date, allocated_until_date | Verfügbarkeits-Berechnung |
| invoices | idx_invoices_status | status | Finanz-Dashboards |
| invoices | idx_invoices_due_date | due_date | Mahnung-Berechnung |
| packing_list_items | idx_packing_list_items_is_picked | is_picked | Packlisten-Status |
| invoice_installments | idx_invoice_installments_due_date | due_date | Zahlung-Planung |
| reminders | idx_reminders_status | status | Reminder-Versand |
| federated_equipment | idx_federated_equipment_is_available | is_available | Sub-Rental-Anfragen |

---

## 5. Constraints & Referenzen

### Foreign Keys mit Cascade-Strategien:

| Relation | ON DELETE | Grund |
|----------|-----------|-------|
| project_equipment → projects | CASCADE | Projekt löschen → Equipment-Zuordnungen weg |
| project_equipment → equipment | CASCADE | Sollte nicht vorkommen (Equipment archivieren) |
| invoice_items → invoices | CASCADE | Rechnung löschen → Positionen weg |
| packing_list_items → packing_lists | CASCADE | Packliste löschen → Positionen weg |
| sub_rental_equipments → sub_rental_requests | CASCADE | Anfrage löschen → Equipment-Details weg |
| federation_api_logs → federation_partners | SET NULL | Partner-Löschung → Logs bleiben für Audit |
| audit_logs → users | SET NULL | Benutzer-Löschung → Logs erhalten |

---

## 6. Migrations-Strategie

### Phase 1: Basis-Schema
```
- User-Management (users, roles, user_roles)
- Storage-Locations (Hierarchie)
- Equipment-Katalog und Inventar
```

### Phase 2: Projekt-Management
```
- Projects, project_equipment, project_freelancers
- Packing-Lists
- Time-Entries
```

### Phase 3: Finanzen
```
- Offers, Invoices, Invoice-Items
- Invoice-Installments
- Reminders, DATEV-Export
```

### Phase 4: Sub-Rental & Federation
```
- Sub-Rental Requests und Conditions
- Federation-Partners
- Federated-Equipment, Federation-API-Logs
```

### Phase 5: Audit & System
```
- Audit-Logs, Sessions, API-Keys
- System-Config, Backups
- Scanner-Sessions und Scans
```

### Migrations durchführen (Go):
```go
// Verwendung von Migrations-Tools wie golang-migrate/migrate
// oder gorm AutoMigrate mit Versionierung

type MigrationFile struct {
  Version   int64
  Name      string
  UpSQL     string
  DownSQL   string
}

// Beispiel: 001_create_users.up.sql
```

---

## 7. ER-Diagramm (ASCII-Art)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          USERS & AUTHENTICATION                             │
├─────────────────────────────────────────────────────────────────────────────┤
│
│  ┌──────────────┐         ┌──────────────┐         ┌──────────────┐
│  │    users     │         │    roles     │         │ user_roles   │
│  ├──────────────┤         ├──────────────┤         ├──────────────┤
│  │ id (PK)      │         │ id (PK)      │         │ id (PK)      │
│  │ email (UNIQ) │         │ name (UNIQ)  │         │ user_id (FK) │
│  │ role         │◄────────│ permissions  │         │ role_id (FK) │
│  │ is_active    │         │              │         │              │
│  └──────────────┘         └──────────────┘         └──────────────┘
│
├─────────────────────────────────────────────────────────────────────────────┤
│                        LAGER & EQUIPMENT                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│
│  ┌──────────────────────┐      ┌────────────────┐
│  │ storage_locations    │      │ equipment_types│
│  ├──────────────────────┤      │   _catalog     │
│  │ id (PK)              │      ├────────────────┤
│  │ parent_id (FK)───┐   │      │ id (PK)        │
│  │ name             │   │      │ category       │
│  │ type             │   │      │ specifications │
│  │ capacity         │   │      │ replacement_val│
│  └──────────────────────┘      └────────────────┘
│         ▲                               │
│         │ parent_id                     │ equipment_type_id
│         │                               ▼
│  ┌──────────────────────────────────────────┐
│  │          equipment                       │
│  ├──────────────────────────────────────────┤
│  │ id (PK)                                  │
│  │ barcode                                  │
│  │ serial_number (wenn serial_tracked=true) │
│  │ quantity_available, quantity_total       │
│  │ condition                                │
│  │ primary_storage_location_id (FK)         │
│  │ purchase_cost, current_value             │
│  └──────────────────────────────────────────┘
│         │                            │
│         │ parent_equipment_id        │ child_equipment_id
│         │                            ▼
│         └───────────►┌──────────────────────┐
│                      │equipment_combinations│
│                      ├──────────────────────┤
│                      │ id (PK)              │
│                      │ quantity             │
│                      │ sequence_order       │
│                      └──────────────────────┘
│
│  ┌──────────────────────┐
│  │equipment_damage      │
│  │  _reports            │
│  ├──────────────────────┤
│  │ id (PK)              │
│  │ equipment_id (FK)    │
│  │ reported_by_user_id  │
│  │ severity_level       │
│  │ repair_status        │
│  │ photo_paths[]        │
│  └──────────────────────┘
│
├─────────────────────────────────────────────────────────────────────────────┤
│                          PROJECTS & JOBS                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│
│  ┌──────────────┐
│  │   projects   │
│  ├──────────────┤
│  │ id (PK)      │
│  │ name         │
│  │ status       │
│  │ start_date   │
│  │ end_date     │
│  │ pm_id (FK→u) │
│  └──────────────┘
│         │
│  ┌──────┴──────┬──────────┬─────────────────┐
│  │             │          │                 │
│  ▼             ▼          ▼                 ▼
│ ┌─────────────────┐  ┌──────────────┐  ┌──────────────┐
│ │ project_        │  │ project_     │  │ packing_     │
│ │ equipment       │  │ freelancers  │  │ lists        │
│ ├─────────────────┤  ├──────────────┤  ├──────────────┤
│ │ id (PK)         │  │ id (PK)      │  │ id (PK)      │
│ │ project_id (FK) │  │ freelancer_id│  │ project_id   │
│ │ equipment_id(FK)│  │ hourly_rate  │  │ name         │
│ │ quantity        │  │ dates        │  └──────────────┘
│ │ allocated_dates │  └──────────────┘         │
│ └─────────────────┘                           │ packing_list_id
│                                               ▼
│                                    ┌──────────────────────┐
│                                    │ packing_list_items   │
│                                    ├──────────────────────┤
│                                    │ id (PK)              │
│                                    │ equipment_id (FK)    │
│                                    │ quantity             │
│                                    │ is_picked            │
│                                    │ picked_by_user_id    │
│                                    └──────────────────────┘
│
│  ┌──────────────────────┐
│  │   time_entries       │
│  ├──────────────────────┤
│  │ id (PK)              │
│  │ project_id (FK)      │
│  │ freelancer_id (FK)   │
│  │ date_worked          │
│  │ hours_worked         │
│  │ hourly_rate          │
│  └──────────────────────┘
│
├─────────────────────────────────────────────────────────────────────────────┤
│                      OFFERS & INVOICES (Finanzen)                           │
├─────────────────────────────────────────────────────────────────────────────┤
│
│  ┌──────────────┐
│  │    offers    │
│  ├──────────────┤
│  │ id (PK)      │
│  │ offer_number │
│  │ project_id   │
│  │ status       │
│  │ total_*      │
│  │ vat_rate     │
│  └──────────────┘
│         │ offer_id
│         ▼
│  ┌──────────────┐        ┌──────────────┐
│  │ offer_items  │        │   invoices   │
│  ├──────────────┤        ├──────────────┤
│  │ id (PK)      │        │ id (PK)      │
│  │ description  │        │ invoice_no   │
│  │ quantity     │◄───────│ offer_id(FK) │
│  │ unit_price   │        │ type         │
│  └──────────────┘        │ status       │
│                          │ total_*      │
│                          │ amount_paid  │
│                          │ due_date     │
│                          └──────────────┘
│                                 │
│                    ┌────────────┼────────────┐
│                    │            │            │
│                    ▼            ▼            ▼
│           ┌──────────────┐ ┌──────────────┐
│           │invoice_items │ │  reminders   │
│           ├──────────────┤ ├──────────────┤
│           │ description  │ │ level        │
│           │ quantity     │ │ status       │
│           │ unit_price   │ │ sent_date    │
│           │              │ │ fee          │
│           └──────────────┘ └──────────────┘
│
│  ┌──────────────────────┐
│  │invoice_installments  │
│  ├──────────────────────┤
│  │ invoice_id (FK)      │
│  │ installment_number   │
│  │ percentage           │
│  │ amount_*             │
│  │ due_date             │
│  │ status               │
│  └──────────────────────┘
│
├─────────────────────────────────────────────────────────────────────────────┤
│                       SUB-RENTAL & FEDERATION                               │
├─────────────────────────────────────────────────────────────────────────────┤
│
│  ┌──────────────────────┐
│  │sub_rental_requests   │
│  ├──────────────────────┤
│  │ id (PK)              │
│  │ project_id (FK)      │
│  │ partner_id (FK)      │
│  │ status               │
│  │ needed_dates         │
│  └──────────────────────┘
│         │
│         ▼
│  ┌──────────────────────┐      ┌──────────────────┐
│  │sub_rental_equipments │      │federation_       │
│  ├──────────────────────┤      │  partners        │
│  │ equipment_id (FK)    │      ├──────────────────┤
│  │ partner_equipment_id │──┐   │ id (PK)          │
│  │ daily_rental_rate    │  │   │ name (UNIQ)      │
│  └──────────────────────┘  │   │ api_key          │
│                            │   │ is_active        │
│                            │   │ is_trusted       │
│                            │   └──────────────────┘
│                            │            │
│                            │            ▼
│                            │   ┌──────────────────┐
│                            │   │sub_rental_       │
│                            │   │ conditions       │
│                            │   ├──────────────────┤
│                            │   │ rental_rate_*    │
│                            │   │ damage_deposit%  │
│                            │   │ min_period       │
│                            │   └──────────────────┘
│                            │
│                            ▼
│  ┌────────────────────────────────────┐
│  │federated_equipment (sync from      │
│  │           partner)                 │
│  ├────────────────────────────────────┤
│  │ partner_equipment_id               │
│  │ availability, prices               │
│  │ last_synced                        │
│  └────────────────────────────────────┘
│
│  ┌────────────────────────────────────┐
│  │sub_rental_handover_reports         │
│  ├────────────────────────────────────┤
│  │ request_id (FK)                    │
│  │ report_type (pickup/return)        │
│  │ equipment_condition (JSONB)        │
│  │ photo_paths[], signature_paths[]   │
│  └────────────────────────────────────┘
│
├─────────────────────────────────────────────────────────────────────────────┤
│                        AUDIT & SYSTEM                                       │
├─────────────────────────────────────────────────────────────────────────────┤
│
│  ┌──────────────────────┐  ┌─────────────────┐
│  │   audit_logs         │  │ federation_     │
│  ├──────────────────────┤  │   api_logs      │
│  │ user_id (FK)         │  ├─────────────────┤
│  │ action               │  │ partner_id (FK) │
│  │ entity_type          │  │ endpoint        │
│  │ entity_id            │  │ request_body    │
│  │ old_values (JSONB)   │  │ response_body   │
│  │ new_values (JSONB)   │  │ error_message   │
│  │ created_at           │  └─────────────────┘
│  └──────────────────────┘
│
│  ┌──────────────────────┐  ┌──────────────────┐
│  │   sessions           │  │   api_keys       │
│  ├──────────────────────┤  ├──────────────────┤
│  │ user_id (FK)         │  │ user_id (FK)     │
│  │ token_hash (UNIQ)    │  │ key_hash (UNIQ)  │
│  │ expires_at           │  │ expires_at       │
│  │ is_active            │  │ permissions      │
│  └──────────────────────┘  └──────────────────┘
│
│  ┌──────────────────────┐  ┌──────────────────┐
│  │scanner_sessions      │  │ scanner_scans    │
│  ├──────────────────────┤  ├──────────────────┤
│  │ user_id (FK)         │  │session_id (FK)   │
│  │ packing_list_id(FK)  │  │ barcode          │
│  │ session_type         │  │ equipment_id(FK) │
│  │ status               │  │ is_valid         │
│  │ items_scanned        │  │ created_at       │
│  └──────────────────────┘  └──────────────────┘
│
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 8. Sicherheits- & Performance-Überlegungen

### Sicherheit:
1. **Passwords**: Nur `password_hash` speichern (bcrypt mit Work-Factor 12+)
2. **API-Keys**: Nur gehashed speichern (`api_key_hash`), echten Key nur bei Erstellung zeigen
3. **Federation**: RSA Public Key für Request-Signing, nicht private keys in DB
4. **Audit-Trail**: Alle sicherheitsrelevanten Aktionen loggen
5. **Row-Level Security** (Falls mehrere Mandanten nötig, aber nicht für diese Architektur)
6. **Soft Deletes**: Keine echte Löschung, ermöglicht Recovery

### Performance:
1. **Denormalisierung**: `equipment.quantity_available`, `equipment.current_value` denormalisiert für Speed
2. **Materialisierte Views**: `equipment_availability` täglich/stündlich aktualisieren
3. **Indizes**: Auf alle Fremdschlüssel, häufig gefilterte Spalten (`status`, `is_active`, Daten-Ranges)
4. **JSONB-Indizes**: `CREATE INDEX idx_metadata ON equipment USING GIN (metadata)` für komplexe Abfragen
5. **Partitionierung**: Optional bei >1M Rows auf `audit_logs` und `federation_api_logs`
6. **Vacuum & Analyze**: Regelmäßig auf `equipment_availability` und anderen aktiven Tabellen

### Datenschutz (DSGVO):
- `deleted_at` ermöglicht "Recht auf Vergessenwerden" ohne Daten zu löschen
- `audit_logs` zeigen wer wann was geändert hat
- API für Daten-Export (Alle persönlichen Daten eines Users)
- Keine Klartexte in Logs, nur Hashes für Keys

---

## 9. Queries für häufige Use-Cases

### 1. Equipment-Verfügbarkeit für Date-Range
```sql
SELECT
  e.id,
  e.name,
  e.quantity_available,
  COALESCE(SUM(pe.quantity), 0) as allocated_qty,
  (e.quantity_available - COALESCE(SUM(pe.quantity), 0)) as available_qty
FROM equipment e
LEFT JOIN project_equipment pe
  ON e.id = pe.equipment_id
  AND pe.allocated_from_date <= '2026-04-01'
  AND pe.allocated_until_date >= '2026-04-01'
WHERE e.deleted_at IS NULL
GROUP BY e.id, e.name, e.quantity_available;
```

### 2. Offene Rechnungen mit Mahnungen
```sql
SELECT
  i.id,
  i.invoice_number,
  i.client_name,
  (i.total_gross - i.amount_paid) as amount_open,
  COALESCE(MAX(r.reminder_level), 0) as last_reminder_level,
  MAX(r.sent_date) as last_reminder_sent
FROM invoices i
LEFT JOIN reminders r ON i.id = r.invoice_id AND r.status != 'payment_received'
WHERE
  i.deleted_at IS NULL
  AND (i.total_gross - i.amount_paid) > 0
  AND i.status NOT IN ('cancelled', 'paid')
GROUP BY i.id
ORDER BY i.due_date ASC;
```

### 3. Projekt-Equipment-Packliste
```sql
SELECT
  p.name,
  e.name,
  pli.quantity,
  pli.is_picked,
  pli.picked_by_user_id,
  sl.name as storage_location
FROM packing_lists pl
JOIN projects p ON pl.project_id = p.id
JOIN packing_list_items pli ON pl.id = pli.packing_list_id
JOIN equipment e ON pli.equipment_id = e.id
LEFT JOIN storage_locations sl ON e.primary_storage_location_id = sl.id
WHERE pl.project_id = $1
ORDER BY pli.sequence_order;
```

### 4. Defekt-Equipment für Wartung
```sql
SELECT
  e.id,
  e.name,
  e.condition,
  dr.damage_date,
  dr.description,
  dr.severity_level,
  dr.estimated_repair_cost
FROM equipment e
JOIN equipment_damage_reports dr ON e.id = dr.equipment_id
WHERE
  e.deleted_at IS NULL
  AND (dr.repair_status IN ('reported', 'in_repair')
  OR e.next_maintenance_date <= CURRENT_DATE)
ORDER BY dr.severity_level DESC, dr.damage_date DESC;
```

### 5. Sub-Rental-Verfügbarkeit von Partner
```sql
SELECT
  partner.name,
  fe.name,
  fe.quantity_available,
  fe.daily_rental_rate,
  fe.last_synced,
  sr.needed_from_date,
  sr.needed_until_date
FROM sub_rental_requests sr
JOIN federation_partners partner ON sr.requesting_from_partner_id = partner.id
JOIN federated_equipment fe ON (
  partner.id = fe.partner_id
  AND sr.equipment_specifications->>'category' = fe.category
)
WHERE
  sr.status = 'requested'
  AND partner.is_active = true
  AND fe.is_available = true;
```

---

## 10. Summe: Normalisierung & Denormalisierung

### Normalisiert (3NF):
- equipment_types_catalog (getrennt von equipment)
- roles und user_roles (viele zu viele, getrennt)
- storage_location_hierarchy (separates normalisiertes Modell für hierarchische Pfade)
- invoice_items und invoice_installments (separate Tabellen, nicht JSONB)

### Bewusst denormalisiert:
- `equipment.quantity_available`, `quantity_total` (redundant, aber kritisch für Verfügbarkeit)
- `equipment.current_value` (kann aus Purchase + Depreciation berechnet werden, aber für Reports wichtig)
- `invoice.amount_paid` (könnte aus Payments summiert werden, aber häufig gelesen)
- `projects.created_at, updated_at` (Denormalisierung für schnelle Filterung)

---

## 11. Konfiguration für Go-Backend

```go
// Beispiel: gorm.Model mit UUID und Soft Delete
type User struct {
  ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
  Email     string    `gorm:"uniqueIndex"`
  Role      string    `gorm:"index"`
  CreatedAt time.Time
  UpdatedAt time.Time
  DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Equipment struct {
  ID                          uuid.UUID `gorm:"type:uuid;primaryKey"`
  Barcode                     string    `gorm:"uniqueIndex;index"`
  Type                        string    `gorm:"index"`
  Condition                   string    `gorm:"index"`
  QuantityAvailable           decimal.Decimal
  CurrentValue                decimal.Decimal
  PrimaryStorageLocationID    uuid.UUID
  CreatedAt                   time.Time
  UpdatedAt                   time.Time
  DeletedAt                   gorm.DeletedAt `gorm:"index"`
}
```

---

## Zusammenfassung

Das Schema unterstützt:
✅ Hierarchische Lagerplätze mit Barcodes/QR-Codes
✅ Equipment mit Kombinationen (Flightcases) und Defekt-Tracking
✅ Projekt-Management mit Packlisten und Freelancer-Zeiterfassung
✅ Angebote → Rechnungen → Mahnungen → DATEV-Export
✅ Sub-Rental mit Zustandsdokumentation und Partner-Konditionen
✅ Federation für Peer-to-Peer Equipment-Sharing
✅ Rollenbasierte Zugriffscontrol mit Audit-Trail
✅ Scanner-Integration für Barcode/QR
✅ DECIMAL für Geldbeträge, JSONB für Flexibilität
✅ Soft-Deletes für Datenwiederherstellung
✅ Performance-Indizes für häufige Queries
✅ 3NF Normalisierung mit bewussten Denormalisierungen
