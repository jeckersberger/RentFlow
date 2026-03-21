export type EquipmentStatus = 'available' | 'reserved' | 'checked_out' | 'maintenance' | 'retired'

export interface Equipment {
  id: string
  name: string
  description?: string
  sku: string
  barcode?: string
  category: string
  location?: string
  status: EquipmentStatus
  price_daily?: number
  price_weekly?: number
  price_monthly?: number
  quantity?: number
  image_url?: string
  dimensions?: {
    length?: number
    width?: number
    height?: number
    unit?: string
  }
  weight?: number
  created_at: string
  updated_at: string
}

export interface CreateEquipmentDTO {
  name: string
  description?: string
  sku: string
  barcode?: string
  category: string
  location?: string
  price_daily?: number
  price_weekly?: number
  price_monthly?: number
  quantity?: number
}

export interface UpdateEquipmentDTO extends Partial<CreateEquipmentDTO> {
  status?: EquipmentStatus
}

export interface EquipmentListResponse {
  data: Equipment[]
  total: number
  page: number
  limit: number
}
