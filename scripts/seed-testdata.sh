#!/bin/bash
# RentFlow Test Data Seeder
# Creates categories, equipment, and sample projects for testing
# Safe to run multiple times (uses ON CONFLICT DO NOTHING)

set -e

POSTGRES_CONTAINER="rentflow-postgres"
POSTGRES_USER="rentflow"

echo "=== RentFlow Test Data Seeder ==="

# Get tenant_id and admin user_id
TENANT_ID=$(docker exec -i $POSTGRES_CONTAINER psql -U $POSTGRES_USER -d auth_service -t -c "SELECT id FROM auth.tenants LIMIT 1;" | tr -d ' \n')
ADMIN_ID=$(docker exec -i $POSTGRES_CONTAINER psql -U $POSTGRES_USER -d auth_service -t -c "SELECT id FROM auth.users LIMIT 1;" | tr -d ' \n')

if [ -z "$TENANT_ID" ]; then
  echo "ERROR: No tenant found. Run setup wizard first."
  exit 1
fi

echo "Tenant: $TENANT_ID"
echo "Admin:  $ADMIN_ID"

# Fix serial_number constraint
echo ""
echo "1. Fixing constraints..."
docker exec -i $POSTGRES_CONTAINER psql -U $POSTGRES_USER -d inventory_service -c \
  "ALTER TABLE inventory.equipment DROP CONSTRAINT IF EXISTS equipment_tenant_id_serial_number_key;" 2>&1 | grep -v "^$"

# Seed Categories
echo "2. Seeding categories..."
docker exec -i $POSTGRES_CONTAINER psql -U $POSTGRES_USER -d inventory_service -q <<SQL
INSERT INTO inventory.categories (id, tenant_id, name, icon, color, sort_order, created_by_user_id)
VALUES
  ('cat-audio-0001', '$TENANT_ID', 'Audio', '🔊', '#3B82F6', 1, '$ADMIN_ID'),
  ('cat-light-0001', '$TENANT_ID', 'Licht', '💡', '#F59E0B', 2, '$ADMIN_ID'),
  ('cat-video-0001', '$TENANT_ID', 'Video', '🎥', '#8B5CF6', 3, '$ADMIN_ID'),
  ('cat-stage-0001', '$TENANT_ID', 'Bühnentechnik', '🏗️', '#EF4444', 4, '$ADMIN_ID'),
  ('cat-cable-0001', '$TENANT_ID', 'Kabel & Zubehör', '🔌', '#6B7280', 5, '$ADMIN_ID'),
  ('cat-power-0001', '$TENANT_ID', 'Strom', '⚡', '#F97316', 6, '$ADMIN_ID')
ON CONFLICT DO NOTHING;
SQL

# Seed Equipment
echo "3. Seeding equipment (20 items)..."
docker exec -i $POSTGRES_CONTAINER psql -U $POSTGRES_USER -d inventory_service -q <<SQL
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, created_by_user_id, created_at, updated_at)
VALUES
  ('eq-001', '$TENANT_ID', 'JBL VTX A12', 'Line Array Lautsprecher', 'cat-audio-0001', 'AU-001', 'RF-SPK-001', 'available', 'excellent', 150.00, 750.00, '$ADMIN_ID', NOW(), NOW()),
  ('eq-002', '$TENANT_ID', 'JBL VTX S28', 'Subwoofer', 'cat-audio-0001', 'AU-002', 'RF-SUB-001', 'available', 'excellent', 120.00, 600.00, '$ADMIN_ID', NOW(), NOW()),
  ('eq-003', '$TENANT_ID', 'Yamaha CL5', 'Digitalmischpult', 'cat-audio-0001', 'AU-003', 'RF-MIX-001', 'available', 'good', 250.00, 1250.00, '$ADMIN_ID', NOW(), NOW()),
  ('eq-004', '$TENANT_ID', 'Shure SM58', 'Dynamisches Mikrofon', 'cat-audio-0001', 'AU-004', 'RF-MIC-001', 'available', 'good', 15.00, 60.00, '$ADMIN_ID', NOW(), NOW()),
  ('eq-005', '$TENANT_ID', 'Shure SM58', 'Dynamisches Mikrofon', 'cat-audio-0001', 'AU-005', 'RF-MIC-002', 'available', 'good', 15.00, 60.00, '$ADMIN_ID', NOW(), NOW()),
  ('eq-006', '$TENANT_ID', 'Sennheiser EW 100 G4', 'Funkmikrofon-Set', 'cat-audio-0001', 'AU-006', 'RF-WIR-001', 'available', 'excellent', 45.00, 200.00, '$ADMIN_ID', NOW(), NOW()),
  ('eq-007', '$TENANT_ID', 'Clay Paky Sharpy Plus', 'Beam Moving Head', 'cat-light-0001', 'LI-001', 'RF-MH-001', 'available', 'excellent', 80.00, 400.00, '$ADMIN_ID', NOW(), NOW()),
  ('eq-008', '$TENANT_ID', 'Clay Paky Sharpy Plus', 'Beam Moving Head', 'cat-light-0001', 'LI-002', 'RF-MH-002', 'available', 'excellent', 80.00, 400.00, '$ADMIN_ID', NOW(), NOW()),
  ('eq-009', '$TENANT_ID', 'GrandMA3 Light', 'Lichtpult', 'cat-light-0001', 'LI-003', 'RF-DMX-001', 'available', 'good', 200.00, 1000.00, '$ADMIN_ID', NOW(), NOW()),
  ('eq-010', '$TENANT_ID', 'Chauvet COLORado Panel Q40', 'LED Wash Panel', 'cat-light-0001', 'LI-004', 'RF-LED-001', 'available', 'good', 40.00, 200.00, '$ADMIN_ID', NOW(), NOW()),
  ('eq-011', '$TENANT_ID', 'Panasonic PT-RZ690', 'Laser-Beamer 6000 ANSI', 'cat-video-0001', 'VI-001', 'RF-BEA-001', 'available', 'excellent', 180.00, 900.00, '$ADMIN_ID', NOW(), NOW()),
  ('eq-012', '$TENANT_ID', 'Samsung 75" Display', 'LED Display 75 Zoll', 'cat-video-0001', 'VI-002', 'RF-DIS-001', 'available', 'good', 120.00, 600.00, '$ADMIN_ID', NOW(), NOW()),
  ('eq-013', '$TENANT_ID', 'Global Truss F34 3m', 'Traverse 3 Meter', 'cat-stage-0001', 'ST-001', 'RF-TRS-001', 'available', 'good', 25.00, 100.00, '$ADMIN_ID', NOW(), NOW()),
  ('eq-014', '$TENANT_ID', 'Global Truss F34 3m', 'Traverse 3 Meter', 'cat-stage-0001', 'ST-002', 'RF-TRS-002', 'available', 'good', 25.00, 100.00, '$ADMIN_ID', NOW(), NOW()),
  ('eq-015', '$TENANT_ID', 'Powerlock 125A Verteiler', 'Stromverteiler 125A', 'cat-power-0001', 'PW-001', 'RF-PWR-001', 'available', 'excellent', 60.00, 300.00, '$ADMIN_ID', NOW(), NOW()),
  ('eq-016', '$TENANT_ID', 'XLR Kabel 10m', 'XLR Kabel 10 Meter', 'cat-cable-0001', 'CA-001', 'RF-CAB-001', 'available', 'good', 5.00, 20.00, '$ADMIN_ID', NOW(), NOW()),
  ('eq-017', '$TENANT_ID', 'XLR Kabel 10m', 'XLR Kabel 10 Meter', 'cat-cable-0001', 'CA-002', 'RF-CAB-002', 'available', 'good', 5.00, 20.00, '$ADMIN_ID', NOW(), NOW()),
  ('eq-018', '$TENANT_ID', 'Powercon Kabel 5m', 'Powercon Kabel 5 Meter', 'cat-cable-0001', 'CA-003', 'RF-CAB-003', 'available', 'good', 5.00, 20.00, '$ADMIN_ID', NOW(), NOW()),
  ('eq-019', '$TENANT_ID', 'Martin MAC Aura XB', 'Wash Moving Head', 'cat-light-0001', 'LI-005', 'RF-MH-003', 'in_maintenance', 'fair', 90.00, 450.00, '$ADMIN_ID', NOW(), NOW()),
  ('eq-020', '$TENANT_ID', 'DI-Box Radial J48', 'Active DI Box', 'cat-audio-0001', 'AU-007', 'RF-DI-001', 'available', 'excellent', 10.00, 40.00, '$ADMIN_ID', NOW(), NOW())
ON CONFLICT DO NOTHING;
SQL

# Seed Projects
echo "4. Seeding projects..."
docker exec -i $POSTGRES_CONTAINER psql -U $POSTGRES_USER -d project_service -q <<SQL
INSERT INTO project.projects (id, tenant_id, name, description, client_name, status, start_date, end_date, budget, currency, created_at, updated_at, created_by_user_id)
VALUES
  ('proj-001', '$TENANT_ID', 'Sommerfest Müller GmbH', 'Outdoor Event mit PA System und Licht', 'Müller GmbH', 'confirmed', '2026-03-28 08:00:00+00', '2026-03-29 22:00:00+00', 5000, 'EUR', NOW(), NOW(), '$ADMIN_ID'),
  ('proj-002', '$TENANT_ID', 'Hochzeit Schmidt', 'Indoor Hochzeitsfeier mit Licht und Ton', 'Familie Schmidt', 'planning', '2026-04-05 10:00:00+00', '2026-04-06 02:00:00+00', 8000, 'EUR', NOW(), NOW(), '$ADMIN_ID'),
  ('proj-003', '$TENANT_ID', 'Firmenjubiläum TechCorp', 'Großes Firmenevent in Halle 5', 'TechCorp AG', 'planning', '2026-04-15 14:00:00+00', '2026-04-16 01:00:00+00', 12000, 'EUR', NOW(), NOW(), '$ADMIN_ID')
ON CONFLICT DO NOTHING;
SQL

echo ""
echo "=== Seed complete! ==="
echo "Created: 6 categories, 20 equipment, 3 projects"
echo ""
echo "Barcodes for scanner testing:"
echo "  RF-SPK-001  JBL VTX A12 (Audio)"
echo "  RF-SUB-001  JBL VTX S28 (Audio)"
echo "  RF-MIX-001  Yamaha CL5 (Audio)"
echo "  RF-MIC-001  Shure SM58 (Audio)"
echo "  RF-MH-001   Clay Paky Sharpy Plus (Licht)"
echo "  RF-DMX-001  GrandMA3 Light (Licht)"
echo "  RF-BEA-001  Panasonic PT-RZ690 (Video)"
echo "  RF-DIS-001  Samsung 75\" Display (Video)"
