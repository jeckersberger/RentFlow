#!/bin/bash
# RentFlow Test Data Seeder
# Creates categories and sample equipment for testing

set -e

POSTGRES_CONTAINER="rentflow-postgres"
POSTGRES_USER="rentflow"

echo "=== Seeding Test Data ==="

# Get tenant_id from auth_service
TENANT_ID=$(docker exec -i $POSTGRES_CONTAINER psql -U $POSTGRES_USER -d auth_service -t -c "SELECT id FROM auth.tenants LIMIT 1;" | tr -d ' \n')

if [ -z "$TENANT_ID" ]; then
  echo "ERROR: No tenant found. Run setup wizard first."
  exit 1
fi

echo "Using Tenant ID: $TENANT_ID"

# Seed Categories
echo "Creating categories..."
docker exec -i $POSTGRES_CONTAINER psql -U $POSTGRES_USER -d inventory_service <<SQL
INSERT INTO inventory.categories (id, tenant_id, name, description, created_at, updated_at)
VALUES
  ('cat-audio-0001', '$TENANT_ID', 'Audio', 'Lautsprecher, Mischpulte, Mikrofone', NOW(), NOW()),
  ('cat-light-0001', '$TENANT_ID', 'Licht', 'Scheinwerfer, Moving Heads, LED', NOW(), NOW()),
  ('cat-video-0001', '$TENANT_ID', 'Video', 'Beamer, Kameras, Bildschirme', NOW(), NOW()),
  ('cat-stage-0001', '$TENANT_ID', 'Bühnentechnik', 'Truss, Traversen, Podeste', NOW(), NOW()),
  ('cat-cable-0001', '$TENANT_ID', 'Kabel & Zubehör', 'Kabel, Adapter, Stecker', NOW(), NOW()),
  ('cat-power-0001', '$TENANT_ID', 'Strom', 'Stromverteiler, Verlängerungen', NOW(), NOW())
ON CONFLICT DO NOTHING;
SQL

# Seed Equipment
echo "Creating test equipment..."
docker exec -i $POSTGRES_CONTAINER psql -U $POSTGRES_USER -d inventory_service <<SQL
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, created_at, updated_at)
VALUES
  ('eq-001', '$TENANT_ID', 'JBL VTX A12', 'Line Array Lautsprecher', 'cat-audio-0001', 'AU-001', 'RF-SPK-001', 'available', 'excellent', 150.00, 750.00, NOW(), NOW()),
  ('eq-002', '$TENANT_ID', 'JBL VTX S28', 'Subwoofer', 'cat-audio-0001', 'AU-002', 'RF-SUB-001', 'available', 'excellent', 120.00, 600.00, NOW(), NOW()),
  ('eq-003', '$TENANT_ID', 'Yamaha CL5', 'Digitalmischpult', 'cat-audio-0001', 'AU-003', 'RF-MIX-001', 'available', 'good', 250.00, 1250.00, NOW(), NOW()),
  ('eq-004', '$TENANT_ID', 'Shure SM58', 'Dynamisches Mikrofon', 'cat-audio-0001', 'AU-004', 'RF-MIC-001', 'available', 'good', 15.00, 60.00, NOW(), NOW()),
  ('eq-005', '$TENANT_ID', 'Shure SM58', 'Dynamisches Mikrofon', 'cat-audio-0001', 'AU-005', 'RF-MIC-002', 'available', 'good', 15.00, 60.00, NOW(), NOW()),
  ('eq-006', '$TENANT_ID', 'Sennheiser EW 100 G4', 'Funkmikrofon-Set', 'cat-audio-0001', 'AU-006', 'RF-WIR-001', 'available', 'excellent', 45.00, 200.00, NOW(), NOW()),
  ('eq-007', '$TENANT_ID', 'Clay Paky Sharpy Plus', 'Beam Moving Head', 'cat-light-0001', 'LI-001', 'RF-MH-001', 'available', 'excellent', 80.00, 400.00, NOW(), NOW()),
  ('eq-008', '$TENANT_ID', 'Clay Paky Sharpy Plus', 'Beam Moving Head', 'cat-light-0001', 'LI-002', 'RF-MH-002', 'available', 'excellent', 80.00, 400.00, NOW(), NOW()),
  ('eq-009', '$TENANT_ID', 'GrandMA3 Light', 'Lichtpult', 'cat-light-0001', 'LI-003', 'RF-DMX-001', 'available', 'good', 200.00, 1000.00, NOW(), NOW()),
  ('eq-010', '$TENANT_ID', 'Chauvet COLORado Panel Q40', 'LED Wash Panel', 'cat-light-0001', 'LI-004', 'RF-LED-001', 'available', 'good', 40.00, 200.00, NOW(), NOW()),
  ('eq-011', '$TENANT_ID', 'Panasonic PT-RZ690', 'Laser-Beamer 6000 ANSI', 'cat-video-0001', 'VI-001', 'RF-BEA-001', 'available', 'excellent', 180.00, 900.00, NOW(), NOW()),
  ('eq-012', '$TENANT_ID', 'Samsung 75" Display', 'LED Display 75 Zoll', 'cat-video-0001', 'VI-002', 'RF-DIS-001', 'available', 'good', 120.00, 600.00, NOW(), NOW()),
  ('eq-013', '$TENANT_ID', 'Global Truss F34 3m', 'Traverse 3 Meter', 'cat-stage-0001', 'ST-001', 'RF-TRS-001', 'available', 'good', 25.00, 100.00, NOW(), NOW()),
  ('eq-014', '$TENANT_ID', 'Global Truss F34 3m', 'Traverse 3 Meter', 'cat-stage-0001', 'ST-002', 'RF-TRS-002', 'available', 'good', 25.00, 100.00, NOW(), NOW()),
  ('eq-015', '$TENANT_ID', 'Powerlock 125A Verteiler', 'Stromverteiler 125A', 'cat-power-0001', 'PW-001', 'RF-PWR-001', 'available', 'excellent', 60.00, 300.00, NOW(), NOW()),
  ('eq-016', '$TENANT_ID', 'XLR Kabel 10m', 'XLR Kabel 10 Meter', 'cat-cable-0001', 'CA-001', 'RF-CAB-001', 'available', 'good', 5.00, 20.00, NOW(), NOW()),
  ('eq-017', '$TENANT_ID', 'XLR Kabel 10m', 'XLR Kabel 10 Meter', 'cat-cable-0001', 'CA-002', 'RF-CAB-002', 'available', 'good', 5.00, 20.00, NOW(), NOW()),
  ('eq-018', '$TENANT_ID', 'Powercon Kabel 5m', 'Powercon Kabel 5 Meter', 'cat-cable-0001', 'CA-003', 'RF-CAB-003', 'available', 'good', 5.00, 20.00, NOW(), NOW()),
  ('eq-019', '$TENANT_ID', 'Martin MAC Aura XB', 'Wash Moving Head', 'cat-light-0001', 'LI-005', 'RF-MH-003', 'in_maintenance', 'fair', 90.00, 450.00, NOW(), NOW()),
  ('eq-020', '$TENANT_ID', 'DI-Box Radial J48', 'Active DI Box', 'cat-audio-0001', 'AU-007', 'RF-DI-001', 'available', 'excellent', 10.00, 40.00, NOW(), NOW())
ON CONFLICT DO NOTHING;
SQL

echo "=== Seed complete! ==="
echo "Created 6 categories and 20 equipment items."
echo "Barcodes: RF-SPK-001, RF-SUB-001, RF-MIX-001, RF-MIC-001, RF-MIC-002, etc."
