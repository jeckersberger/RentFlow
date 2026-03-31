export type EquipmentStatus = 'available' | 'reserved' | 'checked_out' | 'in_maintenance' | 'retired' | 'damaged'

export type EquipmentCondition = 'new' | 'good' | 'fair' | 'poor' | 'defective'

export interface Equipment {
  id: string
  tenant_id: string
  name: string
  description?: string
  sku: string
  serial_number?: string
  barcode?: string
  category_id: string
  location_id?: string
  status: EquipmentStatus
  condition: EquipmentCondition
  purchase_date?: string
  purchase_price?: number
  rental_price_day?: number
  rental_price_week?: number
  weight?: number
  dimensions?: {
    length?: number
    width?: number
    height?: number
    unit?: string
  }
  image_url?: string
  image_refs?: string[]
  tags?: string[]
  custom_fields?: Record<string, string>
  created_at: string
  updated_at: string
  created_by_user_id?: string
  rfid_tag?: string
}

export interface CreateEquipmentDTO {
  name: string
  description?: string
  sku: string
  serial_number?: string
  barcode?: string
  category_id: string
  location_id?: string
  rental_price_day?: number
  rental_price_week?: number
  purchase_price?: number
  weight?: number
  dimensions?: {
    length?: number
    width?: number
    height?: number
    unit?: string
  }
  tags?: string[]
  custom_fields?: Record<string, string>
  rfid_tag?: string
}

export interface UpdateEquipmentDTO extends Partial<CreateEquipmentDTO> {}

export interface EquipmentListResponse {
  data: Equipment[]
  total: number
  limit: number
  offset: number
}

export interface Category {
  id: string
  tenant_id: string
  name: string
  parent_id?: string
  icon: string
  color: string
  sort_order: number
  created_at: string
  created_by_user_id?: string
}

export interface PriceResult {
  days: number
  daily_rate: number
  base_price: number
  volume_discount: number
  custom_discount: number
  subtotal: number
  vat: number
  total: number
  volume_discount_percent: number
  custom_discount_percent: number
  vat_percent: number
}
