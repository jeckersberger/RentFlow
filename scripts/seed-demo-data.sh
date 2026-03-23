#!/bin/bash
# =============================================================================
# RentFlow Comprehensive Demo Data Seeder
# Populates the database with realistic VT rental company data
# Safe to run multiple times (idempotent via ON CONFLICT DO NOTHING)
# =============================================================================

set -e

POSTGRES_CONTAINER="rentflow-postgres"
POSTGRES_USER="rentflow"

echo "============================================="
echo "  RentFlow Demo Data Seeder"
echo "  Comprehensive dataset for VT rental demo"
echo "============================================="
echo ""

# ---------------------------------------------------------------------------
# 0. Get tenant_id and admin user_id
# ---------------------------------------------------------------------------
TENANT_ID=$(docker exec -i $POSTGRES_CONTAINER psql -U $POSTGRES_USER -d auth_service -t -c "SELECT id FROM auth.tenants LIMIT 1;" | tr -d ' \n\r')
ADMIN_ID=$(docker exec -i $POSTGRES_CONTAINER psql -U $POSTGRES_USER -d auth_service -t -c "SELECT id FROM auth.users LIMIT 1;" | tr -d ' \n\r')

if [ -z "$TENANT_ID" ]; then
  echo "ERROR: No tenant found. Run setup wizard first."
  exit 1
fi

echo "Tenant ID: $TENANT_ID"
echo "Admin ID:  $ADMIN_ID"
echo ""

# ---------------------------------------------------------------------------
# Helper: determine correct project schema (project vs projects)
# ---------------------------------------------------------------------------
PROJECT_SCHEMA=$(docker exec -i $POSTGRES_CONTAINER psql -U $POSTGRES_USER -d project_service -t -c \
  "SELECT schema_name FROM information_schema.schemata WHERE schema_name IN ('project','projects') LIMIT 1;" | tr -d ' \n\r')

if [ -z "$PROJECT_SCHEMA" ]; then
  echo "WARNING: No project schema found. Trying 'projects'..."
  PROJECT_SCHEMA="projects"
fi
echo "Project schema: $PROJECT_SCHEMA"
echo ""

# ---------------------------------------------------------------------------
# 1. Categories (6)
# ---------------------------------------------------------------------------
echo "[1/8] Seeding categories..."
docker exec -i $POSTGRES_CONTAINER psql -U $POSTGRES_USER -d inventory_service -q <<SQL
INSERT INTO inventory.categories (id, tenant_id, name, icon, color, sort_order, created_by_user_id)
VALUES
  ('cat-audio-0001', '$TENANT_ID', 'Audio', '🔊', '#3B82F6', 1, '$ADMIN_ID'),
  ('cat-light-0001', '$TENANT_ID', 'Licht', '💡', '#F59E0B', 2, '$ADMIN_ID'),
  ('cat-video-0001', '$TENANT_ID', 'Video', '🎥', '#8B5CF6', 3, '$ADMIN_ID'),
  ('cat-rigging-001', '$TENANT_ID', 'Rigging/Bühnentechnik', '🏗️', '#EF4444', 4, '$ADMIN_ID'),
  ('cat-power-0001', '$TENANT_ID', 'Strom', '⚡', '#F97316', 5, '$ADMIN_ID'),
  ('cat-cable-0001', '$TENANT_ID', 'Kabel & Zubehör', '🔌', '#6B7280', 6, '$ADMIN_ID')
ON CONFLICT DO NOTHING;
SQL
echo "  -> 6 categories"

# ---------------------------------------------------------------------------
# 2. Equipment (50+ items)
# ---------------------------------------------------------------------------
echo "[2/8] Seeding equipment..."

# Fix serial_number constraint if needed
docker exec -i $POSTGRES_CONTAINER psql -U $POSTGRES_USER -d inventory_service -c \
  "ALTER TABLE inventory.equipment DROP CONSTRAINT IF EXISTS equipment_tenant_id_serial_number_key;" 2>/dev/null || true

docker exec -i $POSTGRES_CONTAINER psql -U $POSTGRES_USER -d inventory_service -q <<SQL

-- ===== AUDIO (15 item types, multiple units) =====

-- JBL VTX A12 Line Array (x8)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-au-001', '$TENANT_ID', 'JBL VTX A12', 'Line Array Lautsprecher, 12" 3-Wege', 'cat-audio-0001', 'AU-A12-001', 'RF-SPK-001', 'available', 'excellent', 150.00, 750.00, 42.0, 8500.00, '2023-03-15', '$ADMIN_ID'),
  ('eq-au-002', '$TENANT_ID', 'JBL VTX A12', 'Line Array Lautsprecher, 12" 3-Wege', 'cat-audio-0001', 'AU-A12-002', 'RF-SPK-002', 'available', 'excellent', 150.00, 750.00, 42.0, 8500.00, '2023-03-15', '$ADMIN_ID'),
  ('eq-au-003', '$TENANT_ID', 'JBL VTX A12', 'Line Array Lautsprecher, 12" 3-Wege', 'cat-audio-0001', 'AU-A12-003', 'RF-SPK-003', 'available', 'excellent', 150.00, 750.00, 42.0, 8500.00, '2023-03-15', '$ADMIN_ID'),
  ('eq-au-004', '$TENANT_ID', 'JBL VTX A12', 'Line Array Lautsprecher, 12" 3-Wege', 'cat-audio-0001', 'AU-A12-004', 'RF-SPK-004', 'available', 'good', 150.00, 750.00, 42.0, 8500.00, '2023-03-15', '$ADMIN_ID'),
  ('eq-au-005', '$TENANT_ID', 'JBL VTX A12', 'Line Array Lautsprecher, 12" 3-Wege', 'cat-audio-0001', 'AU-A12-005', 'RF-SPK-005', 'rented', 'excellent', 150.00, 750.00, 42.0, 8500.00, '2023-06-01', '$ADMIN_ID'),
  ('eq-au-006', '$TENANT_ID', 'JBL VTX A12', 'Line Array Lautsprecher, 12" 3-Wege', 'cat-audio-0001', 'AU-A12-006', 'RF-SPK-006', 'rented', 'excellent', 150.00, 750.00, 42.0, 8500.00, '2023-06-01', '$ADMIN_ID'),
  ('eq-au-007', '$TENANT_ID', 'JBL VTX A12', 'Line Array Lautsprecher, 12" 3-Wege', 'cat-audio-0001', 'AU-A12-007', 'RF-SPK-007', 'available', 'good', 150.00, 750.00, 42.0, 8500.00, '2023-06-01', '$ADMIN_ID'),
  ('eq-au-008', '$TENANT_ID', 'JBL VTX A12', 'Line Array Lautsprecher, 12" 3-Wege', 'cat-audio-0001', 'AU-A12-008', 'RF-SPK-008', 'available', 'excellent', 150.00, 750.00, 42.0, 8500.00, '2023-06-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- JBL VTX S28 Subwoofer (x4)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-au-010', '$TENANT_ID', 'JBL VTX S28', 'Subwoofer, Dual 18"', 'cat-audio-0001', 'AU-S28-001', 'RF-SUB-001', 'available', 'excellent', 120.00, 600.00, 68.0, 6200.00, '2023-03-15', '$ADMIN_ID'),
  ('eq-au-011', '$TENANT_ID', 'JBL VTX S28', 'Subwoofer, Dual 18"', 'cat-audio-0001', 'AU-S28-002', 'RF-SUB-002', 'available', 'excellent', 120.00, 600.00, 68.0, 6200.00, '2023-03-15', '$ADMIN_ID'),
  ('eq-au-012', '$TENANT_ID', 'JBL VTX S28', 'Subwoofer, Dual 18"', 'cat-audio-0001', 'AU-S28-003', 'RF-SUB-003', 'rented', 'good', 120.00, 600.00, 68.0, 6200.00, '2023-03-15', '$ADMIN_ID'),
  ('eq-au-013', '$TENANT_ID', 'JBL VTX S28', 'Subwoofer, Dual 18"', 'cat-audio-0001', 'AU-S28-004', 'RF-SUB-004', 'available', 'good', 120.00, 600.00, 68.0, 6200.00, '2023-06-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- Yamaha CL5 Digitalmischpult (x1)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-au-020', '$TENANT_ID', 'Yamaha CL5', 'Digitalmischpult, 72 Mono + 8 Stereo', 'cat-audio-0001', 'AU-CL5-001', 'RF-MIX-001', 'available', 'excellent', 250.00, 1250.00, 38.0, 22000.00, '2022-11-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- Yamaha Rio3224 Stagebox (x2)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-au-025', '$TENANT_ID', 'Yamaha Rio3224', 'Stagebox 32in/24out, Dante', 'cat-audio-0001', 'AU-RIO-001', 'RF-STG-001', 'available', 'excellent', 80.00, 400.00, 11.0, 5800.00, '2022-11-01', '$ADMIN_ID'),
  ('eq-au-026', '$TENANT_ID', 'Yamaha Rio3224', 'Stagebox 32in/24out, Dante', 'cat-audio-0001', 'AU-RIO-002', 'RF-STG-002', 'available', 'good', 80.00, 400.00, 11.0, 5800.00, '2023-02-10', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- Shure SM58 Mikrofon (x20 - we create 10 individual items for demo)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-au-030', '$TENANT_ID', 'Shure SM58', 'Dynamisches Mikrofon', 'cat-audio-0001', 'AU-SM58-001', 'RF-MIC-001', 'available', 'good', 15.00, 60.00, 0.33, 110.00, '2021-06-01', '$ADMIN_ID'),
  ('eq-au-031', '$TENANT_ID', 'Shure SM58', 'Dynamisches Mikrofon', 'cat-audio-0001', 'AU-SM58-002', 'RF-MIC-002', 'available', 'good', 15.00, 60.00, 0.33, 110.00, '2021-06-01', '$ADMIN_ID'),
  ('eq-au-032', '$TENANT_ID', 'Shure SM58', 'Dynamisches Mikrofon', 'cat-audio-0001', 'AU-SM58-003', 'RF-MIC-003', 'available', 'fair', 15.00, 60.00, 0.33, 110.00, '2021-06-01', '$ADMIN_ID'),
  ('eq-au-033', '$TENANT_ID', 'Shure SM58', 'Dynamisches Mikrofon', 'cat-audio-0001', 'AU-SM58-004', 'RF-MIC-004', 'rented', 'good', 15.00, 60.00, 0.33, 110.00, '2022-01-15', '$ADMIN_ID'),
  ('eq-au-034', '$TENANT_ID', 'Shure SM58', 'Dynamisches Mikrofon', 'cat-audio-0001', 'AU-SM58-005', 'RF-MIC-005', 'available', 'excellent', 15.00, 60.00, 0.33, 110.00, '2022-01-15', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- Shure ULXD4Q Funksystem (x4)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-au-040', '$TENANT_ID', 'Shure ULXD4Q', 'Vierfach-Funksystem mit Handsender', 'cat-audio-0001', 'AU-ULX-001', 'RF-WIR-001', 'available', 'excellent', 95.00, 475.00, 3.5, 4200.00, '2023-01-20', '$ADMIN_ID'),
  ('eq-au-041', '$TENANT_ID', 'Shure ULXD4Q', 'Vierfach-Funksystem mit Handsender', 'cat-audio-0001', 'AU-ULX-002', 'RF-WIR-002', 'available', 'excellent', 95.00, 475.00, 3.5, 4200.00, '2023-01-20', '$ADMIN_ID'),
  ('eq-au-042', '$TENANT_ID', 'Shure ULXD4Q', 'Vierfach-Funksystem mit Taschensender', 'cat-audio-0001', 'AU-ULX-003', 'RF-WIR-003', 'rented', 'good', 95.00, 475.00, 3.5, 4200.00, '2023-05-10', '$ADMIN_ID'),
  ('eq-au-043', '$TENANT_ID', 'Shure ULXD4Q', 'Vierfach-Funksystem mit Taschensender', 'cat-audio-0001', 'AU-ULX-004', 'RF-WIR-004', 'available', 'good', 95.00, 475.00, 3.5, 4200.00, '2023-05-10', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- Sennheiser EW 100 G4 (x6)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-au-050', '$TENANT_ID', 'Sennheiser EW 100 G4', 'Funkmikrofon-Set, UHF', 'cat-audio-0001', 'AU-EW1-001', 'RF-SEN-001', 'available', 'excellent', 45.00, 200.00, 1.2, 680.00, '2022-08-01', '$ADMIN_ID'),
  ('eq-au-051', '$TENANT_ID', 'Sennheiser EW 100 G4', 'Funkmikrofon-Set, UHF', 'cat-audio-0001', 'AU-EW1-002', 'RF-SEN-002', 'available', 'good', 45.00, 200.00, 1.2, 680.00, '2022-08-01', '$ADMIN_ID'),
  ('eq-au-052', '$TENANT_ID', 'Sennheiser EW 100 G4', 'Funkmikrofon-Set, UHF', 'cat-audio-0001', 'AU-EW1-003', 'RF-SEN-003', 'available', 'good', 45.00, 200.00, 1.2, 680.00, '2023-01-15', '$ADMIN_ID'),
  ('eq-au-053', '$TENANT_ID', 'Sennheiser EW 100 G4', 'Funkmikrofon-Set, UHF', 'cat-audio-0001', 'AU-EW1-004', 'RF-SEN-004', 'rented', 'excellent', 45.00, 200.00, 1.2, 680.00, '2023-01-15', '$ADMIN_ID'),
  ('eq-au-054', '$TENANT_ID', 'Sennheiser EW 100 G4', 'Funkmikrofon-Set, Lavalier', 'cat-audio-0001', 'AU-EW1-005', 'RF-SEN-005', 'available', 'excellent', 45.00, 200.00, 1.2, 720.00, '2023-06-01', '$ADMIN_ID'),
  ('eq-au-055', '$TENANT_ID', 'Sennheiser EW 100 G4', 'Funkmikrofon-Set, Lavalier', 'cat-audio-0001', 'AU-EW1-006', 'RF-SEN-006', 'available', 'good', 45.00, 200.00, 1.2, 720.00, '2023-06-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- DI-Box Radial J48 (x10 - create 5)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-au-060', '$TENANT_ID', 'DI-Box Radial J48', 'Active DI Box, Phantom-powered', 'cat-audio-0001', 'AU-DI-001', 'RF-DI-001', 'available', 'excellent', 10.00, 40.00, 0.5, 180.00, '2022-03-01', '$ADMIN_ID'),
  ('eq-au-061', '$TENANT_ID', 'DI-Box Radial J48', 'Active DI Box, Phantom-powered', 'cat-audio-0001', 'AU-DI-002', 'RF-DI-002', 'available', 'good', 10.00, 40.00, 0.5, 180.00, '2022-03-01', '$ADMIN_ID'),
  ('eq-au-062', '$TENANT_ID', 'DI-Box Radial J48', 'Active DI Box, Phantom-powered', 'cat-audio-0001', 'AU-DI-003', 'RF-DI-003', 'available', 'good', 10.00, 40.00, 0.5, 180.00, '2022-03-01', '$ADMIN_ID'),
  ('eq-au-063', '$TENANT_ID', 'DI-Box Radial J48', 'Active DI Box, Phantom-powered', 'cat-audio-0001', 'AU-DI-004', 'RF-DI-004', 'available', 'excellent', 10.00, 40.00, 0.5, 180.00, '2023-02-15', '$ADMIN_ID'),
  ('eq-au-064', '$TENANT_ID', 'DI-Box Radial J48', 'Active DI Box, Phantom-powered', 'cat-audio-0001', 'AU-DI-005', 'RF-DI-005', 'available', 'excellent', 10.00, 40.00, 0.5, 180.00, '2023-02-15', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- Crown iT 12000HD Endstufe (x4)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-au-070', '$TENANT_ID', 'Crown iT 12000HD', 'Endstufe, 2x 4500W/4Ohm', 'cat-audio-0001', 'AU-AMP-001', 'RF-AMP-001', 'available', 'excellent', 85.00, 425.00, 18.5, 4800.00, '2022-09-01', '$ADMIN_ID'),
  ('eq-au-071', '$TENANT_ID', 'Crown iT 12000HD', 'Endstufe, 2x 4500W/4Ohm', 'cat-audio-0001', 'AU-AMP-002', 'RF-AMP-002', 'available', 'excellent', 85.00, 425.00, 18.5, 4800.00, '2022-09-01', '$ADMIN_ID'),
  ('eq-au-072', '$TENANT_ID', 'Crown iT 12000HD', 'Endstufe, 2x 4500W/4Ohm', 'cat-audio-0001', 'AU-AMP-003', 'RF-AMP-003', 'rented', 'good', 85.00, 425.00, 18.5, 4800.00, '2023-04-01', '$ADMIN_ID'),
  ('eq-au-073', '$TENANT_ID', 'Crown iT 12000HD', 'Endstufe, 2x 4500W/4Ohm', 'cat-audio-0001', 'AU-AMP-004', 'RF-AMP-004', 'available', 'good', 85.00, 425.00, 18.5, 4800.00, '2023-04-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- XLR Kabel 10m (x10 of 50)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-au-080', '$TENANT_ID', 'XLR Kabel 10m', 'XLR Kabel, Neutrik, 10 Meter', 'cat-audio-0001', 'AU-XLR-001', 'RF-XLR-001', 'available', 'good', 5.00, 20.00, 0.8, 25.00, '2021-01-01', '$ADMIN_ID'),
  ('eq-au-081', '$TENANT_ID', 'XLR Kabel 10m', 'XLR Kabel, Neutrik, 10 Meter', 'cat-audio-0001', 'AU-XLR-002', 'RF-XLR-002', 'available', 'good', 5.00, 20.00, 0.8, 25.00, '2021-01-01', '$ADMIN_ID'),
  ('eq-au-082', '$TENANT_ID', 'XLR Kabel 10m', 'XLR Kabel, Neutrik, 10 Meter', 'cat-audio-0001', 'AU-XLR-003', 'RF-XLR-003', 'available', 'fair', 5.00, 20.00, 0.8, 25.00, '2021-01-01', '$ADMIN_ID'),
  ('eq-au-083', '$TENANT_ID', 'XLR Kabel 10m', 'XLR Kabel, Neutrik, 10 Meter', 'cat-audio-0001', 'AU-XLR-004', 'RF-XLR-004', 'available', 'good', 5.00, 20.00, 0.8, 25.00, '2022-06-01', '$ADMIN_ID'),
  ('eq-au-084', '$TENANT_ID', 'XLR Kabel 10m', 'XLR Kabel, Neutrik, 10 Meter', 'cat-audio-0001', 'AU-XLR-005', 'RF-XLR-005', 'available', 'good', 5.00, 20.00, 0.8, 25.00, '2022-06-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- Multicore 24/8 30m (x2)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-au-090', '$TENANT_ID', 'Multicore 24/8 30m', 'Multicore Kabel, 24 Send / 8 Return, 30m', 'cat-audio-0001', 'AU-MC-001', 'RF-MC-001', 'available', 'good', 35.00, 140.00, 18.0, 850.00, '2021-09-01', '$ADMIN_ID'),
  ('eq-au-091', '$TENANT_ID', 'Multicore 24/8 30m', 'Multicore Kabel, 24 Send / 8 Return, 30m', 'cat-audio-0001', 'AU-MC-002', 'RF-MC-002', 'available', 'good', 35.00, 140.00, 18.0, 850.00, '2022-04-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- dbx DriveRack PA2 (x2)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-au-095', '$TENANT_ID', 'dbx DriveRack PA2', 'Lautsprecher-Management-System', 'cat-audio-0001', 'AU-DBX-001', 'RF-DBX-001', 'available', 'excellent', 40.00, 160.00, 2.7, 550.00, '2022-06-15', '$ADMIN_ID'),
  ('eq-au-096', '$TENANT_ID', 'dbx DriveRack PA2', 'Lautsprecher-Management-System', 'cat-audio-0001', 'AU-DBX-002', 'RF-DBX-002', 'available', 'good', 40.00, 160.00, 2.7, 550.00, '2023-01-10', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- Pioneer CDJ-3000 (x2)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-au-100', '$TENANT_ID', 'Pioneer CDJ-3000', 'DJ Media Player, ProDJ Link', 'cat-audio-0001', 'AU-CDJ-001', 'RF-CDJ-001', 'available', 'excellent', 75.00, 375.00, 5.0, 2300.00, '2023-08-01', '$ADMIN_ID'),
  ('eq-au-101', '$TENANT_ID', 'Pioneer CDJ-3000', 'DJ Media Player, ProDJ Link', 'cat-audio-0001', 'AU-CDJ-002', 'RF-CDJ-002', 'available', 'excellent', 75.00, 375.00, 5.0, 2300.00, '2023-08-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- Pioneer DJM-900NXS2 (x1)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-au-105', '$TENANT_ID', 'Pioneer DJM-900NXS2', 'DJ Mixer, 4-Kanal, 64bit', 'cat-audio-0001', 'AU-DJM-001', 'RF-DJM-001', 'available', 'excellent', 65.00, 325.00, 5.5, 2100.00, '2023-08-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- Allen & Heath SQ-5 (x1)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-au-110', '$TENANT_ID', 'Allen & Heath SQ-5', 'Digitalmischpult, 48 Kanäle, 96kHz', 'cat-audio-0001', 'AU-SQ5-001', 'RF-SQ5-001', 'available', 'excellent', 180.00, 900.00, 12.0, 4500.00, '2024-01-15', '$ADMIN_ID')
ON CONFLICT DO NOTHING;


-- ===== LICHT (12 item types) =====

-- Clay Paky Sharpy Plus (x8)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-li-001', '$TENANT_ID', 'Clay Paky Sharpy Plus', 'Beam Moving Head, 330W', 'cat-light-0001', 'LI-SHP-001', 'RF-MH-001', 'available', 'excellent', 80.00, 400.00, 21.0, 5200.00, '2023-02-01', '$ADMIN_ID'),
  ('eq-li-002', '$TENANT_ID', 'Clay Paky Sharpy Plus', 'Beam Moving Head, 330W', 'cat-light-0001', 'LI-SHP-002', 'RF-MH-002', 'available', 'excellent', 80.00, 400.00, 21.0, 5200.00, '2023-02-01', '$ADMIN_ID'),
  ('eq-li-003', '$TENANT_ID', 'Clay Paky Sharpy Plus', 'Beam Moving Head, 330W', 'cat-light-0001', 'LI-SHP-003', 'RF-MH-003', 'available', 'good', 80.00, 400.00, 21.0, 5200.00, '2023-02-01', '$ADMIN_ID'),
  ('eq-li-004', '$TENANT_ID', 'Clay Paky Sharpy Plus', 'Beam Moving Head, 330W', 'cat-light-0001', 'LI-SHP-004', 'RF-MH-004', 'rented', 'excellent', 80.00, 400.00, 21.0, 5200.00, '2023-02-01', '$ADMIN_ID'),
  ('eq-li-005', '$TENANT_ID', 'Clay Paky Sharpy Plus', 'Beam Moving Head, 330W', 'cat-light-0001', 'LI-SHP-005', 'RF-MH-005', 'available', 'excellent', 80.00, 400.00, 21.0, 5200.00, '2023-07-01', '$ADMIN_ID'),
  ('eq-li-006', '$TENANT_ID', 'Clay Paky Sharpy Plus', 'Beam Moving Head, 330W', 'cat-light-0001', 'LI-SHP-006', 'RF-MH-006', 'available', 'good', 80.00, 400.00, 21.0, 5200.00, '2023-07-01', '$ADMIN_ID'),
  ('eq-li-007', '$TENANT_ID', 'Clay Paky Sharpy Plus', 'Beam Moving Head, 330W', 'cat-light-0001', 'LI-SHP-007', 'RF-MH-007', 'in_maintenance', 'fair', 80.00, 400.00, 21.0, 5200.00, '2023-07-01', '$ADMIN_ID'),
  ('eq-li-008', '$TENANT_ID', 'Clay Paky Sharpy Plus', 'Beam Moving Head, 330W', 'cat-light-0001', 'LI-SHP-008', 'RF-MH-008', 'available', 'excellent', 80.00, 400.00, 21.0, 5200.00, '2023-07-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- Martin MAC Aura XB (x6)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-li-010', '$TENANT_ID', 'Martin MAC Aura XB', 'Wash Moving Head, LED RGBW', 'cat-light-0001', 'LI-MAC-001', 'RF-MAC-001', 'available', 'excellent', 90.00, 450.00, 8.0, 4800.00, '2022-12-01', '$ADMIN_ID'),
  ('eq-li-011', '$TENANT_ID', 'Martin MAC Aura XB', 'Wash Moving Head, LED RGBW', 'cat-light-0001', 'LI-MAC-002', 'RF-MAC-002', 'available', 'excellent', 90.00, 450.00, 8.0, 4800.00, '2022-12-01', '$ADMIN_ID'),
  ('eq-li-012', '$TENANT_ID', 'Martin MAC Aura XB', 'Wash Moving Head, LED RGBW', 'cat-light-0001', 'LI-MAC-003', 'RF-MAC-003', 'rented', 'good', 90.00, 450.00, 8.0, 4800.00, '2022-12-01', '$ADMIN_ID'),
  ('eq-li-013', '$TENANT_ID', 'Martin MAC Aura XB', 'Wash Moving Head, LED RGBW', 'cat-light-0001', 'LI-MAC-004', 'RF-MAC-004', 'available', 'good', 90.00, 450.00, 8.0, 4800.00, '2023-05-01', '$ADMIN_ID'),
  ('eq-li-014', '$TENANT_ID', 'Martin MAC Aura XB', 'Wash Moving Head, LED RGBW', 'cat-light-0001', 'LI-MAC-005', 'RF-MAC-005', 'available', 'excellent', 90.00, 450.00, 8.0, 4800.00, '2023-05-01', '$ADMIN_ID'),
  ('eq-li-015', '$TENANT_ID', 'Martin MAC Aura XB', 'Wash Moving Head, LED RGBW', 'cat-light-0001', 'LI-MAC-006', 'RF-MAC-006', 'in_maintenance', 'fair', 90.00, 450.00, 8.0, 4800.00, '2023-05-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- GrandMA3 Light (x1)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-li-020', '$TENANT_ID', 'GrandMA3 Light', 'Lichtpult, 8192 Parameter', 'cat-light-0001', 'LI-GMA-001', 'RF-GMA-001', 'available', 'excellent', 200.00, 1000.00, 25.0, 28000.00, '2023-09-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- Chauvet COLORado Panel Q40 (x4 of 8)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-li-025', '$TENANT_ID', 'Chauvet COLORado Panel Q40', 'LED Wash Panel, RGBW', 'cat-light-0001', 'LI-CHV-001', 'RF-PAN-001', 'available', 'excellent', 40.00, 200.00, 7.5, 1800.00, '2023-04-01', '$ADMIN_ID'),
  ('eq-li-026', '$TENANT_ID', 'Chauvet COLORado Panel Q40', 'LED Wash Panel, RGBW', 'cat-light-0001', 'LI-CHV-002', 'RF-PAN-002', 'available', 'good', 40.00, 200.00, 7.5, 1800.00, '2023-04-01', '$ADMIN_ID'),
  ('eq-li-027', '$TENANT_ID', 'Chauvet COLORado Panel Q40', 'LED Wash Panel, RGBW', 'cat-light-0001', 'LI-CHV-003', 'RF-PAN-003', 'available', 'excellent', 40.00, 200.00, 7.5, 1800.00, '2023-04-01', '$ADMIN_ID'),
  ('eq-li-028', '$TENANT_ID', 'Chauvet COLORado Panel Q40', 'LED Wash Panel, RGBW', 'cat-light-0001', 'LI-CHV-004', 'RF-PAN-004', 'rented', 'good', 40.00, 200.00, 7.5, 1800.00, '2023-04-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- ETC Source Four (x4 of 12)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-li-030', '$TENANT_ID', 'ETC Source Four', 'Profilscheinwerfer, 750W, 26°', 'cat-light-0001', 'LI-ETC-001', 'RF-ETC-001', 'available', 'good', 25.00, 100.00, 8.2, 950.00, '2020-06-01', '$ADMIN_ID'),
  ('eq-li-031', '$TENANT_ID', 'ETC Source Four', 'Profilscheinwerfer, 750W, 36°', 'cat-light-0001', 'LI-ETC-002', 'RF-ETC-002', 'available', 'good', 25.00, 100.00, 8.2, 950.00, '2020-06-01', '$ADMIN_ID'),
  ('eq-li-032', '$TENANT_ID', 'ETC Source Four', 'Profilscheinwerfer, 750W, 26°', 'cat-light-0001', 'LI-ETC-003', 'RF-ETC-003', 'available', 'fair', 25.00, 100.00, 8.2, 950.00, '2021-02-01', '$ADMIN_ID'),
  ('eq-li-033', '$TENANT_ID', 'ETC Source Four', 'Profilscheinwerfer, 750W, 50°', 'cat-light-0001', 'LI-ETC-004', 'RF-ETC-004', 'available', 'good', 25.00, 100.00, 8.2, 950.00, '2021-02-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- Eurolite LED Bar RGB (x4 of 10)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-li-040', '$TENANT_ID', 'Eurolite LED Bar RGB', 'LED Leiste, 1m, RGB Wash', 'cat-light-0001', 'LI-BAR-001', 'RF-BAR-001', 'available', 'good', 20.00, 80.00, 2.5, 220.00, '2022-01-01', '$ADMIN_ID'),
  ('eq-li-041', '$TENANT_ID', 'Eurolite LED Bar RGB', 'LED Leiste, 1m, RGB Wash', 'cat-light-0001', 'LI-BAR-002', 'RF-BAR-002', 'available', 'good', 20.00, 80.00, 2.5, 220.00, '2022-01-01', '$ADMIN_ID'),
  ('eq-li-042', '$TENANT_ID', 'Eurolite LED Bar RGB', 'LED Leiste, 1m, RGB Wash', 'cat-light-0001', 'LI-BAR-003', 'RF-BAR-003', 'available', 'fair', 20.00, 80.00, 2.5, 220.00, '2022-01-01', '$ADMIN_ID'),
  ('eq-li-043', '$TENANT_ID', 'Eurolite LED Bar RGB', 'LED Leiste, 1m, RGB Wash', 'cat-light-0001', 'LI-BAR-004', 'RF-BAR-004', 'available', 'good', 20.00, 80.00, 2.5, 220.00, '2023-03-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- MDG theONE Hazer (x2)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-li-050', '$TENANT_ID', 'MDG theONE Hazer', 'Hazer, ölbasiert, DMX-steuerbar', 'cat-light-0001', 'LI-HAZ-001', 'RF-HAZ-001', 'available', 'excellent', 60.00, 300.00, 15.0, 3200.00, '2023-03-01', '$ADMIN_ID'),
  ('eq-li-051', '$TENANT_ID', 'MDG theONE Hazer', 'Hazer, ölbasiert, DMX-steuerbar', 'cat-light-0001', 'LI-HAZ-002', 'RF-HAZ-002', 'available', 'good', 60.00, 300.00, 15.0, 3200.00, '2023-08-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- Follow Spot Robert Juliat (x2)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-li-055', '$TENANT_ID', 'Follow Spot Robert Juliat', 'Verfolgerscheinwerfer, 2500W HMI', 'cat-light-0001', 'LI-FOL-001', 'RF-FOL-001', 'available', 'excellent', 55.00, 275.00, 22.0, 4500.00, '2022-05-01', '$ADMIN_ID'),
  ('eq-li-056', '$TENANT_ID', 'Follow Spot Robert Juliat', 'Verfolgerscheinwerfer, 2500W HMI', 'cat-light-0001', 'LI-FOL-002', 'RF-FOL-002', 'available', 'good', 55.00, 275.00, 22.0, 4500.00, '2022-05-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- DMX Kabel 10m (x5 of 30)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-li-060', '$TENANT_ID', 'DMX Kabel 10m', 'DMX Kabel 5-pol, 10m', 'cat-light-0001', 'LI-DMX-001', 'RF-DMX-001', 'available', 'good', 3.00, 12.00, 0.5, 18.00, '2021-01-01', '$ADMIN_ID'),
  ('eq-li-061', '$TENANT_ID', 'DMX Kabel 10m', 'DMX Kabel 5-pol, 10m', 'cat-light-0001', 'LI-DMX-002', 'RF-DMX-002', 'available', 'good', 3.00, 12.00, 0.5, 18.00, '2021-01-01', '$ADMIN_ID'),
  ('eq-li-062', '$TENANT_ID', 'DMX Kabel 10m', 'DMX Kabel 5-pol, 10m', 'cat-light-0001', 'LI-DMX-003', 'RF-DMX-003', 'available', 'fair', 3.00, 12.00, 0.5, 18.00, '2022-03-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- Powercon Kabel 5m (x3 of 20)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-li-065', '$TENANT_ID', 'Powercon Kabel 5m', 'Powercon Kabel TRUE1, 5m', 'cat-light-0001', 'LI-PWC-001', 'RF-PWC-001', 'available', 'good', 5.00, 20.00, 0.6, 22.00, '2021-01-01', '$ADMIN_ID'),
  ('eq-li-066', '$TENANT_ID', 'Powercon Kabel 5m', 'Powercon Kabel TRUE1, 5m', 'cat-light-0001', 'LI-PWC-002', 'RF-PWC-002', 'available', 'good', 5.00, 20.00, 0.6, 22.00, '2022-06-01', '$ADMIN_ID'),
  ('eq-li-067', '$TENANT_ID', 'Powercon Kabel 5m', 'Powercon Kabel TRUE1, 5m', 'cat-light-0001', 'LI-PWC-003', 'RF-PWC-003', 'available', 'fair', 5.00, 20.00, 0.6, 22.00, '2022-06-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- Dimmer Rack 12ch (x2)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-li-070', '$TENANT_ID', 'Dimmer Rack 12ch', 'Dimmer-Rack, 12 Kanäle, je 2.3kW', 'cat-light-0001', 'LI-DIM-001', 'RF-DIM-001', 'available', 'good', 45.00, 225.00, 32.0, 2800.00, '2021-08-01', '$ADMIN_ID'),
  ('eq-li-071', '$TENANT_ID', 'Dimmer Rack 12ch', 'Dimmer-Rack, 12 Kanäle, je 2.3kW', 'cat-light-0001', 'LI-DIM-002', 'RF-DIM-002', 'available', 'good', 45.00, 225.00, 32.0, 2800.00, '2022-02-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- LED Par 64 RGBW (x4 of 16)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-li-075', '$TENANT_ID', 'LED Par 64 RGBW', 'LED PAR, 18x10W RGBW', 'cat-light-0001', 'LI-PAR-001', 'RF-PAR-001', 'available', 'good', 15.00, 60.00, 3.5, 280.00, '2022-01-01', '$ADMIN_ID'),
  ('eq-li-076', '$TENANT_ID', 'LED Par 64 RGBW', 'LED PAR, 18x10W RGBW', 'cat-light-0001', 'LI-PAR-002', 'RF-PAR-002', 'available', 'good', 15.00, 60.00, 3.5, 280.00, '2022-01-01', '$ADMIN_ID'),
  ('eq-li-077', '$TENANT_ID', 'LED Par 64 RGBW', 'LED PAR, 18x10W RGBW', 'cat-light-0001', 'LI-PAR-003', 'RF-PAR-003', 'available', 'excellent', 15.00, 60.00, 3.5, 280.00, '2023-01-01', '$ADMIN_ID'),
  ('eq-li-078', '$TENANT_ID', 'LED Par 64 RGBW', 'LED PAR, 18x10W RGBW', 'cat-light-0001', 'LI-PAR-004', 'RF-PAR-004', 'rented', 'good', 15.00, 60.00, 3.5, 280.00, '2023-01-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;


-- ===== VIDEO (8 item types) =====

-- Panasonic PT-RZ690 Beamer (x2)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-vi-001', '$TENANT_ID', 'Panasonic PT-RZ690', 'Laser-Beamer, 6000 ANSI Lumen, WUXGA', 'cat-video-0001', 'VI-PAN-001', 'RF-BEA-001', 'available', 'excellent', 180.00, 900.00, 18.0, 6500.00, '2023-05-01', '$ADMIN_ID'),
  ('eq-vi-002', '$TENANT_ID', 'Panasonic PT-RZ690', 'Laser-Beamer, 6000 ANSI Lumen, WUXGA', 'cat-video-0001', 'VI-PAN-002', 'RF-BEA-002', 'available', 'excellent', 180.00, 900.00, 18.0, 6500.00, '2023-09-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- Samsung 75" Display (x2)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-vi-005', '$TENANT_ID', 'Samsung 75" Display', 'LED Display 75 Zoll, 4K UHD, inkl. Standfuß', 'cat-video-0001', 'VI-SAM-001', 'RF-DIS-001', 'available', 'excellent', 120.00, 600.00, 35.0, 3800.00, '2023-06-01', '$ADMIN_ID'),
  ('eq-vi-006', '$TENANT_ID', 'Samsung 75" Display', 'LED Display 75 Zoll, 4K UHD, inkl. Standfuß', 'cat-video-0001', 'VI-SAM-002', 'RF-DIS-002', 'rented', 'good', 120.00, 600.00, 35.0, 3800.00, '2024-01-15', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- Blackmagic ATEM Mini Pro (x2)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-vi-010', '$TENANT_ID', 'Blackmagic ATEM Mini Pro', 'Video-Mischer, 4x HDMI, Streaming', 'cat-video-0001', 'VI-BM-001', 'RF-ATM-001', 'available', 'excellent', 45.00, 225.00, 0.9, 550.00, '2023-03-01', '$ADMIN_ID'),
  ('eq-vi-011', '$TENANT_ID', 'Blackmagic ATEM Mini Pro', 'Video-Mischer, 4x HDMI, Streaming', 'cat-video-0001', 'VI-BM-002', 'RF-ATM-002', 'available', 'good', 45.00, 225.00, 0.9, 550.00, '2023-08-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- PTZ Kamera Sony SRG-300H (x3)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-vi-015', '$TENANT_ID', 'PTZ Kamera Sony SRG-300H', 'PTZ Kamera, 30x Zoom, HD-SDI', 'cat-video-0001', 'VI-PTZ-001', 'RF-PTZ-001', 'available', 'excellent', 85.00, 425.00, 2.3, 3200.00, '2023-04-01', '$ADMIN_ID'),
  ('eq-vi-016', '$TENANT_ID', 'PTZ Kamera Sony SRG-300H', 'PTZ Kamera, 30x Zoom, HD-SDI', 'cat-video-0001', 'VI-PTZ-002', 'RF-PTZ-002', 'available', 'good', 85.00, 425.00, 2.3, 3200.00, '2023-04-01', '$ADMIN_ID'),
  ('eq-vi-017', '$TENANT_ID', 'PTZ Kamera Sony SRG-300H', 'PTZ Kamera, 30x Zoom, HD-SDI', 'cat-video-0001', 'VI-PTZ-003', 'RF-PTZ-003', 'available', 'excellent', 85.00, 425.00, 2.3, 3200.00, '2024-01-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- HDMI Kabel 10m (x3 of 10)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-vi-020', '$TENANT_ID', 'HDMI Kabel 10m', 'HDMI 2.0 Kabel, 4K/60Hz, 10m', 'cat-video-0001', 'VI-HDM-001', 'RF-HDM-001', 'available', 'good', 8.00, 32.00, 0.6, 35.00, '2022-01-01', '$ADMIN_ID'),
  ('eq-vi-021', '$TENANT_ID', 'HDMI Kabel 10m', 'HDMI 2.0 Kabel, 4K/60Hz, 10m', 'cat-video-0001', 'VI-HDM-002', 'RF-HDM-002', 'available', 'good', 8.00, 32.00, 0.6, 35.00, '2022-01-01', '$ADMIN_ID'),
  ('eq-vi-022', '$TENANT_ID', 'HDMI Kabel 10m', 'HDMI 2.0 Kabel, 4K/60Hz, 10m', 'cat-video-0001', 'VI-HDM-003', 'RF-HDM-003', 'available', 'fair', 8.00, 32.00, 0.6, 35.00, '2023-06-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- SDI Kabel 20m (x3 of 6)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-vi-025', '$TENANT_ID', 'SDI Kabel 20m', 'SDI Kabel BNC, 3G-SDI, 20m', 'cat-video-0001', 'VI-SDI-001', 'RF-SDI-001', 'available', 'good', 12.00, 48.00, 1.2, 55.00, '2022-03-01', '$ADMIN_ID'),
  ('eq-vi-026', '$TENANT_ID', 'SDI Kabel 20m', 'SDI Kabel BNC, 3G-SDI, 20m', 'cat-video-0001', 'VI-SDI-002', 'RF-SDI-002', 'available', 'good', 12.00, 48.00, 1.2, 55.00, '2022-03-01', '$ADMIN_ID'),
  ('eq-vi-027', '$TENANT_ID', 'SDI Kabel 20m', 'SDI Kabel BNC, 3G-SDI, 20m', 'cat-video-0001', 'VI-SDI-003', 'RF-SDI-003', 'available', 'excellent', 12.00, 48.00, 1.2, 55.00, '2023-09-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- Leinwand 3x2m (x2)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-vi-030', '$TENANT_ID', 'Leinwand 3x2m', 'Aufprojektion, Stativ, 3x2m, matt-weiß', 'cat-video-0001', 'VI-SCR-001', 'RF-SCR-001', 'available', 'good', 35.00, 140.00, 12.0, 450.00, '2021-11-01', '$ADMIN_ID'),
  ('eq-vi-031', '$TENANT_ID', 'Leinwand 3x2m', 'Aufprojektion, Stativ, 3x2m, matt-weiß', 'cat-video-0001', 'VI-SCR-002', 'RF-SCR-002', 'available', 'good', 35.00, 140.00, 12.0, 450.00, '2023-02-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- Video Splitter 1:4 (x3)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-vi-035', '$TENANT_ID', 'Video Splitter 1:4', 'HDMI Splitter 1 Eingang, 4 Ausgänge, 4K', 'cat-video-0001', 'VI-SPL-001', 'RF-SPL-001', 'available', 'excellent', 15.00, 60.00, 0.5, 120.00, '2023-01-01', '$ADMIN_ID'),
  ('eq-vi-036', '$TENANT_ID', 'Video Splitter 1:4', 'HDMI Splitter 1 Eingang, 4 Ausgänge, 4K', 'cat-video-0001', 'VI-SPL-002', 'RF-SPL-002', 'available', 'good', 15.00, 60.00, 0.5, 120.00, '2023-01-01', '$ADMIN_ID'),
  ('eq-vi-037', '$TENANT_ID', 'Video Splitter 1:4', 'HDMI Splitter 1 Eingang, 4 Ausgänge, 4K', 'cat-video-0001', 'VI-SPL-003', 'RF-SPL-003', 'available', 'good', 15.00, 60.00, 0.5, 120.00, '2023-07-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;


-- ===== RIGGING / BÜHNENTECHNIK (8 item types) =====

-- Global Truss F34 3m (x5 of 20)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-ri-001', '$TENANT_ID', 'Global Truss F34 3m', 'Traverse, Aluminium, F34, 3 Meter', 'cat-rigging-001', 'RI-T3M-001', 'RF-TRS-001', 'available', 'good', 25.00, 100.00, 15.0, 380.00, '2020-03-01', '$ADMIN_ID'),
  ('eq-ri-002', '$TENANT_ID', 'Global Truss F34 3m', 'Traverse, Aluminium, F34, 3 Meter', 'cat-rigging-001', 'RI-T3M-002', 'RF-TRS-002', 'available', 'good', 25.00, 100.00, 15.0, 380.00, '2020-03-01', '$ADMIN_ID'),
  ('eq-ri-003', '$TENANT_ID', 'Global Truss F34 3m', 'Traverse, Aluminium, F34, 3 Meter', 'cat-rigging-001', 'RI-T3M-003', 'RF-TRS-003', 'rented', 'good', 25.00, 100.00, 15.0, 380.00, '2021-06-01', '$ADMIN_ID'),
  ('eq-ri-004', '$TENANT_ID', 'Global Truss F34 3m', 'Traverse, Aluminium, F34, 3 Meter', 'cat-rigging-001', 'RI-T3M-004', 'RF-TRS-004', 'available', 'excellent', 25.00, 100.00, 15.0, 380.00, '2022-09-01', '$ADMIN_ID'),
  ('eq-ri-005', '$TENANT_ID', 'Global Truss F34 3m', 'Traverse, Aluminium, F34, 3 Meter', 'cat-rigging-001', 'RI-T3M-005', 'RF-TRS-005', 'available', 'good', 25.00, 100.00, 15.0, 380.00, '2022-09-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- Global Truss F34 2m (x3 of 10)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-ri-010', '$TENANT_ID', 'Global Truss F34 2m', 'Traverse, Aluminium, F34, 2 Meter', 'cat-rigging-001', 'RI-T2M-001', 'RF-T2M-001', 'available', 'good', 20.00, 80.00, 10.0, 280.00, '2020-03-01', '$ADMIN_ID'),
  ('eq-ri-011', '$TENANT_ID', 'Global Truss F34 2m', 'Traverse, Aluminium, F34, 2 Meter', 'cat-rigging-001', 'RI-T2M-002', 'RF-T2M-002', 'available', 'good', 20.00, 80.00, 10.0, 280.00, '2021-06-01', '$ADMIN_ID'),
  ('eq-ri-012', '$TENANT_ID', 'Global Truss F34 2m', 'Traverse, Aluminium, F34, 2 Meter', 'cat-rigging-001', 'RI-T2M-003', 'RF-T2M-003', 'available', 'excellent', 20.00, 80.00, 10.0, 280.00, '2022-09-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- Chainmaster BGV-D8+ 1t (x4)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-ri-020', '$TENANT_ID', 'Chainmaster BGV-D8+ 1t', 'Kettenzug, 1 Tonne, BGV-D8+', 'cat-rigging-001', 'RI-CM-001', 'RF-CM-001', 'available', 'excellent', 55.00, 275.00, 28.0, 3800.00, '2022-01-01', '$ADMIN_ID'),
  ('eq-ri-021', '$TENANT_ID', 'Chainmaster BGV-D8+ 1t', 'Kettenzug, 1 Tonne, BGV-D8+', 'cat-rigging-001', 'RI-CM-002', 'RF-CM-002', 'available', 'excellent', 55.00, 275.00, 28.0, 3800.00, '2022-01-01', '$ADMIN_ID'),
  ('eq-ri-022', '$TENANT_ID', 'Chainmaster BGV-D8+ 1t', 'Kettenzug, 1 Tonne, BGV-D8+', 'cat-rigging-001', 'RI-CM-003', 'RF-CM-003', 'rented', 'good', 55.00, 275.00, 28.0, 3800.00, '2023-03-01', '$ADMIN_ID'),
  ('eq-ri-023', '$TENANT_ID', 'Chainmaster BGV-D8+ 1t', 'Kettenzug, 1 Tonne, BGV-D8+', 'cat-rigging-001', 'RI-CM-004', 'RF-CM-004', 'available', 'good', 55.00, 275.00, 28.0, 3800.00, '2023-03-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- Ladebalken 6m (x4)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-ri-030', '$TENANT_ID', 'Ladebalken 6m', 'Ladebalken, Stahl, 6 Meter', 'cat-rigging-001', 'RI-LB-001', 'RF-LB-001', 'available', 'good', 30.00, 120.00, 25.0, 600.00, '2021-04-01', '$ADMIN_ID'),
  ('eq-ri-031', '$TENANT_ID', 'Ladebalken 6m', 'Ladebalken, Stahl, 6 Meter', 'cat-rigging-001', 'RI-LB-002', 'RF-LB-002', 'available', 'good', 30.00, 120.00, 25.0, 600.00, '2021-04-01', '$ADMIN_ID'),
  ('eq-ri-032', '$TENANT_ID', 'Ladebalken 6m', 'Ladebalken, Stahl, 6 Meter', 'cat-rigging-001', 'RI-LB-003', 'RF-LB-003', 'available', 'fair', 30.00, 120.00, 25.0, 600.00, '2022-07-01', '$ADMIN_ID'),
  ('eq-ri-033', '$TENANT_ID', 'Ladebalken 6m', 'Ladebalken, Stahl, 6 Meter', 'cat-rigging-001', 'RI-LB-004', 'RF-LB-004', 'rented', 'good', 30.00, 120.00, 25.0, 600.00, '2022-07-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- Absperrgitter Mojo (x5 of 20)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-ri-040', '$TENANT_ID', 'Absperrgitter Mojo', 'Mojo Barrier, Aluminium, 1.2m', 'cat-rigging-001', 'RI-MJ-001', 'RF-MJ-001', 'available', 'good', 8.00, 32.00, 22.0, 420.00, '2020-06-01', '$ADMIN_ID'),
  ('eq-ri-041', '$TENANT_ID', 'Absperrgitter Mojo', 'Mojo Barrier, Aluminium, 1.2m', 'cat-rigging-001', 'RI-MJ-002', 'RF-MJ-002', 'available', 'good', 8.00, 32.00, 22.0, 420.00, '2020-06-01', '$ADMIN_ID'),
  ('eq-ri-042', '$TENANT_ID', 'Absperrgitter Mojo', 'Mojo Barrier, Aluminium, 1.2m', 'cat-rigging-001', 'RI-MJ-003', 'RF-MJ-003', 'available', 'fair', 8.00, 32.00, 22.0, 420.00, '2021-03-01', '$ADMIN_ID'),
  ('eq-ri-043', '$TENANT_ID', 'Absperrgitter Mojo', 'Mojo Barrier, Aluminium, 1.2m', 'cat-rigging-001', 'RI-MJ-004', 'RF-MJ-004', 'rented', 'good', 8.00, 32.00, 22.0, 420.00, '2022-05-01', '$ADMIN_ID'),
  ('eq-ri-044', '$TENANT_ID', 'Absperrgitter Mojo', 'Mojo Barrier, Aluminium, 1.2m', 'cat-rigging-001', 'RI-MJ-005', 'RF-MJ-005', 'available', 'good', 8.00, 32.00, 22.0, 420.00, '2022-05-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- Bühnenpodest 2x1m (x4 of 12)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-ri-050', '$TENANT_ID', 'Bühnenpodest 2x1m', 'Bühnenpodest, Aluminium, 2x1m, Höhe variabel', 'cat-rigging-001', 'RI-BP-001', 'RF-BP-001', 'available', 'good', 18.00, 72.00, 35.0, 650.00, '2020-09-01', '$ADMIN_ID'),
  ('eq-ri-051', '$TENANT_ID', 'Bühnenpodest 2x1m', 'Bühnenpodest, Aluminium, 2x1m, Höhe variabel', 'cat-rigging-001', 'RI-BP-002', 'RF-BP-002', 'available', 'good', 18.00, 72.00, 35.0, 650.00, '2020-09-01', '$ADMIN_ID'),
  ('eq-ri-052', '$TENANT_ID', 'Bühnenpodest 2x1m', 'Bühnenpodest, Aluminium, 2x1m, Höhe variabel', 'cat-rigging-001', 'RI-BP-003', 'RF-BP-003', 'available', 'fair', 18.00, 72.00, 35.0, 650.00, '2021-04-01', '$ADMIN_ID'),
  ('eq-ri-053', '$TENANT_ID', 'Bühnenpodest 2x1m', 'Bühnenpodest, Aluminium, 2x1m, Höhe variabel', 'cat-rigging-001', 'RI-BP-004', 'RF-BP-004', 'rented', 'good', 18.00, 72.00, 35.0, 650.00, '2022-01-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- Treppenstufe (x3 of 8)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-ri-060', '$TENANT_ID', 'Treppenstufe', 'Treppenstufe für Bühnenpodest, 1m breit', 'cat-rigging-001', 'RI-TS-001', 'RF-TS-001', 'available', 'good', 12.00, 48.00, 15.0, 280.00, '2020-09-01', '$ADMIN_ID'),
  ('eq-ri-061', '$TENANT_ID', 'Treppenstufe', 'Treppenstufe für Bühnenpodest, 1m breit', 'cat-rigging-001', 'RI-TS-002', 'RF-TS-002', 'available', 'good', 12.00, 48.00, 15.0, 280.00, '2021-04-01', '$ADMIN_ID'),
  ('eq-ri-062', '$TENANT_ID', 'Treppenstufe', 'Treppenstufe für Bühnenpodest, 1m breit', 'cat-rigging-001', 'RI-TS-003', 'RF-TS-003', 'available', 'fair', 12.00, 48.00, 15.0, 280.00, '2022-01-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- Molton schwarz 6x3m (x3 of 6)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-ri-070', '$TENANT_ID', 'Molton schwarz 6x3m', 'Bühnenmolton, schwarz, 6x3m, B1', 'cat-rigging-001', 'RI-MOL-001', 'RF-MOL-001', 'available', 'good', 15.00, 60.00, 8.0, 180.00, '2021-01-01', '$ADMIN_ID'),
  ('eq-ri-071', '$TENANT_ID', 'Molton schwarz 6x3m', 'Bühnenmolton, schwarz, 6x3m, B1', 'cat-rigging-001', 'RI-MOL-002', 'RF-MOL-002', 'available', 'good', 15.00, 60.00, 8.0, 180.00, '2022-06-01', '$ADMIN_ID'),
  ('eq-ri-072', '$TENANT_ID', 'Molton schwarz 6x3m', 'Bühnenmolton, schwarz, 6x3m, B1', 'cat-rigging-001', 'RI-MOL-003', 'RF-MOL-003', 'available', 'fair', 15.00, 60.00, 8.0, 180.00, '2022-06-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;


-- ===== STROM (5 item types) =====

-- Powerlock 125A Verteiler (x2)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-pw-001', '$TENANT_ID', 'Powerlock 125A Verteiler', 'Stromverteiler 125A Powerlock, 6x CEE 63A', 'cat-power-0001', 'PW-PL-001', 'RF-PWR-001', 'available', 'excellent', 60.00, 300.00, 45.0, 3500.00, '2022-03-01', '$ADMIN_ID'),
  ('eq-pw-002', '$TENANT_ID', 'Powerlock 125A Verteiler', 'Stromverteiler 125A Powerlock, 6x CEE 63A', 'cat-power-0001', 'PW-PL-002', 'RF-PWR-002', 'available', 'good', 60.00, 300.00, 45.0, 3500.00, '2023-01-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- CEE 63A Verteiler (x3)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-pw-005', '$TENANT_ID', 'CEE 63A Verteiler', 'Stromverteiler CEE 63A auf 6x CEE 32A', 'cat-power-0001', 'PW-C63-001', 'RF-C63-001', 'available', 'good', 35.00, 140.00, 18.0, 1200.00, '2021-06-01', '$ADMIN_ID'),
  ('eq-pw-006', '$TENANT_ID', 'CEE 63A Verteiler', 'Stromverteiler CEE 63A auf 6x CEE 32A', 'cat-power-0001', 'PW-C63-002', 'RF-C63-002', 'available', 'good', 35.00, 140.00, 18.0, 1200.00, '2022-04-01', '$ADMIN_ID'),
  ('eq-pw-007', '$TENANT_ID', 'CEE 63A Verteiler', 'Stromverteiler CEE 63A auf 6x CEE 32A', 'cat-power-0001', 'PW-C63-003', 'RF-C63-003', 'rented', 'excellent', 35.00, 140.00, 18.0, 1200.00, '2023-02-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- CEE 32A Verlängerung 25m (x3 of 6)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-pw-010', '$TENANT_ID', 'CEE 32A Verlängerung 25m', 'CEE 32A Verlängerung, H07RN-F 5G6, 25m', 'cat-power-0001', 'PW-C32-001', 'RF-C32-001', 'available', 'good', 15.00, 60.00, 8.5, 180.00, '2021-01-01', '$ADMIN_ID'),
  ('eq-pw-011', '$TENANT_ID', 'CEE 32A Verlängerung 25m', 'CEE 32A Verlängerung, H07RN-F 5G6, 25m', 'cat-power-0001', 'PW-C32-002', 'RF-C32-002', 'available', 'good', 15.00, 60.00, 8.5, 180.00, '2022-03-01', '$ADMIN_ID'),
  ('eq-pw-012', '$TENANT_ID', 'CEE 32A Verlängerung 25m', 'CEE 32A Verlängerung, H07RN-F 5G6, 25m', 'cat-power-0001', 'PW-C32-003', 'RF-C32-003', 'available', 'fair', 15.00, 60.00, 8.5, 180.00, '2022-03-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- Stromkabel 16A 10m (x3 of 20)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-pw-015', '$TENANT_ID', 'Stromkabel 16A 10m', 'Schuko-Verlängerung H07RN-F 3G1.5, 10m', 'cat-power-0001', 'PW-S16-001', 'RF-S16-001', 'available', 'good', 5.00, 20.00, 2.0, 30.00, '2021-01-01', '$ADMIN_ID'),
  ('eq-pw-016', '$TENANT_ID', 'Stromkabel 16A 10m', 'Schuko-Verlängerung H07RN-F 3G1.5, 10m', 'cat-power-0001', 'PW-S16-002', 'RF-S16-002', 'available', 'good', 5.00, 20.00, 2.0, 30.00, '2022-06-01', '$ADMIN_ID'),
  ('eq-pw-017', '$TENANT_ID', 'Stromkabel 16A 10m', 'Schuko-Verlängerung H07RN-F 3G1.5, 10m', 'cat-power-0001', 'PW-S16-003', 'RF-S16-003', 'available', 'fair', 5.00, 20.00, 2.0, 30.00, '2022-06-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

-- USV APC 1500VA (x2)
INSERT INTO inventory.equipment (id, tenant_id, name, description, category_id, sku, barcode, status, condition, rental_price_day, rental_price_week, weight, purchase_price, purchase_date, created_by_user_id)
VALUES
  ('eq-pw-020', '$TENANT_ID', 'USV APC 1500VA', 'USV APC Smart-UPS 1500VA, Line-Interactive', 'cat-power-0001', 'PW-USV-001', 'RF-USV-001', 'available', 'excellent', 25.00, 100.00, 22.0, 850.00, '2023-06-01', '$ADMIN_ID'),
  ('eq-pw-021', '$TENANT_ID', 'USV APC 1500VA', 'USV APC Smart-UPS 1500VA, Line-Interactive', 'cat-power-0001', 'PW-USV-002', 'RF-USV-002', 'available', 'good', 25.00, 100.00, 22.0, 850.00, '2024-01-01', '$ADMIN_ID')
ON CONFLICT DO NOTHING;

SQL
echo "  -> 100+ equipment items across 5 categories"

# ---------------------------------------------------------------------------
# 3. Contacts (10)
# ---------------------------------------------------------------------------
echo "[3/8] Seeding contacts..."
docker exec -i $POSTGRES_CONTAINER psql -U $POSTGRES_USER -d project_service -q <<SQL
INSERT INTO ${PROJECT_SCHEMA}.contacts (id, tenant_id, type, company_name, first_name, last_name, email, phone, mobile, website, street, house_number, zip, city, country, vat_id, notes, tags, created_by)
VALUES
  ('con-001', '$TENANT_ID', 'company', 'Festival GmbH', 'Thomas', 'Berger', 'thomas.berger@festival-gmbh.de', '+49 89 1234567', '+49 170 1234567', 'www.festival-gmbh.de', 'Leopoldstraße', '24', '80802', 'München', 'Deutschland', 'DE123456789', 'Stammkunde seit 2020, jährlich Stadtfest + Sommerfestival', '{Stammkunde,Festival,Großkunde}', '$ADMIN_ID'),
  ('con-002', '$TENANT_ID', 'company', 'Stage Entertainment AG', 'Sabine', 'Kowalski', 'sabine.kowalski@stage-entertainment.de', '+49 40 9876543', '+49 171 9876543', 'www.stage-entertainment.de', 'Reeperbahn', '1', '20359', 'Hamburg', 'Deutschland', 'DE987654321', 'Musical-Produktionen, regelmäßige Buchungen', '{Stammkunde,Theater,Musical}', '$ADMIN_ID'),
  ('con-003', '$TENANT_ID', 'company', 'Kongresshaus Zürich', 'Marcel', 'Bühler', 'marcel.buehler@kongresshaus-zh.ch', '+41 44 2061500', '+41 79 3456789', 'www.kongresshaus.ch', 'Claridenstrasse', '5', '8002', 'Zürich', 'Schweiz', 'CHE-123.456.789', 'Internationale Konferenzen, Dolmetschertechnik nötig', '{International,Konferenz,Schweiz}', '$ADMIN_ID'),
  ('con-004', '$TENANT_ID', 'company', 'Messe Frankfurt GmbH', 'Klaus', 'Richter', 'klaus.richter@messefrankfurt.com', '+49 69 75750', '+49 172 5678901', 'www.messefrankfurt.com', 'Ludwig-Erhard-Anlage', '1', '60327', 'Frankfurt', 'Deutschland', 'DE234567890', 'Messestände und Konferenztechnik', '{Messe,Großkunde,Video}', '$ADMIN_ID'),
  ('con-005', '$TENANT_ID', 'company', 'Stadttheater Wien', 'Elisabeth', 'Gruber', 'e.gruber@stadttheater-wien.at', '+43 1 51444', '+43 664 1234567', 'www.stadttheater-wien.at', 'Museumstraße', '7', '1070', 'Wien', 'Österreich', 'ATU12345678', 'Theaterproduktionen, Licht + Ton', '{Theater,Österreich,Stammkunde}', '$ADMIN_ID'),
  ('con-006', '$TENANT_ID', 'person', '', 'Marco', 'Di Lorenzo', 'marco@djmarco.de', '+49 89 5551234', '+49 175 5551234', 'www.djmarco.de', 'Schillerstraße', '42', '80336', 'München', 'Deutschland', '', 'DJ, braucht regelmäßig CDJs + Mixer + kleine PA', '{DJ,Freelancer,München}', '$ADMIN_ID'),
  ('con-007', '$TENANT_ID', 'company', 'Autohaus Müller', 'Stefan', 'Müller', 'stefan.mueller@autohaus-mueller.de', '+49 6221 778899', '+49 173 7788990', 'www.autohaus-mueller.de', 'Eppelheimer Straße', '88', '69115', 'Heidelberg', 'Deutschland', 'DE345678901', 'Jährliche Firmenfeiern, Neuwagenpräsentationen', '{Firmenevent,Heidelberg}', '$ADMIN_ID'),
  ('con-008', '$TENANT_ID', 'person', '', 'Lisa', 'Weber', 'lisa@weber-hochzeiten.de', '+49 621 4455667', '+49 176 4455667', 'www.weber-hochzeiten.de', 'Friedrichsplatz', '12', '68165', 'Mannheim', 'Deutschland', '', 'Hochzeitsplanerin, 5-8 Hochzeiten pro Saison', '{Hochzeit,Freelancer,Mannheim}', '$ADMIN_ID'),
  ('con-009', '$TENANT_ID', 'company', 'Sportverein TSV 1860', 'Markus', 'Eder', 'markus.eder@tsv1860.de', '+49 89 6420000', '+49 174 6420000', 'www.tsv1860.de', 'Grünwalder Straße', '114', '81547', 'München', 'Deutschland', 'DE456789012', 'Vereinsfeste, Hallenveranstaltungen', '{Verein,Sport,München}', '$ADMIN_ID'),
  ('con-010', '$TENANT_ID', 'company', 'Brauerei Hofbräu', 'Hans', 'Wimmer', 'hans.wimmer@hofbraeu.de', '+49 89 92105', '+49 170 9210500', 'www.hofbraeu.de', 'Innere Wiener Straße', '19', '81667', 'München', 'Deutschland', 'DE567890123', 'Oktoberfest, Starkbierfest, große Bühnenproduktionen', '{Großkunde,Oktoberfest,München,Festival}', '$ADMIN_ID')
ON CONFLICT DO NOTHING;
SQL
echo "  -> 10 contacts"

# ---------------------------------------------------------------------------
# 4. Projects (8)
# ---------------------------------------------------------------------------
echo "[4/8] Seeding projects..."
docker exec -i $POSTGRES_CONTAINER psql -U $POSTGRES_USER -d project_service -q <<SQL
INSERT INTO ${PROJECT_SCHEMA}.projects (id, tenant_id, name, description, client_name, client_email, client_phone, venue_street, venue_city, venue_postal_code, venue_country, status, start_date, end_date, setup_date, teardown_date, budget, currency, notes, tags, created_at, updated_at, created_by_user_id)
VALUES
  -- 1. Stadtfest Heidelberg — confirmed, next week
  ('proj-demo-001', '$TENANT_ID', 'Stadtfest Heidelberg 2026', 'Großes Stadtfest auf dem Marktplatz, Hauptbühne + Nebenbühne. Full PA mit Line Array, Lichtdesign inkl. Moving Heads, Haze.', 'Festival GmbH', 'thomas.berger@festival-gmbh.de', '+49 89 1234567', 'Marktplatz', 'Heidelberg', '69117', 'DE', 'confirmed', '2026-03-28 10:00:00+01', '2026-03-29 23:00:00+01', '2026-03-28 06:00:00+01', '2026-03-30 08:00:00+01', 15000.00, 'EUR', 'Stromversorgung über städtische Anschlüsse. Ansprechpartner vor Ort: Hr. Keller (Stadt Heidelberg). Aufbau ab 6 Uhr, Abbau Sonntag früh.', '{Festival,Outdoor,Audio,Licht}', NOW() - INTERVAL '14 days', NOW() - INTERVAL '2 days', '$ADMIN_ID'),

  -- 2. Firmenfeier Autohaus Müller — confirmed, in 2 weeks
  ('proj-demo-002', '$TENANT_ID', 'Firmenfeier Autohaus Müller', 'Jährliche Firmenweihnachtsfeier im Showroom. PA-Anlage für Reden + Musik, Beamer für Präsentation, Ambientebeleuchtung.', 'Autohaus Müller', 'stefan.mueller@autohaus-mueller.de', '+49 6221 778899', 'Eppelheimer Straße 88', 'Heidelberg', '69115', 'DE', 'confirmed', '2026-04-04 17:00:00+01', '2026-04-05 01:00:00+01', '2026-04-04 14:00:00+01', '2026-04-05 09:00:00+01', 4500.00, 'EUR', 'Ca. 120 Gäste. Funkmikrofone für Reden des GF. DJ ab 21 Uhr. Beamer auf Leinwand für Jahresrückblick-Video.', '{Firmenevent,Indoor,Audio,Video}', NOW() - INTERVAL '10 days', NOW() - INTERVAL '3 days', '$ADMIN_ID'),

  -- 3. Festival am See — in_progress (packed), this weekend
  ('proj-demo-003', '$TENANT_ID', 'Festival am See 2026', 'Open-Air Festival am Bodensee, 3 Tage. Full Production: Main Stage PA (JBL VTX A12 Array + Subs), komplett Lichtdesign mit 16 Moving Heads, 2 Verfolger, Video-Regie mit 3 Kameras und Beamer.', 'Festival GmbH', 'thomas.berger@festival-gmbh.de', '+49 89 1234567', 'Seestraße', 'Konstanz', '78462', 'DE', 'in_progress', '2026-03-20 14:00:00+01', '2026-03-22 23:00:00+01', '2026-03-19 08:00:00+01', '2026-03-23 18:00:00+01', 35000.00, 'EUR', 'WICHTIG: Regenplan beachten! Zelte für FOH + Monitorregie. Strom: 125A Powerlock vom Trafo. 2 LKW-Ladungen. Crew: 6 Techniker.', '{Festival,Outdoor,FullProduction,Großevent}', NOW() - INTERVAL '30 days', NOW() - INTERVAL '1 day', '$ADMIN_ID'),

  -- 4. Hochzeit Weber/Schmidt — draft, in 3 weeks
  ('proj-demo-004', '$TENANT_ID', 'Hochzeit Weber/Schmidt', 'Hochzeitsfeier in Scheune, ca. 80 Gäste. Kleine PA für Zeremonie + Party, 2 Funkmikrofone, LED Uplighting, DJ-Setup.', 'Lisa Weber', 'lisa@weber-hochzeiten.de', '+49 176 4455667', 'Landgut am Bach', 'Schwetzingen', '68723', 'DE', 'draft', '2026-04-11 14:00:00+01', '2026-04-12 02:00:00+01', '2026-04-11 10:00:00+01', '2026-04-12 10:00:00+01', 2500.00, 'EUR', 'Kontakt Brautpaar: Weber/Schmidt. Trauung outdoor (Backup indoor). DJ bringt eigenen Laptop. Warmweiße Beleuchtung gewünscht.', '{Hochzeit,Indoor,Audio,Licht}', NOW() - INTERVAL '5 days', NOW() - INTERVAL '1 day', '$ADMIN_ID'),

  -- 5. Messe Frankfurt Standbau — confirmed, next month
  ('proj-demo-005', '$TENANT_ID', 'Messe Frankfurt – TechExpo Stand', 'Messestand 120qm: 3x Samsung 75" Displays, 2x Beamer auf Leinwand, Beschallungsanlage für Vorträge, LED-Ambientebeleuchtung.', 'Messe Frankfurt GmbH', 'klaus.richter@messefrankfurt.com', '+49 69 75750', 'Ludwig-Erhard-Anlage 1', 'Frankfurt', '60327', 'DE', 'confirmed', '2026-04-20 09:00:00+01', '2026-04-23 18:00:00+01', '2026-04-19 08:00:00+01', '2026-04-24 16:00:00+01', 12000.00, 'EUR', '4 Tage Messe. Displays müssen täglich 10h laufen. Netzwerk für Streaming wird vom Veranstalter gestellt. Aufbau nur So 8-20 Uhr.', '{Messe,Indoor,Video,Licht}', NOW() - INTERVAL '21 days', NOW() - INTERVAL '5 days', '$ADMIN_ID'),

  -- 6. Konzert Stadttheater Wien — completed, last week
  ('proj-demo-006', '$TENANT_ID', 'Konzert Stadttheater Wien', 'Klassikkonzert mit Verstärkung. Dezentes Lichtdesign, 12 Funkmikrofone für Orchester-Pickup, Beschallung für 800 Plätze.', 'Stadttheater Wien', 'e.gruber@stadttheater-wien.at', '+43 1 51444', 'Museumstraße 7', 'Wien', '1070', 'AT', 'completed', '2026-03-14 19:00:00+01', '2026-03-14 22:30:00+01', '2026-03-14 08:00:00+01', '2026-03-15 12:00:00+01', 10000.00, 'EUR', 'Sehr guter Ablauf. Akustik im Saal hervorragend. Wiederholung für Herbstsaison geplant. 4 Techniker vor Ort.', '{Konzert,Theater,Indoor,Audio,Licht}', NOW() - INTERVAL '45 days', NOW() - INTERVAL '8 days', '$ADMIN_ID'),

  -- 7. DJ Event Club Nachtwerk — cancelled
  ('proj-demo-007', '$TENANT_ID', 'DJ Event Club Nachtwerk', 'DJ-Nacht im Club Nachtwerk, CDJ-Setup + kleine PA-Ergänzung.', 'Marco Di Lorenzo', 'marco@djmarco.de', '+49 175 5551234', 'Landsberger Straße 185', 'München', '80687', 'DE', 'cancelled', '2026-03-08 22:00:00+01', '2026-03-09 05:00:00+01', '2026-03-08 18:00:00+01', '2026-03-09 10:00:00+01', 1500.00, 'EUR', 'STORNIERT: Club hat kurzfristig abgesagt wegen Wasserschaden. Stornogebühr 30% vereinbart.', '{DJ,Club,Storniert}', NOW() - INTERVAL '25 days', NOW() - INTERVAL '20 days', '$ADMIN_ID'),

  -- 8. Konferenz Zürich — confirmed, in 2 weeks
  ('proj-demo-008', '$TENANT_ID', 'Konferenz Zürich 2026', 'Internationale Konferenz, 2 Tage. Hauptsaal: PA + 2 Beamer + Kamera-Setup für Live-Stream. 3 Breakout-Räume mit je 1 Display.', 'Kongresshaus Zürich', 'marcel.buehler@kongresshaus-zh.ch', '+41 44 2061500', 'Claridenstrasse 5', 'Zürich', '8002', 'CH', 'confirmed', '2026-04-06 08:00:00+01', '2026-04-07 18:00:00+01', '2026-04-05 14:00:00+01', '2026-04-08 12:00:00+01', 18000.00, 'EUR', 'Simultanübersetzung DE/EN/FR. Live-Streaming auf YouTube. 300 Teilnehmer. Kontakt Technik vor Ort: Marcel Bühler.', '{Konferenz,International,Video,Audio,Streaming}', NOW() - INTERVAL '18 days', NOW() - INTERVAL '4 days', '$ADMIN_ID')
ON CONFLICT DO NOTHING;
SQL
echo "  -> 8 projects"

# ---------------------------------------------------------------------------
# 5. Invoices (5)
# We need to handle the immutable_finalized constraint:
# temporarily disable it, insert data, re-enable
# ---------------------------------------------------------------------------
echo "[5/8] Seeding invoices..."
docker exec -i $POSTGRES_CONTAINER psql -U $POSTGRES_USER -d invoice_service -q <<SQL

-- Temporarily drop the immutable_finalized constraint for seeding
ALTER TABLE invoice.invoices DROP CONSTRAINT IF EXISTS immutable_finalized;

-- Insert invoices
INSERT INTO invoice.invoices (id, tenant_id, invoice_number, project_id, client_name, client_address_street, client_address_city, client_address_postcode, client_address_country, client_email, client_tax_id, sub_total, tax_rate, tax_amount, total, currency, status, issue_date, due_date, paid_date, payment_method, notes, hash, created_at, updated_at)
VALUES
  -- RE-2024-001: Stadtfest Heidelberg — sent
  ('inv-demo-001', '$TENANT_ID', 'RE-2026-001', 'proj-demo-001', 'Festival GmbH', 'Leopoldstraße 24', 'München', '80802', 'Deutschland', 'thomas.berger@festival-gmbh.de', 'DE123456789',
   10462.18, 19.00, 1987.82, 12450.00, 'EUR', 'sent',
   NOW() - INTERVAL '10 days', NOW() + INTERVAL '20 days', NULL, NULL,
   'Stadtfest Heidelberg 2026 - Komplettpaket Audio + Licht',
   md5('inv-demo-001-sent'), NOW() - INTERVAL '10 days', NOW() - INTERVAL '10 days'),

  -- RE-2024-002: Firmenfeier — paid
  ('inv-demo-002', '$TENANT_ID', 'RE-2026-002', 'proj-demo-002', 'Autohaus Müller', 'Eppelheimer Straße 88', 'Heidelberg', '69115', 'Deutschland', 'stefan.mueller@autohaus-mueller.de', 'DE345678901',
   2756.30, 19.00, 523.70, 3280.00, 'EUR', 'paid',
   NOW() - INTERVAL '30 days', NOW() - INTERVAL '2 days', NOW() - INTERVAL '5 days', 'bank_transfer',
   'Firmenfeier Autohaus Müller - Audio + Video',
   md5('inv-demo-002-paid'), NOW() - INTERVAL '30 days', NOW() - INTERVAL '5 days'),

  -- RE-2024-003: Festival am See — overdue
  ('inv-demo-003', '$TENANT_ID', 'RE-2026-003', 'proj-demo-003', 'Festival GmbH', 'Leopoldstraße 24', 'München', '80802', 'Deutschland', 'thomas.berger@festival-gmbh.de', 'DE123456789',
   24285.71, 19.00, 4614.29, 28900.00, 'EUR', 'overdue',
   NOW() - INTERVAL '45 days', NOW() - INTERVAL '15 days', NULL, NULL,
   'Festival am See 2026 - Full Production (Audio, Licht, Video, Rigging, Strom)',
   md5('inv-demo-003-overdue'), NOW() - INTERVAL '45 days', NOW() - INTERVAL '45 days'),

  -- RE-2024-004: Konzert Wien — paid
  ('inv-demo-004', '$TENANT_ID', 'RE-2026-004', 'proj-demo-006', 'Stadttheater Wien', 'Museumstraße 7', 'Wien', '1070', 'Österreich', 'e.gruber@stadttheater-wien.at', 'ATU12345678',
   7352.94, 19.00, 1397.06, 8750.00, 'EUR', 'paid',
   NOW() - INTERVAL '10 days', NOW() + INTERVAL '20 days', NOW() - INTERVAL '3 days', 'bank_transfer',
   'Konzert Stadttheater Wien - Audio + Licht',
   md5('inv-demo-004-paid'), NOW() - INTERVAL '10 days', NOW() - INTERVAL '3 days'),

  -- RE-2024-005: DJ Event — cancelled
  ('inv-demo-005', '$TENANT_ID', 'RE-2026-005', 'proj-demo-007', 'Marco Di Lorenzo', 'Schillerstraße 42', 'München', '80336', 'Deutschland', 'marco@djmarco.de', '',
   1008.40, 19.00, 191.60, 1200.00, 'EUR', 'cancelled',
   NOW() - INTERVAL '25 days', NOW() + INTERVAL '5 days', NULL, NULL,
   'DJ Event Club Nachtwerk - STORNIERT (Stornogebühr 30%)',
   md5('inv-demo-005-cancelled'), NOW() - INTERVAL '25 days', NOW() - INTERVAL '20 days')
ON CONFLICT DO NOTHING;

-- Re-add the constraint (only applies to new inserts)
ALTER TABLE invoice.invoices ADD CONSTRAINT immutable_finalized CHECK (status = 'draft') NOT VALID;

-- Seed invoice items for each invoice
INSERT INTO invoice.invoice_items (id, invoice_id, description, quantity, unit, unit_price, total_price, tax_rate, equipment_id)
VALUES
  -- Stadtfest Heidelberg items
  ('ii-001-01', 'inv-demo-001', 'JBL VTX A12 Line Array (8 Stk, 2 Tage)', 16, 'Stk/Tag', 150.00, 2400.00, 19.00, 'eq-au-001'),
  ('ii-001-02', 'inv-demo-001', 'JBL VTX S28 Subwoofer (4 Stk, 2 Tage)', 8, 'Stk/Tag', 120.00, 960.00, 19.00, 'eq-au-010'),
  ('ii-001-03', 'inv-demo-001', 'Yamaha CL5 Digitalmischpult (2 Tage)', 2, 'Tag', 250.00, 500.00, 19.00, 'eq-au-020'),
  ('ii-001-04', 'inv-demo-001', 'Clay Paky Sharpy Plus (8 Stk, 2 Tage)', 16, 'Stk/Tag', 80.00, 1280.00, 19.00, 'eq-li-001'),
  ('ii-001-05', 'inv-demo-001', 'Martin MAC Aura XB (6 Stk, 2 Tage)', 12, 'Stk/Tag', 90.00, 1080.00, 19.00, 'eq-li-010'),
  ('ii-001-06', 'inv-demo-001', 'GrandMA3 Light (2 Tage)', 2, 'Tag', 200.00, 400.00, 19.00, 'eq-li-020'),
  ('ii-001-07', 'inv-demo-001', 'MDG theONE Hazer (2 Stk, 2 Tage)', 4, 'Stk/Tag', 60.00, 240.00, 19.00, 'eq-li-050'),
  ('ii-001-08', 'inv-demo-001', 'Techniker-Pauschale (3 Techniker, 2 Tage)', 6, 'Person/Tag', 350.00, 2100.00, 19.00, NULL),
  ('ii-001-09', 'inv-demo-001', 'Transport Heidelberg (hin + zurück)', 1, 'Pauschale', 450.00, 450.00, 19.00, NULL),

  -- Firmenfeier items
  ('ii-002-01', 'inv-demo-002', 'Allen & Heath SQ-5 (1 Tag)', 1, 'Tag', 180.00, 180.00, 19.00, 'eq-au-110'),
  ('ii-002-02', 'inv-demo-002', 'Sennheiser EW 100 G4 (2 Stk, 1 Tag)', 2, 'Stk/Tag', 45.00, 90.00, 19.00, 'eq-au-050'),
  ('ii-002-03', 'inv-demo-002', 'Panasonic PT-RZ690 Beamer (1 Tag)', 1, 'Tag', 180.00, 180.00, 19.00, 'eq-vi-001'),
  ('ii-002-04', 'inv-demo-002', 'Leinwand 3x2m (1 Tag)', 1, 'Tag', 35.00, 35.00, 19.00, 'eq-vi-030'),
  ('ii-002-05', 'inv-demo-002', 'LED Par 64 RGBW (8 Stk, 1 Tag)', 8, 'Stk/Tag', 15.00, 120.00, 19.00, 'eq-li-075'),
  ('ii-002-06', 'inv-demo-002', 'Pioneer CDJ-3000 (2 Stk, 1 Tag)', 2, 'Stk/Tag', 75.00, 150.00, 19.00, 'eq-au-100'),
  ('ii-002-07', 'inv-demo-002', 'Pioneer DJM-900NXS2 (1 Tag)', 1, 'Tag', 65.00, 65.00, 19.00, 'eq-au-105'),
  ('ii-002-08', 'inv-demo-002', 'Techniker (1 Person, 1 Tag)', 1, 'Person/Tag', 350.00, 350.00, 19.00, NULL),
  ('ii-002-09', 'inv-demo-002', 'Transport Heidelberg', 1, 'Pauschale', 150.00, 150.00, 19.00, NULL),

  -- Festival am See items
  ('ii-003-01', 'inv-demo-003', 'JBL VTX A12 Line Array (8 Stk, 3 Tage)', 24, 'Stk/Tag', 150.00, 3600.00, 19.00, 'eq-au-001'),
  ('ii-003-02', 'inv-demo-003', 'JBL VTX S28 Subwoofer (4 Stk, 3 Tage)', 12, 'Stk/Tag', 120.00, 1440.00, 19.00, 'eq-au-010'),
  ('ii-003-03', 'inv-demo-003', 'Yamaha CL5 + Rio3224 (3 Tage)', 3, 'Tag', 330.00, 990.00, 19.00, 'eq-au-020'),
  ('ii-003-04', 'inv-demo-003', 'Shure ULXD4Q Funksysteme (4 Stk, 3 Tage)', 12, 'Stk/Tag', 95.00, 1140.00, 19.00, 'eq-au-040'),
  ('ii-003-05', 'inv-demo-003', 'Clay Paky Sharpy Plus (8 Stk, 3 Tage)', 24, 'Stk/Tag', 80.00, 1920.00, 19.00, 'eq-li-001'),
  ('ii-003-06', 'inv-demo-003', 'Martin MAC Aura XB (6 Stk, 3 Tage)', 18, 'Stk/Tag', 90.00, 1620.00, 19.00, 'eq-li-010'),
  ('ii-003-07', 'inv-demo-003', 'GrandMA3 Light (3 Tage)', 3, 'Tag', 200.00, 600.00, 19.00, 'eq-li-020'),
  ('ii-003-08', 'inv-demo-003', 'Follow Spot Robert Juliat (2 Stk, 3 Tage)', 6, 'Stk/Tag', 55.00, 330.00, 19.00, 'eq-li-055'),
  ('ii-003-09', 'inv-demo-003', 'PTZ Kamera Sony (3 Stk, 3 Tage)', 9, 'Stk/Tag', 85.00, 765.00, 19.00, 'eq-vi-015'),
  ('ii-003-10', 'inv-demo-003', 'Panasonic PT-RZ690 Beamer (2 Stk, 3 Tage)', 6, 'Stk/Tag', 180.00, 1080.00, 19.00, 'eq-vi-001'),
  ('ii-003-11', 'inv-demo-003', 'Chainmaster BGV-D8+ (4 Stk, 3 Tage)', 12, 'Stk/Tag', 55.00, 660.00, 19.00, 'eq-ri-020'),
  ('ii-003-12', 'inv-demo-003', 'Global Truss F34 (10 Stk, 3 Tage)', 30, 'Stk/Tag', 25.00, 750.00, 19.00, 'eq-ri-001'),
  ('ii-003-13', 'inv-demo-003', 'Powerlock 125A Verteiler (3 Tage)', 3, 'Tag', 60.00, 180.00, 19.00, 'eq-pw-001'),
  ('ii-003-14', 'inv-demo-003', 'Techniker-Crew (6 Personen, 4 Tage inkl. Aufbau)', 24, 'Person/Tag', 380.00, 9120.00, 19.00, NULL),
  ('ii-003-15', 'inv-demo-003', 'Transport Konstanz (2 LKW, hin + zurück)', 1, 'Pauschale', 1800.00, 1800.00, 19.00, NULL),

  -- Konzert Wien items
  ('ii-004-01', 'inv-demo-004', 'Yamaha CL5 + Rio3224 (1 Tag)', 1, 'Tag', 330.00, 330.00, 19.00, 'eq-au-020'),
  ('ii-004-02', 'inv-demo-004', 'Sennheiser EW 100 G4 (6 Stk, 1 Tag)', 6, 'Stk/Tag', 45.00, 270.00, 19.00, 'eq-au-050'),
  ('ii-004-03', 'inv-demo-004', 'Shure ULXD4Q (4 Stk, 1 Tag)', 4, 'Stk/Tag', 95.00, 380.00, 19.00, 'eq-au-040'),
  ('ii-004-04', 'inv-demo-004', 'Martin MAC Aura XB (6 Stk, 1 Tag)', 6, 'Stk/Tag', 90.00, 540.00, 19.00, 'eq-li-010'),
  ('ii-004-05', 'inv-demo-004', 'ETC Source Four (12 Stk, 1 Tag)', 12, 'Stk/Tag', 25.00, 300.00, 19.00, 'eq-li-030'),
  ('ii-004-06', 'inv-demo-004', 'Dimmer Rack 12ch (2 Stk, 1 Tag)', 2, 'Stk/Tag', 45.00, 90.00, 19.00, 'eq-li-070'),
  ('ii-004-07', 'inv-demo-004', 'Techniker (4 Personen, 1 Tag)', 4, 'Person/Tag', 380.00, 1520.00, 19.00, NULL),
  ('ii-004-08', 'inv-demo-004', 'Transport Wien (hin + zurück)', 1, 'Pauschale', 850.00, 850.00, 19.00, NULL),

  -- DJ Event items (cancelled)
  ('ii-005-01', 'inv-demo-005', 'Pioneer CDJ-3000 (2 Stk, 1 Nacht)', 2, 'Stk', 75.00, 150.00, 19.00, 'eq-au-100'),
  ('ii-005-02', 'inv-demo-005', 'Pioneer DJM-900NXS2 (1 Nacht)', 1, 'Stk', 65.00, 65.00, 19.00, 'eq-au-105'),
  ('ii-005-03', 'inv-demo-005', 'JBL VTX A12 (2 Stk, Ergänzung)', 2, 'Stk', 150.00, 300.00, 19.00, 'eq-au-001'),
  ('ii-005-04', 'inv-demo-005', 'JBL VTX S28 Subwoofer (2 Stk)', 2, 'Stk', 120.00, 240.00, 19.00, 'eq-au-010'),
  ('ii-005-05', 'inv-demo-005', 'Transport München', 1, 'Pauschale', 120.00, 120.00, 19.00, NULL),
  ('ii-005-06', 'inv-demo-005', 'Techniker (1 Person, Abend)', 1, 'Pauschale', 250.00, 250.00, 19.00, NULL)
ON CONFLICT DO NOTHING;

-- Update number sequence so next invoice gets correct number
INSERT INTO invoice.number_sequences (tenant_id, sequence_type, next_value)
VALUES ('$TENANT_ID', 'invoice', 6)
ON CONFLICT (tenant_id, sequence_type) DO UPDATE SET next_value = GREATEST(invoice.number_sequences.next_value, 6);

SQL
echo "  -> 5 invoices with line items"

# ---------------------------------------------------------------------------
# 6. Packlists for in_progress project (Festival am See)
# ---------------------------------------------------------------------------
echo "[6/8] Seeding packlists..."
docker exec -i $POSTGRES_CONTAINER psql -U $POSTGRES_USER -d project_service -q <<SQL
INSERT INTO ${PROJECT_SCHEMA}.packlists (id, project_id, tenant_id, name, status, created_at, updated_at)
VALUES
  ('pl-demo-001', 'proj-demo-003', '$TENANT_ID', 'Audio Main Stage', 'packed', NOW() - INTERVAL '3 days', NOW() - INTERVAL '1 day'),
  ('pl-demo-002', 'proj-demo-003', '$TENANT_ID', 'Licht Main Stage', 'packed', NOW() - INTERVAL '3 days', NOW() - INTERVAL '1 day'),
  ('pl-demo-003', 'proj-demo-003', '$TENANT_ID', 'Video + Streaming', 'in_progress', NOW() - INTERVAL '3 days', NOW()),
  ('pl-demo-004', 'proj-demo-003', '$TENANT_ID', 'Rigging + Strom', 'packed', NOW() - INTERVAL '3 days', NOW() - INTERVAL '1 day'),
  ('pl-demo-005', 'proj-demo-001', '$TENANT_ID', 'Stadtfest Komplett', 'draft', NOW() - INTERVAL '2 days', NOW())
ON CONFLICT DO NOTHING;

INSERT INTO ${PROJECT_SCHEMA}.packlist_items (id, packlist_id, equipment_id, equipment_name, quantity, quantity_packed, quantity_returned, status, created_at, updated_at)
VALUES
  -- Audio Main Stage (packed)
  ('pli-001', 'pl-demo-001', 'eq-au-001', 'JBL VTX A12', 8, 8, 0, 'packed', NOW() - INTERVAL '2 days', NOW() - INTERVAL '1 day'),
  ('pli-002', 'pl-demo-001', 'eq-au-010', 'JBL VTX S28', 4, 4, 0, 'packed', NOW() - INTERVAL '2 days', NOW() - INTERVAL '1 day'),
  ('pli-003', 'pl-demo-001', 'eq-au-020', 'Yamaha CL5', 1, 1, 0, 'packed', NOW() - INTERVAL '2 days', NOW() - INTERVAL '1 day'),
  ('pli-004', 'pl-demo-001', 'eq-au-025', 'Yamaha Rio3224', 2, 2, 0, 'packed', NOW() - INTERVAL '2 days', NOW() - INTERVAL '1 day'),
  ('pli-005', 'pl-demo-001', 'eq-au-040', 'Shure ULXD4Q', 4, 4, 0, 'packed', NOW() - INTERVAL '2 days', NOW() - INTERVAL '1 day'),
  ('pli-006', 'pl-demo-001', 'eq-au-070', 'Crown iT 12000HD', 4, 4, 0, 'packed', NOW() - INTERVAL '2 days', NOW() - INTERVAL '1 day'),

  -- Licht Main Stage (packed)
  ('pli-010', 'pl-demo-002', 'eq-li-001', 'Clay Paky Sharpy Plus', 8, 8, 0, 'packed', NOW() - INTERVAL '2 days', NOW() - INTERVAL '1 day'),
  ('pli-011', 'pl-demo-002', 'eq-li-010', 'Martin MAC Aura XB', 6, 6, 0, 'packed', NOW() - INTERVAL '2 days', NOW() - INTERVAL '1 day'),
  ('pli-012', 'pl-demo-002', 'eq-li-020', 'GrandMA3 Light', 1, 1, 0, 'packed', NOW() - INTERVAL '2 days', NOW() - INTERVAL '1 day'),
  ('pli-013', 'pl-demo-002', 'eq-li-050', 'MDG theONE Hazer', 2, 2, 0, 'packed', NOW() - INTERVAL '2 days', NOW() - INTERVAL '1 day'),
  ('pli-014', 'pl-demo-002', 'eq-li-055', 'Follow Spot Robert Juliat', 2, 2, 0, 'packed', NOW() - INTERVAL '2 days', NOW() - INTERVAL '1 day'),

  -- Video + Streaming (in_progress)
  ('pli-020', 'pl-demo-003', 'eq-vi-001', 'Panasonic PT-RZ690', 2, 2, 0, 'packed', NOW() - INTERVAL '2 days', NOW()),
  ('pli-021', 'pl-demo-003', 'eq-vi-015', 'PTZ Kamera Sony SRG-300H', 3, 2, 0, 'in_progress', NOW() - INTERVAL '2 days', NOW()),
  ('pli-022', 'pl-demo-003', 'eq-vi-010', 'Blackmagic ATEM Mini Pro', 1, 1, 0, 'packed', NOW() - INTERVAL '2 days', NOW()),
  ('pli-023', 'pl-demo-003', 'eq-vi-030', 'Leinwand 3x2m', 2, 0, 0, 'pending', NOW() - INTERVAL '1 day', NOW()),

  -- Rigging + Strom (packed)
  ('pli-030', 'pl-demo-004', 'eq-ri-001', 'Global Truss F34 3m', 10, 10, 0, 'packed', NOW() - INTERVAL '2 days', NOW() - INTERVAL '1 day'),
  ('pli-031', 'pl-demo-004', 'eq-ri-020', 'Chainmaster BGV-D8+', 4, 4, 0, 'packed', NOW() - INTERVAL '2 days', NOW() - INTERVAL '1 day'),
  ('pli-032', 'pl-demo-004', 'eq-pw-001', 'Powerlock 125A Verteiler', 2, 2, 0, 'packed', NOW() - INTERVAL '2 days', NOW() - INTERVAL '1 day'),
  ('pli-033', 'pl-demo-004', 'eq-pw-005', 'CEE 63A Verteiler', 3, 3, 0, 'packed', NOW() - INTERVAL '2 days', NOW() - INTERVAL '1 day')
ON CONFLICT DO NOTHING;
SQL
echo "  -> 5 packlists with items"

# ---------------------------------------------------------------------------
# 7. Reservations
# ---------------------------------------------------------------------------
echo "[7/8] Seeding reservations..."
docker exec -i $POSTGRES_CONTAINER psql -U $POSTGRES_USER -d project_service -q <<SQL
INSERT INTO ${PROJECT_SCHEMA}.reservations (id, tenant_id, project_id, equipment_id, start_date, end_date, status, created_at, updated_at)
VALUES
  -- Stadtfest Heidelberg reservations
  ('res-001', '$TENANT_ID', 'proj-demo-001', 'eq-au-001', '2026-03-28 06:00:00+01', '2026-03-30 08:00:00+01', 'confirmed', NOW() - INTERVAL '10 days', NOW() - INTERVAL '10 days'),
  ('res-002', '$TENANT_ID', 'proj-demo-001', 'eq-au-010', '2026-03-28 06:00:00+01', '2026-03-30 08:00:00+01', 'confirmed', NOW() - INTERVAL '10 days', NOW() - INTERVAL '10 days'),
  ('res-003', '$TENANT_ID', 'proj-demo-001', 'eq-au-020', '2026-03-28 06:00:00+01', '2026-03-30 08:00:00+01', 'confirmed', NOW() - INTERVAL '10 days', NOW() - INTERVAL '10 days'),
  ('res-004', '$TENANT_ID', 'proj-demo-001', 'eq-li-001', '2026-03-28 06:00:00+01', '2026-03-30 08:00:00+01', 'confirmed', NOW() - INTERVAL '10 days', NOW() - INTERVAL '10 days'),
  ('res-005', '$TENANT_ID', 'proj-demo-001', 'eq-li-020', '2026-03-28 06:00:00+01', '2026-03-30 08:00:00+01', 'confirmed', NOW() - INTERVAL '10 days', NOW() - INTERVAL '10 days'),

  -- Festival am See (in_progress)
  ('res-010', '$TENANT_ID', 'proj-demo-003', 'eq-au-005', '2026-03-19 08:00:00+01', '2026-03-23 18:00:00+01', 'active', NOW() - INTERVAL '15 days', NOW() - INTERVAL '3 days'),
  ('res-011', '$TENANT_ID', 'proj-demo-003', 'eq-au-006', '2026-03-19 08:00:00+01', '2026-03-23 18:00:00+01', 'active', NOW() - INTERVAL '15 days', NOW() - INTERVAL '3 days'),
  ('res-012', '$TENANT_ID', 'proj-demo-003', 'eq-au-012', '2026-03-19 08:00:00+01', '2026-03-23 18:00:00+01', 'active', NOW() - INTERVAL '15 days', NOW() - INTERVAL '3 days'),
  ('res-013', '$TENANT_ID', 'proj-demo-003', 'eq-li-004', '2026-03-19 08:00:00+01', '2026-03-23 18:00:00+01', 'active', NOW() - INTERVAL '15 days', NOW() - INTERVAL '3 days'),
  ('res-014', '$TENANT_ID', 'proj-demo-003', 'eq-li-012', '2026-03-19 08:00:00+01', '2026-03-23 18:00:00+01', 'active', NOW() - INTERVAL '15 days', NOW() - INTERVAL '3 days'),
  ('res-015', '$TENANT_ID', 'proj-demo-003', 'eq-ri-003', '2026-03-19 08:00:00+01', '2026-03-23 18:00:00+01', 'active', NOW() - INTERVAL '15 days', NOW() - INTERVAL '3 days'),
  ('res-016', '$TENANT_ID', 'proj-demo-003', 'eq-ri-022', '2026-03-19 08:00:00+01', '2026-03-23 18:00:00+01', 'active', NOW() - INTERVAL '15 days', NOW() - INTERVAL '3 days'),
  ('res-017', '$TENANT_ID', 'proj-demo-003', 'eq-pw-007', '2026-03-19 08:00:00+01', '2026-03-23 18:00:00+01', 'active', NOW() - INTERVAL '15 days', NOW() - INTERVAL '3 days'),

  -- Konferenz Zürich
  ('res-020', '$TENANT_ID', 'proj-demo-008', 'eq-vi-001', '2026-04-05 14:00:00+01', '2026-04-08 12:00:00+01', 'confirmed', NOW() - INTERVAL '7 days', NOW() - INTERVAL '7 days'),
  ('res-021', '$TENANT_ID', 'proj-demo-008', 'eq-vi-005', '2026-04-05 14:00:00+01', '2026-04-08 12:00:00+01', 'confirmed', NOW() - INTERVAL '7 days', NOW() - INTERVAL '7 days'),
  ('res-022', '$TENANT_ID', 'proj-demo-008', 'eq-vi-015', '2026-04-05 14:00:00+01', '2026-04-08 12:00:00+01', 'confirmed', NOW() - INTERVAL '7 days', NOW() - INTERVAL '7 days'),
  ('res-023', '$TENANT_ID', 'proj-demo-008', 'eq-au-110', '2026-04-05 14:00:00+01', '2026-04-08 12:00:00+01', 'confirmed', NOW() - INTERVAL '7 days', NOW() - INTERVAL '7 days'),

  -- Messe Frankfurt
  ('res-030', '$TENANT_ID', 'proj-demo-005', 'eq-vi-005', '2026-04-19 08:00:00+01', '2026-04-24 16:00:00+01', 'confirmed', NOW() - INTERVAL '14 days', NOW() - INTERVAL '14 days'),
  ('res-031', '$TENANT_ID', 'proj-demo-005', 'eq-vi-006', '2026-04-19 08:00:00+01', '2026-04-24 16:00:00+01', 'confirmed', NOW() - INTERVAL '14 days', NOW() - INTERVAL '14 days'),
  ('res-032', '$TENANT_ID', 'proj-demo-005', 'eq-vi-001', '2026-04-19 08:00:00+01', '2026-04-24 16:00:00+01', 'confirmed', NOW() - INTERVAL '14 days', NOW() - INTERVAL '14 days'),

  -- Cancelled DJ Event
  ('res-040', '$TENANT_ID', 'proj-demo-007', 'eq-au-100', '2026-03-08 18:00:00+01', '2026-03-09 10:00:00+01', 'cancelled', NOW() - INTERVAL '25 days', NOW() - INTERVAL '20 days'),
  ('res-041', '$TENANT_ID', 'proj-demo-007', 'eq-au-105', '2026-03-08 18:00:00+01', '2026-03-09 10:00:00+01', 'cancelled', NOW() - INTERVAL '25 days', NOW() - INTERVAL '20 days')
ON CONFLICT DO NOTHING;
SQL
echo "  -> 22 reservations"

# ---------------------------------------------------------------------------
# 8. Summary
# ---------------------------------------------------------------------------
echo ""
echo "============================================="
echo "  Demo Data Seeding Complete!"
echo "============================================="
echo ""
echo "Summary:"
echo "  Categories:    6  (Audio, Licht, Video, Rigging, Strom, Kabel)"
echo "  Equipment:     100+ items across all categories"
echo "  Contacts:      10  (8 companies, 2 persons)"
echo "  Projects:      8   (confirmed, in_progress, draft, completed, cancelled)"
echo "  Invoices:      5   (sent, paid, overdue, cancelled)"
echo "  Packlists:     5   (packed, in_progress, draft)"
echo "  Reservations:  22  (confirmed, active, cancelled)"
echo ""
echo "Project Statuses:"
echo "  - Stadtfest Heidelberg      -> confirmed (next week)"
echo "  - Firmenfeier Autohaus      -> confirmed (2 weeks)"
echo "  - Festival am See           -> in_progress (this weekend)"
echo "  - Hochzeit Weber/Schmidt    -> draft (3 weeks)"
echo "  - Messe Frankfurt           -> confirmed (next month)"
echo "  - Konzert Stadttheater Wien -> completed (last week)"
echo "  - DJ Event Club Nachtwerk   -> cancelled"
echo "  - Konferenz Zürich          -> confirmed (2 weeks)"
echo ""
echo "Invoice Statuses:"
echo "  - RE-2026-001: €12.450  -> sent"
echo "  - RE-2026-002: €3.280   -> paid"
echo "  - RE-2026-003: €28.900  -> overdue"
echo "  - RE-2026-004: €8.750   -> paid"
echo "  - RE-2026-005: €1.200   -> cancelled"
echo ""
echo "Barcodes for scanner testing:"
echo "  RF-SPK-001..008  JBL VTX A12 (Audio)"
echo "  RF-SUB-001..004  JBL VTX S28 (Audio)"
echo "  RF-MIX-001       Yamaha CL5 (Audio)"
echo "  RF-MH-001..008   Clay Paky Sharpy Plus (Licht)"
echo "  RF-GMA-001       GrandMA3 Light (Licht)"
echo "  RF-BEA-001..002  Panasonic PT-RZ690 (Video)"
echo "  RF-TRS-001..005  Global Truss F34 3m (Rigging)"
echo "  RF-PWR-001..002  Powerlock 125A (Strom)"
echo ""
