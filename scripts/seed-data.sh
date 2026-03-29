#!/bin/bash
# =============================================================================
# CrateDesk — Seed-Daten fuer Entwicklung und Tests
# Fuegt realistische Testdaten in alle Service-Datenbanken ein.
# Usage: bash scripts/seed-data.sh
# Env:   DB_HOST, DB_PORT, DB_USER, DB_PASSWORD (defaults: localhost, 5432, cratedesk)
# =============================================================================

set -euo pipefail

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-cratedesk}"
DB_PASSWORD="${DB_PASSWORD:-}"
export PGPASSWORD="$DB_PASSWORD"

psql_exec() {
  local db="$1"
  shift
  psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$db" -q -v ON_ERROR_STOP=1 "$@"
}

echo "=== CrateDesk: Seed-Daten einfuegen ==="
echo "    Host: $DB_HOST:$DB_PORT, User: $DB_USER"

# -------------------------------------------------------
# 1. Tenant ID ermitteln
# -------------------------------------------------------
TENANT_ID=$(psql_exec auth_service -tAc "SELECT id FROM tenants LIMIT 1" 2>/dev/null || echo "")
if [ -z "$TENANT_ID" ]; then
  echo "FEHLER: Kein Tenant gefunden. Bitte zuerst /api/v1/setup ausfuehren."
  exit 1
fi
echo "    Tenant: $TENANT_ID"

# -------------------------------------------------------
# 2. Inventory: Kategorien + Equipment
# -------------------------------------------------------
echo "  -> Inventory: Kategorien + Equipment..."
psql_exec inventory_service <<SQL
-- Kategorien (idempotent via ON CONFLICT)
INSERT INTO categories (id, tenant_id, name, icon, color, sort_order) VALUES
  ('a0000001-0000-0000-0000-000000000001', '$TENANT_ID', 'Mikrofone', 'mic', '#3b82f6', 1),
  ('a0000001-0000-0000-0000-000000000002', '$TENANT_ID', 'Lautsprecher', 'speaker', '#8b5cf6', 2),
  ('a0000001-0000-0000-0000-000000000003', '$TENANT_ID', 'Mischpulte', 'sliders', '#10b981', 3),
  ('a0000001-0000-0000-0000-000000000004', '$TENANT_ID', 'Licht', 'lightbulb', '#f59e0b', 4),
  ('a0000001-0000-0000-0000-000000000005', '$TENANT_ID', 'Kabel & Stecker', 'cable', '#6b7280', 5),
  ('a0000001-0000-0000-0000-000000000006', '$TENANT_ID', 'Stative & Rigging', 'maximize', '#ef4444', 6),
  ('a0000001-0000-0000-0000-000000000007', '$TENANT_ID', 'Video', 'monitor', '#06b6d4', 7),
  ('a0000001-0000-0000-0000-000000000008', '$TENANT_ID', 'Strom', 'zap', '#eab308', 8)
ON CONFLICT (tenant_id, name, parent_id) DO NOTHING;

-- Equipment (VT-typische Geraete)
INSERT INTO equipment (id, tenant_id, category_id, name, barcode, serial_number, status, condition, quantity_total, quantity_available, rental_price_day, rental_price_week, replacement_value, manufacturer, model) VALUES
  ('b0000001-0000-0000-0000-000000000001', '$TENANT_ID', 'a0000001-0000-0000-0000-000000000001', 'Shure SM58', 'EQ-001', 'SM58-001', 'available', 'operational', 1, 1, 1500, 6000, 11900, 'Shure', 'SM58'),
  ('b0000001-0000-0000-0000-000000000002', '$TENANT_ID', 'a0000001-0000-0000-0000-000000000001', 'Shure SM58', 'EQ-002', 'SM58-002', 'available', 'operational', 1, 1, 1500, 6000, 11900, 'Shure', 'SM58'),
  ('b0000001-0000-0000-0000-000000000003', '$TENANT_ID', 'a0000001-0000-0000-0000-000000000001', 'Shure SM58', 'EQ-003', 'SM58-003', 'available', 'operational', 1, 1, 1500, 6000, 11900, 'Shure', 'SM58'),
  ('b0000001-0000-0000-0000-000000000004', '$TENANT_ID', 'a0000001-0000-0000-0000-000000000001', 'Shure SM58', 'EQ-004', 'SM58-004', 'available', 'operational', 1, 1, 1500, 6000, 11900, 'Shure', 'SM58'),
  ('b0000001-0000-0000-0000-000000000005', '$TENANT_ID', 'a0000001-0000-0000-0000-000000000001', 'Sennheiser e835', 'EQ-005', 'E835-001', 'available', 'operational', 1, 1, 1200, 5000, 9900, 'Sennheiser', 'e835'),
  ('b0000001-0000-0000-0000-000000000006', '$TENANT_ID', 'a0000001-0000-0000-0000-000000000001', 'Sennheiser e835', 'EQ-006', 'E835-002', 'available', 'operational', 1, 1, 1200, 5000, 9900, 'Sennheiser', 'e835'),
  ('b0000001-0000-0000-0000-000000000007', '$TENANT_ID', 'a0000001-0000-0000-0000-000000000001', 'Shure SM7B', 'EQ-007', 'SM7B-001', 'available', 'operational', 1, 1, 3500, 14000, 39900, 'Shure', 'SM7B'),
  ('b0000001-0000-0000-0000-000000000008', '$TENANT_ID', 'a0000001-0000-0000-0000-000000000001', 'Sennheiser EW 100 G4', 'EQ-008', 'EW100-001', 'available', 'operational', 1, 1, 4500, 18000, 59900, 'Sennheiser', 'EW 100 G4'),
  ('b0000001-0000-0000-0000-000000000009', '$TENANT_ID', 'a0000001-0000-0000-0000-000000000001', 'Sennheiser EW 100 G4', 'EQ-009', 'EW100-002', 'available', 'operational', 1, 1, 4500, 18000, 59900, 'Sennheiser', 'EW 100 G4'),
  ('b0000001-0000-0000-0000-000000000010', '$TENANT_ID', 'a0000001-0000-0000-0000-000000000002', 'QSC K12.2', 'EQ-010', 'K12-001', 'available', 'operational', 1, 1, 7500, 30000, 89900, 'QSC', 'K12.2'),
  ('b0000001-0000-0000-0000-000000000011', '$TENANT_ID', 'a0000001-0000-0000-0000-000000000002', 'QSC K12.2', 'EQ-011', 'K12-002', 'available', 'operational', 1, 1, 7500, 30000, 89900, 'QSC', 'K12.2'),
  ('b0000001-0000-0000-0000-000000000012', '$TENANT_ID', 'a0000001-0000-0000-0000-000000000002', 'QSC KS118', 'EQ-012', 'KS118-001', 'available', 'operational', 1, 1, 8500, 34000, 119900, 'QSC', 'KS118'),
  ('b0000001-0000-0000-0000-000000000013', '$TENANT_ID', 'a0000001-0000-0000-0000-000000000002', 'QSC KS118', 'EQ-013', 'KS118-002', 'available', 'operational', 1, 1, 8500, 34000, 119900, 'QSC', 'KS118'),
  ('b0000001-0000-0000-0000-000000000014', '$TENANT_ID', 'a0000001-0000-0000-0000-000000000002', 'JBL EON615', 'EQ-014', 'EON615-001', 'available', 'good', 1, 1, 5000, 20000, 49900, 'JBL', 'EON615'),
  ('b0000001-0000-0000-0000-000000000015', '$TENANT_ID', 'a0000001-0000-0000-0000-000000000002', 'JBL EON615', 'EQ-015', 'EON615-002', 'available', 'good', 1, 1, 5000, 20000, 49900, 'JBL', 'EON615'),
  ('b0000001-0000-0000-0000-000000000016', '$TENANT_ID', 'a0000001-0000-0000-0000-000000000003', 'Allen & Heath SQ-5', 'EQ-016', 'SQ5-001', 'available', 'operational', 1, 1, 15000, 60000, 349900, 'Allen & Heath', 'SQ-5'),
  ('b0000001-0000-0000-0000-000000000017', '$TENANT_ID', 'a0000001-0000-0000-0000-000000000003', 'Yamaha TF1', 'EQ-017', 'TF1-001', 'available', 'operational', 1, 1, 12000, 48000, 299900, 'Yamaha', 'TF1'),
  ('b0000001-0000-0000-0000-000000000018', '$TENANT_ID', 'a0000001-0000-0000-0000-000000000003', 'Behringer X32', 'EQ-018', 'X32-001', 'available', 'operational', 1, 1, 8000, 32000, 199900, 'Behringer', 'X32'),
  ('b0000001-0000-0000-0000-000000000019', '$TENANT_ID', 'a0000001-0000-0000-0000-000000000004', 'Chauvet DJ SlimPAR Pro H', 'EQ-019', 'SLIM-001', 'available', 'operational', 1, 1, 2500, 10000, 29900, 'Chauvet DJ', 'SlimPAR Pro H'),
  ('b0000001-0000-0000-0000-000000000020', '$TENANT_ID', 'a0000001-0000-0000-0000-000000000004', 'Chauvet DJ SlimPAR Pro H', 'EQ-020', 'SLIM-002', 'available', 'operational', 1, 1, 2500, 10000, 29900, 'Chauvet DJ', 'SlimPAR Pro H'),
  ('b0000001-0000-0000-0000-000000000021', '$TENANT_ID', 'a0000001-0000-0000-0000-000000000004', 'Chauvet DJ SlimPAR Pro H', 'EQ-021', 'SLIM-003', 'available', 'operational', 1, 1, 2500, 10000, 29900, 'Chauvet DJ', 'SlimPAR Pro H'),
  ('b0000001-0000-0000-0000-000000000022', '$TENANT_ID', 'a0000001-0000-0000-0000-000000000004', 'Chauvet DJ SlimPAR Pro H', 'EQ-022', 'SLIM-004', 'available', 'operational', 1, 1, 2500, 10000, 29900, 'Chauvet DJ', 'SlimPAR Pro H'),
  ('b0000001-0000-0000-0000-000000000023', '$TENANT_ID', 'a0000001-0000-0000-0000-000000000004', 'Martin RUSH MH 5 Profile', 'EQ-023', 'RUSH-001', 'available', 'operational', 1, 1, 6000, 24000, 129900, 'Martin', 'RUSH MH 5 Profile'),
  ('b0000001-0000-0000-0000-000000000024', '$TENANT_ID', 'a0000001-0000-0000-0000-000000000004', 'Martin RUSH MH 5 Profile', 'EQ-024', 'RUSH-002', 'available', 'operational', 1, 1, 6000, 24000, 129900, 'Martin', 'RUSH MH 5 Profile'),
  ('b0000001-0000-0000-0000-000000000025', '$TENANT_ID', 'a0000001-0000-0000-0000-000000000004', 'GrandMA3 onPC', 'EQ-025', 'MA3-001', 'available', 'operational', 1, 1, 5000, 20000, 79900, 'MA Lighting', 'GrandMA3 onPC'),
  ('b0000001-0000-0000-0000-000000000026', '$TENANT_ID', 'a0000001-0000-0000-0000-000000000005', 'XLR Kabel 10m', 'EQ-026', 'XLR10-001', 'available', 'operational', 1, 1, 200, 800, 1500, 'Cordial', 'CFM 10 FM'),
  ('b0000001-0000-0000-0000-000000000027', '$TENANT_ID', 'a0000001-0000-0000-0000-000000000005', 'XLR Kabel 10m', 'EQ-027', 'XLR10-002', 'available', 'operational', 1, 1, 200, 800, 1500, 'Cordial', 'CFM 10 FM'),
  ('b0000001-0000-0000-0000-000000000028', '$TENANT_ID', 'a0000001-0000-0000-0000-000000000005', 'Powercon Kabel 5m', 'EQ-028', 'PWR5-001', 'available', 'operational', 1, 1, 300, 1200, 2500, 'Neutrik', 'Powercon'),
  ('b0000001-0000-0000-0000-000000000029', '$TENANT_ID', 'a0000001-0000-0000-0000-000000000006', 'K&M Mikrofonstativ', 'EQ-029', 'KM-001', 'available', 'operational', 1, 1, 500, 2000, 4900, 'K&M', '210/9'),
  ('b0000001-0000-0000-0000-000000000030', '$TENANT_ID', 'a0000001-0000-0000-0000-000000000006', 'K&M Mikrofonstativ', 'EQ-030', 'KM-002', 'available', 'operational', 1, 1, 500, 2000, 4900, 'K&M', '210/9')
ON CONFLICT DO NOTHING;
SQL

# -------------------------------------------------------
# 3. Customers
# -------------------------------------------------------
echo "  -> Customers..."
psql_exec customer_service <<SQL
INSERT INTO customers (id, tenant_id, company_name, customer_number, email, phone, billing_address_street, billing_address_city, billing_address_zip, billing_address_country) VALUES
  ('c0000001-0000-0000-0000-000000000001', '$TENANT_ID', 'Eventhaus Muenchen GmbH', 'K-001', 'info@eventhaus-muenchen.de', '+49 89 12345678', 'Leopoldstr. 42', 'Muenchen', '80802', 'DE'),
  ('c0000001-0000-0000-0000-000000000002', '$TENANT_ID', 'Hochzeitsplaner Schmidt', 'K-002', 'kontakt@hochzeit-schmidt.de', '+49 911 9876543', 'Koenigsstr. 15', 'Nuernberg', '90402', 'DE'),
  ('c0000001-0000-0000-0000-000000000003', '$TENANT_ID', 'Stadthalle Fuerth', 'K-003', 'technik@stadthalle-fuerth.de', '+49 911 5554321', 'Rosenstr. 50', 'Fuerth', '90762', 'DE'),
  ('c0000001-0000-0000-0000-000000000004', '$TENANT_ID', 'DJ Marco Events', 'K-004', 'marco@dj-marco.de', '+49 170 1234567', 'Hauptstr. 8', 'Erlangen', '91054', 'DE'),
  ('c0000001-0000-0000-0000-000000000005', '$TENANT_ID', 'Kulturzentrum K4', 'K-005', 'technik@k4-nuernberg.de', '+49 911 2345678', 'Koenigstor 4', 'Nuernberg', '90402', 'DE'),
  ('c0000001-0000-0000-0000-000000000006', '$TENANT_ID', 'Band Frequency', 'K-006', 'booking@frequency-band.de', '+49 160 9876543', 'Beethovenstr. 12', 'Fuerth', '90762', 'DE'),
  ('c0000001-0000-0000-0000-000000000007', '$TENANT_ID', 'Messebau Weber', 'K-007', 'info@messebau-weber.de', '+49 911 7654321', 'Industriestr. 23', 'Nuernberg', '90441', 'DE'),
  ('c0000001-0000-0000-0000-000000000008', '$TENANT_ID', 'Kirchengemeinde St. Paul', 'K-008', 'pfarramt@st-paul.de', '+49 911 1112222', 'Kirchweg 1', 'Nuernberg', '90409', 'DE'),
  ('c0000001-0000-0000-0000-000000000009', '$TENANT_ID', 'Sportverein TSV 1860', 'K-009', 'vorstand@tsv1860.de', '+49 911 3334444', 'Sportplatzweg 5', 'Feucht', '90537', 'DE'),
  ('c0000001-0000-0000-0000-000000000010', '$TENANT_ID', 'Restaurant Zum Goldenen Hirschen', 'K-010', 'info@goldener-hirsch.de', '+49 911 5556666', 'Marktplatz 3', 'Altdorf', '90518', 'DE')
ON CONFLICT DO NOTHING;

INSERT INTO contacts (id, tenant_id, customer_id, first_name, last_name, email, phone, position, is_primary) VALUES
  ('d0000001-0000-0000-0000-000000000001', '$TENANT_ID', 'c0000001-0000-0000-0000-000000000001', 'Thomas', 'Huber', 'thomas.huber@eventhaus-muenchen.de', '+49 89 12345679', 'Technischer Leiter', true),
  ('d0000001-0000-0000-0000-000000000002', '$TENANT_ID', 'c0000001-0000-0000-0000-000000000002', 'Lisa', 'Schmidt', 'lisa@hochzeit-schmidt.de', '+49 911 9876544', 'Inhaberin', true),
  ('d0000001-0000-0000-0000-000000000003', '$TENANT_ID', 'c0000001-0000-0000-0000-000000000003', 'Martin', 'Braun', 'braun@stadthalle-fuerth.de', '+49 911 5554322', 'Veranstaltungstechniker', true),
  ('d0000001-0000-0000-0000-000000000004', '$TENANT_ID', 'c0000001-0000-0000-0000-000000000005', 'Sarah', 'Mueller', 'sarah@k4-nuernberg.de', '+49 911 2345679', 'Programmleitung', true),
  ('d0000001-0000-0000-0000-000000000005', '$TENANT_ID', 'c0000001-0000-0000-0000-000000000007', 'Klaus', 'Weber', 'klaus@messebau-weber.de', '+49 911 7654322', 'Geschaeftsfuehrer', true)
ON CONFLICT DO NOTHING;
SQL

# -------------------------------------------------------
# 4. Projects
# -------------------------------------------------------
echo "  -> Projects..."
psql_exec project_service <<SQL
INSERT INTO projects (id, tenant_id, name, project_number, status, customer_id, contact_name, contact_email, venue_name, venue_address, start_date, end_date, budget, notes) VALUES
  ('e0000001-0000-0000-0000-000000000001', '$TENANT_ID', 'Firmengala Eventhaus', 'P-2026-001', 'confirmed', 'c0000001-0000-0000-0000-000000000001', 'Thomas Huber', 'thomas.huber@eventhaus-muenchen.de', 'Eventhaus Muenchen', 'Leopoldstr. 42, 80802 Muenchen', '2026-04-15', '2026-04-16', 350000, 'Komplette PA + Licht, 200 Gaeste'),
  ('e0000001-0000-0000-0000-000000000002', '$TENANT_ID', 'Hochzeit Mueller-Bauer', 'P-2026-002', 'confirmed', 'c0000001-0000-0000-0000-000000000002', 'Lisa Schmidt', 'lisa@hochzeit-schmidt.de', 'Schloss Burgthann', 'Schlossstr. 1, 90559 Burgthann', '2026-04-26', '2026-04-27', 180000, 'Outdoor-Zeremonie + Party, DJ Setup'),
  ('e0000001-0000-0000-0000-000000000003', '$TENANT_ID', 'Konzert Frequency', 'P-2026-003', 'active', 'c0000001-0000-0000-0000-000000000006', 'Frequency Band', 'booking@frequency-band.de', 'Kulturzentrum K4', 'Koenigstor 4, 90402 Nuernberg', '2026-04-05', '2026-04-05', 120000, 'Full Band Setup, 16 Kanaele'),
  ('e0000001-0000-0000-0000-000000000004', '$TENANT_ID', 'Messe NuernbergMesse Halle 4', 'P-2026-004', 'draft', 'c0000001-0000-0000-0000-000000000007', 'Klaus Weber', 'klaus@messebau-weber.de', 'NuernbergMesse Halle 4', 'Messezentrum 1, 90471 Nuernberg', '2026-05-10', '2026-05-13', 450000, 'Messestand-Beschallung + Video'),
  ('e0000001-0000-0000-0000-000000000005', '$TENANT_ID', 'Sommerfest TSV', 'P-2026-005', 'draft', 'c0000001-0000-0000-0000-000000000009', 'TSV Vorstand', 'vorstand@tsv1860.de', 'Sportplatz TSV Feucht', 'Sportplatzweg 5, 90537 Feucht', '2026-06-20', '2026-06-21', 80000, 'Outdoor PA, 4 SlimPARs, DJ Pult')
ON CONFLICT DO NOTHING;
SQL

# -------------------------------------------------------
# 5. Invoices
# -------------------------------------------------------
echo "  -> Invoices..."
psql_exec invoice_service <<SQL
INSERT INTO number_sequences (tenant_id, prefix, year, last_number)
VALUES ('$TENANT_ID', 'RE', 2026, 3)
ON CONFLICT (tenant_id, prefix, year) DO UPDATE SET last_number = GREATEST(number_sequences.last_number, 3);

INSERT INTO invoices (id, tenant_id, invoice_number, status, customer_name, customer_email, customer_address, invoice_date, due_date, vat_rate, kleinunternehmer, total_net, total_vat, total_gross) VALUES
  ('f0000001-0000-0000-0000-000000000001', '$TENANT_ID', 'RE-2026-00001', 'sent', 'Eventhaus Muenchen GmbH', 'info@eventhaus-muenchen.de', 'Leopoldstr. 42, 80802 Muenchen', '2026-03-15', '2026-04-14', 0, true, 250000, 0, 250000),
  ('f0000001-0000-0000-0000-000000000002', '$TENANT_ID', 'RE-2026-00002', 'draft', 'Hochzeitsplaner Schmidt', 'kontakt@hochzeit-schmidt.de', 'Koenigsstr. 15, 90402 Nuernberg', '2026-03-28', '2026-04-27', 0, true, 180000, 0, 180000),
  ('f0000001-0000-0000-0000-000000000003', '$TENANT_ID', 'RE-2026-00003', 'paid', 'Band Frequency', 'booking@frequency-band.de', 'Beethovenstr. 12, 90762 Fuerth', '2026-03-01', '2026-03-31', 0, true, 95000, 0, 95000)
ON CONFLICT DO NOTHING;

INSERT INTO invoice_items (id, tenant_id, invoice_id, description, quantity, unit, unit_price, position) VALUES
  ('f1000001-0000-0000-0000-000000000001', '$TENANT_ID', 'f0000001-0000-0000-0000-000000000001', 'PA System QSC K12.2 (2x)', 1, 'Pauschal', 150000, 1),
  ('f1000001-0000-0000-0000-000000000002', '$TENANT_ID', 'f0000001-0000-0000-0000-000000000001', 'Licht SlimPAR (4x)', 1, 'Pauschal', 60000, 2),
  ('f1000001-0000-0000-0000-000000000003', '$TENANT_ID', 'f0000001-0000-0000-0000-000000000001', 'Techniker vor Ort (8h)', 8, 'Stunde', 5000, 3),
  ('f1000001-0000-0000-0000-000000000004', '$TENANT_ID', 'f0000001-0000-0000-0000-000000000002', 'DJ Setup komplett', 1, 'Pauschal', 120000, 1),
  ('f1000001-0000-0000-0000-000000000005', '$TENANT_ID', 'f0000001-0000-0000-0000-000000000002', 'Lichteffekte Hochzeit', 1, 'Pauschal', 60000, 2),
  ('f1000001-0000-0000-0000-000000000006', '$TENANT_ID', 'f0000001-0000-0000-0000-000000000003', 'PA + Monitoring Band', 1, 'Pauschal', 75000, 1),
  ('f1000001-0000-0000-0000-000000000007', '$TENANT_ID', 'f0000001-0000-0000-0000-000000000003', 'Anlieferung + Aufbau', 1, 'Pauschal', 20000, 2)
ON CONFLICT DO NOTHING;

INSERT INTO payments (id, tenant_id, invoice_id, amount, payment_date, payment_method, reference) VALUES
  ('f2000001-0000-0000-0000-000000000001', '$TENANT_ID', 'f0000001-0000-0000-0000-000000000003', 95000, '2026-03-20', 'bank_transfer', 'Ueberweisung RE-2026-00003')
ON CONFLICT DO NOTHING;
SQL

# -------------------------------------------------------
# 6. Expense Categories
# -------------------------------------------------------
echo "  -> Expense Categories..."
psql_exec expense_service <<SQL
INSERT INTO expense_categories (id, tenant_id, name, color) VALUES
  ('g0000001-0000-0000-0000-000000000001', '$TENANT_ID', 'Fahrtkosten', '#3b82f6'),
  ('g0000001-0000-0000-0000-000000000002', '$TENANT_ID', 'Verschleissteile', '#ef4444'),
  ('g0000001-0000-0000-0000-000000000003', '$TENANT_ID', 'Werkzeug', '#10b981'),
  ('g0000001-0000-0000-0000-000000000004', '$TENANT_ID', 'Buero & Verwaltung', '#8b5cf6'),
  ('g0000001-0000-0000-0000-000000000005', '$TENANT_ID', 'Bewirtung', '#f59e0b')
ON CONFLICT DO NOTHING;
SQL

echo "=== CrateDesk: Seed-Daten erfolgreich eingefuegt ==="
echo "    8 Kategorien, 30 Equipment, 10 Kunden, 5 Projekte, 3 Rechnungen, 5 Ausgaben-Kategorien"
