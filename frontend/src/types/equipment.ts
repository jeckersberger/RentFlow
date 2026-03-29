export interface Equipment {
  id: string;
  tenant_id: string;
  name: string;
  description: string;
  category_id: string;
  equipment_type_id: string;
  barcode: string;
  qr_code: string;
  serial_number: string;
  rfid_tag: string;
  status: string;
  condition: string;
  quantity_total: number;
  quantity_available: number;
  rental_price_day: number;
  rental_price_week: number;
  replacement_value: number;
  weight_grams: number;
  purchase_price: number;
  manufacturer: string;
  model: string;
  image_url: string;
  notes: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}
