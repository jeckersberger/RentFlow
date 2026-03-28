export interface Equipment {
  id: string;
  tenant_id: string;
  name: string;
  description: string;
  category_id: string;
  equipment_type_id: string;
  barcode: string;
  rfid_tag: string;
  status: string;
  condition_state: string;
  purchase_price: number;
  rental_price_daily: number;
  serial_number: string;
  notes: string;
  weight_kg: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}
