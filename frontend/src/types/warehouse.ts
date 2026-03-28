export interface Warehouse {
  id: string;
  tenant_id: string;
  name: string;
  code: string;
  address: string;
  capacity_description: string;
  is_active: boolean;
  created_at: string;
}

export interface Zone {
  id: string;
  warehouse_id: string;
  tenant_id: string;
  name: string;
  code: string;
  climate_controlled: boolean;
  max_weight_kg: number;
  notes: string;
  created_at: string;
}

export interface Rack {
  id: string;
  zone_id: string;
  tenant_id: string;
  name: string;
  code: string;
  levels: number;
  bays_per_level: number;
  max_weight_kg: number;
  created_at: string;
}

export interface StockLocation {
  id: string;
  rack_id: string;
  zone_id: string;
  tenant_id: string;
  code: string;
  barcode: string;
  level: number;
  bay: number;
  max_weight_kg: number;
  is_active: boolean;
  created_at: string;
}

export interface Movement {
  id: string;
  tenant_id: string;
  equipment_id: string;
  from_location_id: string;
  to_location_id: string;
  quantity: number;
  reason: string;
  user_id: string;
  notes: string;
  created_at: string;
}

export interface InventoryCheck {
  id: string;
  tenant_id: string;
  zone_id: string;
  status: string;
  started_by: string;
  started_at: string;
  completed_at: string;
  expected_count: number;
  actual_count: number;
  discrepancy_count: number;
  notes: string;
}

export interface InventoryCheckItem {
  id: string;
  check_id: string;
  equipment_id: string;
  expected: boolean;
  found: boolean;
  scanned_at: string;
  notes: string;
}
