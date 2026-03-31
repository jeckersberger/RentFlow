export interface WarehouseLocation {
  id: string
  name: string
  type: 'site' | 'room' | 'rack' | 'shelf'
  parent_id?: string
  children?: WarehouseLocation[]
  equipment_count?: number
  equipment_ids?: string[]
}

export interface WarehouseStructure {
  id: string
  name: string
  locations: WarehouseLocation[]
}
